/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2020-2021. All rights reserved.
 */

// Package huawei implements the query and allocation of the device and the function of the log.
package huawei

import "C"

// HwAscend310Manager manages huawei Ascend310 devices.
type HwAscend310Manager struct {
	ascendCommonFunction
}

// NewHwAscend310Manager used to create ascend 310 manager
func NewHwAscend310Manager() *HwAscend310Manager {
	var nam string
	nam = hiAIAscend310Prefix
	if GetFdFlag {
		nam = hiAIAscendfdPrefix
	}
	return &HwAscend310Manager{ascendCommonFunction{name: nam,
		unHealthyKey: huaweiUnHealthAscend310}}
}
