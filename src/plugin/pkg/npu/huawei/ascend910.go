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
	"time"

	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"
)

// switch error log
var logFlag = true

// HwAscend910Manager manages huawei Ascend910 devices.
type HwAscend910Manager struct {
	timeInterval  string
	checkNum      string
	restoreNum    string
	highThreshold string
	lowThreshold  string
	netDetect     bool
}

// NewHwAscend910Manager is used to create ascend 910 manager
func NewHwAscend910Manager(timeInterval, checkNum, restoreNum, highThreshold, lowThreshold string,
	netDetect bool) *HwAscend910Manager {
	return &HwAscend910Manager{
		timeInterval:  timeInterval,
		checkNum:      checkNum,
		restoreNum:    restoreNum,
		highThreshold: highThreshold,
		lowThreshold:  lowThreshold,
		netDetect:     netDetect,
	}
}

// GetNPUs function discovers all HUAWEI Ascend910 devices available
// on the local node by calling walking `/dev` directory.
func (hnm *HwAscend910Manager) GetNPUs(allDevices *[]npuDevice, allDeviceTypes *[]string) error {
	errs := enableContainerService()
	if errs != nil {
		logger.Error("enable containner Service failed. error", zap.String("error", errs.Error()))
	}
	var ids [hiAIMaxDeviceNum]uint32

	devNum, err := getDeviceList(&ids)
	if err != nil {
		return err
	}
	for i := int32(0); i < devNum; i++ {
		devID := fmt.Sprintf("%s-%d", hiAIAscend910Prefix, ids[i])
		phyID, err := getPhyID(uint32(ids[i]))
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
		logger.Error("get device logicID failed.",
			zap.String("deviceId", DeviceName),
			zap.String("error", err.Error()))
		return pluginapi.Unhealthy
	}

	healthState, err := getDeviceHealth(logicID)
	if err != nil {
		logger.Error("get device healthy state failed.",
			zap.Int32("deviceId", logicID),
			zap.String("error", err.Error()))
		return pluginapi.Unhealthy
	}
	if healthState != 0 {
		unhealthyState(healthState, uint32(logicID), "healthState")
		return pluginapi.Unhealthy
	}

	/*
		if hnm.netDetect {
			netHealthState, err := getDeviceNetworkHealth(logicID)
			if err != nil {
				logger.Error("get device %d network state failed.",
					zap.Int32("deviceId", logicID),
					zap.String("error", err.Error()))

				return pluginapi.Unhealthy
			}
			if netHealthState != 0 {
				unhealthyState(netHealthState, uint32(logicID), "netHealthState")
				return pluginapi.Unhealthy
			}
		}
	*/

	return pluginapi.Healthy

}

// GetDefaultDevs Discovers Huawei Ascend910 devices and sets up device access environment.
func (hnm *HwAscend910Manager) GetDefaultDevs(defaultDeivces *[]string) error {
	return getDefaultDevices(defaultDeivces)
}

// GetDevPath get dev path
func (hnm *HwAscend910Manager) GetDevPath(id string, hostPath *string, containerPath *string) error {
	var majorID string
	var minorID string

	if err := getAscendDeviceID(id, &majorID, &minorID); err != nil {
		return fmt.Errorf("cannot get device exact id from input id string: %s", id)
	}

	*hostPath = fmt.Sprintf("%s%s", "/dev/davinci", majorID)
	*containerPath = *hostPath
	return nil
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
