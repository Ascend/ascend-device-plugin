/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2020-2021. All rights reserved.
 */

// Package huawei implements the query and allocation of the device and the function of the log.
package huawei

// HwAscend310Manager manages huawei Ascend310 devices.
type HwAscend310Manager struct {
	ascendCommonFunction
}

// NewHwAscend310Manager used to create ascend 310 manager
func NewHwAscend310Manager() *HwAscend310Manager {
	name := hiAIAscend310Prefix
	if GetFdFlag {
		name = hiAIAscendfdPrefix
	}
	return &HwAscend310Manager{
		ascendCommonFunction: ascendCommonFunction{
			name:         name,
			unHealthyKey: huaweiUnHealthAscend310},
	}
}
