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
	"strings"
	"sync"
	"time"

	"huawei.com/npu-exporter/v5/common-utils/hwlog"
	"huawei.com/npu-exporter/v5/common-utils/utils"
	"huawei.com/npu-exporter/v5/devmanager/common"
	"k8s.io/apimachinery/pkg/util/sets"
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
	// ManuallySeparateNPU Manually Separate NPU
	ManuallySeparateNPU = "ManuallySeparateNPU"
	// CardUnhealthy fault is caused by card unhealthy
	CardUnhealthy = "CardUnhealthy"
	// CardNetworkUnhealthy  fault is caused by card network unhealthy
	CardNetworkUnhealthy = "CardNetworkUnhealthy"
	// LinkDownFaultCode linkdown fault code
	LinkDownFaultCode = 0x81078603
	// LinkDownFaultCodeStr linkdown fault code string
	LinkDownFaultCodeStr = "81078603"

	faultCodeFilePath = "/usr/local/faultCode.json"
)

var (
	faultTypeCode FaultTypeCode
	// initLogicIDs need init fault code device. add by train or inference
	initLogicIDs []int32
	// logicIDLock operate initLogicIDs lock
	logicIDLock sync.Mutex
	// UseGetDeviceNetWorkHealthApi for indicating whether to use dcmi_get_device_network_health api
	UseGetDeviceNetWorkHealthApi = true
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
	// manuallySeparateNpuMapLock operate manuallySeparateNpuMap lock
	manuallySeparateNpuMapLock sync.Mutex
	// manuallySeparateNpuMap manually separate npu info cache
	manuallySeparateNpuMap = make(map[int32]ManuallyFaultInfo, GeneralMapSize)
	// FaultTypeSet is a set that contains all the fault level
	FaultTypeSet = sets.NewString(NotHandleFault, RestartRequest, RestartBusiness, FreeRestartNPU,
		RestartNPU, PreSeparateNPU, SeparateNPU, ManuallySeparateNPU)
)

// fault customization
var (
	// WaitFlushingCMTime is the time used in waiting flushing reset info CM
	WaitFlushingCMTime time.Duration = DefaultWaitFlushCMTime
	// WaitDeviceResetTime is the time used in waiting device reset
	WaitDeviceResetTime time.Duration = DefaultWaitDeviceResetTime
	// faultFrequencyMap is the cache saving the occur frequency of a fault, key is event id
	faultFrequencyMap = make(map[string]*FaultFrequencyCache, common.MaxErrorCodeCount)
	// faultFrequencyMapLock is the lock of faultFrequencyMap
	faultFrequencyMapLock sync.Mutex
	// LinkDownTimeoutCustomization is the customized timeout for link down event
	LinkDownTimeoutCustomization = ParamOption.LinkdownTimeout
	// LinkUpTimeoutCustomization is the customized timeout for link up event
	LinkUpTimeoutCustomization = int64(DefaultLinkUpTimeout)

	faultSeverityMap = make(map[int64]int8, common.MaxErrorCodeCount)
)

// ManuallyFaultInfo save the info of ManuallySeparateNPU
type ManuallyFaultInfo struct {
	LogicID     int32
	FirstHandle bool
	RecordTime  int64
}

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

// FaultCustomization is the customization info of fault
type FaultCustomization struct {
	GraceTolerance GraceToleranceCustomization
	FaultFrequency []FaultFrequencyCustomization
	FaultDuration  []FaultDurationCustomization
}

// GraceToleranceCustomization is the customization info of grace tolerance
type GraceToleranceCustomization struct {
	WaitFlushingCMTime  int64
	WaitDeviceResetTime int64
}

// FaultFrequencyCustomization is the customization info of fault frequency
type FaultFrequencyCustomization struct {
	EventId []string
	FaultFrequency
}

// FaultFrequencyCache is the cache saving the FaultFrequency
type FaultFrequencyCache struct {
	// key: logicID, value: fault occurrence time (unix time)
	Frequency map[int32][]int64
	FaultFrequency
}

// FaultFrequency is the base info of fault frequency
type FaultFrequency struct {
	TimeWindow int64
	Times      int64
	FaultLevel string
}

// FaultDurationCustomization is the customization info of fault duration
type FaultDurationCustomization struct {
	EventId []string
	FaultDuration
}

// FaultDuration is the base info of fault duration
type FaultDuration struct {
	FaultTimeout   int64
	RecoverTimeout int64
	FaultLevel     string
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

func LoadFaultCustomization(faultCustomizationStr string) {
	var faultCustomization FaultCustomization
	if err := json.Unmarshal([]byte(faultCustomizationStr), &faultCustomization); err != nil {
		hwlog.RunLog.Errorf("load fault customization failed, unmarshal err: %v", err)
		return
	}
	loadGraceToleranceCustomization(faultCustomization.GraceTolerance)
	loadFaultFrequencyCustomization(faultCustomization.FaultFrequency)
	loadFaultDurationCustomization(faultCustomization.FaultDuration)
}

func ResetFaultCustomization() {
	hwlog.RunLog.Debugf("reset fault customization, will clear cache and use default value")
	WaitFlushingCMTime = DefaultWaitFlushCMTime
	WaitDeviceResetTime = DefaultWaitDeviceResetTime
	LinkDownTimeoutCustomization = ParamOption.LinkdownTimeout
	LinkUpTimeoutCustomization = DefaultLinkUpTimeout
	faultFrequencyMapLock.Lock()
	faultFrequencyMap = make(map[string]*FaultFrequencyCache, GeneralMapSize)
	faultFrequencyMapLock.Unlock()
}

func loadFaultDurationCustomization(customization []FaultDurationCustomization) {
	for _, cus := range customization {
		for _, id := range cus.EventId {
			if id != LinkDownFaultCodeStr {
				hwlog.RunLog.Warnf("FaultDuration only support network fault(%s) now, skip event id %s",
					LinkDownFaultCodeStr, id)
				continue
			}
			if cus.FaultTimeout < MinLinkDownTimeout || cus.FaultTimeout > MaxLinkDownTimeout {
				LinkDownTimeoutCustomization = ParamOption.LinkdownTimeout
				hwlog.RunLog.Errorf("LinkDownTimeout exceed limit(%d-%d), use default(%d)",
					MinLinkDownTimeout, MaxLinkDownTimeout, ParamOption.LinkdownTimeout)
			} else {
				LinkDownTimeoutCustomization = cus.FaultTimeout
				hwlog.RunLog.Infof("modify LinkDownTimeout success: %d", cus.FaultTimeout)
			}
			if cus.RecoverTimeout < MinLinkUpTimeout || cus.RecoverTimeout > MaxLinkUpTimeout {
				LinkUpTimeoutCustomization = DefaultLinkUpTimeout
				hwlog.RunLog.Errorf("LinkUpTimeout exceed limit(%d-%d), use default(%d)",
					MinLinkUpTimeout, MaxLinkUpTimeout, DefaultLinkUpTimeout)
			} else {
				LinkUpTimeoutCustomization = cus.RecoverTimeout
				hwlog.RunLog.Infof("modify LinkUpTimeout success: %d", cus.RecoverTimeout)
			}
			return
		}
	}
	LinkUpTimeoutCustomization = DefaultLinkUpTimeout
	LinkDownTimeoutCustomization = ParamOption.LinkdownTimeout
	hwlog.RunLog.Infof("did not find network fault timeout customization, use default LinkDownTimeout: %d, "+
		"LinkupTimeout: %d", ParamOption.LinkdownTimeout, DefaultLinkUpTimeout)
}

func loadGraceToleranceCustomization(customization GraceToleranceCustomization) {
	if customization.WaitDeviceResetTime < MinWaitDeviceResetTime || customization.WaitDeviceResetTime > MaxWaitDeviceResetTime {
		hwlog.RunLog.Errorf("WaitDeviceResetTime exceed limits(%d~%d), use default(%d)",
			MinWaitDeviceResetTime, MaxWaitDeviceResetTime, DefaultWaitDeviceResetTime)
		WaitDeviceResetTime = DefaultWaitDeviceResetTime
	} else {
		hwlog.RunLog.Infof("modify WaitDeviceResetTime(%d) success", customization.WaitDeviceResetTime)
		WaitDeviceResetTime = time.Duration(customization.WaitDeviceResetTime)
	}
	if customization.WaitFlushingCMTime < MinWaitFlushCMTime || customization.WaitFlushingCMTime > MaxWaitFlushCMTime {
		hwlog.RunLog.Errorf("WaitFlushingCMTime exceed limits(%d~%d), use default(%d)",
			MinWaitFlushCMTime, MaxWaitFlushCMTime, DefaultWaitFlushCMTime)
		WaitFlushingCMTime = DefaultWaitFlushCMTime
	} else {
		hwlog.RunLog.Infof("modify WaitFlushingCMTime(%d) success", customization.WaitFlushingCMTime)
		WaitFlushingCMTime = time.Duration(customization.WaitFlushingCMTime)
	}
}

func loadFaultFrequencyCustomization(customizations []FaultFrequencyCustomization) {
	handledEventId := make(sets.String, GeneralMapSize)
	faultFrequencyMapLock.Lock()
	defer faultFrequencyMapLock.Unlock()
	for _, cus := range customizations {
		if !validateFaultFrequencyCustomization(cus) {
			continue
		}
		for _, id := range cus.EventId {
			id = strings.ToLower(id)
			if handledEventId.Has(id) {
				hwlog.RunLog.Warnf("duplicated event id detected when handling FaultFrequency, skip, id: %s", id)
				continue
			}
			handledEventId.Insert(id)
			if cache, ok := faultFrequencyMap[id]; ok {
				cache.TimeWindow = cus.TimeWindow
				cache.Times = cus.Times
				cache.FaultLevel = cus.FaultLevel
				hwlog.RunLog.Infof("update FaultFrequency for event id %s success, TimeWindow: %d, "+
					"Times: %d, FaultLevel: %s", id, cus.TimeWindow, cus.Times, cus.FaultLevel)
			} else {
				faultFrequencyMap[id] = &FaultFrequencyCache{
					Frequency: make(map[int32][]int64, common.MaxErrorCodeCount),
					FaultFrequency: FaultFrequency{
						TimeWindow: cus.TimeWindow,
						Times:      cus.Times,
						FaultLevel: cus.FaultLevel,
					},
				}
				hwlog.RunLog.Infof("insert FaultFrequency for event id %s success, TimeWindow: %d, "+
					"Times: %d, FaultLevel: %s", id, cus.TimeWindow, cus.Times, cus.FaultLevel)
			}
		}
	}
	// delete event id those in cache but not in CM
	cachedEventIds := make([]string, 0, len(faultFrequencyMap))
	for k := range faultFrequencyMap {
		cachedEventIds = append(cachedEventIds, k)
	}
	for _, cachedId := range cachedEventIds {
		if !handledEventId.Has(cachedId) && len(cachedId) != 0 {
			delete(faultFrequencyMap, cachedId)
			hwlog.RunLog.Infof("delete FaultFrequency for event id %s", cachedId)
		}
	}
}

func insertFaultFrequency(logicId int32, eventId int64) {
	faultFrequencyMapLock.Lock()
	defer faultFrequencyMapLock.Unlock()
	eventIdStr := strings.ToLower(strconv.FormatInt(eventId, Hex))
	frequencyCache, ok := faultFrequencyMap[eventIdStr]
	if !ok {
		hwlog.RunLog.Debugf("skip inserting event id %s to fault frequency cache, no config found", eventIdStr)
		return
	}
	_, ok = frequencyCache.Frequency[logicId]
	if !ok {
		frequencyCache.Frequency[logicId] = make([]int64, 0, frequencyCache.Times)
	}
	frequencyCache.Frequency[logicId] = append(frequencyCache.Frequency[logicId], time.Now().Unix())
	hwlog.RunLog.Infof("insert fault frequency success, event id: %s, logic id: %d, unix time: %d, "+
		"occurrence times :%d", eventIdStr, logicId, time.Now().Unix(), len(frequencyCache.Frequency[logicId]))
}

func validateFaultFrequencyCustomization(customization FaultFrequencyCustomization) bool {
	if len(customization.EventId) == 0 {
		hwlog.RunLog.Warnf("empty event id in this FaultFrequency, skip")
		return false
	}
	if customization.TimeWindow > MaxFaultFrequencyTimeWindow || customization.TimeWindow < MinFaultFrequencyTimeWindow {
		hwlog.RunLog.Warnf("TimeWindow(%d) in this FaultFrequency exceeds limit(%d~%d), skip",
			customization.TimeWindow, MinFaultFrequencyTimeWindow, MaxFaultFrequencyTimeWindow)
		return false
	}
	if customization.Times > MaxFaultFrequencyTimes || customization.Times < MinFaultFrequencyTimes {
		hwlog.RunLog.Warnf("Times(%d) in this FaultFrequency exceeds limit(%d~%d), skip",
			customization.Times, MinFaultFrequencyTimes, MaxFaultFrequencyTimes)
		return false
	}
	if !FaultTypeSet.Has(customization.FaultLevel) {
		hwlog.RunLog.Warnf("FaultLevel(%s) in this FaultFrequency is unrecognized, skip",
			customization.FaultLevel)
		return false
	}
	return true
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

// GetFaultType will return the fault type from fault codes, fault frequency and ManuallySeparateNPU cache
func GetFaultType(faultCodes []int64, logicId int32) string {
	faultTypes := make([]string, 0, len(FaultTypeSet))
	faultTypes = append(faultTypes, GetFaultTypeByCode(faultCodes))
	faultTypes = append(faultTypes, GetFaultTypeFromFaultFrequency(logicId))
	if QueryManuallyFaultInfoByLogicID(logicId) {
		faultTypes = append(faultTypes, ManuallySeparateNPU)
	}
	return getMostSeriousFaultType(faultTypes)
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
		faultType := getFaultTypeBySeverity(faultCodes)
		hwlog.RunLog.Debugf("not record fault code: %v, get fault type by severity: %s", faultCodes, faultType)
		return faultType
	}
}

// GetFaultTypeFromFaultFrequency refreshes the cache of FaultFrequency, delete the faults those not in time window,
// and return the fault level if the occurrence times of fault >= the set value
func GetFaultTypeFromFaultFrequency(logicId int32) string {
	faultTypes := make([]string, 0, len(faultFrequencyMap))
	faultFrequencyMapLock.Lock()
	defer faultFrequencyMapLock.Unlock()
	for eventId, frequencyCache := range faultFrequencyMap {
		_, ok := frequencyCache.Frequency[logicId]
		if !ok {
			continue
		}
		timeWindowStart := time.Now().Unix() - frequencyCache.TimeWindow
		// delete the occurrence times those less than the start of time window
		index := 0
		for _, occurrenceTime := range frequencyCache.Frequency[logicId] {
			if occurrenceTime < timeWindowStart {
				hwlog.RunLog.Infof("delete the expired fault occurrence, event id: %s, logic id: %d, "+
					"time window start: %d, occurrence time: %d", eventId, logicId, timeWindowStart, occurrenceTime)
				index++
			} else {
				break
			}
		}
		frequencyCache.Frequency[logicId] = frequencyCache.Frequency[logicId][index:]
		if int64(len(frequencyCache.Frequency[logicId])) >= frequencyCache.Times {
			hwlog.RunLog.Infof("FaultFrequency detected, event id: %s, logic id: %d, fault occurred times: %d, "+
				"fault level: %s", eventId, logicId, len(frequencyCache.Frequency[logicId]), frequencyCache.FaultLevel)
			if frequencyCache.FaultLevel == ManuallySeparateNPU {
				hwlog.RunLog.Infof("detect ManuallySeparateNPU, logic id: %d", logicId)
				SaveManuallyFaultInfo(logicId)
			}
			faultTypes = append(faultTypes, frequencyCache.FaultLevel)
			// every time when FaultFrequency detected, clear all the fault occurrence time in cache
			frequencyCache.Frequency[logicId] = make([]int64, 0, frequencyCache.Times)
		}
	}
	return getMostSeriousFaultType(faultTypes)
}

func getFaultTypeBySeverity(faultCodes []int64) string {
	for _, code := range faultCodes {
		severity, ok := faultSeverityMap[code]
		if !ok {
			hwlog.RunLog.Warnf("detect unknown fault code and no match severity: %d", code)
			return SeparateNPU
		}
		if severity > FaultSeverityMinor {
			return SeparateNPU
		}
	}
	return NotHandleFault
}

func getMostSeriousFaultType(fautTypes []string) string {
	faultTypeSet := sets.NewString(fautTypes...)
	if faultTypeSet.Has(ManuallySeparateNPU) {
		return ManuallySeparateNPU
	} else if faultTypeSet.Has(SeparateNPU) {
		return SeparateNPU
	} else if faultTypeSet.Has(PreSeparateNPU) {
		return PreSeparateNPU
	} else if faultTypeSet.Has(RestartNPU) {
		return RestartNPU
	} else if faultTypeSet.Has(FreeRestartNPU) {
		return FreeRestartNPU
	} else if faultTypeSet.Has(RestartBusiness) {
		return RestartBusiness
	} else if faultTypeSet.Has(RestartRequest) {
		return RestartRequest
	} else if faultTypeSet.Has(NotHandleFault) {
		return NotHandleFault
	}
	return NormalNPU
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
		insertFaultFrequency(device.LogicID, faultCode)
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
			insertFaultFrequency(device.LogicID, faultInfo.EventID)
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
	faultSeverityMap[devFaultInfo.EventID] = devFaultInfo.Assertion
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

// SaveManuallyFaultInfo save manually fault info into manuallySeparateNpuMap
func SaveManuallyFaultInfo(logicID int32) {
	if logicID < 0 || logicID > 15 {
		hwlog.RunLog.Warnf("logic id %d is not valid, logic id must be in [0, 15]", logicID)
		return
	}
	manFaultInfo := ManuallyFaultInfo{
		LogicID:     logicID,
		FirstHandle: true,
		RecordTime:  time.Now().UnixMilli(),
	}
	manuallySeparateNpuMapLock.Lock()
	defer manuallySeparateNpuMapLock.Unlock()
	manuallySeparateNpuMap[logicID] = manFaultInfo
	displayTime := time.Unix(0, manFaultInfo.RecordTime*int64(time.Millisecond))
	hwlog.RunLog.Infof("receive manually fault info: %v, manually separate device logic id: %v, record time: %v, "+
		"manually separate device cache: %v",
		manFaultInfo, manFaultInfo.LogicID, displayTime.Format("2006-01-02 15:04:05.000"), manuallySeparateNpuMap)
}

// QueryManuallyFaultInfoByLogicID query manually fault info based on logic id from manuallySeparateNpuMap
func QueryManuallyFaultInfoByLogicID(logicID int32) bool {
	if logicID < 0 || logicID > 15 {
		hwlog.RunLog.Warnf("logic id %d is invalid, logic id must be in [0, 15]", logicID)
		return false
	}

	manuallySeparateNpuMapLock.Lock()
	_, ok := manuallySeparateNpuMap[logicID]
	manuallySeparateNpuMapLock.Unlock()
	return ok
}

// QueryManuallyFaultNPULogicIDsByHandleStatus query manually fault npu logic ids based on handle status from manuallySeparateNpuMap
func QueryManuallyFaultNPULogicIDsByHandleStatus(handleStatus string) []int32 {
	logicIDs := make([]int32, 0, GeneralMapSize)
	if handleStatus != ManuallySeparateNpuFirstHandle && handleStatus != ManuallySeparateNpuHandled &&
		handleStatus != ManuallySeparateNpuAll {
		hwlog.RunLog.Warnf("manually fault npu handle status %v is invalid, it must be in [%v,%v,%v]", handleStatus,
			ManuallySeparateNpuFirstHandle, ManuallySeparateNpuHandled, ManuallySeparateNpuAll)
		return logicIDs
	}

	manuallySeparateNpuMapLock.Lock()
	defer manuallySeparateNpuMapLock.Unlock()

	switch {
	case handleStatus == ManuallySeparateNpuFirstHandle:
		for _, manuallySeparateNpu := range manuallySeparateNpuMap {
			if manuallySeparateNpu.FirstHandle {
				logicIDs = append(logicIDs, manuallySeparateNpu.LogicID)
			}
		}
		break
	case handleStatus == ManuallySeparateNpuHandled:
		for _, manuallySeparateNpu := range manuallySeparateNpuMap {
			if !manuallySeparateNpu.FirstHandle {
				logicIDs = append(logicIDs, manuallySeparateNpu.LogicID)
			}
		}
		break
	default:
		for _, manuallySeparateNpu := range manuallySeparateNpuMap {
			logicIDs = append(logicIDs, manuallySeparateNpu.LogicID)
		}
	}

	return logicIDs
}

// SetManuallyFaultNPUHandled set manually fault NPU handled
func SetManuallyFaultNPUHandled() {
	manuallySeparateNpuMapLock.Lock()
	defer manuallySeparateNpuMapLock.Unlock()

	for logicId, manuallyFaultInfo := range manuallySeparateNpuMap {
		manuallyFaultInfo.FirstHandle = false
		manuallySeparateNpuMap[logicId] = manuallyFaultInfo
	}
}

// DeleteManuallyFaultInfo delete manually fault info from manuallySeparateNpuMap
func DeleteManuallyFaultInfo(logicID int32) {
	if logicID < 0 || logicID > 15 {
		hwlog.RunLog.Warnf("logic id %d not valid, must be in [0, 15]", logicID)
		return
	}

	manuallySeparateNpuMapLock.Lock()
	defer manuallySeparateNpuMapLock.Unlock()

	if deleteManuallySeparateFaultInfo, ok := manuallySeparateNpuMap[logicID]; ok {
		delete(manuallySeparateNpuMap, logicID)
		hwlog.RunLog.Infof("device logic id %v, manually fault info %v has been removed, manually separate device "+
			"cache: %v", logicID, deleteManuallySeparateFaultInfo, manuallySeparateNpuMap)
	} else {
		hwlog.RunLog.Warnf("device logic id %v manually fault info not exist, no need to remove", logicID)
	}
}

// GetLinkdownLinkupFaultEvents get linkdown/linkup events from event subscription interface
func GetLinkdownLinkupFaultEvents(logicID int32, faultInfos []common.DevFaultInfo) {
	for _, faultInfo := range faultInfos {
		if faultInfo.EventID == LinkDownFaultCode {
			if UseGetDeviceNetWorkHealthApi {
				UseGetDeviceNetWorkHealthApi = false
				hwlog.RunLog.Info("linkdown event exists in event subscription interface, " +
					"dcmi_get_device_network_health api will not be used")
			}
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
			AlarmRaisedTime <= LinkDownTimeoutCustomization*SecondMagnification {
			continue
		}
		device.NetworkRealHealth = v1beta1.Unhealthy
		hwlog.RunLog.Debugf("in linkdown timeout checking, %v(linkup) - %v(linkdown) > %v, NPU %v "+
			"network health set %v, fault queue: %v", timeoutFaultInfoMap[device.LogicID][i*2+1],
			timeoutFaultInfoMap[device.LogicID][i*2], LinkDownTimeoutCustomization*SecondMagnification,
			device.LogicID, device.NetworkRealHealth, timeoutFaultInfoMap[device.LogicID])
		timeoutFaultInfoMap[device.LogicID] = timeoutFaultInfoMap[device.LogicID][2*i+1:]
		*exitTag = false
		return
	}

	if i*2+1 == faultQueueLen {
		currentHostTime := time.Now().UnixMilli()
		if currentHostTime-timeoutFaultInfoMap[device.LogicID][i*2].AlarmRaisedTime <=
			LinkDownTimeoutCustomization*SecondMagnification {
			hwlog.RunLog.Debugf("in linkdown timeout checking, %v(current host time) - %v(linkdown) <= %v, NPU %v "+
				"network health set %v, fault queue: %v", currentHostTime, timeoutFaultInfoMap[device.LogicID][i*2],
				LinkDownTimeoutCustomization*SecondMagnification,
				device.LogicID, device.NetworkRealHealth, timeoutFaultInfoMap[device.LogicID])
			timeoutFaultInfoMap[device.LogicID] = timeoutFaultInfoMap[device.LogicID][2*i:]
		} else {
			device.NetworkRealHealth = v1beta1.Unhealthy
			hwlog.RunLog.Debugf("in linkdown timeout checking, %v(current host time) - %v(linkdown) > %v, NPU %v "+
				"network health set %v, fault queue: %v", currentHostTime, timeoutFaultInfoMap[device.LogicID][i*2],
				LinkDownTimeoutCustomization*SecondMagnification,
				device.LogicID, device.NetworkRealHealth, timeoutFaultInfoMap[device.LogicID])
			timeoutFaultInfoMap[device.LogicID] = timeoutFaultInfoMap[device.LogicID][2*i+1:]
		}
		*exitTag = true
	}

	if 2*i == faultQueueLen {
		hwlog.RunLog.Debugf("in linkdown timeout checking, %v(linkup) - %v(linkdown) <= %v, NPU %v "+
			"network health set %v, fault queue: %v", timeoutFaultInfoMap[device.LogicID][i*2-1],
			timeoutFaultInfoMap[device.LogicID][i*2-2], LinkDownTimeoutCustomization*SecondMagnification,
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
			AlarmRaisedTime <= LinkUpTimeoutCustomization*SecondMagnification {
			continue
		}
		device.NetworkRealHealth = v1beta1.Healthy
		hwlog.RunLog.Debugf("in linkup recover checking, %v(linkdown) - %v(linkup) > %v, NPU %v "+
			"network health set %v, fault queue: %v", timeoutFaultInfoMap[device.LogicID][i*2+1],
			timeoutFaultInfoMap[device.LogicID][i*2], LinkUpTimeoutCustomization*SecondMagnification,
			device.LogicID, device.NetworkRealHealth, timeoutFaultInfoMap[device.LogicID])
		timeoutFaultInfoMap[device.LogicID] = timeoutFaultInfoMap[device.LogicID][2*i+1:]
		*exitTag = false
		return
	}

	if i*2+1 == faultQueueLen {
		currentHostTime := time.Now().UnixMilli()
		if currentHostTime-timeoutFaultInfoMap[device.LogicID][i*2].AlarmRaisedTime <=
			LinkUpTimeoutCustomization*SecondMagnification {
			hwlog.RunLog.Debugf("in linkup recover checking, %v(current host time) - %v(linkup) <= %v, NPU %v "+
				"network health set %v, fault queue: %v", currentHostTime, timeoutFaultInfoMap[device.LogicID][i*2],
				LinkUpTimeoutCustomization*SecondMagnification, device.LogicID, device.NetworkRealHealth,
				timeoutFaultInfoMap[device.LogicID])
			timeoutFaultInfoMap[device.LogicID] = timeoutFaultInfoMap[device.LogicID][2*i:]
		} else {
			device.NetworkRealHealth = v1beta1.Healthy
			hwlog.RunLog.Debugf("in linkup recover checking, %v(current host time) - %v(linkup) > %v, NPU %v "+
				"network health set %v, fault queue: %v", currentHostTime, timeoutFaultInfoMap[device.LogicID][i*2],
				LinkUpTimeoutCustomization*SecondMagnification, device.LogicID, device.NetworkRealHealth,
				timeoutFaultInfoMap[device.LogicID])
			timeoutFaultInfoMap[device.LogicID] = timeoutFaultInfoMap[device.LogicID][2*i+1:]
		}
		*exitTag = true
	}

	if 2*i == faultQueueLen {
		hwlog.RunLog.Debugf("in linkup recover checking, %v(linkdown) - %v(linkup) <= %v, NPU %v "+
			"network health set %v, fault queue: %v", timeoutFaultInfoMap[device.LogicID][i*2-1],
			timeoutFaultInfoMap[device.LogicID][i*2-2], LinkUpTimeoutCustomization*SecondMagnification,
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
	if device.NetworkRealHealth == v1beta1.Unhealthy && device.NetworkHealth == v1beta1.Healthy {
		hwlog.RunLog.Debugf("insert network fault into FaultFrequency, logic id: %d", device.LogicID)
		insertFaultFrequency(device.LogicID, LinkDownFaultCode)
	}

	hwlog.RunLog.Debugf("NPU logic id: %v, network health status: %v, fault queue after linkDown timeout "+
		"check and recover: %v", device.LogicID, device.NetworkHealth, timeoutFaultInfoMap[device.LogicID])

	if device.NetworkHealth != device.NetworkRealHealth {
		hwlog.RunLog.Infof("NPU logic id: %v, after handling, network health status change, now network health set %v",
			device.LogicID, device.NetworkRealHealth)
	}

	device.NetworkHealth = device.NetworkRealHealth
}
