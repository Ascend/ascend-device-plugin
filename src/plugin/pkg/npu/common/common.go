// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.
// Package common, a series of common function
package common

import (
	"fmt"
	"regexp"
	"strings"

	hwutil "huawei.com/npu-exporter/utils"
	"k8s.io/client-go/kubernetes"
)

const (
	kubeEnvMaxLength = 253
	component        = "device-plugin"
	idSplitNum       = 2
	virDeviceLen     = 4
	// VirtualDev represent virtual device
	VirtualDev = "VIRTUAL"

	// VNpuCfgKey is the key for virtual NPU configMap record
	VNpuCfgKey = "VNPUCfg"
)

// NpuDevice npu device description
type NpuDevice struct {
	DevType       string
	PciID         string
	ID            string
	Health        string
	NetworkHealth string
}

// GetDeviceID get phyID and virtualID
func GetDeviceID(deviceName string, ascendRuntimeOptions string) (string, string, error) {

	// hiAIAscend310Prefix: davinci-mini
	// vnpu: davinci-coreNum-vid-devID, like Ascend910-2c-111-0
	// ascend310:  davinci-mini0

	idSplit := strings.Split(deviceName, "-")

	if len(idSplit) < idSplitNum {
		return "", "", fmt.Errorf("id: %s is invalid", deviceName)
	}
	var virID string
	deviceID := idSplit[len(idSplit)-1]
	if ascendRuntimeOptions == VirtualDev && len(idSplit) == virDeviceLen {
		virID = idSplit[idSplitNum]
	}
	return deviceID, virID, nil
}

// CheckNodeName for check node name
func CheckNodeName(nodeName string) error {
	if len(nodeName) > kubeEnvMaxLength {
		return fmt.Errorf("node name length %d is bigger than %d", len(nodeName), kubeEnvMaxLength)
	}
	pattern := `^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`
	reg := regexp.MustCompile(pattern)
	if !reg.MatchString(nodeName) {
		return fmt.Errorf("node name %s is illegal", nodeName)
	}
	return nil
}

// NewKubeClient get client from KUBECONFIG  or not
func NewKubeClient(kubeConfig string) (*kubernetes.Clientset, error) {
	return hwutil.K8sClientFor(kubeConfig, component)
}
