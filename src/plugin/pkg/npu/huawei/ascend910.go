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
	"huawei.com/npu-exporter/hwlog"
	"strings"
)

const (
	// ZeroCore is "0c"
	zeroCore = "0"
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

// GetNPUs Discovers all HUAWEI Ascend910 devices by call dsmi interface
// a physical npu can be split into multiple vnpu
// vnpu is classification by computing power, like Ascend910-4c, Ascend910-8c, Ascend910-16c
// physical npu sets corresponding to the deviTypes, and vnpu is vDeviTypes
// vDeviTypes may is: [Ascend910-4c, Ascend910-4c, Ascend910-8c], also deviTypes may is: [Ascend910, Ascend910]
// one class deviType will generate a socket file, like ascend910-4c.sock or Ascend910.sock, so we deduplicate
func (hnm *HwAscend910Manager) GetNPUs(allDevices *[]npuDevice, allDeviceTypes *[]string, matchingDeviType string) error {
	var ids [hiAIMaxDeviceNum]uint32

	devNum, err := hnm.dmgr.GetDeviceList(&ids)
	if err != nil {
		return err
	}
	phyDevMapVirtualDev := make(map[uint32]string, devNum)
	var deviTypes, vDevID []string
	for i := int32(0); i < devNum; i++ {
		phyID, err := hnm.dmgr.GetPhyID(ids[i])
		if err != nil {
			return err
		}
		cgoDsmiVDevInfos, err := hnm.getVirtualDevice(ids[i])
		if err != nil && !strings.Contains(err.Error(), FunctionNotFound) {
			hwlog.Errorf("Query virtual device info failure!, err: %s", err.Error())
			continue
		}
		var devices []npuDevice
		if cgoDsmiVDevInfos.vDevNum == 0 {
			devices, deviTypes = hnm.assemblePhyDevices(phyID)
			phyDevMapVirtualDev[phyID] = fmt.Sprintf("%d", phyID)
		} else {
			devices, deviTypes, vDevID = hnm.assembleVirtualDevices(phyID, cgoDsmiVDevInfos)
			phyDevMapVirtualDev[phyID] = strings.Join(vDevID, ",")
		}
		*allDevices = append(*allDevices, devices...)
		*allDeviceTypes = append(*allDeviceTypes, deviTypes...)
	}
	hnm.phyDevMapVirtualDev = phyDevMapVirtualDev
	*allDeviceTypes = hnm.removeDuplicate(allDeviceTypes)
	return nil
}

func (hnm *HwAscend910Manager) removeDuplicate(allDeviceTypes *[]string) []string {
	deviceTypesMap := make(map[string]string, len(*allDeviceTypes))
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

func (hnm *HwAscend910Manager) assembleVirtualDevices(phyID uint32, cgoDsmiVDevInfos CgoDsmiVDevInfo) (
	[]npuDevice, []string, []string) {
	var devices []npuDevice
	var vDeviTypes []string
	var vDevID []string
	for _, dsmiSubVDevInfo := range cgoDsmiVDevInfos.cgoDsmiSubVDevInfos {
		if dsmiSubVDevInfo.spec.coreNum == zeroCore {
			continue
		}
		vDeviType := fmt.Sprintf("%s-%sc", hiAIAscend910Prefix, dsmiSubVDevInfo.spec.coreNum)
		devID := fmt.Sprintf("%s-%sc-%d-%d", hiAIAscend910Prefix, dsmiSubVDevInfo.spec.coreNum, dsmiSubVDevInfo.vdevid, phyID)
		device := hnm.AssembleNpuDeviceStruct(vDeviType, devID)
		devices = append(devices, device)
		vDeviTypes = append(vDeviTypes, vDeviType)
		vDevID = append(vDevID, fmt.Sprintf("%d", dsmiSubVDevInfo.vdevid))
	}
	return devices, vDeviTypes, vDevID
}

func (hnm *HwAscend910Manager) getVirtualDevice(logicID uint32) (CgoDsmiVDevInfo, error) {
	cgoDsmiVDevInfos, err := hnm.dmgr.GetVDevicesInfo(logicID)
	if err != nil {
		return CgoDsmiVDevInfo{}, fmt.Errorf("query virtual device info failure: %s", err)
	}
	return cgoDsmiVDevInfos, nil
}
