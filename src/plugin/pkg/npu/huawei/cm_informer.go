// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package huawei using informer update cache for hps.devices
package huawei

import (
	"os"
	"path"

	"huawei.com/npu-exporter/hwlog"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"Ascend-device-plugin/src/plugin/pkg/npu/common"
)

const (
	// waitingTimeTask wait timing task update 10s
	waitingTimeTask = 10
)

func updateHpsCache(hdm *HwDevManager) {
	hwlog.RunLog.Infof("start update multi-virtual device cache after create virtual device")
	var newDevices []common.NpuDevice
	var newDevTypes []string
	if err := hdm.manager.GetNPUs(&newDevices, &newDevTypes, hdm.runMode); err != nil {
		hwlog.RunLog.Errorf("get new NPU devices failed, err: %v\n", err)
		return
	}
	getDiffDevCount(hdm, newDevices)
	registerNewServer(hdm, newDevTypes)
}

func getDiffDevCount(hdm *HwDevManager, newDevices []common.NpuDevice) {
	pwrSuffix := []string{hiAIAscend910Prefix, pwr2CSuffix, pwr4CSuffix, pwr8CSuffix, pwr16CSuffix}
	if hdm.runMode == common.RunMode710 {
		pwrSuffix = []string{hiAIAscend710Prefix, chip710Core1C, chip710Core2C, chip710Core4C}
	}
	oldDevices := hdm.allDevs
	for _, devType := range pwrSuffix {
		listenDevCountIsChange[devType] = isDevCountChange(oldDevices, newDevices, devType)
	}
	hdm.allDevs = newDevices
}

func isDevCountChange(oldDevices, newDevices []common.NpuDevice, devType string) bool {
	return isDevEqual(getSpecDevTypes(oldDevices, devType), getSpecDevTypes(newDevices, devType))
}

func registerNewServer(hdm *HwDevManager, newDevTypes []string) {
	hwlog.RunLog.Infof("starting reRegister new type virtual device server")
	interDevTypes := getInterDevType(hdm.GetDevType(), newDevTypes)
	for devType := range getDiffDevType(hdm.GetDevType(), newDevTypes) {
		sockPath := path.Join(v1beta1.DevicePluginPath, devType)
		if _, err := os.Stat(sockPath); err == nil {
			continue
		}
		go hdm.Serve(devType)
		hdm.allDevTypes = append(hdm.allDevTypes, devType)
	}
	for devType := range interDevTypes {
		ServeUpdateMap[devType] = make(chan int, len(interDevTypes))
		ServeUpdateMap[devType] <- 1
	}
	hwlog.RunLog.Infof("reRegister new type virtual device server complete")
}

func getDiffDevType(devTypes, newDevTypes []string) sets.String {
	return convertToSets(newDevTypes).Difference(convertToSets(devTypes))
}

func getInterDevType(devTypes, newDevTypes []string) sets.String {
	return convertToSets(newDevTypes).Intersection(convertToSets(devTypes))
}

func getSpecDevTypes(devices []common.NpuDevice, devType string) []string {
	var devTypes []string
	for _, device := range devices {
		if device.DevType == devType {
			devTypes = append(devTypes, device.ID)
		}
	}
	return devTypes
}

func isDevEqual(oldDevs, newDevs []string) bool {
	return !convertToSets(oldDevs).Equal(convertToSets(newDevs))
}

func convertToSets(devTypes []string) sets.String {
	devSet := sets.String{}
	for _, devType := range devTypes {
		devSet.Insert(devType)
	}
	return devSet
}
