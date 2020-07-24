/*
* Copyright(C) 2020. Huawei Technologies Co.,Ltd. All rights reserved.
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
* http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package huawei

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog"
	"math"
	"net"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

type pluginAPI struct {
	hps *HwPluginServe
}

// Instance is for annoation
type Instance struct { // Instance
	PodName  string   `json:"pod_name"`  // pod Name
	ServerID string   `json:"server_id"` // serverdId
	Devices  []Device `json:"devices"`   // dev
}

// Device id for Instcance
type Device struct { // Device
	DeviceID string `json:"device_id"` // device id
	DeviceIP string `json:"device_ip"` // device ip
}

var ip string

// Register function is use to register k8s devicePlugin to kubelet.
func Register(k8sSocketPath, pluginSocket, resourceName string) error {
	conn, err := grpc.Dial(k8sSocketPath, grpc.WithInsecure(),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}))
	if err != nil {
		logger.Error("connect to kubelet failed.", zap.String("err", err.Error()))
		return fmt.Errorf("connect to kubelet fail: %v", err)
	}
	defer conn.Close()
	client := pluginapi.NewRegistrationClient(conn)

	reqt := &pluginapi.RegisterRequest{
		Version:      pluginapi.Version,
		Endpoint:     pluginSocket,
		ResourceName: resourceName,
	}
	logger.Info("the device plugin api version is:", zap.String("apiVersion", pluginapi.Version))
	if _, err = client.Register(context.Background(), reqt); err != nil {
		return fmt.Errorf("register to kubelet fail: %v", err)
	}
	return nil
}

// GetDevicePluginOptions is Standard interface to kubelet.
func (s *pluginAPI) GetDevicePluginOptions(ctx context.Context, e *pluginapi.Empty) (*pluginapi.DevicePluginOptions,
	error) {
	return &pluginapi.DevicePluginOptions{}, nil
}

// ListAndWatch: if the server get stop signal ,the ListAndWatch should  stop,to be fix
func (s *pluginAPI) ListAndWatch(emtpy *pluginapi.Empty, stream pluginapi.DevicePlugin_ListAndWatchServer) error {

	logger.Info("device-plugin: ListAndWatch start")
	resp := new(pluginapi.ListAndWatchResponse)
	for _, dev := range s.hps.devices {
		resp.Devices = append(resp.Devices, &pluginapi.Device{ID: dev.ID, Health: dev.Health})
	}

	if err := sendDevToKubelet(resp, stream); err != nil {
		logger.Error("listAndWatch: send device info failed, please check kubelet status.",
			zap.String("err", err.Error()))
	}

	outs := false

	usedDevices := sets.NewString()
	for {
		if outs {
			break
		}
		time.Sleep(sleepTime * time.Second)
		resp := new(pluginapi.ListAndWatchResponse)
		stated := false
		for id, dev := range s.hps.devices {
			state := s.hps.hdm.manager.GetDevState(id)
			if dev.Health != state {
				stated = true
				dev.Health = state
				s.hps.devices[id] = dev
			}
		}
		if useVolcanoType {
			for _, device := range s.hps.devices {
				if device.Health != pluginapi.Healthy {
					continue
				}
				s.hps.healthDevice.Insert(device.ID)
			}
			s.getNodeNpuUsed(&usedDevices)
			freeDevices := s.hps.healthDevice.Difference(usedDevices)
			err := s.hps.kubeInteractor.patchAnnotationOnNode(freeDevices)
			if err != nil {
				logger.Error("patch Annotation failed", zap.Error(err))
			}
		}
		if !stated {
			// close log print
			logFlag = false
		}
		if stated {
			// turn on log print
			logFlag = true
			for _, dev := range s.hps.devices {
				resp.Devices = append(resp.Devices, &pluginapi.Device{ID: dev.ID, Health: dev.Health})
			}
			if err := sendDevToKubelet(resp, stream); err != nil {
				logger.Error("listAndWatch: send device info failed, please check kubelet status.",
					zap.String("err", err.Error()))
			}
		}
	}

	return nil
}

// Allocate is called by kubelet to mount device to k8s pod.
func (s *pluginAPI) Allocate(ctx context.Context, requests *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse,
	error) {

	resps := new(pluginapi.AllocateResponse)
	logger.Info("allocate request:", zap.String("request", requests.String()))

	for _, rqt := range requests.ContainerRequests {
		resp := new(pluginapi.ContainerAllocateResponse)

		var devID []string
		var dlogMountPath string

		var majorID string
		var minorID string
		ascendVisibleDevices := ""
		allocateNum := len(rqt.DevicesIDs)

		// 使用volcano调度
		switch useVolcanoType {
		case true:
			err := s.doWithVolcanoSchedule(allocateNum, majorID, minorID, &ascendVisibleDevices)
			if err != nil {
				logger.Error("doWithVolcanoSchedule failed", zap.Error(err))
				return nil, err
			}
		default:
			setEnvFromKubelet(rqt, majorID, minorID, &ascendVisibleDevices)
		}
		if s.hps.runMode == runMode910 {
			s.mountfile(resp, dlogMountPath, devID)
			responseAnonation(resp, ascendVisibleDevices)
		}
		switch UseAscendDocker {
		case true:
			ascendVisibleDevices = addEnv(ascendVisibleDevices, resp)
		default:
			s.mountDefaultDevice(resp)
			s.mountDevice(resp, ascendVisibleDevices)
		}
		resps.ContainerResponses = append(resps.ContainerResponses, resp)
		logger.Info("allocate responses:", zap.String("request", resps.String()))
	}
	return resps, nil
}

func setEnvFromKubelet(rqt *pluginapi.ContainerAllocateRequest, majorID, minorID string, ascendVisibleDevices *string) {
	// 从kubelet获取id
	for _, id := range rqt.DevicesIDs {
		err := getAscendDeviceID(id, &majorID, &minorID)
		if err != nil {
			logger.Error("getAscendDeviceID", zap.Error(err))
		}
		*ascendVisibleDevices += majorID + ","
	}
	*ascendVisibleDevices = strings.TrimSuffix(*ascendVisibleDevices, ",")
	return
}

func (s *pluginAPI) mountDefaultDevice(resp *pluginapi.ContainerAllocateResponse) {
	// mount default devices
	for _, d := range s.hps.defaultDevs {
		resp.Devices = append(resp.Devices, &pluginapi.DeviceSpec{
			HostPath:      d,
			ContainerPath: d,
			Permissions:   "mrw",
		})
	}
}

func (s *pluginAPI) getNodeNpuUsed(usedDevices *sets.String) {
	var (
		res    []v1.Pod
		pl     *v1.PodList
		err    error
		useNpu string
	)

	kubeClient := s.hps.kubeInteractor.clientset
	node, err := kubeClient.CoreV1().Nodes().Get(s.hps.kubeInteractor.nodeName, metav1.GetOptions{})
	selector := fields.SelectorFromSet(fields.Set{"spec.nodeName": node.Name, "status.phase": string(v1.PodRunning)})
	pl, err = kubeClient.CoreV1().Pods(v1.NamespaceAll).List(metav1.ListOptions{
		FieldSelector: selector.String()})
	if err != nil {
		klog.V(3).Infof("%s 这里 nodeName: %s,err %v", "910", node.Name, err)
		return
	}

	for _, pod := range pl.Items {
		res = append(res, pod)
		tmpNpu, ok := pod.Annotations[huaWeiAscend910]
		if !ok {
			continue
		}
		usedDevices.Insert(tmpNpu)
		klog.V(3).Infof("%s 这里,podName: %s,useNpu %v", "910", pod.Name, pod.Annotations[huaWeiAscend910])
	}

	klog.V(3).Infof("%s 这里,nodeName: %s,useNpu %v", "910", node.Name, useNpu)

	return
}

func addEnv(ascendVisibleDevices string, resp *pluginapi.ContainerAllocateResponse) string {
	// add env
	ascendVisibleDevices = strings.TrimSuffix(ascendVisibleDevices, ",")
	(*resp).Envs[ascendVisibleDevicesEnv] = ascendVisibleDevices
	return ascendVisibleDevices
}

func addAnnotation(devices, podName, serverID string) string {
	// Annotations
	var instance Instance
	instance.PodName = podName
	instance.ServerID = serverID
	err := setDevices(&instance, devices)
	if err != nil {
		logger.Error("Add annotation failed", zap.String("error", err.Error()))
		return ""
	}
	instanceByte, err := json.Marshal(instance)
	if err != nil {
		logger.Error("Transfor marshal failed", zap.String("error", err.Error()))
		return ""
	}
	instanceInfo := string(instanceByte)
	return instanceInfo
}

func setDevices(instance *Instance, devices string) error {
	idSplit := strings.Split(devices, ",")
	sort.Sort(sort.StringSlice(idSplit))
	for _, deviceID := range idSplit {
		logicID64, err := strconv.ParseInt(deviceID, 10, 32)
		if err != nil {
			logger.Error(" Device id trasnsform failes ", zap.String("DeviceName", deviceID))
			return err
		}
		logicID := int32(logicID64)
		deviceIP, err := getDeviceIP(logicID)
		if err != nil {
			logger.Error("Get device ip failed:->", zap.Int("deviceId", int(logicID)), zap.Error(err))
			return err
		}
		var device Device
		device.DeviceID = deviceID
		device.DeviceIP = deviceIP
		instance.Devices = append(instance.Devices, device)
	}
	return nil
}

// PreStartContainer is Standard interface to kubelet with empty implement.
func (s *pluginAPI) PreStartContainer(ctx context.Context,
	r *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	logger.Info("PreStart just call in UT.")
	return &pluginapi.PreStartContainerResponse{}, nil
}

func (s *pluginAPI) mountfile(resp *pluginapi.ContainerAllocateResponse, dlogMountPath string, devID []string) {
	if err := s.hps.hdm.manager.GetLogPath(devID, s.hps.hdm.dlogPath, &dlogMountPath); err != nil {
		logger.Error("get logPath failed.", zap.String("err", err.Error()))
		dlogMountPath = s.hps.hdm.dlogPath
	}
	timeStr := time.Now().Format("20060102150405")
	rankID := "" + timeStr + "-0"
	slogConfigPath := GetSlogConfigFilePath()
	resp.Mounts = append(resp.Mounts, &pluginapi.Mount{
		ContainerPath: slogConfigPath,
		HostPath:      slogConfigPath,
		ReadOnly:      true,
	})
	resp.Mounts = append(resp.Mounts, &pluginapi.Mount{
		ContainerPath: "/var/dlog",
		HostPath:      dlogMountPath,
		ReadOnly:      false,
	})
	// mount log format

	logPath := "/var/log/npu"
	hostLogPath := logPath + "/slog/container/" + rankID
	resp.Mounts = append(resp.Mounts, &pluginapi.Mount{
		ContainerPath: logPath + "/slog",
		HostPath:      hostLogPath,
		ReadOnly:      false,
	})

	hostProfilingPath := logPath + "/profiling/container/" + rankID
	resp.Mounts = append(resp.Mounts, &pluginapi.Mount{
		ContainerPath: logPath + "/profiling",
		HostPath:      hostProfilingPath,
		ReadOnly:      false,
	})

	hostDumpPath := logPath + "/dump/container/" + rankID
	resp.Mounts = append(resp.Mounts, &pluginapi.Mount{
		ContainerPath: logPath + "/dump",
		HostPath:      hostDumpPath,
		ReadOnly:      false,
	})

	hostDockerSlogPath := logPath + "/docker_slog_" + rankID
	resp.Mounts = append(resp.Mounts, &pluginapi.Mount{
		ContainerPath: "/usr/slog",
		HostPath:      hostDockerSlogPath,
		ReadOnly:      false,
	})
}

func sendDevToKubelet(resp *pluginapi.ListAndWatchResponse, stream pluginapi.DevicePlugin_ListAndWatchServer) error {
	logger.Info("ListAndWatch: send devices.", zap.String("resp", resp.String()))
	if err := stream.Send(resp); err != nil {
		return err
	}
	return nil
}

// GetSlogConfigFilePath is used to get slog path
func GetSlogConfigFilePath() string {
	return hiAISlogdConfig
}

func (s *pluginAPI) updatePodAnnotations(pod *v1.Pod, ascendVisibleDevices string) error {
	kubeClient := s.hps.kubeInteractor.clientset
	node, err := kubeClient.CoreV1().Nodes().Get(s.hps.kubeInteractor.nodeName, metav1.GetOptions{})
	var serverID string
	for _, addresses := range node.Status.Addresses {
		if addresses.Type == v1.NodeInternalIP {
			serverID = addresses.Address
		}
	}
	if len(pod.Annotations) == 0 {
		pod.Annotations = map[string]string{}
	}

	pod.Annotations[podPredicateTime] = strconv.FormatUint(math.MaxUint64, 10)
	podDeviceValue := addAnnotation(ascendVisibleDevices, renamePod(pod.Name), serverID)
	pod.Annotations[podDeviceKey] = podDeviceValue
	// TODO(@hzxuzhonghu): use patch instead
	_, err = kubeClient.CoreV1().Pods(pod.Namespace).Update(pod)
	return err
}

func renamePod(podName string) string {
	suffix := strings.Split(podName, "-")
	lastIndex := len(suffix) - 1
	_, err := strconv.Atoi(suffix[lastIndex])
	if err == nil {
		return suffix[lastIndex]
	}
	return "0"
}

func getOldestPod(pods []v1.Pod) *v1.Pod {
	if len(pods) == 0 {
		return nil
	}
	oldest := pods[0]
	for _, pod := range pods {
		if GetPredicateTimeFromPodAnnotation(&oldest) > GetPredicateTimeFromPodAnnotation(&pod) {
			oldest = pod
		}
	}
	return &oldest
}

func (s *pluginAPI) getPendingPodsOnNode() ([]v1.Pod, error) {
	var (
		res []v1.Pod
		pl  *v1.PodList
		err error
	)
	nodeName := s.hps.kubeInteractor.nodeName
	selector := fields.SelectorFromSet(fields.Set{"spec.nodeName": nodeName, "status.phase": string(v1.PodPending)})
	err = wait.PollImmediate(1*time.Second, 10*time.Second, func() (bool, error) {
		pl, err = s.hps.kubeInteractor.clientset.CoreV1().Pods(v1.NamespaceAll).List(metav1.ListOptions{
			FieldSelector: selector.String(),
		})
		if err != nil {
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return nil, fmt.Errorf("kube interactor timedout: %v", err)
	}

	for _, pod := range pl.Items {
		if getNPUResourceNumOfPod(&pod) > 0 && IsGPUAssignedPod(&pod) && !IsShouldDeletePod(&pod) {
			res = append(res, pod)
		}
	}

	return res, nil
}

func (s *pluginAPI) mountDevice(resp *pluginapi.ContainerAllocateResponse, deviceID string) {
	devices := strings.Split(deviceID, ",")
	for _, device := range devices {
		hostPath := fmt.Sprintf("%s%s", "/dev/davinci", device)
		containerPath := hostPath
		resp.Devices = append(resp.Devices, &pluginapi.DeviceSpec{
			HostPath:      hostPath,
			ContainerPath: containerPath,
			Permissions:   "mrw",
		})
	}
}

func IsGPUAssignedPod(pod *v1.Pod) bool {

	_, ok := pod.ObjectMeta.Annotations[huaWeiAscend910]
	if !ok {
		klog.V(4).Infof("no assigned flag",
			pod.Name,
			pod.Namespace)
		return false
	}

	return true
}

func IsShouldDeletePod(pod *v1.Pod) bool {
	if pod.DeletionTimestamp != nil {
		return true
	}
	for _, status := range pod.Status.ContainerStatuses {
		if status.State.Waiting != nil &&
			strings.Contains(status.State.Waiting.Message, "PreStartContainer check failed") {
			return true
		}
	}
	if pod.Status.Reason == "UnexpectedAdmissionError" {
		return true
	}
	return false
}

func GetPredicateTimeFromPodAnnotation(pod *v1.Pod) uint64 {
	if assumeTimeStr, ok := pod.Annotations[podPredicateTime]; ok {
		predicateTime, err := strconv.ParseUint(assumeTimeStr, 10, 64)
		if err == nil {
			return predicateTime
		}
	} else {
		klog.Errorf("volcano没有写时间戳，pod name：%s", pod.Name)
	}

	return math.MaxUint64
}

func getNPUResourceNumOfPod(pod *v1.Pod) uint {
	var total uint
	containers := pod.Spec.Containers
	for _, container := range containers {
		if val, ok := container.Resources.Limits[huaWeiAscend910]; ok {
			total += uint(val.Value())
		}
	}
	return total
}

func getNPUAnnotationOfPOd(pod *v1.Pod, allocateDevice *sets.String, allocateNum int) error {
	annotation, exist := pod.Annotations[huaWeiAscend910]
	if !exist {
		return fmt.Errorf("cannot find the annotation")
	}
	devices := strings.Split(annotation, ",")
	if len(devices) != allocateNum {
		return fmt.Errorf("device num is not equal with annotation num", zap.String("annotation", annotation),
			zap.Int("allocateNum", allocateNum))
	}

	for _, device := range devices {
		allocateDevice.Insert(device)
	}
	return nil
}

func responseAnonation(resp *pluginapi.ContainerAllocateResponse, devices string) {
	// Annotations
	annotation := make(map[string]string)
	var instance Instance
	instance.PodName = "cloud-localhost-"

	instance.ServerID = ""

	err := setDevices(&instance, devices)
	if err != nil {
		logger.Error("Add annotation failed", zap.String("error", err.Error()))
		return
	}
	instanceByte, err := json.Marshal(instance)
	if err != nil {
		logger.Error("Transform marshal failed", zap.String("error", err.Error()))
		return
	}
	instanceInfo := string(instanceByte)
	annotation[podDeviceKey] = instanceInfo
	resp.Annotations = annotation
}

func (s *pluginAPI) doWithVolcanoSchedule(allocateNum int, majorID, minorID string, ascendVisibleDevices *string) error {
	pods, err := s.getPendingPodsOnNode()
	if err != nil {
		klog.Errorf("获取pod列表错误，%s", err)
		return err
	}
	oldPod := getOldestPod(pods)
	if oldPod == nil {
		klog.Infof("暂未获取到合适的pending pod")
		return err
	}
	klog.Infof("筛选的pod name %s , pod annotation %v,pod create time %s", oldPod.Name, oldPod.Annotations, oldPod.CreationTimestamp.Format("2006-01-02 15:04:05"))
	allocateDevice := sets.NewString()
	err = getNPUAnnotationOfPOd(oldPod, &allocateDevice, allocateNum)
	if err != nil {
		logger.Error("get NPU Annotation failed: ", zap.Error(err))
		return err
	}
	for _, id := range allocateDevice.List() {
		err := getAscendDeviceID(id, &majorID, &minorID)
		if err != nil {
			logger.Error("getAscendDeviceID", zap.Error(err))
			return err
		}
		*ascendVisibleDevices += majorID + ","
	}
	*ascendVisibleDevices = strings.TrimSuffix(*ascendVisibleDevices, ",")
	freeDevices := s.hps.healthDevice.Difference(allocateDevice)
	errs := s.hps.kubeInteractor.patchAnnotationOnNode(freeDevices)
	if errs != nil {
		logger.Error("patch Annotations failed", zap.Error(err))
	}
	err = s.updatePodAnnotations(oldPod, *ascendVisibleDevices)
	if err != nil {
		return err
	}
	return nil
}
