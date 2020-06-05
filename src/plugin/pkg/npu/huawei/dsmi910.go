/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2019-2024. All rights reserved.
 * Description: ascend910.go
 * Create: 19-11-20 下午8:52
 */

// Package huawei implements the query and allocation of the device and the function of the log.
package huawei

// #cgo pkg-config: ascend_device_plugin
// #include "dsmi_common_interface.h"
import "C"
import (
	"fmt"
	"unsafe"
)

// to be fix  current just support ipv4
func getDeviceIP(logicID int32) (string, error) {

	var portType C.int = 1
	var portID C.int
	var ipAddress [hiAIMaxDeviceNum]C.ip_addr_t
	var maskAddress [hiAIMaxDeviceNum]C.ip_addr_t
	var retIPAddress string
	var ipString [4]uint8

	err := C.dsmi_get_device_ip_address(C.int(logicID), portType, portID, &ipAddress[C.int(logicID)],
		&maskAddress[C.int(logicID)])
	if err != 0 {
		return ERROR, fmt.Errorf("getDevice IP address failed, error code: %d", int32(err))
	}

	unionPara := ipAddress[C.int(logicID)].u_addr
	for i := 0; i < len(ipString); i++ {
		ipString[i] = uint8(*(*C.uchar)(unsafe.Pointer(&unionPara[i])))
	}

	retIPAddress = fmt.Sprintf("%d.%d.%d.%d", ipString[0], ipString[1], ipString[2], ipString[3])
	return retIPAddress, nil

}

/*
func getDeviceNetworkHealth(logicID int32) (uint32, error) {
	var netHealth C.DSMI_NET_HEALTH_STATUS

	err := C.dsmi_get_network_health(C.int(logicID), &netHealth)
	if err != 0 {
		return unretError, fmt.Errorf("get device network health state failed, error code: %d", int32(err))
	}

	return uint32(netHealth), nil

}
*/
