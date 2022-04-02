/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2020-2022. All rights reserved.
 */

// Package huawei implements the query and allocation of the device and the function of the log.
package huawei

import (
	"fmt"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"huawei.com/npu-exporter/hwlog"

	"Ascend-device-plugin/src/plugin/pkg/npu/common"
)

const (
	chip710Core1C         = "Ascend710-1c"
	chip710Core2C         = "Ascend710-2c"
	chip710Core4C         = "Ascend710-4c"
)

var (
	// Dev710PhyCoreCount like huawei.com/Ascend710-spec:1-8c,2-8c
	Dev710PhyCoreCount []string
)

// HwAscend710Manager manages huawei Ascend710 devices.
type HwAscend710Manager struct {
	ascendCommonFunction
}

// NewHwAscend710Manager used to create ascend 710 manager
func NewHwAscend710Manager() *HwAscend710Manager {
	return &HwAscend710Manager{
		ascendCommonFunction{
			name:         hiAIAscend710Prefix,
			unHealthyKey: huaweiUnHealthAscend710,
		},
	}
}

// GetNPUs Discovers all HUAWEI Ascend710 devices by call dsmi interface
func (hnm *HwAscend710Manager) GetNPUs(allDevices *[]common.NpuDevice, allDeviceTypes *[]string,
	deviType string) error {
	hwlog.RunLog.Infof("--->< deviType: %s", deviType)
	var ids [hiAIMaxDeviceNum]uint32

	devNum, err := hnm.dmgr.GetDeviceList(&ids)
	if err != nil {
		return err
	}
	phyDevMapVirtualDev := make(map[uint32]string, devNum)
	var deviTypes, vDevID, dev710PhyCoreCount []string
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
			devices, deviTypes = hnm.assemblePhyDevices(phyID, hiAIAscend710Prefix)
			phyDevMapVirtualDev[phyID] = fmt.Sprintf("%d", phyID)
		} else {
			devices, deviTypes, vDevID = hnm.assembleVirtualDevices(phyID, cgoDsmiVDevInfos, hiAIAscend710Prefix)
			phyDevMapVirtualDev[phyID] = strings.Join(vDevID, ",")
		}
		*allDevices = append(*allDevices, devices...)
		*allDeviceTypes = append(*allDeviceTypes, deviTypes...)
		dev710PhyCoreCount = append(dev710PhyCoreCount, fmt.Sprintf("%d-%dc-%dc",
			phyID, cgoDsmiVDevInfos.CoreCount, cgoDsmiVDevInfos.CoreNumUnused))
	}
	Dev710PhyCoreCount = dev710PhyCoreCount
	hnm.phyDevMapVirtualDev = phyDevMapVirtualDev
	*allDeviceTypes = hnm.removeDuplicate(allDeviceTypes)
	return nil
}

// DoWithVolcanoListAndWatch ascend710 affinity scheduling
func (hnm *HwAscend710Manager) DoWithVolcanoListAndWatch(hps *HwPluginServe, isStateChange bool) {
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

func (hnm *HwAscend710Manager) groupDevsByStatus(hps *HwPluginServe, isStateChange bool) {
	if !isStateChange {
		return
	}
	if hps.devType == hiAIAscend710Prefix {
		totalUHDevices = sets.String{}
	}
	hps.healthDevice = sets.String{}
	for _, device := range hps.devices {
		if common.IsVirtualDev(device.ID) || device.Health == v1beta1.Healthy {
			hps.healthDevice.Insert(device.ID)
		}
		if device.Health != v1beta1.Healthy {
			totalUHDevices.Insert(device.ID)
		}
	}
}

// GetAnnotationMap Get Annonation
func (hnm *HwAscend710Manager) GetAnnotationMap(allocatableDevices sets.String, _ string) map[string]string {
	var pwrSuffix = []string{hiAIAscend710Prefix, chip710Core1C, chip710Core2C, chip710Core4C}
	var annoMap = make(map[string]string, len(pwrSuffix))
	for _, suffix := range pwrSuffix {
		powerAnnotation := filterTagPowerDevice(allocatableDevices, suffix)
		annotationTag := fmt.Sprintf("%s%s", resourceNamePrefix, suffix)
		annoMap[annotationTag] = powerAnnotation
	}
	return annoMap
}
