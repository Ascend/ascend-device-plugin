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
	"fmt"
	"strconv"
	"sync"
	"time"

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
	// NormalNetwork normal network
	NormalNetwork = "NormalNetwork"
	// PreSeparateNPU pre separate NPU
	PreSeparateNPU = "PreSeparateNPU"
	// CardUnhealthy fault is caused by card unhealthy
	CardUnhealthy = "CardUnhealthy"
	// CardNetworkUnhealthy  fault is caused by card network unhealthy
	CardNetworkUnhealthy = "CardNetworkUnhealthy"
	// CardNetworkDisconnected card network disconnected
	CardNetworkDisconnected = "Disconnected"

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
	NotHandleFaultCodes           []int64
	RestartBusinessCodes          []int64
	RecoverRestartBusinessCodes   []int64
	RestartNPUCodes               []int64
	FreeRestartNPUCodes           []int64
	SeparateNPUCodes              []int64
	LargeModelNotHandleFaultCodes []int64
	LargeModelPreSeparateNPUCodes []int64
	LargeModelSeparateNPUCodes    []int64
	NotHandleFaultNetworkCodes    []string
	PreSeparateNPUNetworkCodes    []string
	SeparateNPUNetworkCodes       []string
}

// faultFileInfo fault code file data
type faultFileInfo struct {
	NotHandleFaultCodes           []string
	RestartBusinessCodes          []string
	RecoverRestartBusinessCodes   []string
	RestartNPUCodes               []string
	FreeRestartNPUCodes           []string
	SeparateNPUCodes              []string
	LargeModelNotHandleFaultCodes []string
	LargeModelPreSeparateNPUCodes []string
	LargeModelSeparateNPUCodes    []string
	NotHandleFaultNetworkCodes    []string
	PreSeparateNPUNetworkCodes    []string
	SeparateNPUNetworkCodes       []string
}

// LoadFaultCodeFromFile load fault code and fault type from faultCode.json
func LoadFaultCodeFromFile() error {
	faultCodeBytes, err := utils.LoadFile(faultCodeFilePath)
	if err != nil {
		return fmt.Errorf("load fault code json failed: %v", err)
	}
	var fileInfo faultFileInfo
	if err = json.Unmarshal(faultCodeBytes, &fileInfo); err != nil {
		return fmt.Errorf("unmarshal fault code byte failed: %v", err)
	}
	faultTypeCode = FaultTypeCode{
		NotHandleFaultCodes:           StringTool.HexStringToInt(fileInfo.NotHandleFaultCodes),
		RestartBusinessCodes:          StringTool.HexStringToInt(fileInfo.RestartBusinessCodes),
		RecoverRestartBusinessCodes:   StringTool.HexStringToInt(fileInfo.RecoverRestartBusinessCodes),
		RestartNPUCodes:               StringTool.HexStringToInt(fileInfo.RestartNPUCodes),
		FreeRestartNPUCodes:           StringTool.HexStringToInt(fileInfo.FreeRestartNPUCodes),
		SeparateNPUCodes:              StringTool.HexStringToInt(fileInfo.SeparateNPUCodes),
		LargeModelNotHandleFaultCodes: StringTool.HexStringToInt(fileInfo.LargeModelNotHandleFaultCodes),
		LargeModelPreSeparateNPUCodes: StringTool.HexStringToInt(fileInfo.LargeModelPreSeparateNPUCodes),
		LargeModelSeparateNPUCodes:    StringTool.HexStringToInt(fileInfo.LargeModelSeparateNPUCodes),
		NotHandleFaultNetworkCodes:    fileInfo.NotHandleFaultNetworkCodes,
		PreSeparateNPUNetworkCodes:    fileInfo.PreSeparateNPUNetworkCodes,
		SeparateNPUNetworkCodes:       fileInfo.SeparateNPUNetworkCodes,
	}
	return nil
}

// GetLargeModelFaultTypeByCode get large model fault type by fault code. if code not record, default PreSeparateNPU
func GetLargeModelFaultTypeByCode(faultCodes []int64) string {
	if len(faultCodes) == 0 {
		return NormalNPU
	}
	if len(faultTypeCode.NotHandleFaultCodes) == 0 && len(faultTypeCode.LargeModelNotHandleFaultCodes) == 0 {
		if err := LoadFaultCodeFromFile(); err != nil {
			return PreSeparateNPU
		}
	}
	switch {
	case Int64Tool.SameElement(faultTypeCode.LargeModelSeparateNPUCodes, faultCodes):
		return SeparateNPU
	case Int64Tool.SameElement(faultTypeCode.LargeModelPreSeparateNPUCodes, faultCodes):
		return PreSeparateNPU
	case Int64Tool.SameElement(faultTypeCode.LargeModelNotHandleFaultCodes, faultCodes):
		return NotHandleFault
	default:
		hwlog.RunLog.Debugf("not record fault code : %d, use default type PreSeparateNPU", faultCodes)
		return PreSeparateNPU
	}
}

// GetNetworkFaultTypeByCode get network fault type by fault code. if code not record, default PreSeparateNPU
func GetNetworkFaultTypeByCode(faultCodes []string) string {
	if len(faultCodes) == 0 {
		return NormalNetwork
	}
	if len(faultTypeCode.NotHandleFaultCodes) == 0 && len(faultTypeCode.PreSeparateNPUNetworkCodes) == 0 {
		if err := LoadFaultCodeFromFile(); err != nil {
			return PreSeparateNPU
		}
	}
	switch {
	case StringTool.SameElement(faultTypeCode.SeparateNPUNetworkCodes, faultCodes):
		return SeparateNPU
	case StringTool.SameElement(faultTypeCode.PreSeparateNPUNetworkCodes, faultCodes):
		return PreSeparateNPU
	case StringTool.SameElement(faultTypeCode.NotHandleFaultNetworkCodes, faultCodes):
		return NotHandleFault
	default:
		hwlog.RunLog.Debugf("not record fault code : %v, use default type PreSeparateNPU", faultCodes)
		return PreSeparateNPU
	}
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
		hwlog.RunLog.Debugf("not record fault code : %d, use default type NotHandleFault", faultCodes)
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
	setAlarmRaisedTime(device)
}

// setAlarmRaisedTime set `AlarmRaisedTime` by device fault code length
func setAlarmRaisedTime(device *NpuDevice) {
	if len(device.FaultCodes) == 0 {
		device.AlarmRaisedTime = 0
	} else if device.AlarmRaisedTime == 0 {
		device.AlarmRaisedTime = time.Now().UnixMilli()
	}
}

// SetNewFaultAndCacheOnceRecoverFault set new fault code and cache once recover fault
func SetNewFaultAndCacheOnceRecoverFault(logicID int32, faultInfos []common.DevFaultInfo, device *NpuDevice) {
	// it must deal with two 'for', because the fault may recover one moment, in this case,
	// the recover message and occur message both in faultInfos, this fault cannot be reports outside.
	for _, faultInfo := range faultInfos {
		if faultInfo.Assertion == common.FaultRecover {
			if Int64Tool.Index(device.FaultCodes, faultInfo.EventID) == -1 {
				recoverFaultMap[logicID] = append(recoverFaultMap[logicID], faultInfo.EventID)
			} else {
				device.FaultCodes = Int64Tool.Remove(device.FaultCodes, faultInfo.EventID)
			}
		}
		if faultInfo.Assertion == common.FaultOnce {
			recoverFaultMap[logicID] = append(recoverFaultMap[logicID], faultInfo.EventID)
		}
	}
	for _, faultInfo := range faultInfos {
		if faultInfo.Assertion == common.FaultOccur || faultInfo.Assertion == common.FaultOnce {
			device.FaultCodes = append(device.FaultCodes, faultInfo.EventID)
		}
	}
	setAlarmRaisedTime(device)
}

// DelOnceRecoverFault delete func 'cacheAfterDelFaultCode' record fault code in the end of cycle
func DelOnceRecoverFault(groupDevice map[string][]*NpuDevice) {
	for _, devices := range groupDevice {
		for _, device := range devices {
			recoverFaults := recoverFaultMap[device.LogicID]
			for _, recoverFault := range recoverFaults {
				device.FaultCodes = Int64Tool.Remove(device.FaultCodes, recoverFault)
			}
			setAlarmRaisedTime(device)
		}
	}
	recoverFaultMap = make(map[int32][]int64, GeneralMapSize)
}

// SaveDevFaultInfo save device fault info , subscribe interface call back function
func SaveDevFaultInfo(devFaultInfo common.DevFaultInfo) {
	hwlog.RunLog.Debugf("receive devFaultInfo: %v, hex code: %v", devFaultInfo,
		strconv.FormatInt(devFaultInfo.EventID, Hex))
	if devFaultInfo.EventID == 0 {
		return
	}
	devFaultInfoMapLock.Lock()
	devFaultInfoMap[devFaultInfo.LogicID] = append(devFaultInfoMap[devFaultInfo.LogicID], devFaultInfo)
	devFaultInfoMapLock.Unlock()
}

// GetAndCleanFaultInfo get device fault info and clean cache
func GetAndCleanFaultInfo() map[int32][]common.DevFaultInfo {
	if len(devFaultInfoMap) == 0 {
		return map[int32][]common.DevFaultInfo{}
	}
	devFaultInfoMapLock.Lock()
	oldDevFaultInfoMap := devFaultInfoMap
	devFaultInfoMap = make(map[int32][]common.DevFaultInfo, GeneralMapSize)
	devFaultInfoMapLock.Unlock()
	return oldDevFaultInfoMap
}
