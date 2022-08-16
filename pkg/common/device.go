// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package common a series of common function
package common

import (
	"fmt"
	"regexp"
	"strings"

	"huawei.com/npu-exporter/hwlog"
	"k8s.io/apimachinery/pkg/util/sets"
)

func getDeviceID(deviceName string, ascendRuntimeOptions string) (string, string, error) {
	// hiAIAscend310Prefix: davinci-mini
	// vnpu: davinci-coreNum-vid-devID, like Ascend910-2c-111-0
	// ascend310:  davinci-mini0
	idSplit := strings.Split(deviceName, GangSepDev)

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
func GetDeviceListID(devices []string, ascendRuntimeOptions string) (map[string]string, error) {
	if len(devices) > MaxDevicesNum {
		return nil, fmt.Errorf("device num excceed max num, when get device list id")
	}
	ascendVisibleDevices := make(map[string]string, len(devices))
	for _, id := range devices {
		deviceID, virID, err := getDeviceID(id, ascendRuntimeOptions)
		if err != nil {
			hwlog.RunLog.Errorf("get device ID err: %v", err)
			return nil, err
		}
		if ascendRuntimeOptions == VirtualDev {
			ascendVisibleDevices[virID] = ""
			continue
		}
		ascendVisibleDevices[deviceID] = ""
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
