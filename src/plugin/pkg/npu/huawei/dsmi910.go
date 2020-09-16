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
