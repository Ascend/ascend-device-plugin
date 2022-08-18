/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2020-2022. All rights reserved.
 */

// Package device implements the query and allocation of the device and the function of the log.
package device

import (
	"fmt"
	"strings"

	"huawei.com/npu-exporter/hwlog"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"Ascend-device-plugin/pkg/common"
)

const (
	// ZeroCore is "0c"
	zeroCore = "0"
	// noVDevFound means not supported in the current scenario.
	noVDevFound = "8255"

	networkDetectOK   = uint32(0)
	networkDetectInit = uint32(6)
)

// HwAscend910Manager manages huawei Ascend910 devices.
type HwAscend910Manager struct {
	AscendTools
}

// NewHwAscend910Manager is used to create ascend 910 manager
func NewHwAscend910Manager() *HwAscend910Manager {
	return &HwAscend910Manager{
		AscendTools: AscendTools{
			name:         common.Ascend910,
			unHealthyKey: common.HuaweiUnHealthAscend910,
			devCount:     common.MaxDevicesNum,
		},
	}
}

// GetNPUs Discovers all HUAWEI Ascend910 devices by call devmanager interface
// a physical npu can be split into multiple vnpu
// vnpu is classification by computing power, like Ascend910-4c, Ascend910-8c, Ascend910-16c
// physical npu sets corresponding to the deviTypes, and vnpu is vDeviTypes
// vDeviTypes may is: [Ascend910-4c, Ascend910-4c, Ascend910-8c], also deviTypes may is: [Ascend910, Ascend910]
// one class deviType will generate a socket file, like ascend910-4c.sock or Ascend910.sock, so we deduplicate
func (hnm *HwAscend910Manager) GetNPUs(allDevices *[]common.NpuDevice, allDeviceTypes *[]string, _ string) error {
	devNum, devList, err := hnm.dmgr.GetDeviceList()
	if err != nil {
		return err
	}
	if devNum > hnm.devCount {
		return fmt.Errorf("invalid device num: %d", devNum)
	}
	for i := int32(0); i < devNum; i++ {
		davinCiDev, err := hnm.getDavinCiDev(devList[i], nil)
		if err != nil {
			return err
		}
		vDevInfos, err := hnm.getVirtualDevice(devList[i])
		if err != nil {
			hwlog.RunLog.Errorf("The virtual device is considered not exist, please check the error: %#v", err)
		}
		if vDevInfos.TotalResource.VDevNum == 0 {
			hnm.assemblePhyDevices(davinCiDev, allDevices, allDeviceTypes)
			continue
		}
		hnm.assembleVirtualDevices(davinCiDev, vDevInfos, allDevices, allDeviceTypes)
	}
	*allDeviceTypes = hnm.removeDuplicate(allDeviceTypes)
	return nil
}

// DoWithVolcanoListAndWatch ascend910 affinity scheduling
func (hnm *HwAscend910Manager) DoWithVolcanoListAndWatch(classifyDevs map[string][]*common.NpuDevice) {
	devStatusSet := hnm.getDevStatesDevSet(classifyDevs)
	if err := hnm.UpdateNodeDeviceInfo(devStatusSet, hnm.updateDeviceInfo); err != nil {
		hwlog.RunLog.Errorf("update device info failed, err: %#v", err)
	}
}

// GetDeviceNetworkState check NPU network health
func (hnm *HwAscend910Manager) GetDeviceNetworkState(logicID int32, device *common.NpuDevice) (string, error) {
	healthCode, err := hnm.dmgr.GetDeviceNetWorkHealth(logicID)
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

func (hnm *HwAscend910Manager) groupDevsByStatus(hps *HwPluginServe) {
	hps.healthDevice = sets.String{}
	for _, device := range hps.devices {
		if device.NetworkHealth != v1beta1.Healthy {
			totalNetworkUnhealthDevices.Insert(device.ID)
		}
		if device.Health == v1beta1.Healthy {
			hps.healthDevice.Insert(device.ID)
			continue
		}
		hnm.setUnHealthyDev(hiAIAscend910Prefix, device)
	}
	hwlog.RunLog.Debugf("healthy device %v", hps.healthDevice)
	hwlog.RunLog.Debugf("total unhealthy devices %v", totalUHDevices)
	hwlog.RunLog.Debugf("total network unhealthy devices %v", totalNetworkUnhealthDevices)
}

// GetAnnotationMap Get Annonation
func (hnm *HwAscend910Manager) GetAnnotationMap(allocatableDevices sets.String, devTypes []string) map[string]string {
	var annoMap = make(map[string]string, len(devTypes))
	for _, suffix := range devTypes {
		powerAnnotation := filterTagPowerDevice(allocatableDevices, suffix)
		annotationTag := fmt.Sprintf("%s%s", resourceNamePrefix, suffix)
		annoMap[annotationTag] = powerAnnotation
	}
	return annoMap
}

func filterTagPowerDevice(allocatableDevices sets.String, suffix string) string {
	var powerAnnotation []string
	for deviceName := range allocatableDevices {
		devType, err := getDeviceType(deviceName)
		if err != nil {
			hwlog.RunLog.Warnf("invalid device name: %s", deviceName)
			continue
		}
		if devType == suffix {
			powerAnnotation = append(powerAnnotation, deviceName)
		}
	}
	return strings.Join(powerAnnotation, ",")
}
