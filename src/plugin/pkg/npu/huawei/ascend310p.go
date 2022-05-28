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
	chip310PCore1C = "Ascend310P-1c"
	chip310PCore2C = "Ascend310P-2c"
	chip310PCore4C = "Ascend310P-4c"
)

// HwAscend310PManager manages huawei Ascend310P devices.
type HwAscend310PManager struct {
	ascendCommonFunction
}

// NewHwAscend310PManager used to create ascend 310P manager
func NewHwAscend310PManager() *HwAscend310PManager {
	return &HwAscend310PManager{
		ascendCommonFunction: ascendCommonFunction{
			name:         hiAIAscend310PPrefix,
			unHealthyKey: huaweiUnHealthAscend310P,
		},
	}
}

// GetNPUs Discovers all HUAWEI Ascend310P devices by call dsmi interface
func (hnm *HwAscend310PManager) GetNPUs(allDevices *[]common.NpuDevice, allDeviceTypes *[]string,
	deviType string) error {
	hwlog.RunLog.Infof("--->< deviType: %s", deviType)
	var ids [hiAIMaxDeviceNum]uint32

	devNum, err := hnm.dmgr.GetDeviceList(&ids)
	if err != nil {
		return err
	}
	if devNum > hiAIMaxDeviceNum {
		return fmt.Errorf("invalid device num: %d", devNum)
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
				hwlog.RunLog.Errorf("Query virtual device info failure!, err: %s", err.Error())
				continue
			}
		}
		var devices []common.NpuDevice
		if cgoDsmiVDevInfos.VDevNum == 0 {
			devices, deviTypes = hnm.assemblePhyDevices(phyID, hiAIAscend310PPrefix)
			phyDevMapVirtualDev[phyID] = fmt.Sprintf("%d", phyID)
		} else {
			devices, deviTypes, vDevID = hnm.assembleVirtualDevices(phyID, cgoDsmiVDevInfos, hiAIAscend310PPrefix)
			phyDevMapVirtualDev[phyID] = strings.Join(vDevID, ",")
		}
		*allDevices = append(*allDevices, devices...)
		*allDeviceTypes = append(*allDeviceTypes, deviTypes...)
	}
	hnm.phyDevMapVirtualDev = phyDevMapVirtualDev
	*allDeviceTypes = hnm.removeDuplicate(allDeviceTypes)
	return nil
}

// DoWithVolcanoListAndWatch ascend310P affinity scheduling
func (hnm *HwAscend310PManager) DoWithVolcanoListAndWatch(hps *HwPluginServe) {
	hnm.groupDevsByStatus(hps)
	usedDevices := sets.NewString()
	getNodeNpuUsed(&usedDevices, hps)
	freeDevices := hps.healthDevice.Difference(usedDevices)
	totalDevices = totalDevices.Union(freeDevices)
	if stateThreadNum == len(hps.hdm.allDevTypes) {
		groupAllocatableDevs := hnm.GetAnnotationMap(totalDevices, hps.hdm.allDevTypes)
		if err := hps.kubeInteractor.patchAnnotationOnNode(groupAllocatableDevs, false, hnm.isVirExist(hps),
			hps.devType); err != nil {
			hwlog.RunLog.Errorf("patch Annotation failed, err: %v", err)
		}
		totalDevices = totalDevices.Intersection(sets.String{})
		stateThreadNum = 0
	}
}

func (hnm *HwAscend310PManager) groupDevsByStatus(hps *HwPluginServe) {
	if hps.devType == hiAIAscend310PPrefix {
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
	hwlog.RunLog.Debugf("healthy device %v", hps.healthDevice)
	hwlog.RunLog.Debugf("total unhealthy devices %v", totalUHDevices)
}

// GetAnnotationMap Get Annonation
func (hnm *HwAscend310PManager) GetAnnotationMap(allocatableDevices sets.String, devTypes []string) map[string]string {
	var annoMap = make(map[string]string, len(devTypes))
	for _, suffix := range devTypes {
		powerAnnotation := filterTagPowerDevice(allocatableDevices, suffix)
		annotationTag := fmt.Sprintf("%s%s", resourceNamePrefix, suffix)
		annoMap[annotationTag] = powerAnnotation
	}
	return annoMap
}
