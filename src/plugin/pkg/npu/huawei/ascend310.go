/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2020-2021. All rights reserved.
 */

// Package huawei implements the query and allocation of the device and the function of the log.
package huawei

import "C"
import (
	"huawei.com/npu-exporter/hwlog"
	"k8s.io/apimachinery/pkg/util/sets"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

// HwAscend310Manager manages huawei Ascend310 devices.
type HwAscend310Manager struct {
	ascendCommonFunction
}

// NewHwAscend310Manager used to create ascend 310 manager
func NewHwAscend310Manager() *HwAscend310Manager {
	return &HwAscend310Manager{}
}

// GetMatchingDeviType to get match device type
func (hnm *HwAscend310Manager) GetMatchingDeviType() string {
	if GetFdFlag {
		return hiAIAscendfdPrefix
	}
	return hiAIAscend310Prefix
}

// DoWithVolcanoListAndWatch ascend310 affinity scheduling
func (hnm *HwAscend310Manager) DoWithVolcanoListAndWatch(hps *HwPluginServe, isStateChange bool) {
	hnm.reloadHealthDevice(isStateChange, hps)
	usedDevices := sets.NewString()
	getNodeNpuUsed(&usedDevices, hps)
	freeDevices := hps.healthDevice.Difference(usedDevices)
	groupAllocatableDevs := groupDevByPower(freeDevices, hps.devType)
	if err := hps.kubeInteractor.patchAnnotationOnNode(groupAllocatableDevs, hiAIAscend310Prefix); err != nil {
		hwlog.RunLog.Errorf("Ascend310 patch Annotation failed, err: %v", err)
	}
}

func (hnm *HwAscend310Manager) reloadHealthDevice(isStateChange bool, hps *HwPluginServe) {
	if !isStateChange {
		return
	}
	hps.healthDevice = sets.String{}
	for _, device := range hps.devices {
		if device.Health == pluginapi.Healthy {
			hps.healthDevice.Insert(device.ID)
		}
	}
}
