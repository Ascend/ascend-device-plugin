/* Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
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

// Package common a series of common function
package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"huawei.com/npu-exporter/v5/common-utils/hwlog"
	"huawei.com/npu-exporter/v5/common-utils/utils"
	"huawei.com/npu-exporter/v5/devmanager/common"
)

const (
	// NotHandleFault not handle fault
	NotHandleFault = "NotHandleFault"
	// RestartBusiness restart business
	RestartBusiness = "RestartBusiness"
	// RecoverRestartBusiness recover and restart business
	RecoverRestartBusiness = "RecoverRestartBusiness"
	// RestartNPU restart NPU
	RestartNPU = "RestartNPU"
	// FreeRestartNPU wait free and restart NPU
	FreeRestartNPU = "FreeRestartNPU"
	// SeparateNPU separate NPU
	SeparateNPU = "SeparateNPU"
	// NormalNPU normal NPU
	NormalNPU = "NormalNPU"
	// OneDeviceMaxFaultNum one device max fault num
	OneDeviceMaxFaultNum = 128

	faultCodeFilePath = "/usr/local/faultCode.json"
)

var (
	faultTypeCode FaultTypeCode
	// initLogicIDs need init fault code device. add by train or inference
	initLogicIDs []int32
	// logicIDLock operate initLogicIDs lock
	logicIDLock sync.Mutex
	// recoverFaultMap fault event info cache
	recoverFaultMap = make(map[int32][]int64, GeneralMapSize)
	// devFaultInfoMap save the subscribe interface return fault
	devFaultInfoMap = make(map[int32][]common.DevFaultInfo, GeneralMapSize)
	// devFaultInfoMapLock operate devFaultInfoMap lock
	devFaultInfoMapLock sync.Mutex
	// SubscribeFailed subscribe failed flag
	SubscribeFailed bool
)

// FaultTypeCode group code by type
type FaultTypeCode struct {
	NotHandleFaultCodes         []int64
	RestartBusinessCodes        []int64
	RecoverRestartBusinessCodes []int64
	RestartNPUCodes             []int64
	FreeRestartNPUCodes         []int64
	SeparateNPUCodes            []int64
}

// LoadFaultCodeFromFile load fault code and fault type from faultCode.json
func LoadFaultCodeFromFile() error {
	faultCodeBytes, err := utils.LoadFile(faultCodeFilePath)
	if err != nil {
		return fmt.Errorf("load fault code json failed: %v", err)
	}
	err = json.Unmarshal(faultCodeBytes, &faultTypeCode)
	if err != nil {
		return fmt.Errorf("unmarshal fault code byte failed: %v", err)
	}
	if len(faultTypeCode.NotHandleFaultCodes) == 0 && len(faultTypeCode.RestartBusinessCodes) == 0 &&
		len(faultTypeCode.RestartNPUCodes) == 0 && len(faultTypeCode.SeparateNPUCodes) == 0 &&
		len(faultTypeCode.FreeRestartNPUCodes) == 0 && len(faultTypeCode.RecoverRestartBusinessCodes) == 0 {
		return errors.New("at least one fault code in faultTypeCode")
	}
	return nil
}

// GetFaultTypeByCode get fault type by fault code. if code not record, default NotHandleFault
func GetFaultTypeByCode(faultCodes []int64) string {
	if len(faultCodes) == 0 {
		return NormalNPU
	}
	switch {
	case Int64Tool.SameElement(faultTypeCode.SeparateNPUCodes, faultCodes):
		return SeparateNPU
	case Int64Tool.SameElement(faultTypeCode.RestartNPUCodes, faultCodes):
		return RestartNPU
	case Int64Tool.SameElement(faultTypeCode.FreeRestartNPUCodes, faultCodes):
		return FreeRestartNPU
	case Int64Tool.SameElement(faultTypeCode.RestartBusinessCodes, faultCodes):
		return RestartBusiness
	case Int64Tool.SameElement(faultTypeCode.RecoverRestartBusinessCodes, faultCodes):
		return RecoverRestartBusiness
	case Int64Tool.SameElement(faultTypeCode.NotHandleFaultCodes, faultCodes):
		return NotHandleFault
	default:
		hwlog.RunLog.Warnf("not record fault code : #s, use default type NotHandleFault", faultCodes)
		return NotHandleFault
	}
}

// SetDeviceInit set should init device's logicID
func SetDeviceInit(logicID int32) {
	logicIDLock.Lock()
	initLogicIDs = append(initLogicIDs, logicID)
	logicIDLock.Unlock()
}

// GetAndCleanLogicID get should init device's logicID and clean cache
func GetAndCleanLogicID() []int32 {
	if len(initLogicIDs) == 0 {
		return nil
	}
	logicIDLock.Lock()
	oldInitLogicIDs := initLogicIDs
	initLogicIDs = []int32{}
	logicIDLock.Unlock()
	return oldInitLogicIDs
}

// SetFaultCodes set fault codes, all fault code write operate should package into this file for safe
func SetFaultCodes(device *NpuDevice, faultCodes []int64) {
	device.FaultCodes = faultCodes
}

// SetNewFaultAndCacheOnceRecoverFault set new fault code and cache once recover fault
func SetNewFaultAndCacheOnceRecoverFault(logicID int32, faultInfos []common.DevFaultInfo, device *NpuDevice) {
	// it must deal with two 'for', because the fault may recover one moment, in this case,
	// the recover message and occur message both in faultInfos, this fault cannot be reports outside.
	for _, faultInfo := range faultInfos {
		if faultInfo.Assertion == common.FaultRecover {
			device.FaultCodes = Int64Tool.Remove(device.FaultCodes, faultInfo.EventID)
		}
	}
	for _, faultInfo := range faultInfos {
		if faultInfo.Assertion == common.FaultOccur || faultInfo.Assertion == common.FaultOnce {
			device.FaultCodes = append(device.FaultCodes, faultInfo.EventID)
		}
	}
	// once fault or recover in one cycle fault should remove in the end, so we cache the fault first.
	cacheAfterDelFaultCode(logicID, faultInfos)
}

func cacheAfterDelFaultCode(logicID int32, faultInfos []common.DevFaultInfo) {
	for _, faultInfo := range faultInfos {
		if faultInfo.Assertion == common.FaultRecover || faultInfo.Assertion == common.FaultOnce {
			recoverFaultMap[logicID] = append(recoverFaultMap[logicID], faultInfo.EventID)
		}
	}
}

// DelOnceRecoverFault delete func 'cacheAfterDelFaultCode' record fault code in the end of cycle
func DelOnceRecoverFault(groupDevice map[string][]*NpuDevice) {
	for _, devices := range groupDevice {
		for _, device := range devices {
			recoverFaults := recoverFaultMap[device.LogicID]
			for _, recoverFault := range recoverFaults {
				device.FaultCodes = Int64Tool.Remove(device.FaultCodes, recoverFault)
			}
		}
	}
	recoverFaultMap = make(map[int32][]int64, GeneralMapSize)
}

// SaveDevFaultInfo save device fault info , subscribe interface call back function
func SaveDevFaultInfo(devFaultInfo common.DevFaultInfo) {
	hwlog.RunLog.Debugf("receive devFaultInfo: %v", devFaultInfo)
	if devFaultInfo.EventID == 0 {
		return
	}
	devFaultInfoMapLock.Lock()
	devFaultInfoMap[devFaultInfo.LogicID] = append(devFaultInfoMap[devFaultInfo.LogicID], devFaultInfo)
	devFaultInfoMapLock.Unlock()
}

// TakeOutDevFaultInfo take out device fault info
func TakeOutDevFaultInfo(logicID int32) []common.DevFaultInfo {
	if len(devFaultInfoMap[logicID]) == 0 {
		return nil
	}
	devFaultInfoMapLock.Lock()
	devFaultInfo := devFaultInfoMap[logicID]
	devFaultInfoMap[logicID] = []common.DevFaultInfo{}
	devFaultInfoMapLock.Unlock()
	return devFaultInfo
}
