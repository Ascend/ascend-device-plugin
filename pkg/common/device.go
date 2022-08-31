// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package common a series of common function
package common

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"huawei.com/mindx/common/hwlog"
	"k8s.io/apimachinery/pkg/util/sets"
)

func getDeviceID(deviceName string, ascendRuntimeOptions string) (string, string, error) {
	// hiAIAscend310Prefix: davinci-mini
	// vnpu: davinci-coreNum-vid-devID, like Ascend910-2c-111-0
	// ascend310:  davinci-mini0
	idSplit := strings.Split(deviceName, MiddelLine)

	if len(idSplit) < PhyDeviceLen {
		return "", "", fmt.Errorf("id: %s is invalid", deviceName)
	}
	phyID := idSplit[len(idSplit)-1]
	// for virtual device, index 2 data means it's id
	var virID string
	if ascendRuntimeOptions == VirtualDev && len(idSplit) == VirDeviceLen {
		virID = idSplit[PhyDeviceLen]
	}
	return phyID, virID, nil
}

// GetDeviceListID get device id by input device name
func GetDeviceListID(devices []string, ascendRuntimeOptions string) ([]string, error) {
	if len(devices) > MaxDevicesNum {
		return nil, fmt.Errorf("device num excceed max num, when get device list id")
	}
	var ascendVisibleDevices []string
	for _, id := range devices {
		deviceID, virID, err := getDeviceID(id, ascendRuntimeOptions)
		if err != nil {
			hwlog.RunLog.Errorf("get device ID err: %v", err)
			return nil, err
		}
		if ascendRuntimeOptions == VirtualDev {
			ascendVisibleDevices = append(ascendVisibleDevices, virID)
			continue
		}
		ascendVisibleDevices = append(ascendVisibleDevices, deviceID)
	}
	return ascendVisibleDevices, nil
}

// IsVirtualDev used to judge whether a physical device or a virtual device
func IsVirtualDev(devType string) bool {
	patternMap := GetPattern()
	reg910 := regexp.MustCompile(patternMap["vir910"])
	reg310P := regexp.MustCompile(patternMap["vir310p"])
	return reg910.MatchString(devType) || reg310P.MatchString(devType)
}

// ToString convert input data to string
func ToString(devices sets.String, sepType string) string {
	return strings.Join(devices.List(), sepType)
}

// ConvertDevListToSets convert devices to Sets
func ConvertDevListToSets(devices, sepType string) sets.String {
	if devices == "" {
		return sets.String{}
	}
	deviceInfo := strings.Split(devices, sepType)
	if len(deviceInfo) > MaxDevicesNum {
		hwlog.RunLog.Error("The number of device exceeds the upper limit")
		return sets.String{}
	}
	if sepType == DotSepDev {
		return labelToSets(deviceInfo)
	}
	return deviceInfoToSets(deviceInfo)
}

// for label, check device format, must 0.1.2 and more
func labelToSets(deviceInfo []string) sets.String {
	deviceSets := sets.String{}
	for _, device := range deviceInfo {
		if _, isValidNum := IsValidNumber(device); isValidNum {
			deviceSets.Insert(device)
		}
	}
	return deviceSets
}

// for device info, check device format, must Ascend910-0,Ascend910-1 and more
func deviceInfoToSets(deviceInfo []string) sets.String {
	// pattern no need to defined as global variable, only used here
	deviceSets := sets.String{}
	for _, device := range deviceInfo {
		if match, err := regexp.MatchString(GetPattern()["ascend910"], device); !match || err != nil {
			hwlog.RunLog.Warnf("current device %s format err: %#v", device, err)
			continue
		}
		deviceSets.Insert(device)
	}
	return deviceSets
}

// IsValidNumber input checkVal is a valid number
func IsValidNumber(checkVal string) (int64, bool) {
	if strings.Contains(checkVal, UnderLine) {
		hwlog.RunLog.Warnf("device id %s invalid", checkVal)
		return -1, false
	}
	conversionRes, err := strconv.ParseInt(checkVal, BaseDec, BitSize)
	if err != nil {
		hwlog.RunLog.Warnf("current device id invalid, err: %#v", err)
		return -1, false
	}
	return conversionRes, true
}
