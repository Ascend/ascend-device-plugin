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

import (
	"fmt"
	"go.uber.org/zap"
	"os"
	"strconv"
	"strings"
)

func getDefaultDevices(defaultDevices *[]string) error {
	// hiAIManagerDevice is required
	if _, err := os.Stat(hiAIManagerDevice); err != nil {
		return err
	}
	*defaultDevices = append(*defaultDevices, hiAIManagerDevice)

	setDeviceByPath(defaultDevices, hiAIHDCDevice)
	setDeviceByPath(defaultDevices, hiAISVMDevice)
	if GetFdFlag {
		setDeviceByPathWhen200RC(defaultDevices)
	}
	return nil
}

func setDeviceByPathWhen200RC(defaultDevices *[]string) {
	setDeviceByPath(defaultDevices, hiAi200RCEventSched)
	setDeviceByPath(defaultDevices, hiAi200RCHiDvpp)
	setDeviceByPath(defaultDevices, hiAi200RCLog)
	setDeviceByPath(defaultDevices, hiAi200RCMemoryBandwidth)
	setDeviceByPath(defaultDevices, hiAi200RCSVM0)
	setDeviceByPath(defaultDevices, hiAi200RCTsAisle)
	setDeviceByPath(defaultDevices, hiAi200RCUpgrade)
}

func setDeviceByPath(defaultDevices *[]string, device string) {
	if _, err := os.Stat(device); err == nil {
		*defaultDevices = append(*defaultDevices, device)
	}
}

func getAscendDeviceID(id string, majorID *string, minorID *string) error {
	idSplit := strings.Split(id, "-")

	if len(idSplit) < idSplitNum {
		return fmt.Errorf("id: %s is invalid", id)
	}

	*majorID = idSplit[1]

	if len(idSplit) > idSplitNum {
		*minorID = idSplit[2]
		*majorID = *minorID
	}
	*minorID = ""

	return nil
}

func getLogicIDByName(DeviceName string, logicID *int32) error {
	var major string
	var minor string
	var phyID int32

	if err := getAscendDeviceID(DeviceName, &major, &minor); err != nil {
		logger.Error("dev ID is invalid", zap.String("deviceID", DeviceName))
		return err
	}

	devidCheck, err := strconv.Atoi(major)
	if err != nil {
		logger.Error("transfer device string to Integer failed", zap.String("deviceID", DeviceName))
		return err
	}
	phyID = int32(devidCheck)
	if phyID > hiAIMaxDeviceNum || phyID < 0 {
		logger.Error("GetDeviceState phyID overflow", zap.Int32("phyID", phyID))
		return fmt.Errorf("GetDevice phyid %d overflow", phyID)
	}

	*logicID = phyID

	return nil

}

func unhealthyState(healthyState uint32, logicID uint32, healthyType string, dmgr DeviceMgrInterface) error {
	phyID, err := dmgr.GetPhyID(logicID)
	if err != nil {
		return fmt.Errorf("get phyID failed %v", err)
	}
	// if logFlag is true,print device error message
	if logFlag {
		logger.Error("device is unHealthy.",
			zap.Uint32("logicID", logicID),
			zap.Uint32("phyID", phyID),
			zap.Uint32(healthyType, healthyState))
	}
	return nil
}

func getPhyIDFromDeviceID(deviceID string, dmgr DeviceMgrInterface) (string, error) {
	devidCheck, err := strconv.Atoi(deviceID)
	if err != nil {
		logger.Error("transfer device string to Integer failed", zap.String("deviceID", deviceID))
		return "", err
	}
	devID := uint32(devidCheck)
	phyID, err := dmgr.GetPhyID(devID)
	if err != nil {
		logger.Error("get PhyID failed", zap.String("deviceID", deviceID))
		return "", err
	}

	return strconv.Itoa(int(phyID)), nil
}
