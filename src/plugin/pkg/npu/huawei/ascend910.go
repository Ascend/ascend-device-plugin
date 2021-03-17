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
// vnpu is classification by computing power, like Ascend910-4c, Ascend910-8c, Ascend910-16c
// physical npu sets corresponding to the deviTypes, and vnpu is vDeviTypes
// vDeviTypes may is: [Ascend910-4c, Ascend910-4c, Ascend910-8c], also deviTypes may is: [Ascend910, Ascend910]
// one class deviType will generate a socket file, like ascend910-4c.sock or Ascend910.sock, so we deduplicate
func (hnm *HwAscend910Manager) GetNPUs(allDevices *[]npuDevice, allDeviceTypes *[]string, matchingDeviType string) error {
	var ids [hiAIMaxDeviceNum]uint32

	devNum, err := hnm.ascendCommonFunction.dmgr.GetDeviceList(&ids)
	if err != nil {
		return err
	}
	var deviTypes []string
	for i := int32(0); i < devNum; i++ {
		phyID, err := hnm.ascendCommonFunction.dmgr.GetPhyID(ids[i])
		if err != nil {
			return err
		}

		totalVDevInfos, err := hnm.queryVirtualDevice(ids[i])
		if err != nil && !strings.Contains(err.Error(), FunctionNotFound) {
			logger.Error("Query virtual device info failure!", zap.String("err",err.Error()))
			continue
		}
		var devices []npuDevice
		if totalVDevInfos.vDevNum == 0 {
			devices, deviTypes = hnm.assemblePhyDevices(ids[i], phyID)
		}else {
			devices, deviTypes = hnm.assembleVirtualDevices(ids[i], phyID, totalVDevInfos)
		}
		*allDevices = append(*allDevices, devices...)
		*allDeviceTypes = append(*allDeviceTypes, deviTypes...)
	}
	*allDeviceTypes = hnm.removeDuplicate(allDeviceTypes)
	return nil
}

func (hnm *HwAscend910Manager) removeDuplicate(allDeviceTypes *[]string) []string {
	deviceTypesMap := make(map[string]string)
	var rmDupDeviceTypes []string
	for _, deviType := range *allDeviceTypes {
		deviceTypesMap[deviType] = deviType
	}
	for _, deviType := range deviceTypesMap {
		rmDupDeviceTypes = append(rmDupDeviceTypes, deviType)
	}
	return rmDupDeviceTypes
}

func (hnm *HwAscend910Manager) assemblePhyDevices(logicID, phyID uint32) ([]npuDevice, []string) {
	var devices []npuDevice
	var deviTypes [] string
	devID := fmt.Sprintf("%s-%d", hiAIAscend910Prefix, logicID)
	device := hnm.ascendCommonFunction.AssembleNpuDeviceStruct(hiAIAscend910Prefix, devID, phyID)
	devices = append(devices, device)
	deviTypes = append(deviTypes, hiAIAscend910Prefix)
	return devices, deviTypes
}

func (hnm *HwAscend910Manager) assembleVirtualDevices(logicID, phyID uint32, totalVDevInfos TotalVDevInfos) ([]npuDevice, []string) {
	var devices []npuDevice
	var vDeviTypes [] string
	for _, vDevInfo := range totalVDevInfos.vDevInfos {
		vDeviType := fmt.Sprintf("%s-%dc", hiAIAscend910Prefix, vDevInfo.coreNum)
		devID := fmt.Sprintf("%s-%dc-%d-%d", hiAIAscend910Prefix, vDevInfo.coreNum, logicID, vDevInfo.id)
		device := hnm.ascendCommonFunction.AssembleNpuDeviceStruct(vDeviType, devID, phyID)
		devices = append(devices, device)
		vDeviTypes = append(vDeviTypes, vDeviType)
	}
	return devices, vDeviTypes
}

func (hnm *HwAscend910Manager) queryVirtualDevice(logicID uint32) (TotalVDevInfos, error) {
	var totalVDevInfos TotalVDevInfos
	if useVolcanoType {
		return totalVDevInfos, nil
	}
	totalVDevInfos, err := hnm.ascendCommonFunction.dmgr.GetVDevicesInfo(logicID)
	if err != nil {
		return TotalVDevInfos{}, fmt.Errorf("query virtual device info failure: %s", err)
	}
	return totalVDevInfos, nil
}
