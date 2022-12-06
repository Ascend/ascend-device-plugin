// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package device a series of device function.
package device

import (
	"fmt"

	"huawei.com/mindx/common/hwlog"

	"Ascend-device-plugin/pkg/common"
)

// HwAscend310Manager manages huawei Ascend310 devices.
type HwAscend310Manager struct {
	AscendTools
}

// NewHwAscend310Manager used to create ascend 310 manager
func NewHwAscend310Manager() *HwAscend310Manager {
	name := common.Ascend310
	if common.ParamOption.GetFdFlag {
		name = common.AscendfdPrefix
	}
	return &HwAscend310Manager{
		AscendTools: AscendTools{
			name:         name,
			unHealthyKey: common.HuaweiUnHealthAscend310,
			devCount:     common.MaxCardNum * common.MaxDevNumInCard,
		},
	}
}

// GetNPUs Discovers all HUAWEI Ascend310 devices by call devmanager interface
func (hnm *HwAscend310Manager) GetNPUs() (common.NpuAllInfo, error) {
	devNum, devList, err := hnm.dmgr.GetDeviceList()
	if err != nil {
		return common.NpuAllInfo{}, err
	}
	if devNum > hnm.devCount {
		return common.NpuAllInfo{}, fmt.Errorf("invalid device num: %d", devNum)
	}
	var allDevices []common.NpuDevice
	var allDeviceTypes []string
	for i := int32(0); i < devNum; i++ {
		phyID, err := hnm.dmgr.GetPhysicIDFromLogicID(devList[i])
		if err != nil {
			return common.NpuAllInfo{}, err
		}
		deviceName := fmt.Sprintf("%s-%d", hnm.name, phyID)
		device := hnm.assembleNpuDeviceStruct(hnm.name, deviceName, devList[i], phyID)
		allDevices = append(allDevices, device)
	}
	allDeviceTypes = append(allDeviceTypes, hnm.name)
	return common.NpuAllInfo{AllDevs: allDevices, AllDevTypes: allDeviceTypes}, nil
}

// DoWithVolcanoListAndWatch ascend310 watch device
func (hnm *HwAscend310Manager) DoWithVolcanoListAndWatch(classifyDevs map[string][]*common.NpuDevice) {
	devStatusSet := hnm.getDevStatesDevSet(classifyDevs)
	if err := hnm.UpdateNodeDeviceInfo(devStatusSet, hnm.updateDeviceInfo); err != nil {
		hwlog.RunLog.Errorf("update device info failed, err: %#v", err)
	}
}

func (hnm *HwAscend310Manager) updateDeviceInfo(_, newDeviceInfo map[string]string,
	devStatusSet common.DevStatusSet) error {
	if newDeviceInfo == nil {
		return fmt.Errorf("invalid new device info")
	}
	newDeviceInfo[common.HuaweiAscend310] = common.ToString(devStatusSet.FreeHealthyDevice[hnm.name],
		common.CommaSepDev)
	newDeviceInfo[hnm.unHealthyKey] = common.ToString(devStatusSet.UnHealthyDevice, common.CommaSepDev)
	return nil
}
