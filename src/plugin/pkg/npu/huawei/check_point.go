// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package huawei get data from kubelet check point file
package huawei

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"huawei.com/npu-exporter/hwlog"
	"huawei.com/npu-exporter/utils"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"k8s.io/kubernetes/pkg/kubelet/cm/devicemanager/checkpoint"

	"Ascend-device-plugin/src/plugin/pkg/npu/common"
)

const (
	kubeletCheckPointFile = "/var/lib/kubelet/device-plugins/kubelet_internal_checkpoint"
	resourceNameSplitLen  = 2
	resourceTypeIndex     = 1
	phyDeviceIDIndex      = 3
)

// CheckpointData check point info
type CheckpointData struct {
	ResourceName string
	Request      []string
	Response     []string
}

func readCheckPoint(filePath string) ([]checkpoint.PodDevicesEntry, error) {
	checkPointBytes, err := utils.ReadLimitBytes(filePath, utils.Size10M)
	if err != nil {
		return nil, fmt.Errorf("there is no file provided")
	}

	registeredDevs := make(map[string][]string, 1)
	var devEntries []checkpoint.PodDevicesEntry
	cp := checkpoint.New(devEntries, registeredDevs)
	if err := cp.UnmarshalCheckpoint(checkPointBytes); err != nil {
		return nil, fmt.Errorf("unmarshal failed, error is %v", err)
	}

	if err := cp.VerifyChecksum(); err != nil {
		return nil, fmt.Errorf("failed to retrieve checkpoint for %v", err)
	}

	podDeviceEntries, _ := cp.GetData()
	return podDeviceEntries, nil
}

func getEnvVisibleDevices(allocResp []byte) []string {
	if len(allocResp) == 0 {
		hwlog.RunLog.Errorf("allocate response is empty")
		return nil
	}

	response := v1beta1.ContainerAllocateResponse{}
	if err := response.Unmarshal(allocResp); err != nil {
		hwlog.RunLog.Errorf("unmarshal failed, error is %v", err)
		return nil
	}
	visibleDevices, ok := response.Envs[ascendVisibleDevicesEnv]
	if !ok {
		hwlog.RunLog.Errorf("ascend visible devices env does not exist")
		return nil
	}

	var validDeviceIDs []string
	deviceIDs := strings.Split(visibleDevices, ",")
	for _, deviceID := range deviceIDs {
		idNum, err := strconv.Atoi(deviceID)
		if err == nil && idNum >= 0 && idNum <= math.MaxUint32 {
			validDeviceIDs = append(validDeviceIDs, deviceID)
			continue
		}
		hwlog.RunLog.Errorf("device id is invalid")
		return nil
	}
	return validDeviceIDs
}

// GetKubeletCheckPoint get check point data from file
func GetKubeletCheckPoint(filePath string) (map[string]CheckpointData, error) {
	podDeviceEntries, err := readCheckPoint(filePath)
	if err != nil {
		return nil, fmt.Errorf("read check point file failed, error is: %v", err)
	}
	checkpointData := map[string]CheckpointData{}
	for _, podDeviceEntry := range podDeviceEntries {
		validDeviceIDs := getEnvVisibleDevices(podDeviceEntry.AllocResp)
		if len(validDeviceIDs) == 0 {
			hwlog.RunLog.Errorf("get env visible devices failed")
			continue
		}

		checkpointData[podDeviceEntry.PodUID] = CheckpointData{
			ResourceName: podDeviceEntry.ResourceName, // like "huawei.com/Ascend310" or "huawei.com/Ascend910-8c"
			Request:      podDeviceEntry.DeviceIDs,
			Response:     validDeviceIDs,
		}
	}
	return checkpointData, nil
}

func checkDevType(devType, runMode string) bool {
	pwrSuffix := map[string]string{hiAIAscend910Prefix: "", pwr2CSuffix: "", pwr4CSuffix: "", pwr8CSuffix: "",
		pwr16CSuffix: ""}
	if runMode == hiAIAscend710Prefix {
		pwrSuffix = map[string]string{hiAIAscend710Prefix: "", chip710Core1C: "", chip710Core2C: "", chip710Core4C: ""}
	} else if runMode == hiAIAscend310Prefix {
		pwrSuffix = map[string]string{hiAIAscend310Prefix: ""}
	}

	_, exist := pwrSuffix[devType]
	return exist
}

// GetAnnotation get annotation from check point data
func GetAnnotation(data CheckpointData, runMode string) ([]string, []string, error) {
	// Request is kubelet allocate devices like "Ascend910-0" or "Ascend910-8c-197-6"
	if len(data.Request) == 0 {
		err := fmt.Errorf("request is empty")
		return nil, nil, err
	}
	// ResourceName like "huawei.com/Ascend910-8c" or "huawei.com/Ascend910"
	resourceData := strings.Split(data.ResourceName, "/")
	if len(resourceData) != resourceNameSplitLen {
		err := fmt.Errorf("resource name: %s is invalid", data.ResourceName)
		return nil, nil, err
	}
	resourceType := resourceData[resourceTypeIndex]
	if !checkDevType(resourceType, runMode) {
		err := fmt.Errorf("resource type: %s is invalid", resourceType)
		return nil, nil, err
	}
	devInfo := strings.Split(data.Request[0], "-")
	if len(devInfo) != common.PhyDeviceLen && len(devInfo) != common.VirDeviceLen {
		err := fmt.Errorf("devices info is invalild")
		return nil, nil, err
	}

	var responseDeviceName []string
	for _, id := range data.Response {
		if len(devInfo) == common.PhyDeviceLen {
			responseDeviceName = append(responseDeviceName, fmt.Sprintf("%s-%s", resourceType, id))
		} else {
			responseDeviceName = append(responseDeviceName, fmt.Sprintf("%s-%s-%s", resourceType, id,
				devInfo[phyDeviceIDIndex]))
		}
	}
	return data.Request, responseDeviceName, nil
}
