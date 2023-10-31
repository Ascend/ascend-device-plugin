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
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/npu-exporter/v5/common-utils/utils"
	"huawei.com/npu-exporter/v5/devmanager/common"
)

// TestLoadFaultCodeFromFile for test LoadFaultCodeFromFile
func TestLoadFaultCodeFromFile(t *testing.T) {
	convey.Convey("test LoadFaultCodeFromFile", t, func() {
		convey.Convey("utils.LoadFile err", func() {
			mockLoadFile := gomonkey.ApplyFuncReturn(utils.LoadFile, nil, errors.New("failed"))
			defer mockLoadFile.Reset()
			convey.So(LoadFaultCodeFromFile(), convey.ShouldNotBeNil)
		})
		convey.Convey("test LoadFaultCodeFromFile", func() {
			mockLoadFile := gomonkey.ApplyFuncReturn(utils.LoadFile, nil, nil)
			defer mockLoadFile.Reset()
			mockUnmarshal := gomonkey.ApplyFuncReturn(json.Unmarshal, errors.New("failed"))
			defer mockUnmarshal.Reset()
			convey.So(LoadFaultCodeFromFile(), convey.ShouldNotBeNil)
		})
	})
}

// TestGetFaultTypeByCode for test GetFaultTypeByCode
func TestGetFaultTypeByCode(t *testing.T) {
	convey.Convey("test GetFaultTypeByCode", t, func() {
		faultCodes := []int64{1}
		convey.Convey("fault type NormalNPU", func() {
			convey.So(GetFaultTypeByCode(nil), convey.ShouldEqual, NormalNPU)
		})
		convey.Convey("fault type NotHandleFault", func() {
			faultTypeCode = FaultTypeCode{NotHandleFaultCodes: faultCodes}
			convey.So(GetFaultTypeByCode(faultCodes), convey.ShouldEqual, NotHandleFault)
		})
		convey.Convey("fault type SeparateNPU", func() {
			faultTypeCode = FaultTypeCode{SeparateNPUCodes: faultCodes}
			convey.So(GetFaultTypeByCode(faultCodes), convey.ShouldEqual, SeparateNPU)
			faultTypeCode = FaultTypeCode{}
			convey.So(GetFaultTypeByCode(faultCodes), convey.ShouldEqual, SeparateNPU)
		})
		convey.Convey("fault type RestartNPU", func() {
			faultTypeCode = FaultTypeCode{RestartNPUCodes: faultCodes}
			convey.So(GetFaultTypeByCode(faultCodes), convey.ShouldEqual, RestartNPU)
		})
		convey.Convey("fault type FreeRestartNPU", func() {
			faultTypeCode = FaultTypeCode{FreeRestartNPUCodes: faultCodes}
			convey.So(GetFaultTypeByCode(faultCodes), convey.ShouldEqual, FreeRestartNPU)
		})
		convey.Convey("fault type RestartBusiness", func() {
			faultTypeCode = FaultTypeCode{RestartBusinessCodes: faultCodes}
			convey.So(GetFaultTypeByCode(faultCodes), convey.ShouldEqual, RestartBusiness)
		})
		convey.Convey("fault type RestartRequestCodes", func() {
			faultTypeCode = FaultTypeCode{RestartRequestCodes: faultCodes}
			convey.So(GetFaultTypeByCode(faultCodes), convey.ShouldEqual, RestartRequest)
		})
	})
}

// TestSetDeviceInit for test SetDeviceInit
func TestSetDeviceInit(t *testing.T) {
	convey.Convey("test SetDeviceInit", t, func() {
		convey.Convey("SetDeviceInit success", func() {
			initLogicIDs = nil
			SetDeviceInit(0)
			convey.So(len(initLogicIDs), convey.ShouldEqual, 1)
		})
	})
}

// TestGetAndCleanLogicID for test GetAndCleanLogicID
func TestGetAndCleanLogicID(t *testing.T) {
	convey.Convey("test GetAndCleanLogicID", t, func() {
		convey.Convey("initLogicIDs is empty", func() {
			initLogicIDs = nil
			convey.So(GetAndCleanLogicID(), convey.ShouldBeNil)
		})
		convey.Convey("initLogicIDs is not empty", func() {
			testIDs := []int32{1}
			initLogicIDs = testIDs
			convey.So(GetAndCleanLogicID(), convey.ShouldResemble, testIDs)
		})
	})
}

// TestSetFaultCodes for test SetFaultCodes
func TestSetFaultCodes(t *testing.T) {
	convey.Convey("test SetFaultCodes", t, func() {
		convey.Convey("SetFaultCodes success", func() {
			device, faultCodes := &NpuDevice{}, []int64{1}
			SetFaultCodes(device, faultCodes)
			convey.So(len(device.FaultCodes), convey.ShouldEqual, len(faultCodes))
		})
	})
}

// TestSetNewFaultAndCacheOnceRecoverFault for test SetNewFaultAndCacheOnceRecoverFault
func TestSetNewFaultAndCacheOnceRecoverFault(t *testing.T) {
	convey.Convey("test SetNewFaultAndCacheOnceRecoverFault", t, func() {
		convey.Convey("SetNewFaultAndCacheOnceRecoverFault success", func() {
			recoverFaultMap = make(map[int32][]int64, GeneralMapSize)
			logicID := int32(0)
			faultInfos := []common.DevFaultInfo{
				{Assertion: common.FaultRecover},
				{Assertion: common.FaultRecover, EventID: 1},
				{Assertion: common.FaultOnce, EventID: 0},
			}
			device := &NpuDevice{FaultCodes: []int64{1}}
			expectedFaultCodes, expectedFaultMapLen := []int64{0}, 2
			SetNewFaultAndCacheOnceRecoverFault(logicID, faultInfos, device)
			convey.So(device.FaultCodes, convey.ShouldResemble, expectedFaultCodes)
			convey.So(len(recoverFaultMap[logicID]), convey.ShouldEqual, expectedFaultMapLen)
		})
	})
}

// TestDelOnceRecoverFault for test DelOnceRecoverFault
func TestDelOnceRecoverFault(t *testing.T) {
	convey.Convey("test DelOnceRecoverFault", t, func() {
		convey.Convey("DelOnceRecoverFault success", func() {
			faultCodes := []int64{1}
			device := &NpuDevice{FaultCodes: faultCodes}
			recoverFaultMap = map[int32][]int64{
				0: faultCodes,
			}
			groupDevice := map[string][]*NpuDevice{
				"test": {device},
			}
			DelOnceRecoverFault(groupDevice)
			convey.So(len(device.FaultCodes), convey.ShouldEqual, 0)
			convey.So(len(recoverFaultMap), convey.ShouldEqual, 0)
		})
	})
}

// TestSaveDevFaultInfo for test SaveDevFaultInfo
func TestSaveDevFaultInfo(t *testing.T) {
	convey.Convey("test SaveDevFaultInfo", t, func() {
		convey.Convey("SaveDevFaultInfo success", func() {
			devFaultInfoMap = make(map[int32][]common.DevFaultInfo, GeneralMapSize)
			SaveDevFaultInfo(common.DevFaultInfo{})
			convey.So(len(devFaultInfoMap), convey.ShouldEqual, 0)
			SaveDevFaultInfo(common.DevFaultInfo{EventID: 1})
			convey.So(len(devFaultInfoMap), convey.ShouldEqual, 1)
		})
	})
}

// TestTakeOutDevFaultInfo for test TakeOutDevFaultInfo
func TestTakeOutDevFaultInfo(t *testing.T) {
	convey.Convey("test TakeOutDevFaultInfo", t, func() {
		convey.Convey("TakeOutDevFaultInfo success", func() {
			devFaultInfoMap = make(map[int32][]common.DevFaultInfo, GeneralMapSize)
			convey.So(len(GetAndCleanFaultInfo()), convey.ShouldEqual, 0)
			testInfo := []common.DevFaultInfo{{EventID: 1}}
			devFaultInfoMap[0] = testInfo
			convey.So(GetAndCleanFaultInfo()[0], convey.ShouldResemble, testInfo)
			convey.So(len(devFaultInfoMap[0]), convey.ShouldEqual, 0)
		})
	})
}

// TestGetNetworkFaultTypeByCode for test GetNetworkFaultTypeByCode
func TestGetNetworkFaultTypeByCode(t *testing.T) {
	convey.Convey("test GetNetworkFaultTypeByCode", t, func() {
		faultCodes := []string{CardNetworkDisconnected}
		convey.Convey("fault type NormalNetwork", func() {
			convey.So(GetNetworkFaultTypeByCode(nil), convey.ShouldEqual, NormalNetwork)
		})
		convey.Convey("fault type NotHandleFault", func() {
			faultTypeCode = FaultTypeCode{
				NotHandleFaultNetworkCodes: faultCodes,
				NotHandleFaultCodes:        []int64{1},
			}
			convey.So(GetNetworkFaultTypeByCode(faultCodes), convey.ShouldEqual, NotHandleFault)
		})
		convey.Convey("fault type SeparateNPU", func() {
			faultTypeCode = FaultTypeCode{
				SeparateNPUNetworkCodes: faultCodes,
				NotHandleFaultCodes:     []int64{1},
			}
			convey.So(GetNetworkFaultTypeByCode(faultCodes), convey.ShouldEqual, SeparateNPU)
		})
		convey.Convey("fault type PreSeparateNPU", func() {
			faultTypeCode = FaultTypeCode{
				PreSeparateNPUNetworkCodes: faultCodes,
				NotHandleFaultCodes:        []int64{1},
			}
			convey.So(GetNetworkFaultTypeByCode(faultCodes), convey.ShouldEqual, PreSeparateNPU)
			faultTypeCode = FaultTypeCode{}
			convey.So(GetNetworkFaultTypeByCode(faultCodes), convey.ShouldEqual, PreSeparateNPU)
		})
		convey.Convey("read json failed", func() {
			faultTypeCode = FaultTypeCode{}
			mockLoadFile := gomonkey.ApplyFuncReturn(utils.LoadFile, nil, errors.New("failed"))
			defer mockLoadFile.Reset()
			convey.So(GetNetworkFaultTypeByCode(faultCodes), convey.ShouldEqual, PreSeparateNPU)
		})
	})
}
