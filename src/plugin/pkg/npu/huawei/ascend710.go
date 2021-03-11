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
import (
	"fmt"
	"go.uber.org/zap"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"strconv"
)

// HwAscend710Manager manages huawei Ascend710 devices.
type HwAscend710Manager struct {
	dmgr DeviceMgrInterface
}

// NewHwAscend710Manager used to create ascend 710 manager
func NewHwAscend710Manager() *HwAscend710Manager {
	return &HwAscend710Manager{}
}

// GetNPUs Discovers all HUAWEI Ascend710 devices available on the local node by calling walking `/dev` directory.
func (hnm *HwAscend710Manager) GetNPUs(allDevices *[]npuDevice, allDeviceTypes *[]string) error {
	var ids [hiAIMaxDeviceNum]uint32

	devNum, err := hnm.dmgr.GetDeviceList(&ids)
	if err != nil {
		return err
	}
	deviType := hnm.getMatchingDeviType()
	logger.Info("--->< ", zap.String("deviType", deviType))

	for i := int32(0); i < devNum; i++ {
		dev := fmt.Sprintf("%s-%d", deviType, ids[i])
		device := assembleNpuDeviceStruct(deviType, dev, placeholder)
		*allDevices = append(*allDevices, device)
	}
	*allDeviceTypes = append(*allDeviceTypes, deviType)

	return nil
}

func (hnm *HwAscend710Manager) getMatchingDeviType() string {
	return hiAIAscend710Prefix
}

// GetDevState is used to get device state
func (hnm *HwAscend710Manager) GetDevState(DeviceName string) string {
	majorID, err := getDeviceID(DeviceName)
	if err != nil {
		if logFlag {
			logger.Error("get device logicID failed.",
				zap.String("deviceId", DeviceName),
				zap.String("error", err.Error()))
		}
		return pluginapi.Unhealthy
	}
	devidCheck, err := strconv.Atoi(majorID)
	if err != nil {
		if logFlag {
			logger.Error("transfer device string to Integer failed", zap.String("deviceID", DeviceName))
		}
		return pluginapi.Unhealthy
	}
	logicID := int32(devidCheck)
	healthState, err := hnm.dmgr.GetDeviceHealth(logicID)
	if err != nil {
		if logFlag {
			logger.Error("get device healthy state failed.",
				zap.Int32("deviceId", logicID),
				zap.String("error", err.Error()))
		}
		return pluginapi.Unhealthy
	}
	if healthState != 0 {
		err = unhealthyState(healthState, uint32(logicID), "healthState", hnm.dmgr)
		if err != nil {
			logger.Error("unhealthyState ", zap.Error(err))
		}
		return pluginapi.Unhealthy
	}
	return pluginapi.Healthy
}

// GetDevPath is used to get device path
func (hnm *HwAscend710Manager) GetDevPath(id, ascendRuntimeOptions string, hostPath *string, containerPath *string) {
	*hostPath = fmt.Sprintf("%s%s", "/dev/davinci", id)
	*containerPath = *hostPath
}

// GetLogPath is used to get log path
func (hnm *HwAscend710Manager) GetLogPath(devID []string, defaultLogPath string, newLogPath *string) error {
	subdir, err := createLogSubDir(devID)
	if err != nil {
		return err
	}
	err = createLogDirectory(&defaultLogPath, subdir)
	if err != nil {
		return err
	}
	*newLogPath = defaultLogPath
	logger.Info("log dir is:", zap.String("logDir", *newLogPath))
	return nil
}

// SetDmgr to set dmgr
func (hnm *HwAscend710Manager) SetDmgr(dmgr DeviceMgrInterface) {
	hnm.dmgr = dmgr
}
