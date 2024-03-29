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

// Package device a series of device function.
package device

import (
	"fmt"

	"huawei.com/npu-exporter/v5/common-utils/hwlog"

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
	for logicIDIdx := 0; logicIDIdx < len(devList); logicIDIdx++ {
		davinCiDev, err := hnm.getDavinCiDev(devList[logicIDIdx])
		if err != nil {
			return common.NpuAllInfo{}, err
		}
		normalDevices := hnm.getNPUsByNormalMode(davinCiDev)
		if common.ShareDev() {
			normalDevices = hnm.getNPUsByShareMode(davinCiDev)
		}
		allDevices = append(allDevices, normalDevices...)
	}
	return common.NpuAllInfo{AllDevs: allDevices, AllDevTypes: []string{hnm.name}}, err
}

func (hnm *HwAscend310Manager) getNPUsByNormalMode(davinCiDev common.DavinCiDev) []common.NpuDevice {
	deviceName := fmt.Sprintf("%s-%d", hnm.name, davinCiDev.PhyID)
	return []common.NpuDevice{hnm.assembleNpuDeviceStruct(hnm.name, deviceName, davinCiDev)}
}

// DoWithVolcanoListAndWatch ascend310 watch device
func (hnm *HwAscend310Manager) DoWithVolcanoListAndWatch(classifyDevs map[string][]*common.NpuDevice) {
	devStatusSet := hnm.getDevStatesDevSet(classifyDevs)
	if err := hnm.UpdateNodeDeviceInfo(devStatusSet, hnm.updateDeviceInfo); err != nil {
		hwlog.RunLog.Errorf("update device info failed, err: %v", err)
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
	var data []byte
	if data = common.MarshalData(devStatusSet.DeviceFault); len(data) == 0 {
		return fmt.Errorf("device fault code marshal failed")
	}
	newDeviceInfo[common.HuaweiFaultCodeAscend310] = string(data)
	return nil
}

// GraceTolerance graceful fault tolerance, not supported currently
func (hnm *HwAscend310Manager) GraceTolerance(map[string][]*common.NpuDevice) {
	return
}
