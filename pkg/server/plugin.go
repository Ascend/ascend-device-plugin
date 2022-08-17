// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package server holds the implementation of registration to kubelet, k8s device plugin interface and grpc service.
package server

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"

	"huawei.com/npu-exporter/devmanager"
	"huawei.com/npu-exporter/hwlog"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/kubeclient"
)

func (ps *PluginServer) stopListAndWatch() {
	if ps.isRunning.Load() {
		ps.stop <- struct{}{}
	}
}

// Notify is called when device status changed, to notify ListAndWatch
func (ps *PluginServer) Notify(devices []*common.NpuDevice) bool {
	if ps == nil {
		hwlog.RunLog.Errorf("invalid interface receiver")
		return false
	}
	if ps.isRunning.Load() {
		ps.deepCopyDevice(devices)
		ps.reciChan <- struct{}{}
		return true
	}
	return false
}

func (ps *PluginServer) generateAllDeviceMap() map[string]string {
	vol2kltMap := make(map[string]string, 1)
	var notInVolDev []string
	allDev := sets.String{}
	klDev := sets.String{}
	ps.allocMapLock.RLock()
	ps.cachedLock.RLock()
	for _, dev := range ps.cachedDevices {
		allDev.Insert(dev.DeviceName)
		d, exist := ps.vol2KlDevMap[dev.DeviceName]
		if !exist {
			notInVolDev = append(notInVolDev, dev.DeviceName)
			continue
		}
		klDev.Insert(d)
		vol2kltMap[dev.DeviceName] = d
	}
	ps.allocMapLock.RUnlock()
	ps.cachedLock.RUnlock()
	notInKlDev := allDev.Difference(klDev).List()
	for index, d := range notInKlDev {
		if index >= len(notInVolDev) {
			hwlog.RunLog.Warnf("found volcano not using device %s in notInVolDev on local %d failed", d, index)
			continue
		}
		vol := notInVolDev[index]
		vol2kltMap[vol] = d
	}
	return vol2kltMap
}

func sendToKubelet(stream v1beta1.DevicePlugin_ListAndWatchServer, resp *v1beta1.ListAndWatchResponse) error {
	return stream.Send(resp)
}

func (ps *PluginServer) responseToKubelet() *v1beta1.ListAndWatchResponse {
	resp := new(v1beta1.ListAndWatchResponse)
	ps.cachedLock.RLock()
	if common.ParamOption.UseVolcanoType && !common.IsVirtualDev(ps.deviceType) {
		vol2kltMap := ps.generateAllDeviceMap()
		for _, device := range ps.cachedDevices {
			d, exist := vol2kltMap[device.DeviceName]
			if !exist {
				hwlog.RunLog.Warnf(" not exist map key, %s  map %+v", device.DeviceName, vol2kltMap)
				continue
			}
			hwlog.RunLog.Infof("ListAndWatch resp devices: %s %s", d, device.Health)
			resp.Devices = append(resp.Devices, &v1beta1.Device{ID: d, Health: device.Health})
		}
	} else {
		for _, device := range ps.cachedDevices {
			hwlog.RunLog.Infof("ListAndWatch resp devices: %s %s", device.DeviceName, device.Health)
			resp.Devices = append(resp.Devices, &v1beta1.Device{ID: device.DeviceName, Health: device.Health})
		}
	}
	ps.cachedLock.RUnlock()
	return resp
}

func (ps *PluginServer) deepCopyDevice(cachedDevices []*common.NpuDevice) {
	ps.cachedLock.Lock()
	ps.cachedDevices = ps.cachedDevices[:0]
	for _, dev := range cachedDevices {
		ps.cachedDevices = append(ps.cachedDevices, common.NpuDevice{DeviceName: dev.DeviceName, Health: dev.Health})
	}
	ps.cachedLock.Unlock()
}

// ListAndWatch is to send device info to kubelet
func (ps *PluginServer) ListAndWatch(empty *v1beta1.Empty, stream v1beta1.DevicePlugin_ListAndWatchServer) error {
	send := func(stream v1beta1.DevicePlugin_ListAndWatchServer) {
		if err := sendToKubelet(stream, ps.responseToKubelet()); err != nil {
			hwlog.RunLog.Errorf("send to kubelet failed, error is %s", err.Error())
		}
	}
	ps.isRunning.Store(true)
	send(stream)
	for {
		select {
		case <-ps.stop:
			ps.isRunning.Store(false)
			return nil
		case _, ok := <-ps.reciChan:
			if ok {
				send(stream)
			}
		}
	}
}

func (ps *PluginServer) deviceExists(id string) bool {
	ps.cachedLock.RLock()
	defer ps.cachedLock.RUnlock()
	for _, d := range ps.cachedDevices {
		if d.DeviceName == id {
			return true
		}
	}
	return false
}

func (ps *PluginServer) checkAllocateRequest(requests *v1beta1.AllocateRequest) error {
	if requests == nil {
		return fmt.Errorf("invalid requests")
	}
	if len(requests.ContainerRequests) > common.MaxContainerLimit {
		return fmt.Errorf("the number of container request %d exceeds the upper limit",
			len(requests.ContainerRequests))
	}
	for _, rqt := range requests.ContainerRequests {
		if len(rqt.DevicesIDs) > common.MaxDevicesNum {
			return fmt.Errorf("the devices can't bigger than %d", common.MaxDevicesNum)
		}
		for _, deviceName := range rqt.DevicesIDs {
			if !ps.deviceExists(deviceName) {
				return fmt.Errorf("plugin doesn't have device %s", deviceName)
			}
			if common.IsVirtualDev(deviceName) && len(rqt.DevicesIDs) > common.MaxRequestVirtualDeviceNum {
				return fmt.Errorf("request more than %d virtual device, current is %d",
					common.MaxRequestVirtualDeviceNum, len(rqt.DevicesIDs))
			}
			if common.IsVirtualDev(deviceName) {
				ps.ascendRuntimeOptions = common.VirtualDev
				return nil
			}
		}
	}
	return nil
}

func getPredicateTimeFromPodAnnotation(pod *v1.Pod) uint64 {
	assumeTimeStr, ok := pod.Annotations[common.PodPredicateTime]
	if !ok {
		hwlog.RunLog.Infof("volcano not write timestamp, pod Name: " + pod.Name)
		return math.MaxUint64
	}
	predicateTime, err := strconv.ParseUint(assumeTimeStr, common.BaseDec, common.BitSize)
	if err != nil {
		hwlog.RunLog.Errorf("parse timestamp failed, %s", err.Error())
		return math.MaxUint64
	}
	return predicateTime
}

func (ps *PluginServer) getOldestPod(pods []v1.Pod) *v1.Pod {
	if len(pods) == 0 {
		return nil
	}
	oldest := pods[0]
	for _, pod := range pods {
		hwlog.RunLog.Debugf("pod %v, predicate time: %v", oldest.Name, pod.Annotations[common.PodPredicateTime])
		if getPredicateTimeFromPodAnnotation(&oldest) > getPredicateTimeFromPodAnnotation(&pod) {
			oldest = pod
		}
	}
	hwlog.RunLog.Debugf("oldest pod %v, predicate time: %v", oldest.Name, oldest.Annotations[common.PodPredicateTime])
	annotation := map[string]string{common.PodPredicateTime: strconv.FormatUint(math.MaxUint64, common.BaseDec)}
	if err := ps.kubeClient.TryUpdatePodAnnotation(&oldest, annotation); err != nil {
		hwlog.RunLog.Errorf("update pod %v failed, err: %v", oldest.Name, err)
		return nil
	}
	return &oldest
}

func (ps *PluginServer) updateAllocMap(realAlloc, kltAlloc []string) {
	if len(realAlloc) != len(kltAlloc) {
		hwlog.RunLog.Errorf("length of klt allocate not equal real allocate")
		return
	}
	ps.allocMapLock.Lock()
	for _, id := range kltAlloc {
		for k, v := range ps.vol2KlDevMap {
			if v == id {
				delete(ps.vol2KlDevMap, k)
			}
		}
	}
	for i, id := range realAlloc {
		ps.vol2KlDevMap[id] = kltAlloc[i]
	}
	ps.allocMapLock.Unlock()
}

// GetRealAllocateDevices is convert kubelet allocate device list to volcano allocate device list
func (ps *PluginServer) GetRealAllocateDevices(kltAllocate []string) ([]string, error) {
	if ps == nil {
		return nil, fmt.Errorf("invalid interface receiver")
	}
	ps.allocMapLock.RLock()
	defer ps.allocMapLock.RUnlock()
	klt2vol := make(map[string]string, len(ps.vol2KlDevMap))
	for k, v := range ps.vol2KlDevMap {
		klt2vol[v] = k
	}
	var realAllocate []string
	for _, id := range kltAllocate {
		realID, exist := klt2vol[id]
		if !exist {
			return nil, fmt.Errorf("cannot found real allocate device by %s", id)
		}
		realAllocate = append(realAllocate, realID)
	}
	return realAllocate, nil
}

func (ps *PluginServer) doWithVolcanoSchedule(requestDevices []string) ([]string, error) {
	conditionFunc := func(pod *v1.Pod) bool {
		allocateDevice, err := common.GetDeviceFromPodAnnotation(pod, ps.deviceType)
		if err != nil {
			return false
		}
		return len(allocateDevice) == len(requestDevices)
	}
	allPods, err := ps.kubeClient.GetPodList()
	if err != nil {
		return nil, err
	}
	pods, err := common.FilterPods(allPods, common.GetPodPhaseBlackList(), ps.deviceType, conditionFunc)
	if err != nil {
		return nil, err
	}
	oldestPod := ps.getOldestPod(pods)
	if oldestPod == nil {
		return nil, fmt.Errorf("not get valid pod")
	}
	allocateDevices, err := common.GetDeviceFromPodAnnotation(oldestPod, ps.deviceType)
	if err != nil {
		return nil, err
	}
	hwlog.RunLog.Infof("vol found: %#v", allocateDevices)
	ps.updateAllocMap(allocateDevices, requestDevices)
	return allocateDevices, nil
}

func (ps *PluginServer) useVolcano(requestDevices []string) ([]string, error) {
	// if virtual device, allocate by k8s
	if common.IsVirtualDev(ps.deviceType) {
		return requestDevices, nil
	}
	return ps.doWithVolcanoSchedule(requestDevices)
}

func getDevPath(id, ascendRuntimeOptions string) (string, string) {
	containerPath := fmt.Sprintf("%s%s", "/dev/davinci", id)
	hostPath := containerPath
	if ascendRuntimeOptions == common.VirtualDev {
		hostPath = fmt.Sprintf("%s%s", "/dev/vdavinci", id)
	}
	return containerPath, hostPath
}

func mountDevice(resp *v1beta1.ContainerAllocateResponse, devices []string, ascendRuntimeOptions string) {
	for deviceID := range devices {
		containerPath, hostPath := getDevPath(fmt.Sprintf("%s", deviceID), ascendRuntimeOptions)
		resp.Devices = append(resp.Devices, &v1beta1.DeviceSpec{
			HostPath:      hostPath,
			ContainerPath: containerPath,
			Permissions:   "rw",
		})
	}
}

func mountDefaultDevice(resp *v1beta1.ContainerAllocateResponse, defaultDevs []string) {
	// mount default devices
	for _, d := range defaultDevs {
		resp.Devices = append(resp.Devices, &v1beta1.DeviceSpec{
			HostPath:      d,
			ContainerPath: d,
			Permissions:   "rw",
		})
	}
}

// Allocate is called by kubelet to mount device to k8s pod.
func (ps *PluginServer) Allocate(ctx context.Context, requests *v1beta1.AllocateRequest) (*v1beta1.AllocateResponse,
	error) {
	if err := ps.checkAllocateRequest(requests); err != nil {
		hwlog.RunLog.Error(err)
		return nil, err
	}
	resps := new(v1beta1.AllocateResponse)
	for _, rqt := range requests.ContainerRequests {
		var err error
		allocateDevices := rqt.DevicesIDs
		hwlog.RunLog.Infof("request: %#v", rqt.DevicesIDs)
		if common.ParamOption.UseVolcanoType {
			allocateDevices, err = ps.useVolcano(rqt.DevicesIDs)
			if err != nil {
				hwlog.RunLog.Error(err)
				return nil, err
			}
		}
		ascendVisibleDevices, err := common.GetDeviceListID(allocateDevices, ps.ascendRuntimeOptions)
		if err != nil {
			hwlog.RunLog.Error(err)
			return nil, err
		}

		resp := new(v1beta1.ContainerAllocateResponse)
		common.SetAscendRuntimeEnv(ascendVisibleDevices, ps.ascendRuntimeOptions, resp)
		if !common.ParamOption.UseAscendDocker {
			mountDefaultDevice(resp, ps.defaultDevs)
			mountDevice(resp, ascendVisibleDevices, ps.ascendRuntimeOptions)
		}
		resps.ContainerResponses = append(resps.ContainerResponses, resp)
	}
	return resps, nil
}

// GetPreferredAllocation implement the kubelet device plugin interface
func (ps *PluginServer) GetPreferredAllocation(context.Context, *v1beta1.PreferredAllocationRequest) (
	*v1beta1.PreferredAllocationResponse, error) {
	return nil, errors.New("not support")
}

// GetDevicePluginOptions is Standard interface to kubelet.
func (ps *PluginServer) GetDevicePluginOptions(ctx context.Context, e *v1beta1.Empty) (*v1beta1.DevicePluginOptions,
	error) {
	return &v1beta1.DevicePluginOptions{}, nil
}

// PreStartContainer is Standard interface to kubelet with empty implement.
func (ps *PluginServer) PreStartContainer(ctx context.Context,
	r *v1beta1.PreStartContainerRequest) (*v1beta1.PreStartContainerResponse, error) {
	hwlog.RunLog.Infof("PreStart just call in UT.")
	return &v1beta1.PreStartContainerResponse{}, nil
}

// NewPluginServer returns an initialized PluginServer
func NewPluginServer(devManager devmanager.DeviceInterface, client *kubeclient.ClientK8s, deviceType string,
	devices []*common.NpuDevice, defaultDevs []string) *PluginServer {
	ps := &PluginServer{
		restart:      true,
		reciChan:     make(chan interface{}),
		devManager:   devManager,
		kubeClient:   client,
		deviceType:   deviceType,
		defaultDevs:  defaultDevs,
		stop:         make(chan interface{}),
		vol2KlDevMap: make(map[string]string, common.MaxDevicesNum),
		isRunning:    common.NewAtomicBool(false),
	}
	ps.deepCopyDevice(devices)
	return ps
}
