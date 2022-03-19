// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.
// Package dsmi, using driver interface
package dsmi

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	hiAIMaxDeviceNum     = 64
	maxChipName          = 32
	deviceIPLength       = 4
	npuTestNum           = 8
	maxAiCoreNum         = 32
	testAiCoreNum        = 20
	defaultVDevNum       = 0
	testVDevNum          = 2
	testComputeCoreNum   = 4
	unHealthyTestLogicID = 3

	// testPhyDevID use for ut, represent device id: 0,2,4,5,6,7
	testPhyDevID = "0234567"
)

type fakeDeviceManager struct{}

// NewFakeDeviceManager fakeDeviceManager
func NewFakeDeviceManager() *fakeDeviceManager {
	return &fakeDeviceManager{}
}

// CreateVirtualDevice for create virtual device
func (d *fakeDeviceManager) CreateVirtualDevice(uint32, string, []string) error {
	return nil
}

// DestroyVirtualDevice for destroy virtual device
func (d *fakeDeviceManager) DestroyVirtualDevice(uint32, uint32) error {
	return nil
}

// GetDeviceCount get ascend910 device quantity
func (d *fakeDeviceManager) GetDeviceCount() (int32, error) {
	return int32(npuTestNum), nil
}

// GetDeviceList device get list
func (d *fakeDeviceManager) GetDeviceList(devices *[hiAIMaxDeviceNum]uint32) (int32, error) {
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

//  GetDeviceHealth get device health by id
func (d *fakeDeviceManager) GetDeviceHealth(logicID int32) (uint32, error) {
	if logicID == unHealthyTestLogicID {
		return uint32(unHealthyTestLogicID), nil
	}
	return uint32(0), nil
}

//  GetDeviceNetworkHealth get device network health by id
func (d *fakeDeviceManager) GetDeviceNetworkHealth(logicID int32) (uint32, error) {
	if logicID == unHealthyTestLogicID {
		return uint32(unHealthyTestLogicID), nil
	}
	return uint32(0), nil
}

// GetPhyID get physic id form logic id
func (d *fakeDeviceManager) GetPhyID(logicID uint32) (uint32, error) {
	return logicID, nil
}

// GetLogicID get logic id form physic id
func (d *fakeDeviceManager) GetLogicID(phyID uint32) (uint32, error) {
	return phyID, nil

}

// ShutDown the function
func (d *fakeDeviceManager) ShutDown() {
	fmt.Printf("use fake DeviceManager function ShutDown")
}

// GetChipInfo for fakeDeviceManager
func (d *fakeDeviceManager) GetChipInfo(logicID int32) (string, error) {
	return "310", nil
}

// GetDeviceIP get deviceIP
func (d *fakeDeviceManager) GetDeviceIP(logicID int32) (string, error) {
	retIPAddress := fmt.Sprintf("%d.%d.%d.%d", 0, 0, 0, logicID)
	return retIPAddress, nil
}

// GetVDevicesInfo for fakeDeviceManager
func (d *fakeDeviceManager) GetVDevicesInfo(logicID uint32) (CgoDsmiVDevInfo, error) {
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
	for i := 0; i < 2; i++ {
		coreNum := fmt.Sprintf("%d", testComputeCoreNum*(i+1))
		cgoDsmiVDevInfos.CgoDsmiSubVDevInfos = append(cgoDsmiVDevInfos.CgoDsmiSubVDevInfos, CgoDsmiSubVDevInfo{
			Status: uint32(0),
			Vdevid: uint32(int(logicID) + i),
			Vfid:   uint32(int(logicID) + i),
			Cid:    uint64(i),
			Spec: CgoDsmiVdevSpecInfo{
				CoreNum: coreNum,
			},
		})
	}
	return cgoDsmiVDevInfos, nil
}

// GetDeviceErrorCode get device error code
func (d *fakeDeviceManager) GetDeviceErrorCode(logicID uint32) error {
	return nil
}
