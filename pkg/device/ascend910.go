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

// Package device a series of device function
package device

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"huawei.com/npu-exporter/v5/common-utils/hwlog"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"Ascend-device-plugin/pkg/common"
)

const (
	networkDetectOK   = uint32(0)
	networkDetectInit = uint32(6)
)

var (
	lastTimeNetworkRecoverDevices sets.String
	hotResetManagerInitOnce       sync.Once
)

// HwAscend910Manager manages huawei Ascend910 devices.
type HwAscend910Manager struct {
	AscendTools
	hotResetManager HotResetManager
}

// NewHwAscend910Manager is used to create ascend 910 manager
func NewHwAscend910Manager() *HwAscend910Manager {
	return &HwAscend910Manager{
		AscendTools: AscendTools{
			name:         common.Ascend910,
			unHealthyKey: common.HuaweiUnHealthAscend910,
			devCount:     common.MaxDevicesNum,
		},
	}
}

// GetNPUs Discovers all HUAWEI Ascend910 devices by call devmanager interface
// a physical npu can be split into multiple vNPU
// vNPU is classification by computing power, like Ascend910-4c, Ascend910-8c, Ascend910-16c
// physical npu sets corresponding to the deviTypes, and vNPU is vDeviTypes
// vDeviTypes may is: [Ascend910-4c, Ascend910-4c, Ascend910-8c], also deviTypes may is: [Ascend910, Ascend910]
// one class deviType will generate a socket file, like ascend910-4c.sock or Ascend910.sock, so we deduplicate
func (hnm *HwAscend910Manager) GetNPUs() (common.NpuAllInfo, error) {
	devNum, devList, err := hnm.dmgr.GetDeviceList()
	if err != nil {
		return common.NpuAllInfo{}, err
	}
	if devNum > hnm.devCount {
		return common.NpuAllInfo{}, fmt.Errorf("invalid device num: %d", devNum)
	}
	var allDevices []common.NpuDevice
	var aiCoreDevices []*common.NpuDevice
	var allDeviceTypes []string
	for i := int32(0); i < devNum; i++ {
		davinCiDev, err := hnm.getDavinCiDev(devList[i])
		if err != nil {
			return common.NpuAllInfo{}, err
		}
		vDevInfos, err := hnm.getVirtualDevice(devList[i])
		if err != nil {
			hwlog.RunLog.Errorf("The virtual device is considered not exist, please check the error: %#v", err)
		}
		if vDevInfos.TotalResource.VDevNum > common.MaxVirtualDeviceNum {
			return common.NpuAllInfo{}, fmt.Errorf("invalid virtual device count")
		}
		if !common.ParamOption.PresetVDevice {
			common.FakeAiCoreDevice(davinCiDev, &aiCoreDevices)
		}
		if vDevInfos.TotalResource.VDevNum == 0 {
			hnm.assemblePhyDevices(davinCiDev, &allDevices, &allDeviceTypes)
			continue
		}
		hnm.assembleVirtualDevices(davinCiDev, vDevInfos, &allDevices, &allDeviceTypes)
	}
	allDeviceTypes = hnm.removeDuplicate(&allDeviceTypes)
	return common.NpuAllInfo{AllDevs: allDevices, AICoreDevs: aiCoreDevices, AllDevTypes: allDeviceTypes}, nil
}

// GraceTolerance process training task with device fault gracefully
func (hnm *HwAscend910Manager) GraceTolerance(classifyDevs map[string][]*common.NpuDevice) {
	hotResetManagerInitOnce.Do(func() {
		hnm.hotResetManager = NewHotResetManager(common.ParamOption.RealCardType)
	})
	if hnm.hotResetManager == nil {
		hwlog.RunLog.Debugf("hot reset manager is nil, devType: %s", common.ParamOption.RealCardType)
		return
	}
	// 1. obtain the current device status and update the cache of hot reset manager
	if err := hnm.updateHotResetCache(classifyDevs); err != nil {
		hwlog.RunLog.Errorf("failed to update hot reset cache, err: %#v", err)
		return
	}
	// 2. performs graceful fault tolerance for tasks to be processed based on the device information in the cache
	if err := hnm.processAllTask(); err != nil {
		hwlog.RunLog.Errorf("failed to process task, err: %#v", err)
		return
	}
	// 3. filter the faulty device in the reset state in the device info cm to avoid rescheduling
	if err := hnm.filterDevStatus(classifyDevs); err != nil {
		hwlog.RunLog.Errorf("failed to filter device status,err: %#v", err)
		return
	}
}

// DoWithVolcanoListAndWatch ascend910 affinity scheduling
func (hnm *HwAscend910Manager) DoWithVolcanoListAndWatch(classifyDevs map[string][]*common.NpuDevice) {
	devStatusSet := hnm.getDevStatesDevSet(classifyDevs)
	if err := hnm.UpdateNodeDeviceInfo(devStatusSet, hnm.updateDeviceInfo); err != nil {
		hwlog.RunLog.Errorf("update device info failed, err: %#v", err)
	}
}

func (tool *AscendTools) getDeviceNetworkState(logicID int32) string {
	healthCode, err := tool.dmgr.GetDeviceNetWorkHealth(logicID)
	if err != nil {
		hwlog.RunLog.Warnf("get logicID %d network health failed, error code is %d", logicID, healthCode)
		return v1beta1.Unhealthy
	}

	switch healthCode {
	case networkDetectOK, networkDetectInit:
		return v1beta1.Healthy
	default:
		hwlog.RunLog.Debugf("%d network status is unhealthy, health code is %d", logicID, healthCode)
		return v1beta1.Unhealthy
	}
}

func (hnm *HwAscend910Manager) updateDeviceInfo(oldDevInfo, newDevInfo map[string]string,
	devStatusSet common.DevStatusSet) error {
	if newDevInfo == nil {
		return fmt.Errorf("invalid new device info")
	}
	nodeFmtDevRecover, nodeFmtDevNetRecover := sets.String{}, sets.String{}
	curNode, err := hnm.getRecoverLabelFromNodeSets(&nodeFmtDevRecover, &nodeFmtDevNetRecover)
	if err != nil {
		return err
	}
	newDevRecoverLabel, newAscend910 := hnm.getHealthAndRecoverDev(devStatusSet, nodeFmtDevRecover,
		common.ConvertDevListToSets(oldDevInfo[common.HuaweiUnHealthAscend910], common.CommaSepDev))
	newNetRecoverSets, newNetUHDevSets := hnm.getNewNetworkRecoverDev(devStatusSet.NetUnHealthyDevice,
		common.ConvertDevListToSets(oldDevInfo[common.HuaweiNetworkUnHealthAscend910], common.CommaSepDev),
		nodeFmtDevNetRecover)
	newDevInfo[common.HuaweiAscend910] = newAscend910
	newDevInfo[common.HuaweiUnHealthAscend910] = common.ToString(devStatusSet.UnHealthyDevice, common.CommaSepDev)
	newDevInfo[common.HuaweiNetworkUnHealthAscend910] = common.ToString(newNetUHDevSets, common.CommaSepDev)
	if common.ParamOption.AutoStowingDevs {
		return nil
	}
	if err := hnm.updateNodeLabel(curNode, newDevRecoverLabel, hnm.getPatchLabel(newNetRecoverSets)); err != nil {
		hwlog.RunLog.Errorf("update node label failed, err: %#v", err)
		return err
	}
	lastTimeNetworkRecoverDevices = newNetRecoverSets
	return nil
}

func (hnm *HwAscend910Manager) updateNodeLabel(curNode *v1.Node, devRecoverLabel, netRecoverLabel string) error {
	newNode := curNode.DeepCopy()
	newNode.Labels[common.HuaweiRecoverAscend910] = devRecoverLabel
	newNode.Labels[common.HuaweiNetworkRecoverAscend910] = netRecoverLabel
	hwlog.RunLog.Debugf("newNode.Labels: %#v", newNode.Labels)
	updatedNode, _, err := hnm.client.PatchNodeState(curNode, newNode)
	if err != nil {
		return err
	}
	hwlog.RunLog.Debugf("updatedNode.Labels: %#v", updatedNode.Labels)
	return nil
}

func (hnm *HwAscend910Manager) getHealthAndRecoverDev(curDevStatusSet common.DevStatusSet, devRecoverDev,
	recordUHDev sets.String) (string, string) {
	device910 := curDevStatusSet.FreeHealthyDevice[common.Ascend910]
	if common.ParamOption.AutoStowingDevs {
		return "", common.ToString(device910, common.CommaSepDev)
	}
	addRecoverSets := recordUHDev.Difference(curDevStatusSet.UnHealthyDevice)
	devRecoverSets := devRecoverDev.Union(addRecoverSets)
	newDevice910 := device910.Difference(devRecoverSets)
	return hnm.getPatchLabel(devRecoverSets), common.ToString(newDevice910, common.CommaSepDev)
}

// getNewNetworkRecoverDev , return new devices to be restored and network unhealthy device in this times
func (hnm *HwAscend910Manager) getNewNetworkRecoverDev(totalNetUHDev, devInfoNetUHRecord,
	labelRecoverRecord sets.String) (sets.String, sets.String) {
	// devInfoNetUHRecord means device info record network unhealthy devices
	// labelRecoverRecord means device's network is ok and to be restored
	// if there is no network unhealthy device and autoStowing devices is true
	if common.ParamOption.AutoStowingDevs {
		return sets.String{}, totalNetUHDev
	}
	// devices recovered between the last check and this check
	recoveredDevSets := lastTimeNetworkRecoverDevices.Difference(labelRecoverRecord)

	newNetworkRecoverDevSets := devInfoNetUHRecord.Difference(totalNetUHDev)
	// remove the device that network is unhealthy in this times
	newNetworkRecoverDevSets = newNetworkRecoverDevSets.Difference(labelRecoverRecord.Intersection(totalNetUHDev))
	// remove the device that recovered
	newNetworkRecoverDevSets = newNetworkRecoverDevSets.Difference(recoveredDevSets)
	newNetworkUnhealthyDevSets := devInfoNetUHRecord.Union(totalNetUHDev).Difference(recoveredDevSets)
	return newNetworkRecoverDevSets, newNetworkUnhealthyDevSets
}

// getPatchLabel get elements one by one from the sets and change the element "Ascend910-x" to "x"
// which will patch to node
func (hnm *HwAscend910Manager) getPatchLabel(chips sets.String) string {
	if chips.Len() == 0 {
		return ""
	}

	var ascendLabel []string
	for devName := range chips {
		devTypeAndID := strings.Split(devName, common.MiddelLine)
		if len(devTypeAndID) != common.LabelDeviceLen {
			continue
		}
		phyID := devTypeAndID[len(devTypeAndID)-1]
		if _, isValidNum := common.IsValidNumber(phyID); !isValidNum {
			continue
		}
		ascendLabel = append(ascendLabel, phyID)
	}

	return strings.Join(ascendLabel, common.DotSepDev)
}

func (hnm *HwAscend910Manager) getRecoverLabelFromNodeSets(devRecoverLabel, netRecoverLabel *sets.String) (
	*v1.Node, error) {
	if common.ParamOption.AutoStowingDevs {
		return nil, nil
	}
	curNode, err := hnm.client.GetNode()
	if err != nil {
		hwlog.RunLog.Error("get node error")
		return nil, err
	}
	if curNode == nil || curNode.Labels == nil {
		return nil, fmt.Errorf("invalid node")
	}
	// devRecoverLabel like Ascend910-0,Ascend910-2,Ascend910-3, means dev healthy exception
	*devRecoverLabel = hnm.toStandardDeviceFmt(common.ConvertDevListToSets(
		curNode.Labels[common.HuaweiRecoverAscend910], common.DotSepDev))
	// netRecoverLabel like Ascend910-0,Ascend910-2,Ascend910-3, means dev network exception
	*netRecoverLabel = hnm.toStandardDeviceFmt(common.ConvertDevListToSets(
		curNode.Labels[common.HuaweiNetworkRecoverAscend910], common.DotSepDev))
	return curNode, nil
}

// toStandardDeviceFmt convert physical id "x" to format "Ascend910-x"
func (hnm *HwAscend910Manager) toStandardDeviceFmt(devices sets.String) sets.String {
	if devices.Len() == 0 {
		return sets.String{}
	}

	standardSets := sets.String{}
	for devID := range devices {
		deviceName := fmt.Sprintf("%s-%s", common.Ascend910, devID)
		standardSets.Insert(deviceName)
	}

	return standardSets
}

func (hnm *HwAscend910Manager) updateHotResetCache(classifyDevs map[string][]*common.NpuDevice) error {
	deviceList, ok := classifyDevs[common.Ascend910]
	if !ok {
		hwlog.RunLog.Error("ascend 910 device list no found")
		return fmt.Errorf("ascend 910 device list not found")
	}
	if err := hnm.hotResetManager.UpdateGlobalDevFaultInfoCache(deviceList); err != nil {
		hwlog.RunLog.Errorf("failed to update global device fault info cache, err: %#v", err)
		return err
	}
	if err := hnm.setTaskDevInfoCache(); err != nil {
		hwlog.RunLog.Errorf("failed to set task device info cache, err: %#v", err)
		return err
	}
	return nil
}

func (hnm *HwAscend910Manager) setTaskDevInfoCache() error {
	podList, err := hnm.client.GetAllPodList()
	if err != nil {
		hwlog.RunLog.Errorf("failed to get pod list when update task, err: %#v", err)
		return err
	}
	newTaskDevListCache := make(map[string][]int32)
	newTaskDevFaultInfoCache := make(map[string][]*common.TaskDevInfo)
	newTaskNamespaceCache := make(map[string]string)
	taskListUsedDevice := make(map[string]struct{})
	for _, pod := range podList.Items {
		annotationTag := fmt.Sprintf("%s%s", common.ResourceNamePrefix, common.Ascend910)
		tmpNpu, ok := pod.Annotations[annotationTag]
		if !ok || len(tmpNpu) == 0 || len(tmpNpu) > common.PodAnnotationMaxLength {
			continue
		}
		devIdList := hnm.hotResetManager.GetDevIdList(tmpNpu)
		if len(devIdList) < hnm.hotResetManager.GetRingNum() {
			continue
		}
		taskName, ok := pod.Annotations[common.ResetTaskNameKey]
		if !ok {
			hwlog.RunLog.Error("failed to get task name by task key")
			continue
		}

		rankIndex, ok := pod.Annotations[common.RankIndexKey]
		if !ok {
			hwlog.RunLog.Error("failed to get rank index by rank index key")
			continue
		}
		taskListUsedDevice[taskName] = struct{}{}
		newTaskDevListCache[taskName] = devIdList
		taskDevFaultInfoList, err := hnm.hotResetManager.GenerateTaskDevFaultInfoList(devIdList, rankIndex)
		if err != nil {
			hwlog.RunLog.Errorf("failed to get task device fault info list, err: %#v", err)
			return err
		}
		newTaskDevFaultInfoCache[taskName] = taskDevFaultInfoList
		newTaskNamespaceCache[taskName] = pod.Namespace
	}
	hnm.hotResetManager.UpdateFreeTask(taskListUsedDevice)
	if err := hnm.hotResetManager.UpdateTaskDevListCache(newTaskDevListCache); err != nil {
		return err
	}
	if err := hnm.hotResetManager.UpdateTaskDevFaultInfoCache(newTaskDevFaultInfoCache); err != nil {
		return err
	}
	if err := hnm.hotResetManager.UpdateTaskNamespaceCache(newTaskNamespaceCache); err != nil {
		return err
	}
	return nil
}

func (hnm *HwAscend910Manager) isTaskInReset(taskName string) (bool, error) {
	namespace, err := hnm.hotResetManager.GetTaskNamespace(taskName)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get task namespace, err: %#v", err)
		return false, err
	}
	resetCM, err := hnm.client.GetConfigMap(common.ResetInfoCMNamePrefix+taskName, namespace)
	if err != nil {
		if errors.IsNotFound(err) {
			hwlog.RunLog.Debugf("task %s does not have reset info cm, skip this choice", taskName)
			return true, nil
		}
		hwlog.RunLog.Errorf("failed to reset info cm, err: %#v", err)
		return false, err
	}
	if hnm.hotResetManager.IsCurNodeTaskInReset(taskName) {
		hwlog.RunLog.Infof("this node task %s is resetting, skip once process", taskName)
		return true, nil
	}
	resetInfoData, err := getResetInfoData(resetCM)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get reset info data, err: %#v", err)
		return false, err
	}
	if len(resetInfoData) == 0 {
		return false, nil
	}
	hwlog.RunLog.Infof("global task %s is resetting, skip once process", taskName)
	return true, nil
}

// filterDevStatus filters the health of the device being reset and
// the network health of the ring that the device is on
func (hnm *HwAscend910Manager) filterDevStatus(classifyDevs map[string][]*common.NpuDevice) error {
	devStatusList, ok := classifyDevs[common.Ascend910]
	if !ok {
		return fmt.Errorf("no ascend 910 device needed filter")
	}
	devInReset := hnm.hotResetManager.GetDevListInReset()
	filteredRingIndex := -1
	for _, devStatus := range devStatusList {
		if _, ok := devInReset[devStatus.LogicID]; !ok || devStatus.Health == v1beta1.Healthy {
			continue
		}

		devStatus.Health = v1beta1.Healthy
		ringNum := hnm.hotResetManager.GetRingNum()
		ringIndex := int(devStatus.LogicID) / ringNum
		if ringIndex != filteredRingIndex {
			starDevIndex := ringIndex * ringNum
			endDevIndex := starDevIndex + ringNum
			for devIndex := starDevIndex; devIndex < endDevIndex; devIndex++ {
				devStatusList[devIndex].NetworkHealth = v1beta1.Healthy
			}
			filteredRingIndex = ringIndex
		}
	}
	return nil
}

func (hnm *HwAscend910Manager) processAllTask() error {
	taskDevFaultInfoList := hnm.hotResetManager.GetAllTaskDevFaultInfoList()
	for taskName := range taskDevFaultInfoList {
		policy, policyLevel, err := hnm.hotResetManager.GetTaskProcessPolicy(taskName)
		if err != nil {
			hwlog.RunLog.Errorf("failed to get task %s process policy, err: %#v", taskName, err)
			continue
		}
		if policyLevel != common.RestartErrorLevel && policyLevel != common.ResetErrorLevel {
			continue
		}
		if resetFlag, err := hnm.isTaskInReset(taskName); err != nil || resetFlag {
			continue
		}
		resetInfo, err := hnm.preProcess(taskName, policy)
		if err != nil {
			return err
		}
		if err := hnm.runProcessTask(taskName, policyLevel, resetInfo); err != nil {
			return err
		}
	}
	return nil
}

func (hnm *HwAscend910Manager) runProcessTask(taskName string, policyLevel int, resetInfo *common.TaskResetInfo) error {
	switch policyLevel {
	case common.RestartErrorLevel:
		go hnm.restartProcess(taskName, resetInfo)
	case common.ResetErrorLevel:
		go hnm.resetProcess(taskName, resetInfo)
	default:
		return fmt.Errorf("invalid processing policy")
	}
	return nil
}

func (hnm *HwAscend910Manager) restartProcess(taskName string, resetInfo *common.TaskResetInfo) {
	defer func() {
		if err := hnm.postProcess(taskName, resetInfo); err != nil {
			hwlog.RunLog.Errorf("failed to unset device in reset, err %#v", err)
		}
	}()
	devFaultInfoList, err := hnm.hotResetManager.GetTaskDevFaultInfoList(taskName)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get task device fault info list, err %#v", err)
		return
	}
	devFaultInfoListInReset := hnm.hotResetManager.DeepCopyDevFaultInfoList(devFaultInfoList)
	time.Sleep(common.WaitFlushCMTime * time.Second)
	if err := hnm.refreshDevFaultInfo(devFaultInfoList); err != nil {
		hwlog.RunLog.Errorf("failed to refresh device fault info, err %#v", err)
		return
	}
	if err := hnm.upgradeRestartProcess(taskName, devFaultInfoList); err != nil {
		hwlog.RunLog.Errorf("failed to exec upgrade restart process, err: %#v", err)
		return
	}
	if err := hnm.updateResetCMStatus(taskName, common.RestartError, common.RecoveredStatus,
		devFaultInfoListInReset); err != nil {
		hwlog.RunLog.Errorf("failed to update reset cm to recovered status, err: %#v", err)
		return
	}
	return
}

// upgradeRestartProcess upgrade the device restart processing to the device reset processing
func (hnm *HwAscend910Manager) upgradeRestartProcess(taskName string, devFaultInfoList []*common.TaskDevInfo) error {
	restartFaultInfoList, err := hnm.hotResetManager.GetNeedRestartDevList(devFaultInfoList)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get need reset device list, err %#v", err)
		return err
	}
	if len(restartFaultInfoList) == 0 {
		return nil
	}
	if err := hnm.resetDeviceOnce(devFaultInfoList); err != nil {
		return err
	}
	resultFaultInfoList, err := hnm.hotResetManager.GetNeedRestartDevList(devFaultInfoList)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get need reset device list, err: %#v", err)
		return err
	}
	if len(resultFaultInfoList) == 0 {
		return nil
	}
	if err := hnm.updateResetCMStatus(taskName, common.RestartError, common.RecoverFailedStatus,
		devFaultInfoList); err != nil {
		hwlog.RunLog.Errorf("failed to update reset cm to recover failed status, err: %#v", err)
		return err
	}
	return fmt.Errorf("failed to restart task, upgrade recovery failed status")
}

func (hnm *HwAscend910Manager) updateResetCMStatus(taskName, policyType, status string,
	devFaultInfoList []*common.TaskDevInfo) error {
	newResetInfo, err := hnm.hotResetManager.GetTaskResetInfo(devFaultInfoList, policyType, status)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get task reset info list, err: %#v", err)
		return err
	}
	namespace, err := hnm.hotResetManager.GetTaskNamespace(taskName)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get task namespace, err: %#v", err)
		return err
	}
	if _, err := hnm.client.WriteResetInfoDataIntoCM(taskName, namespace, newResetInfo); err != nil {
		hwlog.RunLog.Errorf("failed to write reset info to cm, err: %#v", err)
		return err
	}
	time.Sleep(common.WaitFlushCMTime * time.Second)
	return nil
}

func (hnm *HwAscend910Manager) resetProcess(taskName string, resetInfo *common.TaskResetInfo) {
	defer func() {
		if err := hnm.postProcess(taskName, resetInfo); err != nil {
			hwlog.RunLog.Errorf("failed to exec post process, err: %#v", err)
		}
	}()
	devFaultInfoList, err := hnm.hotResetManager.GetTaskDevFaultInfoList(taskName)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get task device fault info list, err: %#v", err)
		return
	}
	devFaultInfoListInReset := hnm.hotResetManager.DeepCopyDevFaultInfoList(devFaultInfoList)
	time.Sleep(common.WaitFlushCMTime * time.Second)
	if err := hnm.resetDeviceOnce(devFaultInfoList); err != nil {
		hwlog.RunLog.Errorf("failed to reset device, err: %#v", err)
		return
	}
	if err := hnm.upgradeResetProcess(taskName, devFaultInfoList); err != nil {
		hwlog.RunLog.Errorf("failed to exec upgrade reset process, err :%#v", err)
		return
	}
	if err := hnm.updateResetCMStatus(taskName, common.RestartError, common.RecoveredStatus,
		devFaultInfoListInReset); err != nil {
		hwlog.RunLog.Errorf("failed to update reset cm to recovered status, err: %#v", err)
		return
	}
	if err := hnm.hotResetManager.UnSetTaskInReset(taskName); err != nil {
		hwlog.RunLog.Errorf("failed to unset task in reset, err: %#v", err)
		return
	}
	return
}

// upgradeResetProcess upgrade the device reset processing to the device isolation processing
func (hnm *HwAscend910Manager) upgradeResetProcess(taskName string, devFaultInfoList []*common.TaskDevInfo) error {
	resultFaultInfoList, err := hnm.hotResetManager.GetNeedResetDevList(devFaultInfoList)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get need reset device list, err: %#v", err)
		return err
	}
	if len(resultFaultInfoList) == 0 {
		return nil
	}
	if err := hnm.updateResetCMStatus(taskName, common.ResetError, common.RecoverFailedStatus,
		devFaultInfoList); err != nil {
		hwlog.RunLog.Errorf("failed to update reset cm to recover failed status, err: %#v", err)
		return err
	}
	return fmt.Errorf("failed to reset task, upgrade recovery failed status")
}

// preProcess write cm info, set task and device in reset
func (hnm *HwAscend910Manager) preProcess(taskName, policy string) (*common.TaskResetInfo, error) {
	devFaultInfoList, err := hnm.hotResetManager.GetTaskDevFaultInfoList(taskName)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get task device fault info list, err: %#v", err)
		return nil, err
	}
	namespace, err := hnm.hotResetManager.GetTaskNamespace(taskName)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get task namespace, err: %#v", err)
		return nil, err
	}
	resetInfo, err := hnm.hotResetManager.GetTaskResetInfo(devFaultInfoList, policy, common.UnrecoveredStatus)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get task reset info list, err: %#v", err)
		return nil, err
	}
	if _, err := hnm.client.WriteResetInfoDataIntoCM(taskName, namespace, resetInfo); err != nil {
		hwlog.RunLog.Errorf("failed to write reset info to cm, err: %#v", err)
		return nil, err
	}
	faultInfo, err := hnm.hotResetManager.GetTaskFaultRankInfo(devFaultInfoList)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get task fault rank info, err: %#v", err)
		return nil, err
	}
	if _, err := hnm.client.WriteFaultInfoDataIntoCM(taskName, namespace, faultInfo); err != nil {
		hwlog.RunLog.Errorf("failed to write fault rank info to cm, err %#v", err)
		return nil, err
	}
	if err := hnm.hotResetManager.SetTaskInReset(taskName); err != nil {
		hwlog.RunLog.Errorf("failed to set task %s in reset", taskName)
		return nil, err
	}
	if err := hnm.hotResetManager.SetAllDevInReset(resetInfo); err != nil {
		hwlog.RunLog.Errorf("failed to set all device in reset, err: %#v", err)
		return nil, err
	}
	return resetInfo, nil
}

// postProcess clear reset info cm and unset the reset status of all device in a task
func (hnm *HwAscend910Manager) postProcess(taskName string, resetInfo *common.TaskResetInfo) error {
	if err := hnm.hotResetManager.UnSetAllDevInReset(resetInfo); err != nil {
		hwlog.RunLog.Errorf("failed to unset all device in reset, err: %#v", err)
		return err
	}

	namespace, err := hnm.hotResetManager.GetTaskNamespace(taskName)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get task namespace, err: %#v", err)
		return err
	}
	if err := hnm.client.ClearResetInfo(taskName, namespace); err != nil {
		hwlog.RunLog.Errorf("failed to clear reset info, err: %#v", err)
		return err
	}
	return nil
}
func (hnm *HwAscend910Manager) refreshDevFaultInfo(devFaultInfo []*common.TaskDevInfo) error {
	for _, devInfo := range devFaultInfo {
		_, errorCode, err := hnm.GetDmgr().GetDeviceAllErrorCode(devInfo.LogicId)
		if err != nil {
			hwlog.RunLog.Errorf("failed to get device %d healthy", devInfo.LogicId)
			return err
		}
		devInfo.Policy = hnm.hotResetManager.GetDevProcessPolicy(common.GetFaultTypeByCode(errorCode))
		devInfo.ErrorCode = errorCode
	}
	return nil
}

func (hnm *HwAscend910Manager) resetDeviceOnce(devFaultInfoList []*common.TaskDevInfo) error {
	resetFaultInfoList, err := hnm.hotResetManager.GetNeedResetDevList(devFaultInfoList)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get need reset device list, err: %#v", err)
		return err
	}
	if err := hnm.execResetDevice(resetFaultInfoList); err != nil {
		hwlog.RunLog.Errorf("failed to exec reset device list, err: %#v", err)
		return err
	}
	time.Sleep(common.WaitFlushCMTime * time.Second)
	for _, devInfo := range devFaultInfoList {
		common.SetDeviceInit(devInfo.LogicId)
	}
	if err := hnm.refreshDevFaultInfo(devFaultInfoList); err != nil {
		hwlog.RunLog.Errorf("failed to refresh device fault info, err: %#v", err)
		return err
	}
	return nil
}

func (hnm *HwAscend910Manager) execResetDevice(devList map[int32]struct{}) error {
	errList := make([]error, 0, len(devList))
	for devLogicId := range devList {
		cardId, deviceId, err := hnm.GetDmgr().GetCardIDDeviceID(devLogicId)
		if err != nil {
			hwlog.RunLog.Errorf("failed to get reset device card id and device id, err %#v", err)
			return err
		}
		if err := hnm.tryResetDevice(cardId, deviceId); err != nil {
			errList = append(errList, err)
		}
	}
	if len(errList) == 0 {
		return nil
	}
	return errList[0]
}

func (hnm *HwAscend910Manager) tryResetDevice(cardId, deviceId int32) error {
	var realError error
	for i := 0; i < common.ResetRetryTimes; i++ {
		err := hnm.GetDmgr().SetDeviceReset(cardId, deviceId)
		if err == nil {
			hwlog.RunLog.Infof("reset cardId %d success", cardId)
			return nil
		}
		hwlog.RunLog.Errorf("cardId(%d) failed to reset device, err: %#v", cardId, err)
		realError = err
	}
	return realError
}
