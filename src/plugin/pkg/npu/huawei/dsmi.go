/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2019-2024. All rights reserved.
 * Description: dsmi.go
 * Create: 19-11-20 下午8:52
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

// to be fix
func getDeviceDie(logicID int32, dieID *[dieIDNum]uint32) error {
	var deviceDie C.struct_dsmi_soc_die_stru

	err := C.dsmi_get_device_die(C.int(logicID), &deviceDie)
	if err != 0 {
		return fmt.Errorf("get logic id failed ,error code is : %d", int32(err))
	}

	for i := 0; i < dieIDNum; i++ {
		dieID[i] = uint32(deviceDie.soc_die[i])
	}
	return nil

}
