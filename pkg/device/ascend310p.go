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

	"huawei.com/mindx/common/hwlog"

	"Ascend-device-plugin/pkg/common"
)

// HwAscend310PManager manages huawei Ascend310P devices.
type HwAscend310PManager struct {
	AscendTools
}

// NewHwAscend310PManager used to create ascend 310P manager
func NewHwAscend310PManager() *HwAscend310PManager {
	return &HwAscend310PManager{
		AscendTools: AscendTools{
			name:         common.Ascend310P,
			unHealthyKey: common.HuaweiUnHealthAscend310P,
			devCount:     common.MaxDevicesNum,
		},
	}
}

// GetNPUs Discovers all HUAWEI Ascend310P devices by call devmanager interface
func (hnm *HwAscend310PManager) GetNPUs(allDevices *[]common.NpuDevice, allDeviceTypes *[]string) error {
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
		if vDevInfos.TotalResource.VDevNum > common.MaxVirtualDeviceNum {
			return fmt.Errorf("invalid virtual device count")
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

// DoWithVolcanoListAndWatch ascend310P affinity scheduling
func (hnm *HwAscend310PManager) DoWithVolcanoListAndWatch(classifyDevs map[string][]*common.NpuDevice) {
	devStatusSet := hnm.getDevStatesDevSet(classifyDevs)
	if err := hnm.UpdateNodeDeviceInfo(devStatusSet, hnm.updateDeviceInfo); err != nil {
		hwlog.RunLog.Errorf("update device info failed, err: %#v", err)
	}
}

func (hnm *HwAscend310PManager) getTemplateName2DeviceTypeMap() map[string]string {
	return map[string]string{
		"vir04":          common.Ascend310Pc4,
		"vir04_3c":       common.Ascend310Pc4Cpu3,
		"vir02":          common.Ascend310Pc2,
		"vir02_1c":       common.Ascend310Pc2Cpu1,
		"vir01":          common.Ascend310Pc1,
		"vir04_4c_dvpp":  common.Ascend310Pc4Cpu4Dvpp,
		"vir04_3c_ndvpp": common.Ascend310Pc4Cpu3Ndvpp,
	}
}

func (hnm *HwAscend310PManager) updateDeviceInfo(_, newDevInfo map[string]string,
	devStatusSet common.DevStatusSet) error {
	if newDevInfo == nil {
		return fmt.Errorf("invalid new device info")
	}
	newDevInfo[common.HuaweiAscend310P] = common.ToString(devStatusSet.FreeHealthyDevice[hnm.name],
		common.CommaSepDev)
	newDevInfo[hnm.unHealthyKey] = common.ToString(devStatusSet.UnHealthyDevice, common.CommaSepDev)
	return nil
}
