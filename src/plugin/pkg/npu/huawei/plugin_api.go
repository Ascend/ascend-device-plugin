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
	"go.uber.org/atomic"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
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

	"sync"
)

type pluginAPI struct {
	hps                  *HwPluginServe
	outbreak             *atomic.Bool
	ascendRuntimeOptions string
}

// Instance is for annotation
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
var totalDevices sets.String
var stateThreadNum int

// Register function is use to register k8s devicePlugin to kubelet.
func (hps *HwPluginServe) Register(k8sSocketPath, pluginSocket, resourceName string) error {
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
		s.hps.healthDevice.Insert(dev.ID)
	}

	if err := sendDevToKubelet(resp, stream); err != nil {
		logger.Error("listAndWatch: send device info failed, please check kubelet status.",
			zap.String("err", err.Error()))
	}

	for {
		if s.outbreak.Load() {
			break
		}
		time.Sleep(sleepTime * time.Second)
		isStateChange := s.isDeviceStatusChange()
		if useVolcanoType {
			s.doWithVolcanoListAndWatch(isStateChange)
		}
		if !isStateChange {
			// close log print
			logFlag = false
		}
		if isStateChange {
			// turn on log print
			logFlag = true
			resp.Devices = resp.Devices[:0]
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

func (s *pluginAPI) isDeviceStatusChange() bool {
	if IsVirtualDev(s.hps.devType) {
		return s.listenVirtualDevices()
	}
	return s.listenPhysicalDevices()
}

func (s *pluginAPI) listenVirtualDevices() bool {
	isStateChange := false
	var deviceIDs [hiAIMaxDeviceNum]uint32
	devNum, err := s.hps.hdm.dmgr.GetDeviceList(&deviceIDs)
	if err != nil {
		logger.Error("Get device list fail")
		return isStateChange
	}
	for idx := int32(0); idx < devNum; idx++ {
		phyID, err := s.hps.hdm.dmgr.GetPhyID(deviceIDs[idx])
		if err != nil {
			logger.Error("Get PhyID fail")
			return isStateChange
		}
		deviceName := fmt.Sprintf("%s-%d", hiAIAscend910Prefix, phyID)
		healthStatus := s.hps.hdm.manager.GetDevState(deviceName, s.hps.hdm.dmgr)
		for devID, device := range s.hps.devices {
			if s.isPhyDevOwnThisVirtualDevice(device, phyID) && healthStatus != device.Health {
				isStateChange = true
				device.Health = healthStatus
				s.hps.devices[devID] = device
			}
		}
	}
	return isStateChange
}

func (s *pluginAPI) isPhyDevOwnThisVirtualDevice(device *npuDevice, phyID uint32) bool {
	return strings.Split(device.ID, "-")[logicIDIndexInVirtualDevID910] == fmt.Sprintf("%d", phyID)
}

func (s *pluginAPI) listenPhysicalDevices() bool {
	isStateChange := false
	for id, dev := range s.hps.devices {
		state := s.hps.hdm.manager.GetDevState(id, s.hps.hdm.dmgr)
		if dev.Health != state {
			isStateChange = true
			dev.Health = state
			s.hps.devices[id] = dev
		}
	}
	return isStateChange
}

func (s *pluginAPI) doWithVolcanoListAndWatch(isStateChange bool) {
	if isStateChange {
		s.hps.healthDevice = sets.String{}
		for _, device := range s.hps.devices {
			if device.Health != pluginapi.Healthy {
				continue
			}
			s.hps.healthDevice.Insert(device.ID)
		}
	}
	var m sync.Mutex
	m.Lock()
	usedDevices := sets.NewString()
	s.getNodeNpuUsed(&usedDevices)
	freeDevices := s.hps.healthDevice.Difference(usedDevices)
	totalDevices = totalDevices.Union(freeDevices)
	stateThreadNum += interval
	if stateThreadNum == len(s.hps.hdm.allDevTypes) {
		if err := s.hps.kubeInteractor.patchAnnotationOnNode(totalDevices, ""); err != nil {
			logger.Error("patch Annotation failed", zap.Error(err))
		}
		totalDevices = totalDevices.Intersection(sets.String{})
		stateThreadNum = resetZero
	}
	m.Unlock()
}

// Allocate is called by kubelet to mount device to k8s pod.
func (s *pluginAPI) Allocate(ctx context.Context, requests *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse,
	error) {

	resps := new(pluginapi.AllocateResponse)
	logger.Info("allocate request:", zap.String("request", requests.String()))
	requestErrs := s.setAscendRuntimeOptions(requests)
	if requestErrs != nil {
		return nil, requestErrs
	}
	for _, rqt := range requests.ContainerRequests {
		resp := new(pluginapi.ContainerAllocateResponse)

		var devID []string
		var dlogMountPath string

		allocateNum := len(rqt.DevicesIDs)
		ascendVisibleDevicesMap, errs := s.setEnvFromKubelet(rqt)
		if errs != nil {
			logger.Error("plugin doesn't have device", zap.Error(errs))
			return nil, errs
		}
		// 使用volcano调度
		if useVolcanoType {
			ascendVisibleDevicesMap, errs = s.doWithVolcanoSchedule(allocateNum)
			if errs != nil {
				return nil, errs
			}
		}

		if s.hps.runMode == runMode910 {
			s.mountfile(resp, dlogMountPath, devID)
			s.responseAnonation(resp, ascendVisibleDevicesMap)
		}
		addEnv(ascendVisibleDevicesMap, s.ascendRuntimeOptions, resp)
		if !UseAscendDocker {
			s.mountDefaultDevice(resp)
			s.mountDevice(resp, ascendVisibleDevicesMap)
		}
		resps.ContainerResponses = append(resps.ContainerResponses, resp)
		logger.Info("allocate responses:", zap.String("request", resps.String()))
	}
	return resps, nil
}

func (s *pluginAPI) setAscendRuntimeOptions(requests *pluginapi.AllocateRequest) error {
	for _, rqt := range requests.ContainerRequests {
		for _, deviceName := range rqt.DevicesIDs {
			if IsVirtualDev(deviceName) && len(rqt.DevicesIDs) > interval {
				return fmt.Errorf("request more than one virtual device, current is %d", len(rqt.DevicesIDs))
			}
			if IsVirtualDev(deviceName) {
				s.ascendRuntimeOptions = virtualDev
				return nil
			}
		}
	}
	return nil
}

func (s *pluginAPI) setEnvFromKubelet(rqt *pluginapi.ContainerAllocateRequest) (map[string]string, error) {
	// get id from kubelet
	ascendVisibleDevices := make(map[string]string, MaxVirtualDevNum)
	for _, id := range rqt.DevicesIDs {
		_, ok := s.hps.devices[id]
		if !ok {
			return nil, fmt.Errorf("plugin doesn't have device %s", id)
		}
		deviceID, virID, err := getDeviceID(id, s.ascendRuntimeOptions)
		if err != nil {
			logger.Error("getDeviceID", zap.Error(err))
			return nil, err
		}
		var deviceIP string
		if s.ascendRuntimeOptions == virtualDev {
			ascendVisibleDevices[virID] = defaultDeviceIP
			continue
		}

		if strings.Contains(s.hps.devType, hiAIAscend910Prefix) {
			deviceIP, err = s.getDeviceIP(deviceID)
			if err != nil {
				logger.Error("Get device ip failed:->", zap.String("deviceId", deviceID), zap.Error(err))
				return nil, err
			}
		}
		ascendVisibleDevices[deviceID] = deviceIP
	}
	logger.Info("Found ascendVisibleDevices: ", zap.Any("ascendVisibleDevices", ascendVisibleDevices))
	return ascendVisibleDevices, nil
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
		err    error
		useNpu []string
	)

	kubeClient := s.hps.kubeInteractor.clientset
	node, err := kubeClient.CoreV1().Nodes().Get(s.hps.kubeInteractor.nodeName, metav1.GetOptions{})
	if err != nil {
		logger.Error("get node from k8s error", zap.Error(err))
		return
	}
	getFailed := s.getNPUByStatus(kubeClient, node.Name, string(v1.PodRunning), &useNpu)
	if getFailed {
		return
	}

	getFailed = s.getNPUByStatus(kubeClient, node.Name, string(v1.PodPending), &useNpu)
	if getFailed {
		return
	}
	for _, device := range useNpu {
		usedDevices.Insert(device)
	}
	logger.Debug(fmt.Sprintf("nodeName: %s", node.Name), zap.Strings("useNpu", useNpu))
	return
}

func (s *pluginAPI) getNPUByStatus(kubeClient kubernetes.Interface, nodeName, status string, useNpu *[]string) bool {
	selector := fields.SelectorFromSet(fields.Set{"spec.nodeName": nodeName, "status.phase": status})
	podList, err := kubeClient.CoreV1().Pods(v1.NamespaceAll).List(metav1.ListOptions{
		FieldSelector: selector.String()})
	if err != nil {
		logger.Error(fmt.Sprintf("nodeName: %s", nodeName), zap.Error(err))
		return true
	}
	for _, pod := range podList.Items {
		annotationTag := fmt.Sprintf("%s%s", resourceNamePrefix, s.hps.devType)
		tmpNpu, ok := pod.Annotations[annotationTag]
		if !ok {
			continue
		}
		*useNpu = append(*useNpu, tmpNpu)
	}
	logger.Debug("useNpu: " + strings.Join(*useNpu, ","))
	return false
}

func addEnv(devices map[string]string, ascendRuntimeOptions string, resp *pluginapi.ContainerAllocateResponse) {
	// add env
	var ascendVisibleDevices []string
	if len((*resp).Envs) == 0 {
		(*resp).Envs = make(map[string]string)
	}
	for deviceID := range devices {
		ascendVisibleDevices = append(ascendVisibleDevices, deviceID)
	}

	(*resp).Envs[ascendVisibleDevicesEnv] = strings.Join(ascendVisibleDevices, ",")
	(*resp).Envs[ascendRuntimeOptionsEnv] = ascendRuntimeOptions
}

func (s *pluginAPI) addAnnotation(devices map[string]string, podName, serverID string) string {
	// Annotations
	var instance Instance
	instance.PodName = podName
	instance.ServerID = serverID
	s.setDevices(&instance, devices)
	instanceByte, err := json.Marshal(instance)
	if err != nil {
		logger.Error("Transfor marshal failed", zap.String("error", err.Error()))
		return ""
	}
	instanceInfo := string(instanceByte)
	return instanceInfo
}

func (s *pluginAPI) setDevices(instance *Instance, devices map[string]string) {
	var sortDevicesKey []string
	for deviceID := range devices {
		sortDevicesKey = append(sortDevicesKey, deviceID)
	}
	sort.Strings(sortDevicesKey)
	for _, deviceID := range sortDevicesKey {
		var device Device
		if s.ascendRuntimeOptions == virtualDev {
			s.setVirtualDevices(instance, device, deviceID)
			continue
		}
		device.DeviceID = deviceID
		device.DeviceIP = devices[deviceID]
		instance.Devices = append(instance.Devices, device)
	}
}

func (s *pluginAPI) setVirtualDevices(instance *Instance, device Device, deviceID string) {
	for phyID, virIds := range s.hps.hdm.manager.GetPhyDevMapVirtualDev() {
		if strings.Contains(virIds, deviceID) {
			device.DeviceID = fmt.Sprintf("%d", phyID)
			device.DeviceIP = defaultDeviceIP
			instance.Devices = append(instance.Devices, device)
		}
	}
}

func (s *pluginAPI) getDeviceIP(phyID string) (string, error) {
	transPhyID, err := strconv.ParseInt(phyID, 10, 32)
	if err != nil {
		logger.Error(" Device id trasnsform failes ", zap.String("DeviceName", phyID))
		return "", err
	}

	logicID, err := s.hps.hdm.dmgr.GetLogicID(uint32(transPhyID))
	if err != nil {
		return ERROR, fmt.Errorf("transfor phyID %s to logicID failed, error code : %v", phyID, err)
	}
	return s.hps.hdm.dmgr.GetDeviceIP(int32(logicID))
}

// PreStartContainer is Standard interface to kubelet with empty implement.
func (s *pluginAPI) PreStartContainer(ctx context.Context,
	r *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	logger.Info("PreStart just call in UT.")
	return &pluginapi.PreStartContainerResponse{}, nil
}

func (s *pluginAPI) mountfile(resp *pluginapi.ContainerAllocateResponse, dlogMountPath string, devID []string) {
	if err := s.hps.hdm.manager.GetLogPath(devID, s.hps.hdm.dlogPath, s.ascendRuntimeOptions, &dlogMountPath); err != nil {
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

func (s *pluginAPI) updatePodAnnotations(pod *v1.Pod, ascendVisibleDevices map[string]string) error {
	kubeClient := s.hps.kubeInteractor.clientset
	node, err := kubeClient.CoreV1().Nodes().Get(s.hps.kubeInteractor.nodeName, metav1.GetOptions{})
	if err != nil {
		return err
	}
	var serverID string
	for _, addresses := range node.Status.Addresses {
		if addresses.Type == v1.NodeInternalIP {
			serverID = addresses.Address
		}
	}
	if len(pod.Annotations) == 0 {
		pod.Annotations = map[string]string{}
	}

	podDeviceValue := s.addAnnotation(ascendVisibleDevices, pod.Name, serverID)
	pod2, err := s.updatePod(pod, podDeviceValue)
	for i := 0; err != nil && i < retryTime; i++ {
		logger.Info("try again ...")
		pod2, err = s.updatePod(pod2, podDeviceValue)
		if err == nil {
			return nil
		}
		time.Sleep(interval * time.Second)
	}

	return err
}

func (s *pluginAPI) updatePod(pod *v1.Pod, podDeviceValue string) (*v1.Pod, error) {
	pod1, err := s.hps.kubeInteractor.clientset.CoreV1().Pods(pod.Namespace).Get(pod.Name, metav1.GetOptions{})
	if err != nil {
		logger.Error("query pod info failed,", zap.Error(err))
		return nil, fmt.Errorf("query pod info failed,%v", err)
	}
	pod1.Annotations[podPredicateTime] = strconv.FormatUint(math.MaxUint64, 10)
	pod1.Annotations[podDeviceKey] = podDeviceValue
	pod2, err := s.hps.kubeInteractor.clientset.CoreV1().Pods(pod.Namespace).Update(pod1)
	if err != nil {
		logger.Error("update pod failed,%v", zap.Error(err))
		return nil, fmt.Errorf("update pod failed,%v", err)
	}
	return pod2, nil
}

func getOldestPod(pods []v1.Pod) *v1.Pod {
	if len(pods) == 0 {
		return nil
	}
	oldest := pods[0]
	for _, pod := range pods {
		if getPredicateTimeFromPodAnnotation(&oldest) > getPredicateTimeFromPodAnnotation(&pod) {
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
	err = wait.PollImmediate(interval*time.Second, timeout*time.Second, func() (bool, error) {
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
		if s.getNPUResourceNumOfPod(&pod) > 0 && s.isAscendAssignedPod(&pod) && !s.isShouldDeletePod(&pod) {
			res = append(res, pod)
		}
	}
	return res, nil
}

func (s *pluginAPI) mountDevice(resp *pluginapi.ContainerAllocateResponse, devices map[string]string) {
	for deviceID := range devices {
		containerPath, hostPath := s.hps.hdm.manager.GetDevPath(fmt.Sprintf("%s", deviceID), s.ascendRuntimeOptions)
		resp.Devices = append(resp.Devices, &pluginapi.DeviceSpec{
			HostPath:      hostPath,
			ContainerPath: containerPath,
			Permissions:   "mrw",
		})
	}
}

func (s *pluginAPI) isAscendAssignedPod(pod *v1.Pod) bool {
	annotationTag := fmt.Sprintf("%s%s", resourceNamePrefix, s.hps.devType)
	_, ok := pod.ObjectMeta.Annotations[annotationTag]
	if !ok {
		logger.Info("no assigned flag",
			zap.String("pod name", pod.Name),
			zap.String("pod.NameSpace", pod.Namespace))
		return false
	}
	return true
}

func (s *pluginAPI) isShouldDeletePod(pod *v1.Pod) bool {
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

func getPredicateTimeFromPodAnnotation(pod *v1.Pod) uint64 {
	if assumeTimeStr, ok := pod.Annotations[podPredicateTime]; ok {
		predicateTime, err := strconv.ParseUint(assumeTimeStr, 10, 64)
		if err == nil {
			return predicateTime
		}
	}
	logger.Info("volcano not write timestamp, pod name：" + pod.Name)
	return math.MaxUint64
}

func (s *pluginAPI) getNPUResourceNumOfPod(pod *v1.Pod) uint {
	var total uint
	containers := pod.Spec.Containers
	annotationTag := fmt.Sprintf("%s%s", resourceNamePrefix, s.hps.devType)
	for _, container := range containers {
		if val, ok := container.Resources.Limits[v1.ResourceName(annotationTag)]; ok {
			total += uint(val.Value())
		}
	}
	return total
}

func (s *pluginAPI) getNPUAnnotationOfPod(pod *v1.Pod, allocateDevice *sets.String, allocateNum int) error {
	annotationTag := fmt.Sprintf("%s%s", resourceNamePrefix, s.hps.devType)
	annotation, exist := pod.Annotations[annotationTag]
	if !exist {
		return fmt.Errorf("cannot find the annotation")
	}
	devices := strings.Split(annotation, ",")
	if len(devices) != allocateNum {
		return fmt.Errorf("device num %v is not equal with annotation num%v",
			zap.Int("allocateNum", allocateNum), zap.String("annotation", annotation),
		)
	}

	for _, device := range devices {
		allocateDevice.Insert(device)
	}
	return nil
}

func (s *pluginAPI) responseAnonation(resp *pluginapi.ContainerAllocateResponse, devices map[string]string) {
	// Annotations
	annotation := make(map[string]string)
	var instance Instance
	instance.PodName = "cloud-localhost-"

	instance.ServerID = ""

	s.setDevices(&instance, devices)
	instanceByte, err := json.Marshal(instance)
	if err != nil {
		logger.Error("Transform marshal failed", zap.String("error", err.Error()))
		return
	}
	instanceInfo := string(instanceByte)
	annotation[podDeviceKey] = instanceInfo
	resp.Annotations = annotation
}

func (s *pluginAPI) doWithVolcanoSchedule(allocateNum int) (map[string]string, error) {
	ascendVisibleDevices := make(map[string]string, MaxVirtualDevNum)
	pods, err := s.getPendingPodsOnNode()
	if err != nil {
		logger.Error("get pod list err", zap.Error(err))
		return nil, err
	}
	oldPod := getOldestPod(pods)
	if oldPod == nil {
		logger.Info("not get pending pod")
		return nil, err
	}
	allocateDevice := sets.NewString()
	err = s.getNPUAnnotationOfPod(oldPod, &allocateDevice, allocateNum)
	if err != nil {
		logger.Error("get NPU Annotation failed: ", zap.Error(err))
		return nil, err
	}
	errors := s.getAscendVisiDevsWithVolcano(allocateDevice, &ascendVisibleDevices)
	if errors != nil {
		logger.Error("get ascend devs with volcano failed: ", zap.Error(err))
	}

	usedDevices := sets.NewString()
	s.getNodeNpuUsed(&usedDevices)
	freeDevices := s.hps.healthDevice.Difference(usedDevices)
	errs := s.hps.kubeInteractor.patchAnnotationOnNode(freeDevices, s.hps.devType)
	if errs != nil {
		logger.Error("patch Annotations failed", zap.Error(err))
		return nil, err
	}
	err = s.updatePodAnnotations(oldPod, ascendVisibleDevices)
	if err != nil {
		return nil, err
	}
	return ascendVisibleDevices, nil
}

func (s *pluginAPI) getAscendVisiDevsWithVolcano(allocateDevice sets.String, devices *map[string]string) error {
	for _, id := range allocateDevice.List() {

		deviceID, virID, err := getDeviceID(id, s.ascendRuntimeOptions)
		if err != nil {
			logger.Error("get phyID", zap.Error(err))
			return err
		}
		if s.ascendRuntimeOptions == virtualDev {
			(*devices)[virID] = defaultDeviceIP
			continue
		}
		deviceIP, errs := s.getDeviceIP(deviceID)
		if errs != nil {
			logger.Error("Get device ip failed:->", zap.String("deviceId", deviceID), zap.Error(errs))
			return errs
		}
		(*devices)[deviceID] = deviceIP

	}
	return nil
}
