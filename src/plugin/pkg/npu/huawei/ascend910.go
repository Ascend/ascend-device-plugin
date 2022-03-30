/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2020-2022. All rights reserved.
 */

// Package huawei implements the query and allocation of the device and the function of the log.
package huawei

import (
	"fmt"
	"strings"

	"huawei.com/npu-exporter/hwlog"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"Ascend-device-plugin/src/plugin/pkg/npu/common"
)

const (
	// ZeroCore is "0c"
	zeroCore = "0"
	// noVDevFound means not supported in the current scenario.
	noVDevFound = "8255"

	networkDetectOK   = uint32(0)
	networkDetectInit = uint32(6)
)

const (
	virtualDevicesPattern = "Ascend910-(2|4|8|16)c"
	pwr2CSuffix           = "Ascend910-2c"
	pwr4CSuffix           = "Ascend910-4c"
	pwr8CSuffix           = "Ascend910-8c"
	pwr16CSuffix          = "Ascend910-16c"
)

var (
	// Dev910PhyCoreCount like huawei.com/Ascend910-spec:1-32c,2-30c
	Dev910PhyCoreCount []string
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
func (hnm *HwAscend910Manager) GetNPUs(allDevices *[]common.NpuDevice, allDeviceTypes *[]string, _ string) error {
	var ids [hiAIMaxDeviceNum]uint32

	devNum, err := hnm.dmgr.GetDeviceList(&ids)
	if err != nil {
		return err
	}
	phyDevMapVirtualDev := make(map[uint32]string, devNum)
	var deviTypes, vDevID, dev910PhyCoreCount []string
	for i := int32(0); i < devNum; i++ {
		phyID, err := hnm.dmgr.GetPhyID(ids[i])
		if err != nil {
			return err
		}
		cgoDsmiVDevInfos, err := hnm.getVirtualDevice(ids[i])
		if err != nil && !strings.Contains(err.Error(), FunctionNotFound) {
			if !strings.Contains(err.Error(), noVDevFound) {
				hwlog.RunLog.Errorf("Query virtual device info failure!, err: %s", err.Error())
				continue
			}
		}
		var devices []common.NpuDevice
		if cgoDsmiVDevInfos.VDevNum == 0 {
			devices, deviTypes = hnm.assemblePhyDevices(phyID, hiAIAscend910Prefix)
			phyDevMapVirtualDev[phyID] = fmt.Sprintf("%d", phyID)
		} else {
			devices, deviTypes, vDevID = hnm.assembleVirtualDevices(phyID, cgoDsmiVDevInfos, hiAIAscend910Prefix)
			phyDevMapVirtualDev[phyID] = strings.Join(vDevID, ",")
		}
		*allDevices = append(*allDevices, devices...)
		*allDeviceTypes = append(*allDeviceTypes, deviTypes...)
		dev910PhyCoreCount = append(dev910PhyCoreCount, fmt.Sprintf("%d-%dc-%dc",
			phyID, cgoDsmiVDevInfos.CoreCount, cgoDsmiVDevInfos.CoreNumUnused))
	}
	Dev910PhyCoreCount = dev910PhyCoreCount
	hnm.phyDevMapVirtualDev = phyDevMapVirtualDev
	*allDeviceTypes = hnm.removeDuplicate(allDeviceTypes)
	return nil
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
		groupAllocatableDevs := hnm.GetAnnotationMap(totalDevices, hps.devType)
		if err := hps.kubeInteractor.patchAnnotationOnNode(groupAllocatableDevs); err != nil {
			hwlog.RunLog.Errorf("patch Annotation failed, err: %v", err)
		}
		totalDevices = totalDevices.Intersection(sets.String{})
		stateThreadNum = 0
	}
	m.Unlock()
}

// GetDeviceNetworkState check NPU network health
func (hnm *HwAscend910Manager) GetDeviceNetworkState(logicID int32, device *common.NpuDevice) (string, error) {
	healthCode, err := hnm.dmgr.GetDeviceNetworkHealth(logicID)
	if err != nil {
		return "", err
	}

	switch healthCode {
	case networkDetectOK, networkDetectInit:
		return v1beta1.Healthy, nil
	default:
		hwlog.RunLog.Debugf("%s network status is unhealthy, code value is %v", device.ID, healthCode)
		return v1beta1.Unhealthy, nil
	}
}

func (hnm *HwAscend910Manager) groupDevsByStatus(hps *HwPluginServe, isStateChange bool) {
	if !isStateChange {
		return
	}
	if hps.devType == hiAIAscend910Prefix {
		totalUHDevices = sets.String{}
		totalNetworkUnhealthDevices = sets.String{}
	}
	hps.healthDevice = sets.String{}
	for _, device := range hps.devices {
		if device.NetworkHealth != v1beta1.Healthy {
			totalNetworkUnhealthDevices.Insert(device.ID)
		}

		if IsVirtualDev(device.ID) || device.Health == v1beta1.Healthy {
			hps.healthDevice.Insert(device.ID)
			continue
		}

		if device.Health != v1beta1.Healthy {
			totalUHDevices.Insert(device.ID)
		}
	}
}

// GetAnnotationMap Get Annonation
func (hnm *HwAscend910Manager) GetAnnotationMap(allocatableDevices sets.String, _ string) map[string]string {
	var pwrSuffix = []string{hiAIAscend910Prefix, pwr2CSuffix, pwr4CSuffix, pwr8CSuffix, pwr16CSuffix}
	var annoMap = make(map[string]string, len(pwrSuffix))
	for _, suffix := range pwrSuffix {
		powerAnnotation := filterTagPowerDevice(allocatableDevices, suffix)
		annotationTag := fmt.Sprintf("%s%s", resourceNamePrefix, suffix)
		annoMap[annotationTag] = powerAnnotation
	}
	return annoMap
}

func filterTagPowerDevice(allocatableDevices sets.String, suffix string) string {
	var powerAnnotation []string
	for deviceName := range allocatableDevices {
		switch suffix {
		case hiAIAscend910Prefix, hiAIAscend710Prefix:
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
