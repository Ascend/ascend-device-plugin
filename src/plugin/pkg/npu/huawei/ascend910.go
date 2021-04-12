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

// Package huawei implements the query and allocation of the device and the function of the log.
package huawei

import (
	"fmt"
	"go.uber.org/zap"
	"strings"
)

const (
	// ZERO_CORE is "0c"
	ZERO_CORE = "0"
)

// switch error log
var logFlag = true

// HwAscend910Manager manages huawei Ascend910 devices.
type HwAscend910Manager struct {
	ascendCommonFunction
}

// NewHwAscend910Manager is used to create ascend 910 manager
func NewHwAscend910Manager() *HwAscend910Manager {
	return &HwAscend910Manager{}
}

// GetNPUs function discovers all HUAWEI Ascend910 devices available
// on the local node by calling walking `/dev` directory.
// a physical npu can be split into multiple vnpu
// vnpu is classification by computing power, like Ascend-910-4c,Ascend-910-8c,Ascend-910-16c
// physical npu sets corresponding to the deviTypes ,and npu id vDeviTypes
// vDeviTypes may is: [Ascend-910-4c,Ascend-910-8c,Ascend-910-16c], also deviTypes: [Ascend-910, Ascend-910]
// one class deviType will generate a socker file, like ascend-910-4c.sock
func (hnm *HwAscend910Manager) GetNPUs(allDevices *[]npuDevice, allDeviceTypes *[]string, deviType string) error {
	var ids [hiAIMaxDeviceNum]uint32

	devNum, errInfo := hnm.dmgr.GetDeviceList(&ids)
	if errInfo != nil {
		return errInfo
	}
	var deviTypes []string
	for i := int32(0); i < devNum; i++ {
		phyID, err := hnm.dmgr.GetPhyID(ids[i])
		if err != nil {
			return err
		}

		cgoDsmiVDevInfos, err := hnm.getVirtualDevice(ids[i])
		if err != nil && !strings.Contains(err.Error(), FunctionNotFound) {
			logger.Error("Query virtual device info failure!", zap.String("err", err.Error()))
			continue
		}
		var devices []npuDevice
		if cgoDsmiVDevInfos.vDevNum == 0 {
			devices, deviTypes = hnm.assemblePhyDevices(phyID)
		} else {
			devices, deviTypes = hnm.assembleVirtualDevices(phyID, cgoDsmiVDevInfos)
		}
		*allDevices = append(*allDevices, devices...)
		*allDeviceTypes = append(*allDeviceTypes, deviTypes...)
	}
	*allDeviceTypes = hnm.removeDuplicate(allDeviceTypes)
	return nil
}

func (hnm *HwAscend910Manager) removeDuplicate(allDeviceTypes *[]string) []string {
	deviceTypesMap := make(map[string]string, 10)
	var rmDupDeviceTypes []string
	for _, deviType := range *allDeviceTypes {
		deviceTypesMap[deviType] = deviType
	}
	for _, deviType := range deviceTypesMap {
		rmDupDeviceTypes = append(rmDupDeviceTypes, deviType)
	}
	return rmDupDeviceTypes
}

func (hnm *HwAscend910Manager) assemblePhyDevices(phyID uint32) ([]npuDevice, []string) {
	var devices []npuDevice
	var deviTypes []string
	devID := fmt.Sprintf("%s-%d", hiAIAscend910Prefix, phyID)
	device := hnm.AssembleNpuDeviceStruct(hiAIAscend910Prefix, devID)
	devices = append(devices, device)
	deviTypes = append(deviTypes, hiAIAscend910Prefix)
	return devices, deviTypes
}

func (hnm *HwAscend910Manager) assembleVirtualDevices(phyID uint32, vInfos CgoDsmiVDevInfo) (
	[]npuDevice, []string) {
	var devices []npuDevice
	var vDeviTypes []string
	for _, dsmiSubVdevInfo := range vInfos.cgoDsmiSubVDevInfos {
		if dsmiSubVdevInfo.spec.coreNum == ZERO_CORE {
			continue
		}
		vDevType := fmt.Sprintf("%s-%sc", hiAIAscend910Prefix, dsmiSubVdevInfo.spec.coreNum)
		devID := fmt.Sprintf("%s-%sc-%d-%d", hiAIAscend910Prefix, dsmiSubVdevInfo.spec.coreNum,
			dsmiSubVdevInfo.vdevid, phyID)
		device := hnm.AssembleNpuDeviceStruct(vDevType, devID)
		devices = append(devices, device)
		vDeviTypes = append(vDeviTypes, vDevType)
	}
	return devices, vDeviTypes
}

func (hnm *HwAscend910Manager) getVirtualDevice(logicID uint32) (CgoDsmiVDevInfo, error) {
	var cgoDsmiVDevInfos CgoDsmiVDevInfo
	if useVolcanoType {
		return cgoDsmiVDevInfos, nil
	}
	cgoDsmiVDevInfos, err := hnm.dmgr.GetVDevicesInfo(logicID)
	if err != nil {
		return CgoDsmiVDevInfo{}, fmt.Errorf("query virtual device info failure: %s\n", err)
	}
	return cgoDsmiVDevInfos, nil
}
