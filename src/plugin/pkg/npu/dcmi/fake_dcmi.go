// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package dcmi using driver interface
package dcmi

import (
	"fmt"
	"math"
)

const (
	npuTestNum         = 8
	npuTestDeviceNum   = 1
	testAiCoreNum      = 20
	testVDevNum        = 2
	testComputeCoreNum = 4
)

// FakeDriverManager FakeDriverManager
type FakeDriverManager struct{}

// NewFakeDriverManager FakeDriverManager
func NewFakeDriverManager() *FakeDriverManager {
	return &FakeDriverManager{}
}

// ShutDown clean the dynamically loaded resource
func (d *FakeDriverManager) ShutDown() error {
	return nil
}

// GetCardList get card list
func (d *FakeDriverManager) GetCardList() (int32, []int32, error) {
	var cardIDList []int32
	for i := 0; i < npuTestNum; i++ {
		cardIDList = append(cardIDList, int32(i))
	}
	return npuTestNum, cardIDList, nil
}

// GetDeviceNumInCard get device number in the npu card
func (d *FakeDriverManager) GetDeviceNumInCard(cardID int32) (int32, error) {
	return int32(npuTestDeviceNum), nil
}

// GetDeviceLogicID get device logicID
func (d *FakeDriverManager) GetDeviceLogicID(cardID, deviceID int32) (uint32, error) {
	return uint32(cardID), nil
}

// SetDestroyVirtualDevice destroy virtual device
func (d *FakeDriverManager) SetDestroyVirtualDevice(cardID, deviceID int32, vDevID uint32) error {
	return nil
}

// CreateVirtualDevice create virtual device
func (d *FakeDriverManager) CreateVirtualDevice(cardID, deviceID, vDevID int32, aiCore uint32) (CgoDcmiCreateVDevOut,
	error) {
	return CgoDcmiCreateVDevOut{}, nil
}

// GetDeviceVDevResource get virtual device resource info
func (d *FakeDriverManager) GetDeviceVDevResource(cardID, deviceID int32, vDevID uint32) (CgoVDevQueryStru, error) {
	vDevResource := CgoVDevQueryStru{}
	vDevResource.queryInfo.status = uint32(0)
	vDevResource.queryInfo.vfid = vDevID
	vDevResource.queryInfo.containerID = uint64(vDevID)
	vDevResource.queryInfo.computing.aic = float32(testComputeCoreNum * vDevID)
	return vDevResource, nil
}

// GetDeviceTotalResource get device total resource info
func (d *FakeDriverManager) GetDeviceTotalResource(cardID, deviceID int32) (CgoDcmiSocTotalResource, error) {
	var totalResource CgoDcmiSocTotalResource
	totalResource.vDevNum = testVDevNum
	for i := 0; i < testVDevNum; i++ {
		totalResource.vDevID = append(totalResource.vDevID, uint32(int(cardID)+i))
	}
	return totalResource, nil
}

// GetDeviceFreeResource get device free resource info
func (d *FakeDriverManager) GetDeviceFreeResource(cardID, deviceID int32) (CgoDcmiSocFreeResource, error) {
	var freeSource CgoDcmiSocFreeResource
	freeSource.computing.aic = testAiCoreNum
	return freeSource, nil
}

// GetDeviceInfo get device resource info
func (d *FakeDriverManager) GetDeviceInfo(cardID, deviceID int32) (CgoVDevInfo, error) {
	cgoDcmiSocTotalResource, err := d.GetDeviceTotalResource(cardID, deviceID)
	if err != nil {
		return CgoVDevInfo{}, fmt.Errorf("get device tatal resource failed, error is: %v", err)
	}

	cgoDcmiSocFreeResource, err := d.GetDeviceFreeResource(cardID, deviceID)
	if err != nil {
		return CgoVDevInfo{}, fmt.Errorf("get device free resource failed, error is: %v", err)
	}

	dcmiVDevInfo := CgoVDevInfo{
		VDevNum:       cgoDcmiSocTotalResource.vDevNum,
		CoreNumUnused: cgoDcmiSocFreeResource.computing.aic,
	}
	for i := 0; i < len(cgoDcmiSocTotalResource.vDevID); i++ {
		dcmiVDevInfo.VDevID = append(dcmiVDevInfo.VDevID, cgoDcmiSocTotalResource.vDevID[i])
	}
	for _, vDevID := range cgoDcmiSocTotalResource.vDevID {
		cgoVDevQueryStru, err := d.GetDeviceVDevResource(cardID, deviceID, vDevID)
		if err != nil {
			return CgoVDevInfo{}, fmt.Errorf("get device vitrual resource failed, error is: %v", err)
		}
		dcmiVDevInfo.Status = append(dcmiVDevInfo.Status, cgoVDevQueryStru.queryInfo.status)
		dcmiVDevInfo.VfID = append(dcmiVDevInfo.VfID, cgoVDevQueryStru.queryInfo.vfid)
		dcmiVDevInfo.CID = append(dcmiVDevInfo.CID, cgoVDevQueryStru.queryInfo.containerID)
		dcmiVDevInfo.CoreNum = append(dcmiVDevInfo.CoreNum, cgoVDevQueryStru.queryInfo.computing.aic)
	}
	return dcmiVDevInfo, nil
}

// GetCardIDDeviceID get card id and device id from logic id
func (d *FakeDriverManager) GetCardIDDeviceID(logicID uint32) (int32, int32, error) {
	if logicID > uint32(math.MaxInt8) {
		return retError, retError, fmt.Errorf("input invalid logicID: %d", logicID)
	}

	_, cards, err := d.GetCardList()
	if err != nil {
		return retError, retError, fmt.Errorf("get card list failed, error is: %v", err)
	}

	for _, cardID := range cards {
		deviceNum, err := d.GetDeviceNumInCard(cardID)
		if err != nil {
			fmt.Printf("get device num in card failed, error is: %v\n", err)
			continue
		}
		for deviceID := int32(0); deviceID < deviceNum; deviceID++ {
			logicIDGet, err := d.GetDeviceLogicID(cardID, deviceID)
			if err != nil {
				fmt.Printf("get device logic id failed, error is: %v\n", err)
				continue
			}
			if logicID == logicIDGet {
				return cardID, deviceID, nil
			}
		}
	}
	errInfo := fmt.Errorf("the card id and device id corresponding to the logic id are not found")
	return retError, retError, errInfo
}

// CreateVDevice create virtual device by logic id
func (d *FakeDriverManager) CreateVDevice(logicID uint32, aiCore uint32) (uint32, error) {
	cardID, deviceID, err := d.GetCardIDDeviceID(logicID)
	if err != nil {
		return unretError, fmt.Errorf("get card id and device id failed, error is: %v", err)
	}

	cgoDcmiSocFreeResource, err := d.GetDeviceFreeResource(cardID, deviceID)
	if err != nil {
		return unretError, fmt.Errorf("get virtual device info failed, error is: %v", err)
	}

	if cgoDcmiSocFreeResource.computing.aic < float32(aiCore) {
		return unretError, fmt.Errorf("the remaining core resource is insufficient, free core: %f",
			cgoDcmiSocFreeResource.computing.aic)
	}

	var vDevID int32
	createVDevOut, err := d.CreateVirtualDevice(cardID, deviceID, vDevID, aiCore)
	if err != nil {
		return unretError, fmt.Errorf("create virtual device failed, error is: %v", err)
	}
	return createVDevOut.VDevID, nil
}

// GetVDeviceInfo get virtual device info by logic id
func (d *FakeDriverManager) GetVDeviceInfo(logicID uint32) (CgoVDevInfo, error) {
	cardID, deviceID, err := d.GetCardIDDeviceID(logicID)
	if err != nil {
		return CgoVDevInfo{}, fmt.Errorf("get card id and device id failed, error is: %v", err)
	}

	dcmiVDevInfo, err := d.GetDeviceInfo(cardID, deviceID)
	if err != nil {
		return CgoVDevInfo{}, fmt.Errorf("get virtual device info failed, error is: %v", err)
	}
	return dcmiVDevInfo, nil
}

// DestroyVDevice destroy spec virtual device by logic id
func (d *FakeDriverManager) DestroyVDevice(logicID, vDevID uint32) error {
	cardID, deviceID, err := d.GetCardIDDeviceID(logicID)
	if err != nil {
		return fmt.Errorf("get card id and device id failed, error is: %v", err)
	}

	if err = d.SetDestroyVirtualDevice(cardID, deviceID, vDevID); err != nil {
		return fmt.Errorf("destroy virtual device failed, error is: %v", err)
	}
	return nil
}
