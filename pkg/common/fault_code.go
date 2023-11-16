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
	"sort"
	"strconv"
	"sync"
	"time"

	"huawei.com/npu-exporter/v5/common-utils/hwlog"
	"huawei.com/npu-exporter/v5/common-utils/utils"
	"huawei.com/npu-exporter/v5/devmanager/common"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

const (
	// NotHandleFault not handle fault
	NotHandleFault = "NotHandleFault"
	// RestartRequest restart request
	RestartRequest = "RestartRequest"
	// RestartBusiness restart business
	RestartBusiness = "RestartBusiness"
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
	// LinkDownFaultCode linkdown fault code
	LinkDownFaultCode = 0x81078603
	// ResetFinishFaultCode reset finish fault code
	ResetFinishFaultCode = 0x8C2FA009

	faultCodeFilePath = "/usr/local/faultCode.json"
)

var (
	faultTypeCode FaultTypeCode
	// initLogicIDs need init fault code device. add by train or inference
	initLogicIDs []int32
	// logicIDLock operate initLogicIDs lock
	logicIDLock sync.Mutex
	// timeoutFaultInfoMap timeout event info cache
	timeoutFaultInfoMap = make(map[int32][]common.DevFaultInfo, GeneralMapSize)
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
	NotHandleFaultCodes        []int64
	RestartRequestCodes        []int64
	RestartBusinessCodes       []int64
	RestartNPUCodes            []int64
	FreeRestartNPUCodes        []int64
	PreSeparateNPUCodes        []int64
	SeparateNPUCodes           []int64
	NotHandleFaultNetworkCodes []string
	PreSeparateNPUNetworkCodes []string
	SeparateNPUNetworkCodes    []string
}

// faultFileInfo fault code file data
type faultFileInfo struct {
	NotHandleFaultCodes        []string
	RestartRequestCodes        []string
	RestartBusinessCodes       []string
	RestartNPUCodes            []string
	FreeRestartNPUCodes        []string
	SeparateNPUCodes           []string
	PreSeparateNPUCodes        []string
	NotHandleFaultNetworkCodes []string
	PreSeparateNPUNetworkCodes []string
	SeparateNPUNetworkCodes    []string
}

// DevFaultInfoBasedTimeAscend sort fault queue based on alarmRaisedTime in ascending order
type DevFaultInfoBasedTimeAscend []common.DevFaultInfo

func (devFault DevFaultInfoBasedTimeAscend) Len() int {
	return len(devFault)
}

func (devFault DevFaultInfoBasedTimeAscend) Swap(i, j int) {
	devFault[i], devFault[j] = devFault[j], devFault[i]
}

func (devFault DevFaultInfoBasedTimeAscend) Less(i, j int) bool {
	return devFault[i].AlarmRaisedTime < devFault[j].AlarmRaisedTime
}

// LoadFaultCodeFromFile load fault code and fault type from faultCode.json
func LoadFaultCodeFromFile() error {
	faultCodeBytes, err := utils.LoadFile(faultCodeFilePath)
	if err != nil {
		return fmt.Errorf("load fault code json failed: %v", err)
	}
	return LoadFaultCode(faultCodeBytes)
}

func LoadFaultCode(faultCodeBytes []byte) error {
	var fileInfo faultFileInfo
	if err := json.Unmarshal(faultCodeBytes, &fileInfo); err != nil {
		return fmt.Errorf("unmarshal fault code byte failed: %v", err)
	}
	faultTypeCode = FaultTypeCode{
		NotHandleFaultCodes:        StringTool.HexStringToInt(fileInfo.NotHandleFaultCodes),
		RestartRequestCodes:        StringTool.HexStringToInt(fileInfo.RestartRequestCodes),
		RestartBusinessCodes:       StringTool.HexStringToInt(fileInfo.RestartBusinessCodes),
		RestartNPUCodes:            StringTool.HexStringToInt(fileInfo.RestartNPUCodes),
		FreeRestartNPUCodes:        StringTool.HexStringToInt(fileInfo.FreeRestartNPUCodes),
		PreSeparateNPUCodes:        StringTool.HexStringToInt(fileInfo.PreSeparateNPUCodes),
		SeparateNPUCodes:           StringTool.HexStringToInt(fileInfo.SeparateNPUCodes),
		NotHandleFaultNetworkCodes: fileInfo.NotHandleFaultNetworkCodes,
		PreSeparateNPUNetworkCodes: fileInfo.PreSeparateNPUNetworkCodes,
		SeparateNPUNetworkCodes:    fileInfo.SeparateNPUNetworkCodes,
	}
	return nil
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

// GetFaultTypeByCode get fault type by fault code. if code not record, default SeparateNPU0
func GetFaultTypeByCode(faultCodes []int64) string {
	if len(faultCodes) == 0 {
		return NormalNPU
	}
	switch {
	case Int64Tool.SameElement(faultTypeCode.SeparateNPUCodes, faultCodes):
		return SeparateNPU
	case Int64Tool.SameElement(faultTypeCode.PreSeparateNPUCodes, faultCodes):
		return PreSeparateNPU
	case Int64Tool.SameElement(faultTypeCode.RestartNPUCodes, faultCodes):
		return RestartNPU
	case Int64Tool.SameElement(faultTypeCode.FreeRestartNPUCodes, faultCodes):
		return FreeRestartNPU
	case Int64Tool.SameElement(faultTypeCode.RestartBusinessCodes, faultCodes):
		return RestartBusiness
	case Int64Tool.SameElement(faultTypeCode.RestartRequestCodes, faultCodes):
		return RestartRequest
	case Int64Tool.SameElement(faultTypeCode.NotHandleFaultCodes, faultCodes):
		return NotHandleFault
	default:
		hwlog.RunLog.Debugf("not record fault code : %d, use default type SeparateNPU", faultCodes)
		return SeparateNPU
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
	if device == nil {
		hwlog.RunLog.Error("param device is nil")
		return
	}
	newFaultCodes := make([]int64, 0, common.MaxErrorCodeCount)
	for _, faultCode := range faultCodes {
		if faultCode == LinkDownFaultCode {
			continue
		}
		newFaultCodes = append(newFaultCodes, faultCode)
	}
	device.FaultCodes = newFaultCodes
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
	if device == nil {
		hwlog.RunLog.Error("param device is nil")
		return
	}
	// it must deal with two 'for', because the fault may recover one moment, in this case,
	// the recover message and occur message both in faultInfos, this fault cannot be reports outside.
	for _, faultInfo := range faultInfos {
		if faultInfo.EventID == LinkDownFaultCode {
			continue
		}
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
		if faultInfo.EventID == LinkDownFaultCode {
			continue
		}
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
	hwlog.RunLog.Infof("receive devFaultInfo: %v, hex code: %v", devFaultInfo,
		strconv.FormatInt(devFaultInfo.EventID, Hex))
	if devFaultInfo.EventID == 0 {
		return
	}
    if devFaultInfo.EventID == ResetFinishFaultCode {
        SetDeviceInit(devFaultInfo.LogicID)
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

// GetLinkdownLinkupFaultEvents get linkdown/linkup events from event subscription interface
func GetLinkdownLinkupFaultEvents(logicID int32, faultInfos []common.DevFaultInfo) {
	for _, faultInfo := range faultInfos {
		if faultInfo.EventID == LinkDownFaultCode {
			timeoutFaultInfoMap[logicID] = append(timeoutFaultInfoMap[logicID], faultInfo)
		}
	}
}

// GetCurrentDeviceNetWorkHealth Query the NPU network status at the current time
func GetCurrentDeviceNetWorkHealth(logicID int32, deviceNetWorkHealth string) {
	// If the NPU network is healthy, the network status is regarded as linkup
	// If the NPU network is unhealthy, the network status is regarded as linkdown
	var assertion int8
	if deviceNetWorkHealth == v1beta1.Unhealthy {
		assertion = common.FaultOccur
	} else {
		assertion = common.FaultRecover
	}

	devFaultInfo := common.DevFaultInfo{
		EventID:         LinkDownFaultCode,
		LogicID:         logicID,
		Assertion:       assertion,
		AlarmRaisedTime: time.Now().UnixMilli(),
	}
	timeoutFaultInfoMap[logicID] = append(timeoutFaultInfoMap[logicID], devFaultInfo)
}

// mergeContinuousElementBasedAssertion merge continuous element based on assertion
func mergeContinuousElementBasedAssertion(devFaultInfo *[]common.DevFaultInfo) {
	for i := 1; i < len(*devFaultInfo); i++ {
		currentEvent := (*devFaultInfo)[i]
		previousEvent := (*devFaultInfo)[i-1]

		if currentEvent.Assertion == previousEvent.Assertion {
			*devFaultInfo = append((*devFaultInfo)[:i], (*devFaultInfo)[i+1:]...)
			i--
		}
	}
}

// SortMergeFaultQueue sort fault queue based on alarmRaisedTime and merge continuous element based on assertion
func SortMergeFaultQueue(device *NpuDevice) {
	if device == nil {
		hwlog.RunLog.Error("param device is nil")
		return
	}
	faultInfos := timeoutFaultInfoMap[device.LogicID]

	sort.Sort(DevFaultInfoBasedTimeAscend(faultInfos))
	mergeContinuousElementBasedAssertion(&faultInfos)
	timeoutFaultInfoMap[device.LogicID] = faultInfos

	// If the first element is linkup in fault queue when the NPU network is healthy, clear the first element
	if device.NetworkRealHealth == v1beta1.Healthy && len(timeoutFaultInfoMap[device.LogicID]) > 0 &&
		timeoutFaultInfoMap[device.LogicID][0].Assertion == common.FaultRecover {
		timeoutFaultInfoMap[device.LogicID] = timeoutFaultInfoMap[device.LogicID][1:]
	}

	// If the first element is linkdown in fault queue when the NPU network is unhealthy, clear the first element
	if device.NetworkRealHealth == v1beta1.Unhealthy && len(timeoutFaultInfoMap[device.LogicID]) > 0 &&
		timeoutFaultInfoMap[device.LogicID][0].Assertion == common.FaultOccur {
		timeoutFaultInfoMap[device.LogicID] = timeoutFaultInfoMap[device.LogicID][1:]
	}

	hwlog.RunLog.Debugf("NPU logic id: %v, network health status: %v, fault queue after sort and merge: %v",
		device.LogicID, device.NetworkHealth, timeoutFaultInfoMap[device.LogicID])
}

func checkLinkdownTimeoutWhenNetworkHealth(device *NpuDevice, exitTag *bool) {
	faultQueueLen := len(timeoutFaultInfoMap[device.LogicID])
	if faultQueueLen == 0 {
		hwlog.RunLog.Debugf("NPU logic id: %v, fault queue is empty, "+
			"no need to check whether NPU linkdown timeout when NPU network is healthy", device.LogicID)
		*exitTag = true
		return
	}

	var i int
	for i = 0; i < faultQueueLen/2; i++ {
		if timeoutFaultInfoMap[device.LogicID][i*2+1].AlarmRaisedTime-timeoutFaultInfoMap[device.LogicID][i*2].
			AlarmRaisedTime <= ParamOption.LinkdownTimeout*SecondMagnification {
			continue
		}
		device.NetworkRealHealth = v1beta1.Unhealthy
		hwlog.RunLog.Debugf("in linkdown timeout checking, %v(linkup) - %v(linkdown) > %v, NPU %v "+
			"network health set %v, fault queue: %v", timeoutFaultInfoMap[device.LogicID][i*2+1],
			timeoutFaultInfoMap[device.LogicID][i*2], ParamOption.LinkdownTimeout*SecondMagnification,
			device.LogicID, device.NetworkRealHealth, timeoutFaultInfoMap[device.LogicID])
		timeoutFaultInfoMap[device.LogicID] = timeoutFaultInfoMap[device.LogicID][2*i+1:]
		*exitTag = false
		return
	}

	if i*2+1 == faultQueueLen {
		currentHostTime := time.Now().UnixMilli()
		if currentHostTime-timeoutFaultInfoMap[device.LogicID][i*2].AlarmRaisedTime <=
			ParamOption.LinkdownTimeout*SecondMagnification {
			hwlog.RunLog.Debugf("in linkdown timeout checking, %v(current host time) - %v(linkdown) <= %v, NPU %v "+
				"network health set %v, fault queue: %v", currentHostTime, timeoutFaultInfoMap[device.LogicID][i*2],
				ParamOption.LinkdownTimeout*SecondMagnification,
				device.LogicID, device.NetworkRealHealth, timeoutFaultInfoMap[device.LogicID])
			timeoutFaultInfoMap[device.LogicID] = timeoutFaultInfoMap[device.LogicID][2*i:]
		} else {
			device.NetworkRealHealth = v1beta1.Unhealthy
			hwlog.RunLog.Debugf("in linkdown timeout checking, %v(current host time) - %v(linkdown) > %v, NPU %v "+
				"network health set %v, fault queue: %v", currentHostTime, timeoutFaultInfoMap[device.LogicID][i*2],
				ParamOption.LinkdownTimeout*SecondMagnification,
				device.LogicID, device.NetworkRealHealth, timeoutFaultInfoMap[device.LogicID])
			timeoutFaultInfoMap[device.LogicID] = timeoutFaultInfoMap[device.LogicID][2*i+1:]
		}
		*exitTag = true
	}

	if 2*i == faultQueueLen {
		hwlog.RunLog.Debugf("in linkdown timeout checking, %v(linkup) - %v(linkdown) <= %v, NPU %v "+
			"network health set %v, fault queue: %v", timeoutFaultInfoMap[device.LogicID][i*2-1],
			timeoutFaultInfoMap[device.LogicID][i*2-2], ParamOption.LinkdownTimeout*SecondMagnification,
			device.LogicID, device.NetworkRealHealth, timeoutFaultInfoMap[device.LogicID])
		timeoutFaultInfoMap[device.LogicID] = timeoutFaultInfoMap[device.LogicID][2*i:]
		*exitTag = true
	}
}

func checkLinkupRecoverWhenNetworkUnhealth(device *NpuDevice, exitTag *bool) {
	faultQueueLen := len(timeoutFaultInfoMap[device.LogicID])
	if faultQueueLen == 0 {
		hwlog.RunLog.Debugf("NPU logic id: %v, fault queue is empty, "+
			"no need to check whether NPU linkup recover when NPU network is unhealthy", device.LogicID)
		*exitTag = true
		return
	}

	var i int
	for i = 0; i < faultQueueLen/2; i++ {
		if timeoutFaultInfoMap[device.LogicID][i*2+1].AlarmRaisedTime-timeoutFaultInfoMap[device.LogicID][i*2].
			AlarmRaisedTime <= int64(LinkupRecoverTime*SecondMagnification) {
			continue
		}
		device.NetworkRealHealth = v1beta1.Healthy
		hwlog.RunLog.Debugf("in linkup recover checking, %v(linkdown) - %v(linkup) > %v, NPU %v "+
			"network health set %v, fault queue: %v", timeoutFaultInfoMap[device.LogicID][i*2+1],
			timeoutFaultInfoMap[device.LogicID][i*2], LinkupRecoverTime*SecondMagnification,
			device.LogicID, device.NetworkRealHealth, timeoutFaultInfoMap[device.LogicID])
		timeoutFaultInfoMap[device.LogicID] = timeoutFaultInfoMap[device.LogicID][2*i+1:]
		*exitTag = false
		return
	}

	if i*2+1 == faultQueueLen {
		currentHostTime := time.Now().UnixMilli()
		if currentHostTime-timeoutFaultInfoMap[device.LogicID][i*2].AlarmRaisedTime <=
			int64(LinkupRecoverTime*SecondMagnification) {
			hwlog.RunLog.Debugf("in linkup recover checking, %v(current host time) - %v(linkup) <= %v, NPU %v "+
				"network health set %v, fault queue: %v", currentHostTime, timeoutFaultInfoMap[device.LogicID][i*2],
				LinkupRecoverTime*SecondMagnification, device.LogicID, device.NetworkRealHealth,
				timeoutFaultInfoMap[device.LogicID])
			timeoutFaultInfoMap[device.LogicID] = timeoutFaultInfoMap[device.LogicID][2*i:]
		} else {
			device.NetworkRealHealth = v1beta1.Healthy
			hwlog.RunLog.Debugf("in linkup recover checking, %v(current host time) - %v(linkup) > %v, NPU %v "+
				"network health set %v, fault queue: %v", currentHostTime, timeoutFaultInfoMap[device.LogicID][i*2],
				LinkupRecoverTime*SecondMagnification, device.LogicID, device.NetworkRealHealth,
				timeoutFaultInfoMap[device.LogicID])
			timeoutFaultInfoMap[device.LogicID] = timeoutFaultInfoMap[device.LogicID][2*i+1:]
		}
		*exitTag = true
	}

	if 2*i == faultQueueLen {
		hwlog.RunLog.Debugf("in linkup recover checking, %v(linkdown) - %v(linkup) <= %v, NPU %v "+
			"network health set %v, fault queue: %v", timeoutFaultInfoMap[device.LogicID][i*2-1],
			timeoutFaultInfoMap[device.LogicID][i*2-2], LinkupRecoverTime*SecondMagnification,
			device.LogicID, device.NetworkRealHealth, timeoutFaultInfoMap[device.LogicID])
		timeoutFaultInfoMap[device.LogicID] = timeoutFaultInfoMap[device.LogicID][2*i:]
		*exitTag = true
	}
}

// LinkDownTimeoutCheck check whether the NPU linkdown timeout happened and NPU network recovered
func LinkDownTimeoutCheck(device *NpuDevice) {
	if device == nil {
		hwlog.RunLog.Error("param device is nil")
		return
	}
	// check whether the NPU linkdown timeout happened based on the fault queue
	// check whether the NPU network needs to be restored based on the fault queue
	timeoutFaultInfoMapLen := len(timeoutFaultInfoMap[device.LogicID])
	if device.NetworkRealHealth == v1beta1.Unhealthy {
		hwlog.RunLog.Infof("NPU logic id: %v, network health status is %v", device.LogicID, device.NetworkRealHealth)
	}

	if timeoutFaultInfoMapLen == 0 && device.NetworkHealth == device.NetworkRealHealth {
		hwlog.RunLog.Debugf("NPU logic id: %v, fault queue is empty and NPU network health status not change, "+
			"no need to check whether NPU linkdown timeout, or whether need to recover NPU network health", device.LogicID)
		return
	}

	exitTag := false

	for !exitTag {
		if device.NetworkRealHealth == v1beta1.Healthy {
			checkLinkdownTimeoutWhenNetworkHealth(device, &exitTag)
		} else {
			checkLinkupRecoverWhenNetworkUnhealth(device, &exitTag)
		}
	}

	hwlog.RunLog.Debugf("NPU logic id: %v, network health status: %v, fault queue after linkDown timeout "+
		"check and recover: %v", device.LogicID, device.NetworkHealth, timeoutFaultInfoMap[device.LogicID])

	if device.NetworkHealth != device.NetworkRealHealth {
		hwlog.RunLog.Infof("NPU logic id: %v, after handling, network health status change, now network health set %v",
			device.LogicID, device.NetworkRealHealth)
	}

	device.NetworkHealth = device.NetworkRealHealth
}
