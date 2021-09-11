/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2020-2021. All rights reserved.
 */

// Package huawei implements the query and allocation of the device and the function of the log.
package huawei

import (
	"fmt"
	"huawei.com/npu-exporter/hwlog"
	"k8s.io/apimachinery/pkg/util/sets"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"strings"
)

const (
	// ZeroCore is "0c"
	zeroCore = "0"
	// noVDevFound means not supported in the current scenario.
	noVDevFound = "65534"

	networkDetectOK   = uint32(0)
	networkDetectInit = uint32(6)
)

// HwAscend910Manager manages huawei Ascend910 devices.
type HwAscend910Manager struct {
	ascendCommonFunction
}

// NewHwAscend910Manager is used to create ascend 910 manager
func NewHwAscend910Manager() *HwAscend910Manager {
	return &HwAscend910Manager{}
}

// GetNPUs Discovers all HUAWEI Ascend910 devices by call dsmi interface
// a physical npu can be split into multiple vnpu
// vnpu is classification by computing power, like Ascend910-4c, Ascend910-8c, Ascend910-16c
// physical npu sets corresponding to the deviTypes, and vnpu is vDeviTypes
// vDeviTypes may is: [Ascend910-4c, Ascend910-4c, Ascend910-8c], also deviTypes may is: [Ascend910, Ascend910]
// one class deviType will generate a socket file, like ascend910-4c.sock or Ascend910.sock, so we deduplicate
func (hnm *HwAscend910Manager) GetNPUs(allDevices *[]npuDevice, allDeviceTypes *[]string, matchingDeviType string) error {
	var ids [hiAIMaxDeviceNum]uint32

	devNum, err := hnm.dmgr.GetDeviceList(&ids)
	if err != nil {
		return err
	}
	phyDevMapVirtualDev := make(map[uint32]string, devNum)
	var deviTypes, vDevID []string
	for i := int32(0); i < devNum; i++ {
		phyID, err := hnm.dmgr.GetPhyID(ids[i])
		if err != nil {
			return err
		}
		cgoDsmiVDevInfos, err := hnm.getVirtualDevice(ids[i])
		if err != nil && !strings.Contains(err.Error(), FunctionNotFound) {
			if !strings.Contains(err.Error(), noVDevFound) {
				hwlog.Errorf("Query virtual device info failure!, err: %s", err.Error())
				continue
			}
		}
		var devices []npuDevice
		if cgoDsmiVDevInfos.vDevNum == 0 {
			devices, deviTypes = hnm.assemblePhyDevices(phyID)
			phyDevMapVirtualDev[phyID] = fmt.Sprintf("%d", phyID)
		} else {
			devices, deviTypes, vDevID = hnm.assembleVirtualDevices(phyID, cgoDsmiVDevInfos)
			phyDevMapVirtualDev[phyID] = strings.Join(vDevID, ",")
		}
		*allDevices = append(*allDevices, devices...)
		*allDeviceTypes = append(*allDeviceTypes, deviTypes...)
	}
	hnm.phyDevMapVirtualDev = phyDevMapVirtualDev
	*allDeviceTypes = hnm.removeDuplicate(allDeviceTypes)
	return nil
}

func (hnm *HwAscend910Manager) removeDuplicate(allDeviceTypes *[]string) []string {
	deviceTypesMap := make(map[string]string, len(*allDeviceTypes))
	var rmDupDeviceTypes []string
	for _, deviType := range *allDeviceTypes {
		deviceTypesMap[deviType] = deviType
	}
	for _, deviType := range deviceTypesMap {
		rmDupDeviceTypes = append(rmDupDeviceTypes, deviType)
	}
	return rmDupDeviceTypes
}

func (hnm *HwAscend910Manager) assemblePhyDevices(phyID uint32) ([]npuDevice, []string) {
	var devices []npuDevice
	var deviTypes []string
	devID := fmt.Sprintf("%s-%d", hiAIAscend910Prefix, phyID)
	device := hnm.AssembleNpuDeviceStruct(hiAIAscend910Prefix, devID)
	devices = append(devices, device)
	deviTypes = append(deviTypes, hiAIAscend910Prefix)
	return devices, deviTypes
}

func (hnm *HwAscend910Manager) assembleVirtualDevices(phyID uint32, cgoDsmiVDevInfos CgoDsmiVDevInfo) (
	[]npuDevice, []string, []string) {
	var devices []npuDevice
	var vDeviTypes []string
	var vDevID []string
	for _, dsmiSubVDevInfo := range cgoDsmiVDevInfos.cgoDsmiSubVDevInfos {
		if dsmiSubVDevInfo.spec.coreNum == zeroCore {
			continue
		}
		vDeviType := fmt.Sprintf("%s-%sc", hiAIAscend910Prefix, dsmiSubVDevInfo.spec.coreNum)
		devID := fmt.Sprintf("%s-%sc-%d-%d", hiAIAscend910Prefix, dsmiSubVDevInfo.spec.coreNum, dsmiSubVDevInfo.vdevid, phyID)
		device := hnm.AssembleNpuDeviceStruct(vDeviType, devID)
		devices = append(devices, device)
		vDeviTypes = append(vDeviTypes, vDeviType)
		vDevID = append(vDevID, fmt.Sprintf("%d", dsmiSubVDevInfo.vdevid))
	}
	return devices, vDeviTypes, vDevID
}

func (hnm *HwAscend910Manager) getVirtualDevice(logicID uint32) (CgoDsmiVDevInfo, error) {
	cgoDsmiVDevInfos, err := hnm.dmgr.GetVDevicesInfo(logicID)
	if err != nil {
		return CgoDsmiVDevInfo{}, fmt.Errorf("query virtual device info failure: %s", err)
	}
	return cgoDsmiVDevInfos, nil
}

// DoWithVolcanoListAndWatch ascend910 affinity scheduling
func (hnm *HwAscend910Manager) DoWithVolcanoListAndWatch(hps *HwPluginServe, isStateChange bool) {
	hnm.groupDevsByStatus(hps, isStateChange)
	m.Lock()
	usedDevices := sets.NewString()
	getNodeNpuUsed(&usedDevices, hps)
	freeDevices := hps.healthDevice.Difference(usedDevices)
	totalDevices = totalDevices.Union(freeDevices)
	stateThreadNum += interval
	if stateThreadNum == len(hps.hdm.allDevTypes) {
		groupAllocatableDevs := groupDevByPower(totalDevices, hps.devType)
		if err := hps.kubeInteractor.patchAnnotationOnNode(groupAllocatableDevs, hiAIAscend910Prefix); err != nil {
			hwlog.Errorf("patch Annotation failed, err: %v", err)
		}
		totalDevices = totalDevices.Intersection(sets.String{})
		stateThreadNum = resetZero
	}
	m.Unlock()
}

// GetDeviceNetworkState check NPU network health
func (hnm *HwAscend910Manager) GetDeviceNetworkState(logicID int32, device *npuDevice) (string, error) {
	healthCode, err := hnm.dmgr.GetDeviceNetworkHealth(logicID)
	if err != nil {
		return "", err
	}

	switch healthCode {
	case networkDetectOK, networkDetectInit:
		return pluginapi.Healthy, nil
	default:
		hwlog.Warnf("%s network status is unhealthy, code value is %v", device.ID, healthCode)
		return pluginapi.Unhealthy, nil
	}
}

func (hnm *HwAscend910Manager) groupDevsByStatus(hps *HwPluginServe, isStateChange bool) {
	if !isStateChange {
		return
	}
	if hps.devType == hiAIAscend910Prefix && isStateChange {
		totalUHDevices = sets.String{}
		totalNetworkUnhealthDevices = sets.String{}
	}
	hps.healthDevice = sets.String{}
	for _, device := range hps.devices {
		if device.networkHealth != pluginapi.Healthy {
			totalNetworkUnhealthDevices.Insert(device.ID)
		}

		if IsVirtualDev(device.ID) || device.Health == pluginapi.Healthy {
			hps.healthDevice.Insert(device.ID)
			continue
		}

		if device.Health != pluginapi.Healthy {
			totalUHDevices.Insert(device.ID)
		}
	}
}

func groupDevByPower(allocatableDevices sets.String, devType string) map[string]string {
	var pwrSuffix = []string{hiAIAscend910Prefix, pwr2CSuffix, pwr4CSuffix, pwr8CSuffix, pwr16CSuffix}
	var groupAllocatableDevs = make(map[string]string, len(pwrSuffix))
	if devType == hiAIAscend310Prefix {
		chipAnnotation := filterTagPowerDevice(allocatableDevices, hiAIAscend310Prefix)
		annotationTag := fmt.Sprintf("%s%s", resourceNamePrefix, hiAIAscend310Prefix)
		groupAllocatableDevs[annotationTag] = chipAnnotation
		return groupAllocatableDevs
	}
	for _, suffix := range pwrSuffix {
		powerAnnotation := filterTagPowerDevice(allocatableDevices, suffix)
		annotationTag := fmt.Sprintf("%s%s", resourceNamePrefix, suffix)
		groupAllocatableDevs[annotationTag] = powerAnnotation
	}
	return groupAllocatableDevs
}

func filterTagPowerDevice(allocatableDevices sets.String, suffix string) string {
	var powerAnnotation []string
	for deviceName := range allocatableDevices {
		switch suffix {
		case hiAIAscend910Prefix:
			if !IsVirtualDev(deviceName) {
				powerAnnotation = append(powerAnnotation, deviceName)
			}
		default:
			if strings.Contains(deviceName, suffix) {
				powerAnnotation = append(powerAnnotation, deviceName)
			}
		}
	}
	return strings.Join(powerAnnotation, ",")
}
