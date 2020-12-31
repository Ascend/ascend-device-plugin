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
	"time"

	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

// switch error log
var logFlag = true

// HwAscend910Manager manages huawei Ascend910 devices.
type HwAscend910Manager struct {
	dmgr DeviceMgrInterface
}

// NewHwAscend910Manager is used to create ascend 910 manager
func NewHwAscend910Manager() *HwAscend910Manager {
	return &HwAscend910Manager{}
}

// GetNPUs function discovers all HUAWEI Ascend910 devices available
// on the local node by calling walking `/dev` directory.
func (hnm *HwAscend910Manager) GetNPUs(allDevices *[]npuDevice, allDeviceTypes *[]string) error {
	var ids [hiAIMaxDeviceNum]uint32

	devNum, err := hnm.dmgr.GetDeviceList(&ids)
	if err != nil {
		return err
	}
	for i := int32(0); i < devNum; i++ {
		devID := fmt.Sprintf("%s-%d", hiAIAscend910Prefix, ids[i])
		phyID, err := hnm.dmgr.GetPhyID(ids[i])
		if err != nil {
			return err
		}
		logger.Info("Found Huawei Ascend910:", zap.String("logicID", devID), zap.Uint32("phyID", phyID))
		device := npuDevice{
			devType: hiAIAscend910Prefix,
			pciID:   "",
			ID:      devID,
			Health:  pluginapi.Healthy,
		}
		*allDevices = append(*allDevices, device)
	}
	*allDeviceTypes = append(*allDeviceTypes, hiAIAscend910Prefix)

	return nil
}

// GetDevState get device state
func (hnm *HwAscend910Manager) GetDevState(DeviceName string) string {
	var logicID int32

	err := getLogicIDByName(DeviceName, &logicID)
	if err != nil {
		if logFlag {
			logger.Error("get device logicID failed.",
				zap.String("deviceId", DeviceName),
				zap.String("error", err.Error()))
		}
		return pluginapi.Unhealthy
	}
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
		unhealthyState(healthState, uint32(logicID), "healthState", hnm.dmgr)
		return pluginapi.Unhealthy
	}
	return pluginapi.Healthy

}

// GetDefaultDevs Discovers Huawei Ascend910 devices and sets up device access environment.
func (hnm *HwAscend910Manager) GetDefaultDevs(defaultDeivces *[]string) error {
	return getDefaultDevices(defaultDeivces)
}

// GetDevPath get dev path
func (hnm *HwAscend910Manager) GetDevPath(id string, hostPath *string, containerPath *string) {
	*hostPath = fmt.Sprintf("%s%s", "/dev/davinci", id)
	*containerPath = *hostPath
}

// GetLogPath get log path
func (hnm *HwAscend910Manager) GetLogPath(devID []string, defaultLogPath string, newLogPath *string) error {

	*newLogPath = defaultLogPath
	var subdir = "/device"
	for _, item := range devID {
		var major string
		var minor string
		if err := getAscendDeviceID(item, &major, &minor); err != nil {
			logger.Error("dev ID is invalid", zap.String("deviceID", item))
			return fmt.Errorf("dev ID %s is invalid", item)
		}
		subdir += fmt.Sprintf("-%s", major)
	}
	*newLogPath += subdir
	t := time.Now()
	*newLogPath += t.UTC().Format("_2006-01-02-15-04-05.999")
	if _, err := os.Stat(*newLogPath); os.IsNotExist(err) {
		if err := os.MkdirAll(*newLogPath, os.ModePerm); err != nil {
			logger.Error("create directory %s failed.",
				zap.String("path", *newLogPath),
				zap.String("err", err.Error()))
			return fmt.Errorf("create directory %s failed: %s", *newLogPath, err)
		}
	}
	logger.Info("log dir is:", zap.String("logDir", *newLogPath))
	return nil
}

// SetDmgr to set dmgr
func (hnm *HwAscend910Manager) SetDmgr(dmgr DeviceMgrInterface) {
	hnm.dmgr = dmgr
}
