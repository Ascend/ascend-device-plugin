/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2020-2021. All rights reserved.
 */

// Package huawei implements the query and allocation of the device and the function of the log.
package huawei

import "C"
import (
	"huawei.com/npu-exporter/hwlog"
	"k8s.io/apimachinery/pkg/util/sets"
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
	usedDevices := sets.NewString()
	getNodeNpuUsed(&usedDevices, hps)
	freeDevices := hps.healthDevice.Difference(usedDevices)
	groupAllocatableDevs := groupDevByPower(freeDevices, hps.devType)
	if err := hps.kubeInteractor.patchAnnotationOnNode(groupAllocatableDevs, hiAIAscend310Prefix); err != nil {
		hwlog.Errorf("Ascend310 patch Annotation failed, err: %v", err)
	}
}
