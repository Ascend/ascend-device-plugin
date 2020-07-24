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

package huawei

// #cgo pkg-config: ascend_device_plugin
// #include "dsmi_common_interface.h"
import "C"
import (
	"fmt"
)

const (
	// ERROR return string error
	ERROR = "error"

	// RetError return error when the function failed
	retError = -1
	// UnRetError return error
	unretError = 100
)

// ChipInfo chip info
type ChipInfo struct {
	ChipType string
	ChipName string
	ChipVer  string
}

func enableContainerService() error {
	err := C.dsmi_enable_container_service()
	if err != 0 {
		return fmt.Errorf("enable container service faild , error code: %d", int32(err))
	}
	return nil
}

// get ascend910 device quantity
func getDeviceCount() (int32, error) {
	var count C.int

	err := C.dsmi_get_device_count(&count)
	if err != 0 {
		return retError, fmt.Errorf("get device quantity failed, error code: %d", int32(err))
	}
	return int32(count), nil
}

// device get list
func getDeviceList(devices *[hiAIMaxDeviceNum]uint32) (int32, error) {
	devNum, err := getDeviceCount()
	if err != nil {
		return devNum, err
	}

	var ids [hiAIMaxDeviceNum]C.int
	if err := C.dsmi_list_device(&ids[0], C.int(devNum)); err != 0 {
		return retError, fmt.Errorf("unable to get device list, return error: %d", int32(err))
	}
	// transfer device list
	var i int32
	for i = 0; i < devNum; i++ {
		(*devices)[i] = uint32(ids[i])
	}

	return devNum, nil
}

// get device health by id
func getDeviceHealth(logicID int32) (uint32, error) {
	var health C.uint

	err := C.dsmi_get_device_health(C.int(logicID), &health)
	if err != 0 {
		return unretError, fmt.Errorf("get device %d health state failed, error code: %d", logicID, int32(err))
	}

	return uint32(health), nil

}

// get physic id form logic id
func getPhyID(logicID uint32) (uint32, error) {
	var phyID C.uint

	err := C.dsmi_get_phyid_from_logicid(C.uint(logicID), &phyID)
	if err != 0 {
		return unretError, fmt.Errorf("get phy id failed ,error code is: %d", int32(err))
	}

	return uint32(phyID), nil
}

// get logic id form physic id
func getLogicID(phyID uint32) (uint32, error) {
	var logicID C.uint

	err := C.dsmi_get_logicid_from_phyid(C.uint(phyID), &logicID)
	if err != 0 {
		return unretError, fmt.Errorf("get logic id failed ,error code is : %d", int32(err))
	}

	return uint32(logicID), nil

}

func getChipInfo(logicID int32) (*ChipInfo, error) {
	var chipInfo C.struct_dsmi_chip_info_stru
	err := C.dsmi_get_chip_info(C.int(logicID), &chipInfo)
	if err != 0 {
		return nil, fmt.Errorf("get device HBM information failed, error code: %d", int32(err))
	}
	var name []rune
	var ctype []rune
	var ver []rune
	for i, v := range chipInfo.chip_name {
		if v != 0 {
			name = append(name, rune(v))
		}
		c := chipInfo.chip_type[i]
		if c != 0 {
			ctype = append(ctype, rune(c))
		}

		r := chipInfo.chip_ver[i]
		if r != 0 {
			ver = append(ver, rune(r))
		}
	}

	chip := &ChipInfo{
		ChipName: string(name),
		ChipType: string(ctype),
		ChipVer:  string(ver),
	}
	return chip, nil
}
