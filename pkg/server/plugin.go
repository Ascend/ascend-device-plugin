/* Copyright(C) 2022. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Package server holds the implementation of registration to kubelet, k8s device plugin interface and grpc service.
package server

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"huawei.com/npu-exporter/v5/common-utils/hwlog"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/device"
)

func (ps *PluginServer) stopListAndWatch() {
	if ps.isRunning.Load() {
		ps.stop <- struct{}{}
	}
}

// Notify is called when device status changed, to notify ListAndWatch
func (ps *PluginServer) Notify(devices []*common.NpuDevice) bool {
	if ps == nil {
		hwlog.RunLog.Error("invalid interface receiver")
		return false
	}
	if ps.isRunning.Load() {
		ps.deepCopyDevice(devices)
		ps.reciChan <- struct{}{}
		return true
	}
	return false
}

// getUnhealthyAICore
// for example:
// aicore-0, aicore-22: Ascend310P-2c-100-0
// if Ascend310P-0 is unhealthy, aicore-0, aicore-22 is unhealthy
// Ascend310P-0 has 8 aicore, and should select 6 free aicore to be set unhealthy
func (ps *PluginServer) getUnhealthyAICore() sets.String {
	// get chip health status and all AICore devs
	unhealthyPhyID := sets.Int{}
	allAICore := make(sets.String, len(ps.cachedDevices))
	for _, device := range ps.cachedDevices {
		if device.Health == v1beta1.Unhealthy {
			unhealthyPhyID.Insert(int(device.PhyID))
		}
		allAICore.Insert(device.DeviceName)
	}
	// get real used AICore devs
	realUsedAICore, err := ps.GetRealUsedAICore()
	if err != nil {
		hwlog.RunLog.Errorf("failed to get real used AICore device, %v", err)
		return sets.String{}
	}
	// if chip is unhealthy, the real device has same phyid is unhealthy, and the klt ai core is unhealthy
	unhealthyAICore := sets.String{}
	usedAICore := sets.String{}
	for k, r := range realUsedAICore {
		phyID, _, err := common.GetDeviceID(r, "")
		if err != nil {
			hwlog.RunLog.Warn(err)
			continue
		}
		if unhealthyPhyID.Has(phyID) {
			unhealthyAICore.Insert(k)
		}
		usedAICore.Insert(k)
	}
	if unhealthyPhyID.Len() > math.MaxInt/int(ps.manager.GetChipAICore()) {
		hwlog.RunLog.Errorf("the num of unhealthy device %d is invalid", unhealthyPhyID.Len())
		return unhealthyAICore
	}
	leftUnhealthyAICoreNum := unhealthyPhyID.Len()*int(ps.manager.GetChipAICore()) - unhealthyAICore.Len()
	if leftUnhealthyAICoreNum < 0 {
		hwlog.RunLog.Errorf("num of left unhealthy ai core %d is less than 0", leftUnhealthyAICoreNum)
		return unhealthyAICore
	}
	// get free ai core device
	freeAICore := allAICore.Difference(usedAICore)
	if freeAICore.Len() < leftUnhealthyAICoreNum {
		hwlog.RunLog.Errorf("free ai core device num is %d, while need %d", freeAICore.Len(), leftUnhealthyAICoreNum)
		return unhealthyAICore
	}
	// if unhealthy ai core dev is not enough, select free ai core device randomly
	freeList := freeAICore.List()
	for count := 0; count < leftUnhealthyAICoreNum; count++ {
		unhealthyAICore.Insert(freeList[count])
	}
	return unhealthyAICore
}

// GetRealUsedAICore get real used aicore from pod
func (ps *PluginServer) GetRealUsedAICore() (map[string]string, error) {
	podList := ps.manager.GetKubeClient().GetActivePodListCache()
	podDeviceInfo, err := ps.GetKltAndRealAllocateDev(podList)
	if err != nil {
		return nil, fmt.Errorf("failed to get klt and real allocate device, %w", err)
	}
	usedAICore := make(map[string]string, len(podDeviceInfo))
	for _, deviceInfo := range podDeviceInfo {
		hwlog.RunLog.Debugf("pod info name: %s, status:%s, uid:%s", deviceInfo.Pod.Name,
			deviceInfo.Pod.Status.Phase, deviceInfo.Pod.UID)
		if len(deviceInfo.RealDevice) == 0 {
			continue
		}
		for _, coreName := range deviceInfo.KltDevice {
			usedAICore[coreName] = deviceInfo.RealDevice[0]
		}
	}
	return usedAICore, nil
}

func (ps *PluginServer) generateAllDeviceMap() map[string]string {
	vol2kltMap := make(map[string]string, 1)
	var notInVolDev []string
	allDev := sets.String{}
	klDev := sets.String{}
	ps.allocMapLock.RLock()
	ps.cachedLock.RLock()
	vol2KlDevMap := make(map[string]string, len(ps.klt2RealDevMap))
	for k, r := range ps.klt2RealDevMap {
		vol2KlDevMap[r] = k
	}
	for _, dev := range ps.cachedDevices {
		allDev.Insert(dev.DeviceName)
		d, exist := vol2KlDevMap[dev.DeviceName]
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
	if !common.ParamOption.PresetVDevice {
		unhealthyDev := ps.getUnhealthyAICore()
		for _, device := range ps.cachedDevices {
			if unhealthyDev.Has(device.DeviceName) {
				device.Health = v1beta1.Unhealthy
			} else {
				device.Health = v1beta1.Healthy
			}
			hwlog.RunLog.Infof("ListAndWatch resp devices: %s %s", device.DeviceName, device.Health)
			resp.Devices = append(resp.Devices, &v1beta1.Device{ID: device.DeviceName, Health: device.Health})
		}
	} else if common.ParamOption.UseVolcanoType && !common.IsVirtualDev(ps.deviceType) {
		vol2kltMap := ps.generateAllDeviceMap()
		for _, device := range ps.cachedDevices {
			d, exist := vol2kltMap[device.DeviceName]
			if !exist {
				hwlog.RunLog.Warnf(" not exist map key, %s  map %+v", device.DeviceName, vol2kltMap)
				continue
			}
			hwlog.RunLog.Infof("ListAndWatch resp devices: inner device: %s %s, real device: %s %s", d,
				device.Health, device.DeviceName, device.Health)
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
		ps.cachedDevices = append(ps.cachedDevices, common.NpuDevice{
			DeviceName: dev.DeviceName,
			Health:     dev.Health,
			PhyID:      dev.PhyID,
		})
	}
	ps.cachedLock.Unlock()
}

// ListAndWatch is to send device info to kubelet
func (ps *PluginServer) ListAndWatch(empty *v1beta1.Empty, stream v1beta1.DevicePlugin_ListAndWatchServer) error {
	send := func(stream v1beta1.DevicePlugin_ListAndWatchServer) {
		if err := sendToKubelet(stream, ps.responseToKubelet()); err != nil {
			hwlog.RunLog.Errorf("send to kubelet failed, error is %#v", err)
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
		if len(rqt.DevicesIDs) > common.MaxDevicesNum*common.MinAICoreNum {
			return fmt.Errorf("the devices can't bigger than %d", common.MaxDevicesNum)
		}
		for _, deviceName := range rqt.DevicesIDs {
			if len(deviceName) > common.MaxDeviceNameLen {
				return fmt.Errorf("length of device name %d is invalid", len(deviceName))
			}
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
		hwlog.RunLog.Warnf("volcano not write timestamp, pod Name: %s", pod.Name)
		return math.MaxUint64
	}
	if len(assumeTimeStr) > common.PodAnnotationMaxLength {
		hwlog.RunLog.Warnf("timestamp fmt invalid, pod Name: %s", pod.Name)
		return math.MaxUint64
	}
	predicateTime, err := strconv.ParseUint(assumeTimeStr, common.BaseDec, common.BitSize)
	if err != nil {
		hwlog.RunLog.Errorf("parse timestamp failed, %#v", err)
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
		hwlog.RunLog.Debugf("pod %s, predicate time: %s", pod.Name, pod.Annotations[common.PodPredicateTime])
		if getPredicateTimeFromPodAnnotation(&oldest) > getPredicateTimeFromPodAnnotation(&pod) {
			oldest = pod
		}
	}
	hwlog.RunLog.Debugf("oldest pod %#v, predicate time: %#v", oldest.Name,
		oldest.Annotations[common.PodPredicateTime])
	annotation := map[string]string{common.PodPredicateTime: strconv.FormatUint(math.MaxUint64, common.BaseDec)}
	if err := ps.manager.GetKubeClient().TryUpdatePodCacheAnnotation(&oldest, annotation); err != nil {
		hwlog.RunLog.Errorf("update pod %s failed, err: %#v", oldest.Name, err)
		return nil
	}
	return &oldest
}

func (ps *PluginServer) updateAllocMap(realAlloc, kltAlloc []string) {
	if common.ParamOption.PresetVDevice {
		ps.updatePresetAllocMap(realAlloc, kltAlloc)
	} else {
		ps.updateDynamicAllocMap(realAlloc, kltAlloc)
	}
}

func (ps *PluginServer) updateDynamicAllocMap(realAlloc, kltAlloc []string) {
	// real device exist, delete
	if len(realAlloc) == 0 {
		hwlog.RunLog.Warn("not allocate any device")
		return
	}
	// delete klt allocate device in key
	for _, id := range kltAlloc {
		if _, exist := ps.klt2RealDevMap[id]; exist {
			delete(ps.klt2RealDevMap, id)
		}
	}
	// delete real allocate device in value
	for _, id := range realAlloc {
		for k, v := range ps.klt2RealDevMap {
			if v == id {
				delete(ps.klt2RealDevMap, k)
			}
		}
	}
	isVirtualDev := common.IsVirtualDev(realAlloc[0])
	if isVirtualDev && len(realAlloc) > 1 {
		hwlog.RunLog.Warnf("virtual device only support allocate one, %v", realAlloc)
		return
	}
	// for virtual device, N ai core : 1 real device
	// aicore-0, aicore-1 : Ascend910-2c-100-0
	if isVirtualDev {
		for _, id := range kltAlloc {
			ps.klt2RealDevMap[id] = realAlloc[0]
		}
		return
	}
	// for physical device, M ai core : N real device
	// aicore-0,..., aicore-31 : Ascend910-0
	// aicore-32,..., aicore-63 : Ascend910-1
	chipAICore := ps.manager.GetChipAICore()
	if int(chipAICore)*len(realAlloc) != len(kltAlloc) {
		hwlog.RunLog.Warnf("klt allocate core not equal real allocate %v", realAlloc)
		return
	}
	realIdx := 0
	for kltIdx, id := range kltAlloc {
		ps.klt2RealDevMap[id] = realAlloc[realIdx]
		if ((kltIdx + 1) % int(chipAICore)) == 0 {
			realIdx++
		}
	}
}

func (ps *PluginServer) updatePresetAllocMap(realAlloc, kltAlloc []string) {
	if len(realAlloc) != len(kltAlloc) {
		hwlog.RunLog.Error("number of devices of klt allocate not equal real allocate")
		return
	}
	ps.allocMapLock.Lock()
	for _, id := range kltAlloc {
		if _, exist := ps.klt2RealDevMap[id]; exist {
			delete(ps.klt2RealDevMap, id)
		}
	}
	for i, id := range kltAlloc {
		ps.klt2RealDevMap[id] = realAlloc[i]
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
	realAllocate := sets.String{}
	if !common.ParamOption.UseVolcanoType {
		return kltAllocate, nil
	}
	for _, id := range kltAllocate {
		realID, exist := ps.klt2RealDevMap[id]
		if !exist {
			return nil, fmt.Errorf("cannot found real allocate device by %s", id)
		}
		realAllocate.Insert(realID)
	}
	return realAllocate.List(), nil
}

// GetKltAndRealAllocateDev get kubelet and real allocate device of pod
func (ps *PluginServer) GetKltAndRealAllocateDev(podList []v1.Pod) ([]PodDeviceInfo, error) {
	prClient := NewPodResource()
	podDevice, err := prClient.GetPodResource()
	if err != nil {
		return nil, fmt.Errorf("get pod resource failed, %#v", err)
	}
	var podDeviceInfo []PodDeviceInfo
	for _, pod := range podList {
		podKey := pod.Namespace + common.UnderLine + pod.Name
		podResource, exist := podDevice[podKey]
		if !exist {
			continue
		}
		if podResource.ResourceName != common.ResourceNamePrefix+ps.deviceType {
			hwlog.RunLog.Debugf("podKey %s resource name %s not equal device type %s", podKey,
				podResource.ResourceName, ps.deviceType)
			continue
		}
		if common.ParamOption.PresetVDevice && common.IsVirtualDev(ps.deviceType) {
			podDeviceInfo = append(podDeviceInfo, PodDeviceInfo{Pod: pod, KltDevice: podResource.DeviceIds,
				RealDevice: podResource.DeviceIds})
			continue
		}
		realDeviceList, err := ps.GetRealAllocateDevices(podResource.DeviceIds)
		if err != nil {
			realDevice, exist := pod.Annotations[common.ResourceNamePrefix+common.PodRealAlloc]
			if exist {
				realDeviceList = strings.Split(realDevice, common.CommaSepDev)
				ps.updateAllocMap(realDeviceList, podResource.DeviceIds)
			} else {
				hwlog.RunLog.Warnf("%s not found real allocate device", podKey)
				continue
			}
		}
		podDeviceInfo = append(podDeviceInfo, PodDeviceInfo{Pod: pod, KltDevice: podResource.DeviceIds,
			RealDevice: realDeviceList})
	}
	return podDeviceInfo, nil
}

// DestroyNotUsedVNPU destroy not used virtual device
func (ps *PluginServer) DestroyNotUsedVNPU() error {
	allDevInfo, err := ps.manager.GetNPUs()
	if err != nil {
		return err
	}
	podList := ps.manager.GetKubeClient().GetAllPodListCache()
	podDeviceInfo, err := ps.GetKltAndRealAllocateDev(podList)
	if err != nil {
		return err
	}
	usedDevice := ps.removeVGroup(podDeviceInfo)
	var needToDestroy []string
	for _, dev := range allDevInfo.AllDevs {
		if !usedDevice.Has(dev.DeviceName) {
			needToDestroy = append(needToDestroy, dev.DeviceName)
		}
	}
	for _, dev := range needToDestroy {
		if !common.IsVirtualDev(dev) {
			continue
		}
		if err = ps.manager.DestroyVirtualDevice(dev); err == nil {
			hwlog.RunLog.Infof("destroy virtual device %s success", dev)
		} else {
			hwlog.RunLog.Infof("destroy virtual device %s failed, %v", dev, err)
		}
	}
	return nil
}

func (ps *PluginServer) removeVGroup(podDeviceInfo []PodDeviceInfo) sets.String {
	usedDevice := sets.String{}
	for _, deviceInfo := range podDeviceInfo {
		usedDevice.Insert(deviceInfo.RealDevice...)
	}
	noVGroupDevice := sets.String{}
	for dev := range usedDevice {
		vDevAndGroup := strings.Split(dev, common.UnderLine)
		if len(vDevAndGroup) == 1 || len(vDevAndGroup) == common.VGroupAndDevLen {
			noVGroupDevice.Insert(vDevAndGroup[0])
		}
	}
	return noVGroupDevice
}

func checkAnnotationAllocateValid(requestDevices []string, deviceType string, pod *v1.Pod, chipAICore int32) bool {
	if predicateTime, ok := pod.Annotations[common.PodPredicateTime]; ok {
		if predicateTime == strconv.FormatUint(math.MaxUint64, common.BaseDec) {
			hwlog.RunLog.Debugf("The pod has been mounted to a device, pod name: %s", pod.Name)
			return false
		}
	}
	if common.ParamOption.PresetVDevice {
		allocateDevice, err := common.GetDeviceFromPodAnnotation(pod, deviceType)
		if err != nil {
			return false
		}
		return len(allocateDevice) == len(requestDevices)
	}
	// for dynamic segment
	annotation, err := common.GetPodAnnotationByDeviceType(pod, deviceType)
	if err != nil {
		hwlog.RunLog.Warn(err)
		return false
	}
	deviceInfos := strings.Split(annotation, common.MiddelLine)
	// for vnpu, like huawei.com/npu-core:0-vir02
	if len(deviceInfos) > 1 {
		_, template, err := common.GetVNPUSegmentInfo(deviceInfos)
		if err != nil {
			hwlog.RunLog.Warn(err)
			return false
		}
		aiCore, err := common.GetAICore(template)
		if err != nil {
			hwlog.RunLog.Warn(err)
			return false
		}
		return len(requestDevices) == aiCore
	}
	// for physical npu, huawei.com/npu-core:0,1,2,3
	phyDevices := strings.Split(deviceInfos[0], common.CommaSepDev)
	return len(requestDevices) == len(phyDevices)*int(chipAICore)
}

// getAICoreFromPodAnnotation get ai core count from pod annotation
// Annotation
// huawei.com/npu-core:0,1,2,3
// huawei.com/npu-core:0-vir02
func (ps *PluginServer) getAICoreFromPodAnnotation(pod *v1.Pod, deviceType string) ([]string, error) {
	if err := ps.DestroyNotUsedVNPU(); err != nil {
		return nil, err
	}
	annotation, err := common.GetPodAnnotationByDeviceType(pod, deviceType)
	if err != nil {
		return nil, err
	}
	deviceInfos := strings.Split(annotation, common.MiddelLine)
	if len(deviceInfos) > 1 {
		phyID, templateName, err := common.GetVNPUSegmentInfo(deviceInfos)
		if err != nil {
			return nil, err
		}
		deviceName, err := ps.manager.CreateVirtualDevice(phyID, templateName)
		if err != nil {
			return nil, err
		}
		ps.ascendRuntimeOptions = common.VirtualDev
		// like Ascend910-2c-100-0
		return []string{deviceName}, nil
	}
	ps.ascendRuntimeOptions = ""
	var phyDevs []string
	ids := strings.Split(deviceInfos[0], common.CommaSepDev)
	for _, id := range ids {
		phyDevs = append(phyDevs, fmt.Sprintf("%s-%s", ps.manager.GetName(), id))
	}
	inValidIDList := ps.isValidRequestID(ids)
	if len(inValidIDList) != 0 {
		hwlog.RunLog.Errorf("volcano allocated id %s is invalid", inValidIDList)
		return nil, fmt.Errorf(common.NoNPUResource)
	}
	// like Ascend910-0,Ascend910-1,Ascend910-2,Ascend910-3
	return phyDevs, nil
}

func (ps *PluginServer) isValidRequestID(phyDevs []string) []string {
	var inValidIDList []string
	for _, phyID := range phyDevs {
		if ps.isValidPhyID(phyID) {
			continue
		}
		inValidIDList = append(inValidIDList, phyID)
	}
	return inValidIDList
}

func (ps *PluginServer) isValidPhyID(phyID string) bool {
	for _, cacheDev := range ps.cachedDevices {
		if phyID == strconv.Itoa(int(cacheDev.PhyID)) {
			return true
		}
	}
	return false
}

func (ps *PluginServer) doWithVolcanoSchedule(requestDevices []string) ([]string, error) {
	conditionFunc := func(pod *v1.Pod) bool {
		return checkAnnotationAllocateValid(requestDevices, ps.deviceType, pod, ps.manager.GetChipAICore())
	}
	var filteredPods []v1.Pod
	var allPods []v1.Pod
	for i := 0; i < common.GetPodFromInformerTime; i++ {
		if i == common.GetPodFromInformerTime-1 {
			// in the last time of retry, get the pod from api server instead of cache
			noneCachedPod, err := ps.manager.GetKubeClient().GetActivePodList()
			if err != nil {
				hwlog.RunLog.Errorf("get active pod from api server failed")
				return nil, err
			}
			allPods = noneCachedPod
		} else {
			allPods = ps.manager.GetKubeClient().GetActivePodListCache()
		}
		filteredPods = common.FilterPods(allPods, ps.deviceType, conditionFunc)
		if len(filteredPods) != 0 {
			break
		}
		hwlog.RunLog.Warnf("no pod passed the filter, request device: %v, retry: %d", requestDevices, i)
		time.Sleep(time.Second)
	}
	oldestPod := ps.getOldestPod(filteredPods)
	if oldestPod == nil {
		return nil, fmt.Errorf("not get valid pod")
	}
	var allocateDevices []string
	var err error
	if !common.ParamOption.PresetVDevice {
		common.LockAllDeviceInfo()
		allocateDevices, err = ps.getAICoreFromPodAnnotation(oldestPod, ps.deviceType)
		common.UnlockAllDeviceInfo()
	} else {
		allocateDevices, err = common.GetDeviceFromPodAnnotation(oldestPod, ps.deviceType)
	}
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

func mountDevice(resp *v1beta1.ContainerAllocateResponse, devices []int, ascendRuntimeOptions string) {
	for _, deviceID := range devices {
		containerPath, hostPath := getDevPath(fmt.Sprintf("%d", deviceID), ascendRuntimeOptions)
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
			ContainerPath: getDeviceContainerPath(d),
			Permissions:   "rw",
		})
	}
}

func getDeviceContainerPath(hostPath string) string {
	if hostPath == common.HiAIManagerDeviceDocker {
		return common.HiAIManagerDevice
	}
	return hostPath
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
		if !common.ParamOption.PresetVDevice {
			hwlog.RunLog.Infof("request num: %d", len(rqt.DevicesIDs))
		} else {
			hwlog.RunLog.Infof("request: %#v", rqt.DevicesIDs)
		}
		if common.ParamOption.UseVolcanoType {
			allocateDevices, err = ps.useVolcano(rqt.DevicesIDs)
			if err != nil {
				hwlog.RunLog.Error(err)
				return nil, err
			}
		}
		_, ascendVisibleDevices, err := common.GetDeviceListID(allocateDevices, ps.ascendRuntimeOptions)
		if err != nil {
			hwlog.RunLog.Error(err)
			return nil, err
		}

		resp := new(v1beta1.ContainerAllocateResponse)
		if !common.ParamOption.UseAscendDocker {
			hwlog.RunLog.Info("device-plugin will use origin mount way")
			mountDefaultDevice(resp, ps.defaultDevs)
			mountDevice(resp, ascendVisibleDevices, ps.ascendRuntimeOptions)
		} else {
			common.SetAscendRuntimeEnv(ascendVisibleDevices, ps.ascendRuntimeOptions, resp)
			hwlog.RunLog.Info("device-plugin will use ascend-docker to mount")
		}
		resps.ContainerResponses = append(resps.ContainerResponses, resp)
	}
	return resps, nil
}

// GetPreferredAllocation implement the kubelet device plugin interface
func (ps *PluginServer) GetPreferredAllocation(context.Context, *v1beta1.PreferredAllocationRequest) (
	*v1beta1.PreferredAllocationResponse, error) {
	return nil, fmt.Errorf("not support")
}

// GetDevicePluginOptions is Standard interface to kubelet.
func (ps *PluginServer) GetDevicePluginOptions(ctx context.Context, e *v1beta1.Empty) (*v1beta1.DevicePluginOptions,
	error) {
	return &v1beta1.DevicePluginOptions{}, nil
}

// PreStartContainer is Standard interface to kubelet with empty implement.
func (ps *PluginServer) PreStartContainer(ctx context.Context,
	r *v1beta1.PreStartContainerRequest) (*v1beta1.PreStartContainerResponse, error) {
	hwlog.RunLog.Info("PreStart just call in UT.")
	return &v1beta1.PreStartContainerResponse{}, nil
}

// NewPluginServer returns an initialized PluginServer
func NewPluginServer(deviceType string, devices []*common.NpuDevice, defaultDevs []string,
	manager device.DevManager) *PluginServer {
	ps := &PluginServer{
		restart:        true,
		reciChan:       make(chan interface{}),
		deviceType:     deviceType,
		defaultDevs:    defaultDevs,
		stop:           make(chan interface{}),
		klt2RealDevMap: make(map[string]string, common.MaxDevicesNum),
		isRunning:      common.NewAtomicBool(false),
		manager:        manager,
	}
	ps.deepCopyDevice(devices)
	return ps
}
