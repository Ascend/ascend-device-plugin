// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package common a series of common function
package common

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"huawei.com/npu-exporter/utils"
	"k8s.io/client-go/kubernetes"
)

const (
	kubeEnvMaxLength = 253
	component        = "device-plugin"
	// PhyDeviceLen is the length of physical device
	PhyDeviceLen = 2
	// VirDeviceLen is the length of virtual device
	VirDeviceLen = 4
	// VirtualDev represent virtual device
	VirtualDev = "VIRTUAL"

	// VNpuCfgKey is the key for virtual NPU configMap record
	VNpuCfgKey            = "VNPUCfg"
	virtualDevicesPattern = "Ascend910-(2|4|8|16)c"
	virtual710DevsPattern = "Ascend710-(1|2|4)c"
)

var (
	// NodeName is node name variable
	NodeName string
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

	if len(idSplit) < PhyDeviceLen {
		return "", "", fmt.Errorf("id: %s is invalid", deviceName)
	}
	var virID string
	deviceID := idSplit[len(idSplit)-1]
	// for virtual device, index 2 data means it's id
	if ascendRuntimeOptions == VirtualDev && len(idSplit) == VirDeviceLen {
		virID = idSplit[PhyDeviceLen]
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
	return utils.K8sClientFor(kubeConfig, component)
}

// GetNodeNameFromEnv get node name from env
func GetNodeNameFromEnv() error {
	nodeName := os.Getenv("NODE_NAME")
	if err := CheckNodeName(nodeName); err != nil {
		return fmt.Errorf("check node name failed: %v", err)
	}
	NodeName = nodeName
	return nil
}

// IsVirtualDev used to judge whether a physical device or a virtual device
func IsVirtualDev(devType string) bool {
	reg910 := regexp.MustCompile(virtualDevicesPattern)
	reg710 := regexp.MustCompile(virtual710DevsPattern)
	return reg910.MatchString(devType) || reg710.MatchString(devType)
}
