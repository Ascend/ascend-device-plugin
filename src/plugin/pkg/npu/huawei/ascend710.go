/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2020-2021. All rights reserved.
 */

// Package huawei implements the query and allocation of the device and the function of the log.
package huawei

import "C"

// HwAscend710Manager manages huawei Ascend710 devices.
type HwAscend710Manager struct {
	ascendCommonFunction
}

// NewHwAscend710Manager used to create ascend 710 manager
func NewHwAscend710Manager() *HwAscend710Manager {
	return &HwAscend710Manager{}
}
