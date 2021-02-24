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

import "C"
import (
	"fmt"
	"go.uber.org/zap"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

// HwAscend310Manager manages huawei Ascend310 devices.
type HwAscend310Manager struct {
	dmgr DeviceMgrInterface
}

// NewHwAscend310Manager used to create ascend 310 manager
func NewHwAscend310Manager() *HwAscend310Manager {
	return &HwAscend310Manager{}
}

// GetNPUs Discovers all HUAWEI Ascend310 devices available on the local node by calling walking `/dev` directory.
func (hnm *HwAscend310Manager) GetNPUs(allDevices *[]npuDevice, allDeviceTypes *[]string) error {
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

func (hnm *HwAscend310Manager) getMatchingDeviType() string {
	if GetFdFlag {
		return hiAIAscendfdPrefix
	}
	return hiAIAscend310Prefix
}

// GetDevState is used to get device state
func (hnm *HwAscend310Manager) GetDevState(DeviceName string) string {
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

// GetDefaultDevs Discovers Huawei Ascend310 devices and sets up device access environment.
func (hnm *HwAscend310Manager) GetDefaultDevs(defaultDeivces *[]string) error {
	return getDefaultDevices(defaultDeivces)
}

// GetDevPath is used to get device path
func (hnm *HwAscend310Manager) GetDevPath(id, ascendRuntimeOptions string, hostPath *string, containerPath *string) {
	*hostPath = fmt.Sprintf("%s%s", "/dev/davinci", id)
	*containerPath = *hostPath
}

// GetLogPath is used to get log path
func (hnm *HwAscend310Manager) GetLogPath(devID []string, defaultLogPath string, newLogPath *string) error {

	*newLogPath = defaultLogPath
	var subdir = "/device"
	for _, item := range devID {
		major, err := getDeviceID(item)
		if err != nil {
			logger.Error("getdevice", zap.String("devid", item))
			return fmt.Errorf("dev ID %s is invalid", item)
		}
		subdir += fmt.Sprintf("-%s", major)
	}
	*newLogPath += subdir
	t := time.Now()
	*newLogPath += t.UTC().Format("_2006-01-02-15-04-05.999")
	if _, err := os.Stat(*newLogPath); os.IsNotExist(err) {
		if err := os.MkdirAll(*newLogPath, os.ModePerm); err != nil {
			log.Printf("create directory %s failed: %s.\n", *newLogPath, err)
			return fmt.Errorf("create directory %s failed: %s", *newLogPath, err)
		}
	}
	log.Printf("log dir: %s.\n", *newLogPath)
	return nil
}

// ContainerAssignmentNotify is used to notify contain
func ContainerAssignmentNotify() error {

	return nil
}

func getDeviceID(deviceName string) (string, error) {

	// hiAIAscend310Prefix: davinci-mini
	// vnpu: davinci-mini0-0
	// ascend310:  davinci-mini0

	idSplit := strings.Split(deviceName, "-")

	if len(idSplit) < idSplitNum {
		return "", fmt.Errorf("deviceName: %s is invalid", deviceName)
	}

	majorID := idSplit[len(idSplit)-1]
	return majorID, nil
}

// SetDmgr to set dmgr
func (hnm *HwAscend310Manager) SetDmgr(dmgr DeviceMgrInterface) {
	hnm.dmgr = dmgr
}
