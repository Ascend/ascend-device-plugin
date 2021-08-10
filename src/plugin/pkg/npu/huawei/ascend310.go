/*
* Copyright(C) 2020. Huawei Technologies Co.,Ltd. All rights reserved.
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
* http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
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
	if err := hps.kubeInteractor.patchAnnotationOnNode(groupAllocatableDevs, hps.devType); err != nil {
		hwlog.Errorf("Ascend310 patch Annotation failed, err: %v", err)
	}
}
