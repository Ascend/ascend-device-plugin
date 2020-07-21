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

package huawei

import (
	"go.uber.org/zap"
	"testing"
)

// deviceGetCount function unit test
func TestDeviceGetCount(t *testing.T) {

	deviceCount, err := getDeviceCount()
	if err != nil {
		t.Errorf("%s", err)
	}
	// result
	logger.Info("the AI server found device", zap.Int32("device quantity", deviceCount))
}

func TestGetDeviceList(t *testing.T) {
	var ids [hiAIMaxDeviceNum]uint32

	devNum, err := getDeviceList(&ids)
	if err != nil {
		t.Errorf("%s", err)
	}

	logger.Info("device quantity:", zap.Int32("device quantity", devNum), zap.Any("device list", ids))
}

// test the get device health function
func TestGetDeviceHealth(t *testing.T) {
	devNum, err := getDeviceCount()
	if err != nil {
		t.Errorf("%s", err)
	}

	var i int32
	for i = 0; i < devNum; i++ {
		health, err := getDeviceHealth(i)
		if err != nil {
			t.Errorf("%s", err)
		}
		logger.Info("the device healthy state", zap.Int32("deviceID", i), zap.Uint32("healthy state", health))
	}

}

// test get device ip address
func TestGetDeviceIp(t *testing.T) {
	devNum, err := getDeviceCount()
	if err != nil {
		t.Errorf("%s", err)
	}

	var i int32
	for i = 0; i < devNum; i++ {
		retIPAddress, err := getDeviceIP(i)
		if err != nil {
			t.Errorf("%s", err)
		}
		logger.Info("the device ip address is:",
			zap.Int32("deviceID", i),
			zap.String("healthy state", retIPAddress))
	}

}

func TestGetPhyID(t *testing.T) {

	devNum, err := getDeviceCount()
	if err != nil {
		t.Errorf("%s", err)
	}

	var i uint32
	for i = 0; i < uint32(devNum); i++ {
		phyID, err := getPhyID(i)
		if err != nil {
			t.Errorf("%s", err)
		}
		logger.Info("get device PhyID", zap.Uint32("deviceID", i),
			zap.Uint32("phyID", phyID))
	}

}

func TestGetLogicID(t *testing.T) {
	devNum, err := getDeviceCount()
	if err != nil {
		t.Errorf("%s", err)
	}

	var i uint32
	for i = 0; i < uint32(devNum); i++ {
		logicID, err := getLogicID(i)
		if err != nil {
			t.Errorf("%s", err)
		}
		logger.Info("get device logicID", zap.Uint32("deviceID", i),
			zap.Uint32("logicID", logicID))
	}
}

func TestGetChipInfo(t *testing.T) {
	devNum, err := getDeviceCount()
	if err != nil {
		t.Errorf("%s", err)
	}

	var i int32
	for i = 0; i < devNum; i++ {
		chipinfo, err := GetChipInfo(i)
		if err != nil {
			t.Errorf("%s", err)
		}
		logger.Info("the device healthy state", zap.Int32("deviceID", i), zap.String("chipNmae: ",
			chipinfo.ChipName), zap.String("chipType: ", chipinfo.ChipType),
			zap.String("chipVer: ", chipinfo.ChipVer))
	}
}
