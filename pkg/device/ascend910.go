// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package device a series of device function
package device

import (
	"fmt"
	"strconv"
	"strings"

	"huawei.com/npu-exporter/hwlog"
	"k8s.io/api/core/v1"
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
)

// HwAscend910Manager manages huawei Ascend910 devices.
type HwAscend910Manager struct {
	AscendTools
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
func (hnm *HwAscend910Manager) GetNPUs(allDevices *[]common.NpuDevice, allDeviceTypes *[]string, _ string) error {
	devNum, devList, err := hnm.dmgr.GetDeviceList()
	if err != nil {
		return err
	}
	if devNum > hnm.devCount {
		return fmt.Errorf("invalid device num: %d", devNum)
	}
	for i := int32(0); i < devNum; i++ {
		davinCiDev, err := hnm.getDavinCiDev(devList[i], hnm.getTemplateName2DeviceTypeMap())
		if err != nil {
			return err
		}
		vDevInfos, err := hnm.getVirtualDevice(devList[i])
		if err != nil {
			hwlog.RunLog.Errorf("The virtual device is considered not exist, please check the error: %#v", err)
		}
		if vDevInfos.TotalResource.VDevNum == 0 {
			hnm.assemblePhyDevices(davinCiDev, allDevices, allDeviceTypes)
			continue
		}
		hnm.assembleVirtualDevices(davinCiDev, vDevInfos, allDevices, allDeviceTypes)
	}
	*allDeviceTypes = hnm.removeDuplicate(allDeviceTypes)
	return nil
}

// DoWithVolcanoListAndWatch ascend910 affinity scheduling
func (hnm *HwAscend910Manager) DoWithVolcanoListAndWatch(classifyDevs map[string][]*common.NpuDevice) {
	devStatusSet := hnm.getDevStatesDevSet(classifyDevs)
	if err := hnm.UpdateNodeDeviceInfo(devStatusSet, hnm.updateDeviceInfo); err != nil {
		hwlog.RunLog.Errorf("update device info failed, err: %#v", err)
	}
}

// GetDeviceNetworkState check NPU network health
func (hnm *HwAscend910Manager) GetDeviceNetworkState(logicID int32, device *common.NpuDevice) (string, error) {
	healthCode, err := hnm.dmgr.GetDeviceNetWorkHealth(logicID)
	if err != nil {
		return "", err
	}

	switch healthCode {
	case networkDetectOK, networkDetectInit:
		return v1beta1.Healthy, nil
	default:
		hwlog.RunLog.Debugf("%s network status is unhealthy, error code is %d", device.DeviceName, healthCode)
		return v1beta1.Unhealthy, nil
	}
}

func (hnm *HwAscend910Manager) getTemplateName2DeviceTypeMap() map[string]string {
	return map[string]string{
		"vir16": common.Ascend910c16,
		"vir08": common.Ascend910c8,
		"vir04": common.Ascend910c4,
		"vir02": common.Ascend910c2,
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
	newNetRecoverLabel, newNetUHDevSets := hnm.getNewNetworkRecoverDev(devStatusSet,
		common.ConvertDevListToSets(oldDevInfo[common.HuaweiNetworkUnHealthAscend910], common.CommaSepDev),
		nodeFmtDevNetRecover)

	newDevInfo[common.HuaweiAscend910] = newAscend910
	newDevInfo[common.HuaweiUnHealthAscend910] = common.ToString(devStatusSet.UnHealthyDevice, common.CommaSepDev)
	newDevInfo[common.HuaweiNetworkUnHealthAscend910] = common.ToString(newNetUHDevSets, common.CommaSepDev)
	if common.ParamOption.AutoStowingDevs {
		return nil
	}
	if err := hnm.updateNodeLabel(curNode, newDevRecoverLabel, newNetRecoverLabel); err != nil {
		hwlog.RunLog.Errorf("update node label failed, err: %#v", err)
		return err
	}
	lastTimeNetworkRecoverDevices = newNetUHDevSets
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
func (hnm *HwAscend910Manager) getNewNetworkRecoverDev(diffStatusSet common.DevStatusSet, nnu, nnr sets.String) (
	string, sets.String) {
	// nnu means node annotation network unhealthy devices
	// nnr means device's network is ok and to be restored
	// this time network unhealthy devices
	tud := diffStatusSet.NetUnHealthyDevice
	// if there is no network unhealthy device and autoStowingDevs is true
	if common.ParamOption.AutoStowingDevs {
		return "", tud
	}

	// devices recovered between the last check and this check
	recoveredDevSets := lastTimeNetworkRecoverDevices.Difference(nnr)

	newNetworkRecoverDevSets := sets.String{}
	newNetworkRecoverDevSets = newNetworkRecoverDevSets.Union(nnu.Difference(tud))
	// remove the device that network is unhealthy in this times
	newNetworkRecoverDevSets = newNetworkRecoverDevSets.Difference(nnr.Intersection(tud))
	// remove the device that recovered
	newNetworkRecoverDevSets = newNetworkRecoverDevSets.Difference(recoveredDevSets)

	newNetworkUnhealthyDevSets := nnu.Union(tud).Difference(recoveredDevSets)

	return hnm.getPatchLabel(newNetworkRecoverDevSets), newNetworkUnhealthyDevSets
}

// getPatchLabel get elements one by one from the sets and change the element "Ascend910-x" to "x"
// which will patch to node
func (hnm *HwAscend910Manager) getPatchLabel(chips sets.String) string {
	if chips.Len() == 0 {
		return ""
	}

	var ascendLabel []string
	for devName := range chips {
		devTypeAndID := strings.Split(devName, common.GangSepDev)
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
