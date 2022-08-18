/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2020-2021. All rights reserved.
 */

// Package device implements the query and allocation of the device and the function of the log.
package device

import (
	"fmt"

	"huawei.com/npu-exporter/hwlog"

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
