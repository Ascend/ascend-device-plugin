/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2022. All rights reserved.
 */

// Package device implements the query and allocation of the device and the function of the log.
package device

import (
	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/kubeclient"
	"encoding/json"
	"fmt"

	"huawei.com/npu-exporter/devmanager"
	npuCommon "huawei.com/npu-exporter/devmanager/common"
	"huawei.com/npu-exporter/hwlog"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

// AscendTools struct definition
type AscendTools struct {
	client       *kubeclient.ClientK8s
	dmgr         devmanager.DeviceInterface
	name         string
	unHealthyKey string
	devCount     int32
	healthDevice sets.String
}

type devManager interface {
	GetNPUs(*[]common.NpuDevice, *[]string) error
	DoWithVolcanoListAndWatch(map[string][]*common.NpuDevice)
	SetDmgr(devmanager.DeviceInterface)
	GetDmgr() devmanager.DeviceInterface
	SetKubeClient(*kubeclient.ClientK8s)
	GetKubeClient() *kubeclient.ClientK8s
}

// SetDmgr set devmanager
func (tool *AscendTools) SetDmgr(dmgr devmanager.DeviceInterface) {
	tool.dmgr = dmgr
}

// GetDmgr get devmanager
func (tool *AscendTools) GetDmgr() devmanager.DeviceInterface {
	return tool.dmgr
}

// SetKubeClient set ClientK8s
func (tool *AscendTools) SetKubeClient(client *kubeclient.ClientK8s) {
	tool.client = client
}

// GetKubeClient get ClientK8s
func (tool *AscendTools) GetKubeClient() *kubeclient.ClientK8s {
	return tool.client
}

func (tool *AscendTools) assembleNpuDeviceStruct(deviType, deviceName string, logicID, phyID int32) common.NpuDevice {
	hwlog.RunLog.Infof("Found Huawei Ascend, deviceType: %s, deviceName: %s", deviType, deviceName)
	return common.NpuDevice{
		DevType:       deviType,
		DeviceName:    deviceName,
		Health:        v1beta1.Healthy,
		NetworkHealth: v1beta1.Healthy,
		LogicID:       logicID,
		PhyID:         phyID,
	}
}

func (tool *AscendTools) assemblePhyDevices(davinCiDev common.DavinCiDev, devices *[]common.NpuDevice,
	deviceTypes *[]string) {
	deviceName := fmt.Sprintf("%s-%d", tool.name, davinCiDev.PhyID)
	device := tool.assembleNpuDeviceStruct(tool.name, deviceName, davinCiDev.LogicID, davinCiDev.PhyID)
	*deviceTypes = append(*deviceTypes, tool.name)
	*devices = append(*devices, device)
}

func (tool *AscendTools) assembleVirtualDevices(davinCiDev common.DavinCiDev, vDevInfos npuCommon.VirtualDevInfo,
	devices *[]common.NpuDevice, vDeviceTypes *[]string) {
	for _, subVDevInfo := range vDevInfos.VDevInfo {
		vDeviType, deviceName, err := tool.assembleSpecVirtualDevice(davinCiDev.PhyID, subVDevInfo,
			davinCiDev.TemplateName)
		if err != nil {
			hwlog.RunLog.Error(err)
			continue
		}
		device := tool.assembleNpuDeviceStruct(vDeviType, deviceName, davinCiDev.LogicID, davinCiDev.PhyID)
		*devices = append(*devices, device)
		*vDeviceTypes = append(*vDeviceTypes, vDeviType)
	}
}

func (tool *AscendTools) assembleSpecVirtualDevice(phyID int32, vDevInfo npuCommon.CgoVDevQueryStru,
	templateList map[string]string) (string, string, error) {
	coreNum := int32(vDevInfo.QueryInfo.Computing.Aic)
	if coreNum <= 0 {
		return "", "", fmt.Errorf("invalid vdev info, ai core is 0")
	}
	vDeviType, exist := templateList[vDevInfo.QueryInfo.Name]
	if !exist {
		return "", "", fmt.Errorf("check templatename failed, templatename is %s", vDevInfo.QueryInfo.Name)
	}
	devID := fmt.Sprintf("%s-%d-%d", vDeviType, vDevInfo.VDevID, phyID)
	return vDeviType, devID, nil
}

func (tool *AscendTools) removeDuplicate(allDeviceTypes *[]string) []string {
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

func (tool *AscendTools) getDeviceListFromConfigMap() (map[string]string, error) {
	deviceInfo, err := tool.client.GetConfigMap()
	if err != nil || deviceInfo == nil {
		return nil, fmt.Errorf("get configmap failed. %#v", err)
	}
	deviceInfoData, err := getDeviceInfoData(deviceInfo)
	if err != nil || deviceInfoData == nil {
		return nil, fmt.Errorf("get invalid device list. %#v", err)
	}
	return deviceInfoData, nil
}

func getDeviceInfoData(deviceInfo *v1.ConfigMap) (map[string]string, error) {
	data, ok := deviceInfo.Data[common.DeviceInfoCMDataKey]
	if !ok {
		return nil, fmt.Errorf("%s not exist", common.DeviceInfoCMDataKey)
	}
	var nodeDeviceInfo common.NodeDeviceInfoCache
	if err := json.Unmarshal([]byte(data), &nodeDeviceInfo); err != nil {
		return nil, fmt.Errorf("unmarshal configmap data failed, err: %#v", err)
	}
	return nodeDeviceInfo.DeviceInfo.DeviceList, nil
}
