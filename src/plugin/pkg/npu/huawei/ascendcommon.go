/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2020-2021. All rights reserved.
 */

// Package huawei implements the query and allocation of the device and the function of the log.
package huawei

import (
	"encoding/json"
	"fmt"
	"huawei.com/npu-exporter/hwlog"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
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

type antStu struct {
	// Metadata metadata
	Metadata `json:"metadata,omitempty"`
}

// Metadata patch metadata
type Metadata struct {
	// Annotation Annotaions
	Annotation map[string]string `json:"annotations,omitempty"`
}

// ascendCommonFunction struct definition
type ascendCommonFunction struct {
	dmgr                DeviceMgrInterface
	phyDevMapVirtualDev map[uint32]string
	name                string
	unHealthyKey        string
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
		hwlog.RunLog.Errorf("dev ID is invalid, deviceID: %s", DeviceName)
		return phyID, err
	}

	devidCheck, err := strconv.Atoi(deviceID)
	if err != nil {
		hwlog.RunLog.Errorf("transfer device string to Integer failed, deviceID: %s", DeviceName)
		return phyID, err
	}
	phyID = uint32(devidCheck)
	if phyID > hiAIMaxDeviceNum || phyID < 0 {
		hwlog.RunLog.Errorf("GetDeviceState phyID overflow, phyID: %d", phyID)
		return phyID, fmt.Errorf("GetDevice phyid %d overflow", phyID)
	}

	return phyID, nil
}

func getUnHealthDev(listenUHDev, annotationUHDev, labelsRecoverDev, device910 sets.String) (
	sets.String, string) {
	var newAscend910 []string
	if autoStowingDevs {
		for device := range device910 {
			newAscend910 = append(newAscend910, device)
		}
		return sets.String{}, strings.Join(newAscend910, ",")
	}
	addRecoverDev := annotationUHDev.Difference(listenUHDev)
	newLabelsRecoverDev := labelsRecoverDev.Union(addRecoverDev)
	newDevice910 := device910.Difference(newLabelsRecoverDev)
	for device := range newDevice910 {
		newAscend910 = append(newAscend910, device)
	}
	return newLabelsRecoverDev, strings.Join(newAscend910, ",")
}

// getNewNetworkRecoverDev
// return new devices to be restored and network unhealthy device in this times
func getNewNetworkRecoverDev(nnu, nnr sets.String) (sets.String, sets.String) {
	// nnu means node annotation network unhealthy devices
	// nnr means device's network is ok and to be restored

	// this time network unhealthy devices
	tud := totalNetworkUnhealthDevices
	// if there is no network unhealthy device and autoStowingDevs is true
	if autoStowingDevs {
		return sets.String{}, tud
	}

	// devices recovered between the last check and this check
	recoveredDevSets := lastTimeNetworkRecoverDevices.Difference(nnr)

	newNetworkRecoverDevSets := sets.String{}
	newNetworkRecoverDevSets = newNetworkRecoverDevSets.Union(nnu.Difference(tud))
	// remove the device that network is unhealthy in this times
	newNetworkRecoverDevSets = newNetworkRecoverDevSets.Difference(nnr.Intersection(tud))
	// remove the device that recovered
	newNetworkRecoverDevSets = newNetworkRecoverDevSets.Difference(recoveredDevSets)

	newNetworkUnhealthDevSets := nnu.Union(tud).Difference(recoveredDevSets)

	return newNetworkRecoverDevSets, newNetworkUnhealthDevSets
}

func unhealthyState(healthyState uint32, logicID uint32, healthyType string, dmgr DeviceMgrInterface) error {
	phyID, err := dmgr.GetPhyID(logicID)
	if err != nil {
		return fmt.Errorf("get phyID failed %v", err)
	}
	if errs := dmgr.GetDeviceErrorCode(logicID); errs != nil {
		return errs
	}
	// if logFlag is true,print device error message
	if logFlag {
		hwlog.RunLog.Errorf("device is unHealthy, "+
			"logicID: %d, phyID: %d, %s: %d", logicID, phyID, healthyType, healthyState)
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
	if ascendRuntimeOptions == virtualDev && len(idSplit) == virDeviceLen {
		virID = idSplit[idSplitNum]
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
func VerifyPath(verifyPath string) (string, bool) {
	absVerifyPath, err := filepath.Abs(verifyPath)
	if err != nil {
		hwlog.RunLog.Errorf("abs current path failed")
		return "", false
	}
	pathInfo, err := os.Stat(absVerifyPath)
	if err != nil {
		hwlog.RunLog.Errorf("file path not exist")
		return "", false
	}
	realPath, err := filepath.EvalSymlinks(absVerifyPath)
	if err != nil || absVerifyPath != realPath {
		hwlog.RunLog.Errorf("Symlinks is not allowed")
		return "", false
	}
	stat, ok := pathInfo.Sys().(*syscall.Stat_t)
	if !ok || stat.Uid != rootUID || stat.Gid != rootGID {
		hwlog.RunLog.Errorf("Non-root owner group of the path")
		return "", false
	}
	return realPath, true
}

// AssembleNpuDeviceStruct is used to create a struct of npuDevice
func (adc *ascendCommonFunction) AssembleNpuDeviceStruct(deviType, devID string) npuDevice {
	hwlog.RunLog.Infof("Found Huawei Ascend, deviceType: %s, deviceID: %s", deviType, devID)
	return npuDevice{
		devType:       deviType,
		pciID:         "",
		ID:            devID,
		Health:        pluginapi.Healthy,
		networkHealth: pluginapi.Healthy,
	}
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

// GetDevState get device state
func (adc *ascendCommonFunction) GetDevState(DeviceName string, dmgr DeviceMgrInterface) string {
	phyID, err := getPhyIDByName(DeviceName)
	if err != nil {
		if logFlag {
			hwlog.RunLog.Errorf("get device phyID failed, deviceId: %s, err: %s", DeviceName, err.Error())
		}
		return pluginapi.Unhealthy
	}

	logicID, err := dmgr.GetLogicID(phyID)
	if err != nil {
		if logFlag {
			hwlog.RunLog.Errorf("get device logicID failed, deviceId: %s, err: %s", DeviceName, err.Error())
		}
		return pluginapi.Unhealthy
	}

	healthState, err := dmgr.GetDeviceHealth(int32(logicID))
	if err != nil {
		if logFlag {
			hwlog.RunLog.Errorf("get device healthy state failed, deviceId: %d, err: %s", int32(logicID), err.Error())
		}
		return pluginapi.Unhealthy
	}
	switch healthState {
	case normalState, generalAlarm:
		return pluginapi.Healthy
	default:
		err = unhealthyState(healthState, logicID, "healthState", dmgr)
		if err != nil {
			hwlog.RunLog.Errorf("unhealthyState, err: %v", err)
		}
		return pluginapi.Unhealthy
	}
}

// GetNPUs Discovers all HUAWEI Ascend310/Ascend710 devices by call dsmi interface
func (adc *ascendCommonFunction) GetNPUs(allDevices *[]npuDevice, allDeviceTypes *[]string, deviType string) error {
	hwlog.RunLog.Infof("--->< deviType: %s", deviType)

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
	return adc.name
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

// DoWithVolcanoListAndWatch ascend710 do nothing
func (adc *ascendCommonFunction) DoWithVolcanoListAndWatch(hps *HwPluginServe, isStateChange bool) {
	adc.reloadHealthDevice(isStateChange, hps)
	usedDevices := sets.NewString()
	getNodeNpuUsed(&usedDevices, hps)
	freeDevices := hps.healthDevice.Difference(usedDevices)
	annoMap := adc.GetAnnotationMap(freeDevices, hps.devType)
	annoMap[adc.unHealthyKey] = filterTagPowerDevice(hps.unHealthDevice, adc.name)
	if err := hps.kubeInteractor.patchNode(func(_ *v1.Node) []byte {
		as := antStu{Metadata{Annotation: annoMap}}
		bt, err := json.Marshal(as)
		if err != nil {
			hwlog.RunLog.Warnf("patch node error, %v", err)
			return nil
		}
		return bt
	}); err != nil {
		hwlog.RunLog.Errorf("%s patch Annotation failed, err: %v", adc.name, err)
	}
	return
}

// GetDeviceNetworkState check Ascend910 only
func (adc *ascendCommonFunction) GetDeviceNetworkState(_ int32, _ *npuDevice) (string, error) {
	return "", nil
}

func (adc *ascendCommonFunction) reloadHealthDevice(isStateChange bool, hps *HwPluginServe) {
	if !isStateChange {
		return
	}
	hps.healthDevice = sets.String{}
	hps.unHealthDevice = sets.String{}
	for _, device := range hps.devices {
		if device.Health == pluginapi.Healthy {
			hps.healthDevice.Insert(device.ID)
			continue
		}
		hps.unHealthDevice.Insert(device.ID)
	}
}

// GetAnnotationMap Get AnnotationMap
func (adc *ascendCommonFunction) GetAnnotationMap(allocatableDevices sets.String, _ string) map[string]string {
	var antMap = make(map[string]string, initMapCap)
	chipAnnotation := filterTagPowerDevice(allocatableDevices, adc.name)
	annotationTag := fmt.Sprintf("%s%s", resourceNamePrefix, adc.name)
	antMap[annotationTag] = chipAnnotation
	return antMap
}
