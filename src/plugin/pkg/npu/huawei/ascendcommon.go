/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2019-2024. All rights reserved.
 * Description: ascend910.go
 * Create: 19-11-20 下午8:52
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

func getDefaultDevices(defaultDeivces *[]string) error {
	if _, err := os.Stat(hiAIHDCDevice); err != nil {
		return err
	}

	if _, err := os.Stat(hiAIManagerDevice); err != nil {
		return err
	}

	*defaultDeivces = append(*defaultDeivces, hiAIHDCDevice, hiAIManagerDevice)

	if _, err := os.Stat(hiAISVMDevice); err == nil {
		*defaultDeivces = append(*defaultDeivces, hiAISVMDevice)
	}

	return nil
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

func unhealthyState(healthyState uint32, logicID uint32, healthyType string) {
	phyID, err := getPhyID(logicID)
	if err != nil {
		logger.Error("get phyID failed", zap.String("err", err.Error()))
	}
	// if logFlag is true,print device error message
	if logFlag {
		logger.Error("device is unHealthy.",
			zap.Uint32("logicID", logicID),
			zap.Uint32("phyID", phyID),
			zap.Uint32(healthyType, healthyState))
	}
}
