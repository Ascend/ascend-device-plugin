/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2020-2022. All rights reserved.
 */

// Package huawei the device-plugin frame interface
package huawei

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"huawei.com/npu-exporter/hwlog"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"Ascend-device-plugin/pkg/common"
)

type pluginAPI struct {
	hps                  *HwPluginServe
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

var (
	totalDevices                  sets.String
	stateThreadNum                int
	m                             sync.Mutex
	firstTimeList                 = true
	dpStartReset                  sync.Once
	totalUHDevices                sets.String
	totalNetworkUnhealthDevices   sets.String
	lastTimeNetworkRecoverDevices sets.String
	listenDevCountIsChange        = make(map[string]bool, initMapCap)
	podPhaseBlackList             = map[v1.PodPhase]int{v1.PodFailed: 0, v1.PodSucceeded: 0}
)

const (
	// envNum is the number of env variables that will be written to the container
	envNum = 2

	// podNameMaxLength is the max pod name length
	podNameMaxLength = 253

	// podNameSpaceMaxLength is the max pod namespace length
	podNameSpaceMaxLength = 63

	// maxDevicesNum is max number of devices
	maxDevicesNum = 64

	// retryPodUpdateCount is max number of retry update pod annotation
	retryPodUpdateCount = 3

	maxPodLimit       = 110
	maxContainerLimit = 300000
)

// Register function is use to register k8s devicePlugin to kubelet.
func (hps *HwPluginServe) Register() error {
	realKubeletSockPath, isOk := VerifyPath(v1beta1.KubeletSocket)
	if !isOk {
		return fmt.Errorf("check kubelet socket file path failed")
	}
	conn, err := grpc.Dial(realKubeletSockPath, grpc.WithInsecure(),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}))
	if err != nil {
		hwlog.RunLog.Errorf("connect to kubelet failed, err: %s", err.Error())
		return fmt.Errorf("connect to kubelet fail: %v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			hwlog.RunLog.Errorf("close kubelet connect failed, err: %s", err.Error())
		}
	}()

	client := v1beta1.NewRegistrationClient(conn)
	reqt := &v1beta1.RegisterRequest{
		Version:      v1beta1.Version,
		Endpoint:     fmt.Sprintf("%s.sock", hps.devType),
		ResourceName: resourceNamePrefix + hps.devType,
	}
	if _, err = client.Register(context.Background(), reqt); err != nil {
		return fmt.Errorf("register to kubelet fail: %v", err)
	}
	hps.vol2KlDevMap = make(map[string]string, maxDevicesNum)
	return nil
}

// GetDevicePluginOptions is Standard interface to kubelet.
func (s *pluginAPI) GetDevicePluginOptions(ctx context.Context, e *v1beta1.Empty) (*v1beta1.DevicePluginOptions,
	error) {
	return &v1beta1.DevicePluginOptions{}, nil
}

// ListAndWatch: if the server get stop signal ,the ListAndWatch should  stop,to be fix
func (s *pluginAPI) ListAndWatch(emtpy *v1beta1.Empty, stream v1beta1.DevicePlugin_ListAndWatchServer) error {

	hwlog.RunLog.Infof("device-plugin: ListAndWatch start")
	resp := new(v1beta1.ListAndWatchResponse)
	s.updateKubeletDevInfo(resp, stream)
	for {
		if s.hps.outbreak.Load() {
			break
		}
		sleepTime := listAndWatchPeriod - sleep2ListW
		if sleepTime < 0 {
			sleepTime = 0
		}
		time.Sleep(time.Duration(sleepTime) * time.Second)
		m.Lock()
		stateThreadNum += interval
		s.isDeviceStatusChange()
		if useVolcanoType {
			dpStartReset.Do(func() {
				s.hps.kubeInteractor.annotationReset()
			})
			s.updatePodRealAllocate(podPhaseBlackList)
			s.hps.hdm.manager.DoWithVolcanoListAndWatch(s.hps)
		}
		resp.Devices = resp.Devices[:0]
		s.updateKubeletDevInfo(resp, stream)
		m.Unlock()
	}
	s.hps.stopCh <- struct{}{}
	return nil
}

func (s *pluginAPI) updateKubeletDevInfo(resp *v1beta1.ListAndWatchResponse,
	stream v1beta1.DevicePlugin_ListAndWatchServer) {
	s.initK8sInfo(resp, stream)
	var notInVolDev []string
	allDev := sets.String{}
	klDev := sets.String{}
	for _, dev := range s.hps.devices {
		allDev.Insert(dev.ID)
		d, exist := s.hps.vol2KlDevMap[dev.ID]
		if !exist {
			notInVolDev = append(notInVolDev, dev.ID)
			continue
		}
		klDev.Insert(d)
	}
	if len(klDev) < len(allDev) && len(notInVolDev) != 0 {
		notInKlDev := allDev.Difference(klDev).List()
		for index, d := range notInKlDev {
			if index >= len(notInVolDev) {
				hwlog.RunLog.Warnf("found volcano not using device %s in notInVolDev on local %d failed", d, index)
				continue
			}
			vol := notInVolDev[index]
			s.hps.vol2KlDevMap[vol] = d
		}
	}
	for _, dev := range s.hps.devices {
		d, exist := s.hps.vol2KlDevMap[dev.ID]
		if !exist {
			hwlog.RunLog.Warnf(" not exist map key, %s  map %+v", dev.ID, s.hps.vol2KlDevMap)
			continue
		}
		if useVolcanoType && s.hps.kubeInteractor.isSkipRecoverLabel(d, s.hps.runMode) {
			continue
		}
		resp.Devices = append(resp.Devices, &v1beta1.Device{ID: d, Health: dev.Health})
	}
	time.Sleep(sleep2ListW * time.Second)
	if err := sendDevToKubelet(resp, stream); err != nil {
		hwlog.RunLog.Errorf("listAndWatch: send device info failed, please "+
			"check kubelet status, err: %s", err.Error())
	}
}

func (s *pluginAPI) initK8sInfo(resp *v1beta1.ListAndWatchResponse, stream v1beta1.DevicePlugin_ListAndWatchServer) {
	if !firstTimeList {
		return
	}
	totalNetworkUnhealthDevices = sets.String{}
	totalUHDevices = sets.String{}
	for _, dev := range s.hps.devices {
		resp.Devices = append(resp.Devices, &v1beta1.Device{ID: dev.ID, Health: dev.Health})
		s.hps.healthDevice.Insert(dev.ID)
	}
	if err := sendDevToKubelet(resp, stream); err != nil {
		hwlog.RunLog.Errorf("listAndWatch: send device info failed, please "+
			"check kubelet status, err: %s", err.Error())
	}
	firstTimeList = false
}

func (s *pluginAPI) isDeviceStatusChange() bool {
	if common.IsVirtualDev(s.hps.devType) {
		return s.listenVirtualDevices()
	}
	return s.listenPhysicalDevices()
}

func (s *pluginAPI) listenVirtualDevices() bool {
	isStateChange := false
	devNum, devList, err := s.hps.hdm.dmgr.GetDeviceList()
	if err != nil {
		hwlog.RunLog.Errorf("Get device list fail, error is %v", err)
		return isStateChange
	}
	if devNum > hiAIMaxDeviceNum {
		hwlog.RunLog.Error("Get device list fail")
		return isStateChange
	}
	for idx := int32(0); idx < devNum; idx++ {
		phyID, err := s.hps.hdm.dmgr.GetPhysicIDFromLogicID(devList[idx])
		if err != nil {
			hwlog.RunLog.Errorf("Get PhyID fail")
			return isStateChange
		}
		deviceName := fmt.Sprintf("%s-%d", hiAIAscend910Prefix, phyID)
		healthStatus := s.hps.hdm.manager.GetDevState(deviceName, s.hps.hdm.dmgr)
		for devID, device := range s.hps.devices {
			if s.isPhyDevOwnThisVirtualDevice(device, phyID) && healthStatus != device.Health {
				isStateChange = true
				s.hps.devices[devID].Health = healthStatus
			}
		}
	}
	return isStateChange
}

func (s *pluginAPI) isPhyDevOwnThisVirtualDevice(device *common.NpuDevice, phyID int32) bool {
	return strings.Split(device.ID, "-")[logicIDIndexInVirtualDevID910] == fmt.Sprintf("%d", phyID)
}

func (s *pluginAPI) listenPhysicalDevices() bool {
	isStateChange := false
	for id, dev := range s.hps.devices {
		state := s.hps.hdm.manager.GetDevState(id, s.hps.hdm.dmgr)
		if dev.Health != state {
			isStateChange = true
			s.hps.devices[id].Health = state
		}

		// If the device type is Ascend910, check the network health status of the device.
		if s.hps.devType == hiAIAscend910Prefix {
			isStateChange = s.checkDeviceNetworkHealthStatus(dev) || isStateChange
		}
	}
	return isStateChange
}

// check device network health status
// only for Ascend910 and Non-virtual device
func (s *pluginAPI) checkDeviceNetworkHealthStatus(device *common.NpuDevice) bool {
	// virtual devices do not check network health status.
	if common.IsVirtualDev(s.hps.devType) {
		return false
	}

	phyID, err := GetPhyIDByName(device.ID)
	if err != nil {
		hwlog.RunLog.Error(err)
		return false
	}

	logicID, err := s.hps.hdm.dmgr.GetLogicIDFromPhysicID(int32(phyID))
	if err != nil {
		hwlog.RunLog.Error(err)
		return false
	}

	healthStatus, err := s.hps.hdm.manager.GetDeviceNetworkState(int32(logicID), device)
	if err != nil {
		hwlog.RunLog.Error(err)
		return false
	}

	if healthStatus != device.NetworkHealth {
		device.NetworkHealth = healthStatus
		return true
	}

	return false
}

func (s *pluginAPI) patchUnusedDevice() error {
	usedDevices := sets.NewString()
	getNodeNpuUsed(&usedDevices, s.hps)
	freeDevices := s.hps.healthDevice.Difference(usedDevices)
	groupAllocatableDevs := s.hps.hdm.manager.GetAnnotationMap(freeDevices, []string{s.hps.devType})
	if err := s.hps.kubeInteractor.patchAnnotationOnNode(groupAllocatableDevs, true, s.hps.devType); err != nil {
		return fmt.Errorf("patch Annotations failed, error is: %v", err)
	}
	return nil
}

func (s *pluginAPI) useVolcano(ascendVisibleDevicesMap *map[string]string, allocateNum int) error {
	var err error
	if !common.IsVirtualDev(s.hps.devType) {
		*ascendVisibleDevicesMap, err = s.doWithVolcanoSchedule(allocateNum)
		if err != nil {
			hwlog.RunLog.Errorf("do with volcano schedule failed, error is %v", err)
			return err
		}
	}
	if err = s.patchUnusedDevice(); err != nil {
		hwlog.RunLog.Errorf("patch Annotations failed, error: %v", err)
		return err
	}
	return nil
}

func logAllocateRequest(requestDevices []string) {
	if len(requestDevices) == 0 {
		hwlog.RunLog.Errorf("allocate request device's length is 0")
		return
	}
	rqtDevice := requestDevices[0]
	devType, err := getDeviceType(rqtDevice)
	if err != nil {
		hwlog.RunLog.Errorf("allocate request invalid device name: %s", rqtDevice)
		return
	}
	if common.IsVirtualDev(rqtDevice) {
		deviceID, virID, err := common.GetDeviceID(rqtDevice, common.VirtualDev)
		if err != nil {
			hwlog.RunLog.Errorf("dev ID is invalid, deviceID: %s", rqtDevice)
			return
		}
		hwlog.RunLog.Infof("allocate request %s:{%s}", devType, fmt.Sprintf("%s-%s", virID, deviceID))
		return
	}
	var devIDs []string
	for _, requestDevice := range requestDevices {
		deviceID, _, err := common.GetDeviceID(requestDevice, physicalDev)
		if err != nil {
			hwlog.RunLog.Errorf("dev ID is invalid, deviceID: %s", requestDevice)
			return
		}
		devIDs = append(devIDs, deviceID)
	}
	hwlog.RunLog.Infof("allocate request %s:{%s}", devType, strings.Join(devIDs, ","))
}

// Allocate is called by kubelet to mount device to k8s pod.
func (s *pluginAPI) Allocate(ctx context.Context, requests *v1beta1.AllocateRequest) (*v1beta1.AllocateResponse,
	error) {

	resps := new(v1beta1.AllocateResponse)
	hwlog.RunLog.Info("allocate request")
	requestErrs := s.setAscendRuntimeOptions(requests)
	if requestErrs != nil {
		return nil, requestErrs
	}

	for _, rqt := range requests.ContainerRequests {
		logAllocateRequest(rqt.DevicesIDs)
		resp := new(v1beta1.ContainerAllocateResponse)
		ascendVisibleDevicesMap, errs := s.getDeviceListIP(rqt.DevicesIDs)
		if errs != nil {
			hwlog.RunLog.Errorf("plugin doesn't have device, err: %v", errs)
			return nil, errs
		}

		if useVolcanoType {
			if err := s.useVolcano(&ascendVisibleDevicesMap, len(rqt.DevicesIDs)); err != nil {
				hwlog.RunLog.Errorf("use volcano schedule failed, error is %v", err)
				return nil, err
			}
		}
		if s.hps.runMode == common.RunMode910 || s.hps.runMode == common.RunMode310P {
			s.responseAnnotation(resp, ascendVisibleDevicesMap)
		}
		addEnv(ascendVisibleDevicesMap, s.ascendRuntimeOptions, resp)
		if !UseAscendDocker {
			s.mountDefaultDevice(resp)
			s.mountDevice(resp, ascendVisibleDevicesMap)
		}
		resps.ContainerResponses = append(resps.ContainerResponses, resp)
		hwlog.RunLog.Info("allocate response")
	}
	return resps, nil
}

func (s *pluginAPI) setAscendRuntimeOptions(requests *v1beta1.AllocateRequest) error {
	if len(requests.ContainerRequests) > maxContainerLimit {
		return fmt.Errorf("the number of container request %d exceeds the upper limit",
			len(requests.ContainerRequests))
	}
	for _, rqt := range requests.ContainerRequests {
		if len(rqt.DevicesIDs) > maxDevicesNum {
			return fmt.Errorf("the devices can't bigger than %d", maxDevicesNum)
		}
		for _, deviceName := range rqt.DevicesIDs {
			if common.IsVirtualDev(deviceName) && len(rqt.DevicesIDs) > interval {
				return fmt.Errorf("request more than one virtual device, current is %d", len(rqt.DevicesIDs))
			}
			if common.IsVirtualDev(deviceName) {
				s.ascendRuntimeOptions = common.VirtualDev
				return nil
			}
		}
	}
	return nil
}

func (s *pluginAPI) mountDefaultDevice(resp *v1beta1.ContainerAllocateResponse) {
	// mount default devices
	for _, d := range s.hps.hdm.defaultDevs {
		resp.Devices = append(resp.Devices, &v1beta1.DeviceSpec{
			HostPath:      d,
			ContainerPath: d,
			Permissions:   "rw",
		})
	}
}

func getNodeNpuUsed(usedDevices *sets.String, hps *HwPluginServe) {
	var useNpu []string
	kubeClient := hps.kubeInteractor.clientset
	node, err := kubeClient.CoreV1().Nodes().Get(context.Background(), hps.kubeInteractor.nodeName, metav1.GetOptions{})
	if err != nil {
		hwlog.RunLog.Errorf("get node from k8s error: %v", err)
		return
	}
	if getFailed := getNPUByStatus(hps, &useNpu); getFailed {
		return
	}
	for _, device := range useNpu {
		usedDevices.Insert(device)
	}
	hwlog.RunLog.Debugf(fmt.Sprintf("nodeName: %s, useNpus: %v", node.Name, useNpu))
	return
}

func getNPUByStatus(hps *HwPluginServe, useNpu *[]string) bool {
	podList, err := getPodList(hps.kubeInteractor)
	if err != nil {
		hwlog.RunLog.Errorf(fmt.Sprintf("nodeName: %s, err: %v", hps.kubeInteractor.nodeName, err))
		return true
	}
	for index, pod := range podList.Items {
		if index >= maxPodLimit {
			hwlog.RunLog.Error("The number of pods exceeds the upper limit")
			return true
		}
		if pod.Status.Phase == v1.PodSucceeded {
			continue
		}
		annotationTag := fmt.Sprintf("%s%s", resourceNamePrefix, hps.devType)
		tmpNpu, ok := pod.Annotations[annotationTag]
		if !ok {
			continue
		}
		tmpNpuList := strings.Split(tmpNpu, ",")
		if len(tmpNpuList) == 0 || len(tmpNpuList) > maxDevicesNum {
			hwlog.RunLog.Warnf("invalid annotation, len is %d", len(tmpNpu))
			continue
		}
		*useNpu = append(*useNpu, tmpNpuList...)
		hwlog.RunLog.Debugf(" pod Name %s  getNPUByStatus vol : %#v", pod.Name, tmpNpu)
	}
	hwlog.RunLog.Debugf("useNpu: " + strings.Join(*useNpu, ","))
	return false
}

func addEnv(devices map[string]string, ascendRuntimeOptions string, resp *v1beta1.ContainerAllocateResponse) {
	// add env
	var ascendVisibleDevices []string
	if len((*resp).Envs) == 0 {
		(*resp).Envs = make(map[string]string, envNum)
	}
	for deviceID := range devices {
		ascendVisibleDevices = append(ascendVisibleDevices, deviceID)
	}

	(*resp).Envs[ascendVisibleDevicesEnv] = strings.Join(ascendVisibleDevices, ",")
	(*resp).Envs[ascendRuntimeOptionsEnv] = ascendRuntimeOptions

	hwlog.RunLog.Infof("allocate resp env: %s; %s", strings.Join(ascendVisibleDevices, ","), ascendRuntimeOptions)
}

func (s *pluginAPI) addAnnotation(devices map[string]string, podName, serverID string) string {
	// Annotations
	var instance Instance
	instance.PodName = podName
	instance.ServerID = serverID
	s.setDevices(&instance, devices)
	instanceByte, err := json.Marshal(instance)
	if err != nil {
		hwlog.RunLog.Errorf("Transform marshal failed, err: %s", err.Error())
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
		if s.ascendRuntimeOptions == common.VirtualDev {
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
	transPhyID, err := strconv.ParseInt(phyID, baseDec, bitSize32)
	if err != nil {
		hwlog.RunLog.Errorf(" Device id transform failed, DeviceName: %s", phyID)
		return "", err
	}

	logicID, err := s.hps.hdm.dmgr.GetLogicIDFromPhysicID(int32(transPhyID))
	if err != nil {
		return "error", fmt.Errorf("transfor phyID %s to logicID failed, error code : %v", phyID, err)
	}
	return s.hps.hdm.dmgr.GetDeviceIPAddress(logicID)
}

// PreStartContainer is Standard interface to kubelet with empty implement.
func (s *pluginAPI) PreStartContainer(ctx context.Context,
	r *v1beta1.PreStartContainerRequest) (*v1beta1.PreStartContainerResponse, error) {
	hwlog.RunLog.Infof("PreStart just call in UT.")
	return &v1beta1.PreStartContainerResponse{}, nil
}

func sendDevToKubelet(resp *v1beta1.ListAndWatchResponse, stream v1beta1.DevicePlugin_ListAndWatchServer) error {
	for i := 0; i < len(resp.Devices); i++ {
		hwlog.RunLog.Infof("ListAndWatch resp devices: %s %s", resp.Devices[i].ID, resp.Devices[i].Health)
	}
	if err := stream.Send(resp); err != nil {
		return err
	}
	return nil
}

func tryUpdatePodAnnotation(hps *HwPluginServe, pod *v1.Pod, annotation map[string]string) error {
	for i := 0; i < retryPodUpdateCount; i++ {
		podNew, err := hps.kubeInteractor.clientset.CoreV1().Pods(pod.Namespace).Get(context.Background(), pod.Name,
			metav1.GetOptions{})
		if err != nil {
			hwlog.RunLog.Errorf("query pod info failed, error is: %v", err)
			continue
		}
		if podNew == nil {
			err = fmt.Errorf("query pod info failed, pod is nil")
			continue
		}
		for k, v := range annotation {
			podNew.Annotations[k] = v
		}

		if _, err = hps.kubeInteractor.clientset.CoreV1().Pods(podNew.Namespace).Update(context.Background(), podNew,
			metav1.UpdateOptions{}); err == nil {
			return nil
		}
		hwlog.RunLog.Errorf("update pod annotation failed, error is %v", err)
	}
	return fmt.Errorf("exceeded maximum number of retries")
}

func (s *pluginAPI) getPodConfigurationAnnotation(podName string, ascendVisibleDevices map[string]string) (string,
	error) {
	kubeClient := s.hps.kubeInteractor.clientset
	node, err := kubeClient.CoreV1().Nodes().Get(context.Background(),
		s.hps.kubeInteractor.nodeName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	var serverID string
	for index, addresses := range node.Status.Addresses {
		if index > maxPodLimit {
			hwlog.RunLog.Error("The number of node status in exceeds the upper limit")
			break
		}
		if addresses.Type == v1.NodeInternalIP && net.ParseIP(addresses.Address) != nil {
			serverID = addresses.Address
			break
		}
	}

	podDeviceValue := s.addAnnotation(ascendVisibleDevices, podName, serverID)
	return podDeviceValue, nil
}
func getOldestPod(pods []v1.Pod, hps *HwPluginServe) *v1.Pod {
	if len(pods) == 0 {
		return nil
	}
	oldest := pods[0]
	for _, pod := range pods {
		hwlog.RunLog.Debugf("pod %v, predicate time: %v", oldest.Name, pod.Annotations[podPredicateTime])
		if getPredicateTimeFromPodAnnotation(&oldest) > getPredicateTimeFromPodAnnotation(&pod) {
			oldest = pod
		}
	}
	hwlog.RunLog.Debugf("oldest pod %v, predicate time: %v", oldest.Name, oldest.Annotations[podPredicateTime])
	annotation := map[string]string{podPredicateTime: strconv.FormatUint(math.MaxUint64, baseDec)}
	if err := tryUpdatePodAnnotation(hps, &oldest, annotation); err != nil {
		hwlog.RunLog.Errorf("update pod %v failed, err: %v", oldest.Name, err)
		return nil
	}
	return &oldest
}

func (s *pluginAPI) filterPods(blackList map[v1.PodPhase]int, conditionFunc func(pod *v1.Pod) bool) ([]v1.Pod, error) {
	pods, err := getPodList(s.hps.kubeInteractor)
	if err != nil {
		return nil, err
	}
	var res []v1.Pod
	for index, pod := range pods.Items {
		if index >= maxPodLimit {
			hwlog.RunLog.Error("The number of pods exceeds the upper limit")
			return res, nil
		}
		hwlog.RunLog.Debugf("pod: %v, %v", pod.Name, pod.Status.Phase)
		if err := s.checkPodNameAndSpace(pod.Name, podNameMaxLength); err != nil {
			hwlog.RunLog.Errorf("pod name syntax illegal, err: %v", err)
			continue
		}
		if err := s.checkPodNameAndSpace(pod.Namespace, podNameSpaceMaxLength); err != nil {
			hwlog.RunLog.Errorf("pod namespace syntax illegal, err: %v", err)
			continue
		}
		if _, exist := blackList[pod.Status.Phase]; exist {
			continue
		}
		if conditionFunc != nil && !conditionFunc(&pod) {
			continue
		}
		if s.getNPUResourceNumOfPod(&pod) > 0 && s.isAscendAssignedPod(&pod) && !s.isShouldDeletePod(&pod) {
			res = append(res, pod)
		}
	}
	return res, nil
}

func (s *pluginAPI) mountDevice(resp *v1beta1.ContainerAllocateResponse, devices map[string]string) {
	for deviceID := range devices {
		containerPath, hostPath := s.hps.hdm.manager.GetDevPath(fmt.Sprintf("%s", deviceID), s.ascendRuntimeOptions)
		resp.Devices = append(resp.Devices, &v1beta1.DeviceSpec{
			HostPath:      hostPath,
			ContainerPath: containerPath,
			Permissions:   "rw",
		})
	}
}

func (s *pluginAPI) isAscendAssignedPod(pod *v1.Pod) bool {
	annotationTag := fmt.Sprintf("%s%s", resourceNamePrefix, s.hps.devType)
	_, ok := pod.ObjectMeta.Annotations[annotationTag]
	if !ok {
		hwlog.RunLog.Debugf("no assigned flag, pod Name: %s, pod NameSpace: %s", pod.Name, pod.Namespace)
		return false
	}
	return true
}

func (s *pluginAPI) isShouldDeletePod(pod *v1.Pod) bool {
	if pod.DeletionTimestamp != nil {
		return true
	}
	if len(pod.Status.ContainerStatuses) > maxContainerLimit {
		hwlog.RunLog.Error("The number of container exceeds the upper limit")
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
		predicateTime, err := strconv.ParseUint(assumeTimeStr, baseDec, bitSize)
		if err == nil {
			return predicateTime
		}
	}
	hwlog.RunLog.Infof("volcano not write timestamp, pod Name: " + pod.Name)
	return math.MaxUint64
}

func (s *pluginAPI) getNPUResourceNumOfPod(pod *v1.Pod) int64 {
	containers := pod.Spec.Containers
	if len(containers) > maxContainerLimit {
		hwlog.RunLog.Error("The number of container exceeds the upper limit")
		return int64(0)
	}
	var total int64
	annotationTag := fmt.Sprintf("%s%s", resourceNamePrefix, s.hps.devType)
	for _, container := range containers {
		if val, ok := container.Resources.Limits[v1.ResourceName(annotationTag)]; ok {
			limitsDevNum := val.Value()
			if limitsDevNum < 0 || limitsDevNum > int64(maxDevicesNum) {
				hwlog.RunLog.Errorf("apply devices number should be in the range of [0, %d]", maxDevicesNum)
				return int64(0)
			}
			if limitsDevNum > math.MaxInt64-total {
				hwlog.RunLog.Errorf("apply devices number overflow")
				return int64(0)
			}

			total += limitsDevNum
		}
	}
	return total
}

func (s *pluginAPI) getVolAllocateDevice(pod *v1.Pod) ([]string, error) {
	annotationTag := fmt.Sprintf("%s%s", resourceNamePrefix, s.hps.devType)
	annotation, exist := pod.Annotations[annotationTag]
	if !exist {
		return nil, fmt.Errorf("cannot find the annotation")
	}
	return strings.Split(annotation, ","), nil
}

func (s *pluginAPI) responseAnnotation(resp *v1beta1.ContainerAllocateResponse, devices map[string]string) {
	// Annotations
	annotation := make(map[string]string, 1)
	var instance Instance
	instance.PodName = "cloud-localhost-"

	instance.ServerID = ""

	s.setDevices(&instance, devices)
	instanceByte, err := json.Marshal(instance)
	if err != nil {
		hwlog.RunLog.Errorf("Transform marshal failed, err: %s", err.Error())
		return
	}
	instanceInfo := string(instanceByte)
	if s.hps.runMode == common.RunMode910 {
		annotation[pod910DeviceKey] = instanceInfo
	} else {
		annotation[pod310PDeviceKey] = instanceInfo
	}
	resp.Annotations = annotation
}

func (s *pluginAPI) updatePodAnnotation(pod *v1.Pod, kltRequestDevices, dpResponseDevices []string) error {
	ascendVisibleDevices, err := s.getDeviceListIP(dpResponseDevices)
	if err != nil {
		return fmt.Errorf("get ascend devices ip failed, err: %v", err)
	}
	configuration, err := s.getPodConfigurationAnnotation(pod.Name, ascendVisibleDevices)
	if err != nil {
		return fmt.Errorf("get pod configuration failed: %v", err)
	}

	annotation := map[string]string{pod2kl: strings.Join(kltRequestDevices, ","),
		podRealAlloc: strings.Join(dpResponseDevices, ",")}
	if s.hps.runMode == common.RunMode910 {
		annotation[pod910DeviceKey] = configuration
	} else if s.hps.runMode == common.RunMode310P {
		annotation[pod310PDeviceKey] = configuration
	}
	return tryUpdatePodAnnotation(s.hps, pod, annotation)
}

func (s *pluginAPI) updatePodRealAllocate(blackList map[v1.PodPhase]int) {
	checkpointData, err := GetKubeletCheckPoint(kubeletCheckPointFile)
	if err != nil {
		hwlog.RunLog.Errorf("get check point info failed, error is %v", err)
		return
	}
	pods, err := s.filterPods(blackList, nil)
	if err != nil {
		hwlog.RunLog.Errorf("list pod failed, error is %v", err)
		return
	}

	s.hps.vol2KlDevMap = make(map[string]string, maxDevicesNum)
	for _, pod := range pods {
		hwlog.RunLog.Debugf("pods: %v, %v, %v", pod.Name, pod.Status.Phase, pod.UID)
		data, exist := checkpointData[string(pod.UID)]
		if !exist {
			continue
		}
		hwlog.RunLog.Debugf("found pod name: %v, uid: %v, resourceName: %v, request: %v, "+
			"response: %v", pod.Name, pod.UID, data.ResourceName, data.Request, data.Response)

		kltRequestDevices, dpResponseDevices, err := GetAnnotation(data, s.hps.devType)
		if err != nil {
			hwlog.RunLog.Debugf("get annotation failed: %v", err)
			continue
		}
		hwlog.RunLog.Debugf("get annotation kltDevValue: %v, dpDevValue: %v", kltRequestDevices, dpResponseDevices)

		if len(kltRequestDevices) != len(dpResponseDevices) {
			hwlog.RunLog.Warnf("klt len not equ vol , klt : %v vol : %v", kltRequestDevices, dpResponseDevices)
			continue
		}
		for index, vol := range dpResponseDevices {
			s.hps.vol2KlDevMap[vol] = kltRequestDevices[index]
		}
		if _, exist := pod.Annotations[podRealAlloc]; exist {
			continue
		}
		if err := s.updatePodAnnotation(&pod, kltRequestDevices, dpResponseDevices); err != nil {
			hwlog.RunLog.Errorf("update pod annotation failed, error is %v", err)
		}
	}
}

func (s *pluginAPI) doWithVolcanoSchedule(allocateNum int) (map[string]string, error) {
	conditionFunc := func(pod *v1.Pod) bool {
		allocateDevice, err := s.getVolAllocateDevice(pod)
		if err != nil {
			return false
		}
		return len(allocateDevice) == allocateNum
	}
	pods, err := s.filterPods(podPhaseBlackList, conditionFunc)
	if err != nil {
		hwlog.RunLog.Errorf("get pod list err: %v", err)
		return nil, err
	}

	oldPod := getOldestPod(pods, s.hps)
	if oldPod == nil {
		return nil, fmt.Errorf("not get pending pod")
	}
	allocateDevice, err := s.getVolAllocateDevice(oldPod)
	if err != nil {
		return nil, fmt.Errorf("get NPU Annotation failed, err: %v", err)
	}
	ascendVisibleDevices, err := s.getDeviceListIP(allocateDevice)
	if err != nil {
		return nil, fmt.Errorf("get ascend devs with volcano failed, err: %v", err)
	}
	return ascendVisibleDevices, nil
}

func (s *pluginAPI) checkPodNameAndSpace(podPara string, maxLength int) error {
	if len(podPara) > maxLength {
		return fmt.Errorf("para length %d is bigger than %d", len(podPara), maxLength)
	}
	pattern := "^[a-z0-9]+[a-z0-9\\-]*[a-z0-9]+$"
	if maxLength == podNameMaxLength {
		pattern = "^[a-z0-9]+([a-z0-9\\-.]*)[a-z0-9]+$"
	}

	reg := regexp.MustCompile(pattern)
	if !reg.MatchString(podPara) {
		return fmt.Errorf("podPara is illegal")
	}
	return nil
}

func (s *pluginAPI) getDeviceListIP(devices []string) (map[string]string, error) {
	ascendVisibleDevices := make(map[string]string, MaxVirtualDevNum)
	for _, id := range devices {
		if _, ok := s.hps.devices[id]; !ok {
			return nil, fmt.Errorf("plugin doesn't have device %s", id)
		}
		deviceID, virID, err := common.GetDeviceID(id, s.ascendRuntimeOptions)
		if err != nil {
			hwlog.RunLog.Errorf("getDeviceID err: %v", err)
			return nil, err
		}
		if s.ascendRuntimeOptions == common.VirtualDev {
			ascendVisibleDevices[virID] = defaultDeviceIP
			continue
		}
		var deviceIP string
		if strings.Contains(s.hps.devType, hiAIAscend910Prefix) {
			deviceIP, err = s.getDeviceIP(deviceID)
			if err != nil {
				hwlog.RunLog.Errorf("Get device ip failed, deviceId: %s, err: %v", deviceID, err)
				return nil, err
			}
		}
		ascendVisibleDevices[deviceID] = deviceIP
	}
	return ascendVisibleDevices, nil
}

// GetPreferredAllocation implement the kubelet device plugin interface
func (s *pluginAPI) GetPreferredAllocation(context.Context, *v1beta1.PreferredAllocationRequest) (
	*v1beta1.PreferredAllocationResponse, error) {
	return nil, errors.New("not support")
}
