// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package dsmi using driver interface
package dsmi

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	hiAIMaxDeviceNum   = 64
	maxChipName        = 32
	deviceIPLength     = 4
	npuTestNum         = 8
	maxAiCoreNum       = 32
	testAiCoreNum      = 20
	defaultVDevNum     = 0
	testVDevNum        = 2
	testComputeCoreNum = 4
	// UnHealthyTestLogicID use for ut, represent the device is unhealthy whose logicID is 3
	UnHealthyTestLogicID = 3

	// testPhyDevID use for ut, represent device id: 0,2,4,5,6,7
	testPhyDevID = "0234567"
)

// FakeDeviceManager fakeDeviceManager
type FakeDeviceManager struct{}

// NewFakeDeviceManager FakeDeviceManager
func NewFakeDeviceManager() *FakeDeviceManager {
	return &FakeDeviceManager{}
}

// EnableContainerService for enableContainerService
func (d *FakeDeviceManager) EnableContainerService() error {
	return nil
}

// CreateVirtualDevice for create virtual device
func (d *FakeDeviceManager) CreateVirtualDevice(uint32, string, []string) error {
	return nil
}

// DestroyVirtualDevice for destroy virtual device
func (d *FakeDeviceManager) DestroyVirtualDevice(uint32, uint32) error {
	return nil
}

// GetDeviceCount get ascend910 device quantity
func (d *FakeDeviceManager) GetDeviceCount() (int32, error) {
	return int32(npuTestNum), nil
}

// GetDeviceList device get list
func (d *FakeDeviceManager) GetDeviceList(devices *[hiAIMaxDeviceNum]uint32) (int32, error) {
	devNum, err := d.GetDeviceCount()
	if err != nil {
		return devNum, err
	}
	// transfer device list
	var i int32
	for i = 0; i < devNum; i++ {
		(*devices)[i] = uint32(i)
	}

	return devNum, nil
}

// GetDeviceHealth get device health by id
func (d *FakeDeviceManager) GetDeviceHealth(logicID int32) (uint32, error) {
	if logicID == UnHealthyTestLogicID {
		return uint32(UnHealthyTestLogicID), nil
	}
	return uint32(0), nil
}

// GetDeviceNetworkHealth get device network health by id
func (d *FakeDeviceManager) GetDeviceNetworkHealth(logicID int32) (uint32, error) {
	if logicID == UnHealthyTestLogicID {
		return uint32(UnHealthyTestLogicID), nil
	}
	return uint32(0), nil
}

// GetPhyID get physic id form logic id
func (d *FakeDeviceManager) GetPhyID(logicID uint32) (uint32, error) {
	return logicID, nil
}

// GetLogicID get logic id form physic id
func (d *FakeDeviceManager) GetLogicID(phyID uint32) (uint32, error) {
	return phyID, nil

}

// ShutDown the function
func (d *FakeDeviceManager) ShutDown() {
	fmt.Printf("use fake DeviceManager function ShutDown")
}

// GetChipInfo for fakeDeviceManager
func (d *FakeDeviceManager) GetChipInfo(logicID int32) (string, error) {
	return "310", nil
}

// GetDeviceIP get deviceIP
func (d *FakeDeviceManager) GetDeviceIP(logicID int32) (string, error) {
	retIPAddress := fmt.Sprintf("%d.%d.%d.%d", 0, 0, 0, logicID)
	return retIPAddress, nil
}

// GetVDevicesInfo for fakeDeviceManager
func (d *FakeDeviceManager) GetVDevicesInfo(logicID uint32) (CgoDsmiVDevInfo, error) {
	var cgoDsmiVDevInfos CgoDsmiVDevInfo
	if strings.Contains(testPhyDevID, strconv.Itoa(int(logicID))) {
		cgoDsmiVDevInfos = CgoDsmiVDevInfo{
			VDevNum:       uint32(defaultVDevNum),
			CoreNumUnused: uint32(maxAiCoreNum),
		}
		return cgoDsmiVDevInfos, nil
	}
	cgoDsmiVDevInfos = CgoDsmiVDevInfo{
		VDevNum:       uint32(testVDevNum),
		CoreNumUnused: uint32(testAiCoreNum),
	}
	for i := 0; i < testVDevNum; i++ {
		coreNum := fmt.Sprintf("%d", testComputeCoreNum*(i+1))
		cgoDsmiVDevInfos.CgoDsmiSubVDevInfos = append(cgoDsmiVDevInfos.CgoDsmiSubVDevInfos, CgoDsmiSubVDevInfo{
			Status: uint32(0),
			VDevID: uint32(int(logicID) + i),
			VfID:   uint32(int(logicID) + i),
			CID:    uint64(i),
			Spec: CgoDsmiVdevSpecInfo{
				CoreNum: coreNum,
			},
		})
	}
	return cgoDsmiVDevInfos, nil
}

// GetDeviceErrorCode get device error code
func (d *FakeDeviceManager) GetDeviceErrorCode(logicID uint32) error {
	return nil
}
