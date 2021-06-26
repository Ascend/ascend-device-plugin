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
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
	"syscall"
)

const (
	// VirtualDev represent virtual device
	virtualDev = "VIRTUAL"

	// PhysicalDev represent physical device
	physicalDev = ""

	// NormalState health state
	normalState = uint32(0)

	// GeneralAlarm health state
	generalAlarm = uint32(1)

	// Default device ip
	defaultDeviceIP = "127.0.0.1"

	// rootUID and rootGID is user group
	rootUID = 0
	rootGID = 0
)

// ascendCommonFunction struct definition
type ascendCommonFunction struct {
	dmgr                DeviceMgrInterface
	phyDevMapVirtualDev map[uint32]string
}

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

func getPhyIDByName(DeviceName string) (uint32, error) {
	var phyID uint32

	deviceID, _, err := getDeviceID(DeviceName, physicalDev)
	if err != nil {
		logger.Error("dev ID is invalid", zap.String("deviceID", DeviceName))
		return phyID, err
	}

	devidCheck, err := strconv.Atoi(deviceID)
	if err != nil {
		logger.Error("transfer device string to Integer failed", zap.String("deviceID", DeviceName))
		return phyID, err
	}
	phyID = uint32(devidCheck)
	if phyID > hiAIMaxDeviceNum || phyID < 0 {
		logger.Error("GetDeviceState phyID overflow", zap.Uint32("phyID", phyID))
		return phyID, fmt.Errorf("GetDevice phyid %d overflow", phyID)
	}

	return phyID, nil
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

func getDeviceID(deviceName string, ascendRuntimeOptions string) (string, string, error) {

	// hiAIAscend310Prefix: davinci-mini
	// vnpu: davinci-coreNum-vid-devID
	// ascend310:  davinci-mini0

	idSplit := strings.Split(deviceName, "-")

	if len(idSplit) < idSplitNum {
		return "", "", fmt.Errorf("id: %s is invalid", deviceName)
	}
	var virID string
	deviceID := idSplit[len(idSplit)-1]
	if ascendRuntimeOptions == virtualDev {
		virID = idSplit[len(idSplit)-2]
	}
	return deviceID, virID, nil
}

// IsVirtualDev used to judge whether a physical device or a virtual device
func IsVirtualDev(devType string) bool {
	pattern := virtualDevicesPattern
	reg := regexp.MustCompile(pattern)
	return reg.MatchString(devType)
}

// VerifyPath used to verify the validity of the path
func VerifyPath(verifyPath string) bool {
	absVerifyPath, err := filepath.Abs(verifyPath)
	if err != nil {
		logger.Error("abs current path failed: ", zap.Error(err))
		return false
	}
	pathInfo, err := os.Stat(absVerifyPath)
	if err != nil || os.IsNotExist(err) {
		logger.Error("file path not exist: ", zap.Error(err))
		return false
	}
	realPath, err := filepath.EvalSymlinks(absVerifyPath)
	if err != nil || absVerifyPath != realPath {
		logger.Error("Symlinks is not allowed: ")
		return false
	}
	stat, ok:= pathInfo.Sys().(*syscall.Stat_t)
	if !ok || stat.Uid != rootUID || stat.Gid != rootGID {
		logger.Error("Non-root owner group of the path")
		return false
	}
	return true
}

// AssembleNpuDeviceStruct is used to create a struct of npuDevice
func (adc *ascendCommonFunction) AssembleNpuDeviceStruct(deviType, devID string) npuDevice {
	logger.Info("Found Huawei Ascend:", zap.String("deviType", deviType), zap.String("deviceID", devID))
	return npuDevice{
		devType: deviType,
		pciID:   "",
		ID:      devID,
		Health:  pluginapi.Healthy,
	}
}

// CreateLogDirectory is used to create log complete path
func (adc *ascendCommonFunction) CreateLogDirectory(newLogPath *string, subdir string) error {
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
	return nil
}

// CreateLogSubDir is used to create log sub path
func (adc *ascendCommonFunction) CreateLogSubDir(devID []string, ascendRuntimeOptions string) (string, error) {
	var subdir = "/device"
	for _, item := range devID {
		deviceID, _, err := getDeviceID(item, ascendRuntimeOptions)
		if err != nil {
			logger.Error("dev ID is invalid", zap.String("deviceID", item))
			return subdir, fmt.Errorf("dev ID %s is invalid", item)
		}
		subdir += fmt.Sprintf("-%s", deviceID)
	}
	return subdir, nil
}

// GetDevPath is used to get device path
func (adc *ascendCommonFunction) GetDevPath(id, ascendRuntimeOptions string) (string, string) {
	containerPath := fmt.Sprintf("%s%s", "/dev/davinci", id)
	hostPath := containerPath
	if ascendRuntimeOptions == virtualDev {
		hostPath = fmt.Sprintf("%s%s", "/dev/vdavinci", id)
	}
	return containerPath, hostPath
}

// GetLogPath is used to get log path
func (adc *ascendCommonFunction) GetLogPath(devID []string, defaultLogPath, ascendRuntimeOptions string, newLogPath *string) error {
	subdir, err := adc.CreateLogSubDir(devID, ascendRuntimeOptions)
	if err != nil {
		return err
	}
	err = adc.CreateLogDirectory(&defaultLogPath, subdir)
	if err != nil {
		return err
	}
	*newLogPath = defaultLogPath
	logger.Info("log dir is:", zap.String("logDir", *newLogPath))
	return nil
}

// GetDevState get device state
func (adc *ascendCommonFunction) GetDevState(DeviceName string, dmgr DeviceMgrInterface) string {
	phyID, err := getPhyIDByName(DeviceName)
	if err != nil {
		if logFlag {
			logger.Error("get device phyID failed.", zap.String("deviceId", DeviceName),
				zap.String("error", err.Error()))
		}
		return pluginapi.Unhealthy
	}

	logicID, err := dmgr.GetLogicID(phyID)
	if err != nil {
		if logFlag {
			logger.Error("get device logicID failed.", zap.String("deviceId", DeviceName),
				zap.String("error", err.Error()))
		}
		return pluginapi.Unhealthy
	}

	healthState, err := dmgr.GetDeviceHealth(int32(logicID))
	if err != nil {
		if logFlag {
			logger.Error("get device healthy state failed.", zap.Int32("deviceId", int32(logicID)),
				zap.String("error", err.Error()))
		}
		return pluginapi.Unhealthy
	}
	switch healthState {
	case normalState:
		return pluginapi.Healthy
	case generalAlarm:
		logger.Warn("device health state", zap.Uint32("healthState", healthState))
		return pluginapi.Healthy
	default:
		err = unhealthyState(healthState, logicID, "healthState", dmgr)
		if err != nil {
			logger.Error("unhealthyState ", zap.Error(err))
		}
		return pluginapi.Unhealthy
	}
}

// GetNPUs Discovers all HUAWEI Ascend310/Ascend710 devices by call dsmi interface
func (adc *ascendCommonFunction) GetNPUs(allDevices *[]npuDevice, allDeviceTypes *[]string, deviType string) error {
	logger.Info("--->< ", zap.String("deviType", deviType))

	var ids [hiAIMaxDeviceNum]uint32
	devNum, getDevListErrInfo := adc.dmgr.GetDeviceList(&ids)
	if getDevListErrInfo != nil {
		return getDevListErrInfo
	}

	for i := int32(0); i < devNum; i++ {
		phyID, err := adc.dmgr.GetPhyID(ids[i])
		if err != nil {
			return err
		}
		dev := fmt.Sprintf("%s-%d", deviType, phyID)
		device := adc.AssembleNpuDeviceStruct(deviType, dev)
		*allDevices = append(*allDevices, device)
	}
	*allDeviceTypes = append(*allDeviceTypes, deviType)

	return nil
}

// GetMatchingDeviType to get match device type
func (adc *ascendCommonFunction) GetMatchingDeviType() string {
	return hiAIAscend710Prefix
}

// SetDmgr to set dmgr
func (adc *ascendCommonFunction) SetDmgr(dmgr DeviceMgrInterface) {
	adc.dmgr = dmgr
}

// GetDmgr to get dmgr
func (adc *ascendCommonFunction) GetDmgr() DeviceMgrInterface {
	return adc.dmgr
}

// GetPhyDevMapVirtualDev get phy devices and virtual devices mapping
func (adc *ascendCommonFunction) GetPhyDevMapVirtualDev() map[uint32]string {
	return adc.phyDevMapVirtualDev
}
