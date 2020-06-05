/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2019-2024. All rights reserved.
 * Description: dsmi_test.go
 * Create: 19-11-20 下午8:52
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
