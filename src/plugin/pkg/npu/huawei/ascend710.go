/*
* Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.
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

// HwAscend710Manager manages huawei Ascend710 devices.
type HwAscend710Manager struct {
	ascendCommonFunction
}

// NewHwAscend710Manager used to create ascend 710 manager
func NewHwAscend710Manager() *HwAscend710Manager {
	return &HwAscend710Manager{}
}