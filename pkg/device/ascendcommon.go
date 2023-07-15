/* Copyright(C) 2022. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Package device a series of device function
package device

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"huawei.com/npu-exporter/v5/common-utils/hwlog"
	"huawei.com/npu-exporter/v5/devmanager"
	npuCommon "huawei.com/npu-exporter/v5/devmanager/common"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/kubeclient"
)

// isFirstFlushFault for device fault init
var isFirstFlushFault = true

var subscribeToPollingTime = 5 * time.Minute.Milliseconds()

var faultMode = make(map[int32]string, common.GeneralMapSize)

const subscribe = "subscribe"

const polling = "polling"

// AscendTools struct definition
type AscendTools struct {
	client       *kubeclient.ClientK8s
	dmgr         devmanager.DeviceInterface
	name         string
	unHealthyKey string
	devCount     int32
	healthDevice sets.String
}

// DevManager interface for manager device
type DevManager interface {
	GetNPUs() (common.NpuAllInfo, error)
	DoWithVolcanoListAndWatch(map[string][]*common.NpuDevice)
	GraceTolerance(map[string][]*common.NpuDevice)
	SetDmgr(devmanager.DeviceInterface)
	GetDmgr() devmanager.DeviceInterface
	GetChipAICore() int32
	GetName() string
	SetKubeClient(*kubeclient.ClientK8s)
	GetKubeClient() *kubeclient.ClientK8s
	UpdateHealth(map[string][]*common.NpuDevice, []*common.NpuDevice, string)
	GetChange(map[string][]*common.NpuDevice, map[string][]*common.NpuDevice) map[string]bool
	AddPodAnnotation(*v1.Pod, []string, []string, string, string) error
	AppendVGroupInfo([]string)
	CheckDeviceTypeLabel() error
	CreateVirtualDevice(int32, string) (string, error)
	DestroyVirtualDevice(string) error
	GetChipAiCoreCount() (int32, error)
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

// GetChipAICore get ai core
func (tool *AscendTools) GetChipAICore() int32 {
	return common.ParamOption.AiCoreCount
}

// GetName get chip name
func (tool *AscendTools) GetName() string {
	return tool.name
}

// UpdateNodeDeviceInfo update device info
func (tool *AscendTools) UpdateNodeDeviceInfo(devStatusSet common.DevStatusSet,
	updateDeviceInfoFunc func(map[string]string, map[string]string, common.DevStatusSet) error) error {
	waitErr := wait.PollImmediate(common.Interval*time.Second, common.Timeout*time.Second, func() (bool, error) {
		deviceList, err := tool.getDeviceListFromConfigMap()
		if err != nil {
			hwlog.RunLog.Warnf("get device list from config map failed, %#v", err)
			tool.client.ResetDeviceInfo()
			return false, nil
		}
		newDeviceList := common.MapDeepCopy(deviceList)
		if err := updateDeviceInfoFunc(deviceList, newDeviceList, devStatusSet); err != nil {
			hwlog.RunLog.Errorf("update device info failed, err: %#v", err)
			return false, nil
		}
		tool.delVirDevInfo(newDeviceList)
		if _, err := tool.client.WriteDeviceInfoDataIntoCM(newDeviceList); err != nil {
			hwlog.RunLog.Errorf("write device info failed: %#v", err)
			return false, nil
		}

		return true, nil
	})
	return waitErr
}

func (tool *AscendTools) delVirDevInfo(newDeviceList map[string]string) {
	for annotationTag := range common.GetAllDeviceInfoTypeList() {
		if _, ok := newDeviceList[annotationTag]; !ok {
			continue
		}
		if common.IsVirtualDev(annotationTag) {
			delete(newDeviceList, annotationTag)
		}
	}
}

func (tool *AscendTools) assembleNpuDeviceStruct(deviType, deviceName string,
	davinCiDev common.DavinCiDev) common.NpuDevice {
	hwlog.RunLog.Debugf("Found Huawei Ascend, deviceType: %s, deviceName: %s", deviType, deviceName)
	return common.NpuDevice{
		DevType:       deviType,
		DeviceName:    deviceName,
		Health:        v1beta1.Healthy,
		NetworkHealth: v1beta1.Healthy,
		LogicID:       davinCiDev.LogicID,
		PhyID:         davinCiDev.PhyID,
		CardID:        davinCiDev.CardID,
	}
}

func (tool *AscendTools) assemblePhyDevices(davinCiDev common.DavinCiDev, devices *[]common.NpuDevice,
	deviceTypes *[]string) {
	deviceName := fmt.Sprintf("%s-%d", tool.name, davinCiDev.PhyID)
	device := tool.assembleNpuDeviceStruct(tool.name, deviceName, davinCiDev)
	*deviceTypes = append(*deviceTypes, tool.name)
	*devices = append(*devices, device)
}

func (tool *AscendTools) assembleVirtualDevices(davinCiDev common.DavinCiDev, vDevInfos npuCommon.VirtualDevInfo,
	devices *[]common.NpuDevice, vDeviceTypes *[]string) {
	for _, subVDevInfo := range vDevInfos.VDevInfo {
		vDeviType, deviceName, err := tool.assembleSpecVirtualDevice(davinCiDev.PhyID, subVDevInfo)
		if err != nil {
			hwlog.RunLog.Error(err)
			continue
		}
		device := tool.assembleNpuDeviceStruct(vDeviType, deviceName, davinCiDev)
		*devices = append(*devices, device)
		*vDeviceTypes = append(*vDeviceTypes, vDeviType)
	}
}

func (tool *AscendTools) assembleSpecVirtualDevice(phyID int32, vDevInfo npuCommon.CgoVDevQueryStru) (string,
	string, error) {
	coreNum := int32(vDevInfo.QueryInfo.Computing.Aic)
	if coreNum <= 0 {
		return "", "", fmt.Errorf("invalid vdev info, ai core is 0")
	}
	vDeviType, exist := common.GetTemplateName2DeviceTypeMap()[vDevInfo.QueryInfo.Name]
	if !exist {
		return "", "", fmt.Errorf("check templatename failed, templatename is %s", vDevInfo.QueryInfo.Name)
	}
	vDeviType = fmt.Sprintf("%s-%s", tool.name, vDeviType)
	devID := fmt.Sprintf("%s-%d-%d", vDeviType, vDevInfo.VDevID, phyID)
	return vDeviType, devID, nil
}

func (tool *AscendTools) assemble310PMixedPhyDevices(davinCiDev common.DavinCiDev, devices *[]common.NpuDevice,
	deviceTypes *[]string) error {
	cardID, deviceID, err := tool.dmgr.GetCardIDDeviceID(davinCiDev.LogicID)
	if err != nil {
		return fmt.Errorf("get cardID and deviceID failed: LogicID[%#v]", davinCiDev.LogicID)
	}
	productType, err := tool.dmgr.GetProductType(cardID, deviceID)
	if err != nil {
		return fmt.Errorf("get product type failed:cardID[%#v] deviceID[%#v]", cardID, deviceID)
	}
	ProductTypeMap := common.Get310PProductType()
	if _, ok := ProductTypeMap[productType]; !ok {
		return fmt.Errorf("%#v not found", productType)
	}
	deviceName := fmt.Sprintf("%s-%d", ProductTypeMap[productType], davinCiDev.PhyID)
	device := tool.assembleNpuDeviceStruct(ProductTypeMap[productType], deviceName, davinCiDev)
	*deviceTypes = append(*deviceTypes, ProductTypeMap[productType])
	*devices = append(*devices, device)
	return nil
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
	deviceInfo, err := tool.client.GetConfigMap(tool.client.DeviceInfoName, common.DeviceInfoCMNameSpace)
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
	if len(data) > common.CMDataMaxLength {
		return nil, fmt.Errorf("configMap data size is out of memory")
	}
	var nodeDeviceInfo common.NodeDeviceInfoCache
	if err := json.Unmarshal([]byte(data), &nodeDeviceInfo); err != nil {
		return nil, fmt.Errorf("unmarshal configmap data failed, err: %#v", err)
	}
	if nodeDeviceInfo.CheckCode != common.MakeDataHash(nodeDeviceInfo.DeviceInfo) {
		return nil, fmt.Errorf("configmap check hash code error")
	}
	return nodeDeviceInfo.DeviceInfo.DeviceList, nil
}

func getResetInfoData(resetInfo *v1.ConfigMap) ([]*common.TaskDevInfo, error) {
	data, ok := resetInfo.Data[common.ResetInfoCMDataKey]
	if !ok {
		return nil, fmt.Errorf("%s not exist", common.ResetInfoCMDataKey)
	}
	if len(data) > common.CMDataMaxLength {
		return nil, fmt.Errorf("configmap data size is out of memory")
	}
	var taskResetInfo common.TaskResetInfo
	if err := json.Unmarshal([]byte(data), &taskResetInfo); err != nil {
		return nil, fmt.Errorf("unmarshal configmap data failed, err: %#v", err)
	}
	if taskResetInfo.UpdateTime == 0 {
		hwlog.RunLog.Info(" reset configmap is initializing")
		return nil, nil
	}
	checkCode, ok := resetInfo.Data[common.ResetInfoCMCheckCodeKey]
	if !ok {
		return nil, fmt.Errorf("%s not exist", common.ResetInfoCMCheckCodeKey)
	}
	if checkCode != common.MakeDataHash(taskResetInfo) {
		return nil, fmt.Errorf("configmap check hash code error")
	}
	return taskResetInfo.RankList, nil
}

func (tool *AscendTools) getRealUsedDevices() sets.String {
	podList, err := tool.client.GetActivePodList()
	if err != nil {
		hwlog.RunLog.Warn(err)
		return sets.String{}
	}
	usedDevice := sets.String{}
	for _, pod := range podList {
		realDevice, exist := pod.Annotations[common.ResourceNamePrefix+common.PodRealAlloc]
		if !exist {
			continue
		}
		usedDevice.Insert(strings.Split(realDevice, common.CommaSepDev)...)
	}
	return usedDevice
}

func (tool *AscendTools) getDevStatesDevSet(classifyDevs map[string][]*common.NpuDevice) common.DevStatusSet {
	totalFreeDevices := make(map[string]sets.String, len(classifyDevs))
	totalUHDevices, totalNetUHDevices, allTypeUsedDevice := sets.String{}, sets.String{}, sets.String{}
	if !common.ParamOption.PresetVDevice {
		allTypeUsedDevice = tool.getRealUsedDevices()
	}
	for devType, classifyDev := range classifyDevs {
		healthDevices, uhDevices, netUnHDevices := tool.groupDevsByStatus(classifyDev, tool.name)
		usedDevices := tool.client.GetPodsUsedNpu(devType)
		totalFreeDevices[devType] = healthDevices.Difference(usedDevices)
		if !common.ParamOption.PresetVDevice {
			totalFreeDevices[devType] = totalFreeDevices[devType].Difference(allTypeUsedDevice)
		}
		totalUHDevices = totalUHDevices.Union(uhDevices)
		totalNetUHDevices = totalNetUHDevices.Union(netUnHDevices)
	}
	return common.DevStatusSet{
		FreeHealthyDevice:  totalFreeDevices,
		UnHealthyDevice:    totalUHDevices,
		NetUnHealthyDevice: totalNetUHDevices,
	}
}

func (tool *AscendTools) groupDevsByStatus(subClassDevices []*common.NpuDevice, runMode string) (
	sets.String, sets.String, sets.String) {
	healthDevice, totalUHDevices, totalNetworkUHDevices := sets.String{}, sets.String{}, sets.String{}
	for _, device := range subClassDevices {
		if device.NetworkHealth != v1beta1.Healthy {
			totalNetworkUHDevices.Insert(device.DeviceName)
		}
		if device.Health == v1beta1.Healthy {
			healthDevice.Insert(device.DeviceName)
			continue
		}
		if !common.IsVirtualDev(device.DeviceName) {
			totalUHDevices.Insert(device.DeviceName)
			continue
		}
		dev := fmt.Sprintf("%s-%d", runMode, device.PhyID)
		if !totalUHDevices.Has(dev) {
			totalUHDevices.Insert(dev)
		}
	}
	hwlog.RunLog.Debugf("healthy device %#v", healthDevice)
	hwlog.RunLog.Debugf("total unhealthy devices %#v", totalUHDevices)
	hwlog.RunLog.Debugf("total network unhealthy devices %#v", totalNetworkUHDevices)
	return healthDevice, totalUHDevices, totalNetworkUHDevices
}

func (tool *AscendTools) getDavinCiDev(logicID int32) (common.DavinCiDev, error) {
	phyID, err := tool.dmgr.GetPhysicIDFromLogicID(logicID)
	if err != nil {
		return common.DavinCiDev{}, err
	}
	cardID, _, err := tool.dmgr.GetCardIDDeviceID(logicID)
	if err != nil {
		return common.DavinCiDev{}, err
	}
	return common.DavinCiDev{
		LogicID: logicID,
		PhyID:   phyID,
		CardID:  cardID,
	}, nil
}

func (tool *AscendTools) getVirtualDevice(logicID int32) (npuCommon.VirtualDevInfo, error) {
	virtualDevInfos, err := tool.dmgr.GetVirtualDeviceInfo(logicID)
	if err != nil {
		return npuCommon.VirtualDevInfo{}, fmt.Errorf("query virtual device info failure: %s", err)
	}
	return virtualDevInfos, nil
}

func (tool *AscendTools) getDeviceIP(deviceType string, phyID int) (string, error) {
	if common.IsVirtualDev(deviceType) {
		return common.DefaultDeviceIP, nil
	}
	logicID, err := tool.dmgr.GetLogicIDFromPhysicID(int32(phyID))
	if err != nil {
		return "", fmt.Errorf("transfor phyID %d to logicID failed, err: %#v", phyID, err)
	}
	chip, err := tool.dmgr.GetChipInfo(logicID)
	if err != nil {
		return "", fmt.Errorf("get logicID %d chip info failed, err: %#v", logicID, err)
	}
	if strings.Contains(chip.Name, common.VirMark) {
		return common.DefaultDeviceIP, nil
	}
	return tool.dmgr.GetDeviceIPAddress(logicID)
}

func (tool *AscendTools) getDeviceListIP(devices []string, deviceType string) (map[int]string, error) {
	ascendRuntimeOptions := ""
	if common.IsVirtualDev(deviceType) {
		ascendRuntimeOptions = common.VirtualDev
	}
	_, ascendDevices, err := common.GetDeviceListID(devices, ascendRuntimeOptions)
	if err != nil {
		hwlog.RunLog.Errorf("get device list id err: %#v", err)
		return nil, err
	}
	devicesWithIP := make(map[int]string, len(devices))
	for _, id := range ascendDevices {
		if !strings.Contains(deviceType, common.Ascend910) {
			devicesWithIP[id] = ""
			continue
		}
		deviceIP, err := tool.getDeviceIP(deviceType, id)
		if err != nil {
			hwlog.RunLog.Errorf("get device %d ip err: %#v", id, err)
			return nil, err
		}
		devicesWithIP[id] = deviceIP
	}
	return devicesWithIP, nil
}

// AddPodAnnotation get ip of device list
func (tool *AscendTools) AddPodAnnotation(pod *v1.Pod, kltRequestDevices, dpResponseDevices []string,
	deviceType, serverID string) error {
	ascendRuntimeOptions := ""
	if common.IsVirtualDev(deviceType) {
		ascendRuntimeOptions = common.VirtualDev
	}
	phyDevMapVirtualDev, _, err := common.GetDeviceListID(dpResponseDevices, ascendRuntimeOptions)
	if err != nil {
		hwlog.RunLog.Errorf("get device list id err: %#v", err)
		return err
	}
	ascendVisibleDevices, err := tool.getDeviceListIP(dpResponseDevices, deviceType)
	if err != nil {
		return fmt.Errorf("get ascend devices ip failed, err: %#v", err)
	}
	configuration := common.GetPodConfiguration(phyDevMapVirtualDev, ascendVisibleDevices, pod.Name, serverID,
		deviceType)
	if !common.ParamOption.PresetVDevice {
		tool.AppendVGroupInfo(dpResponseDevices)
	}
	annotation := make(map[string]string, 1)
	if !common.IsVirtualDev(deviceType) {
		annotation[common.ResourceNamePrefix+common.Pod2kl] = strings.Join(kltRequestDevices, common.CommaSepDev)
		annotation[common.ResourceNamePrefix+common.PodRealAlloc] = strings.Join(dpResponseDevices, common.CommaSepDev)
	}
	if tool.name == common.Ascend910 {
		annotation[common.Pod910DeviceKey] = configuration
	}
	return tool.client.TryUpdatePodAnnotation(pod, annotation)
}

// UpdateHealth update group device healthy
func (tool *AscendTools) UpdateHealth(groupDevice map[string][]*common.NpuDevice,
	aiCoreDevs []*common.NpuDevice, runMode string) {
	// update health of device
	tool.writeNewFaultCode(groupDevice, runMode)

	setHealthyIfDuoCard(groupDevice)
	setAICoreHealthyIfVNpu(groupDevice, aiCoreDevs)
}

func (tool *AscendTools) GetChange(groupDevice, oldGroupDevice map[string][]*common.NpuDevice) map[string]bool {
	isStateChange := make(map[string]bool, len(groupDevice))
	for devType, devices := range groupDevice {
		for idx, device := range devices {
			if device.Health != oldGroupDevice[devType][idx].Health {
				isStateChange[devType] = true
			}
		}
	}
	return isStateChange
}

func setAICoreHealthyIfVNpu(groupDevice map[string][]*common.NpuDevice, aiCoreDevs []*common.NpuDevice) {
	if common.ParamOption.PresetVDevice {
		return
	}
	logicDeviceMap := make(map[int32]*common.NpuDevice, common.GeneralMapSize)
	for _, devices := range groupDevice {
		for _, device := range devices {
			logicDeviceMap[device.LogicID] = device
		}
	}
	for _, device := range aiCoreDevs {
		device.Health = logicDeviceMap[device.LogicID].Health
		device.FaultCodes = logicDeviceMap[device.LogicID].FaultCodes
		device.AlarmRaisedTime = logicDeviceMap[device.LogicID].AlarmRaisedTime
		device.NetworkHealth = logicDeviceMap[device.LogicID].NetworkHealth
	}
}

func setHealthyIfDuoCard(groupDevice map[string][]*common.NpuDevice) {
	if !common.IsContainAtlas300IDuo() {
		return
	}
	if common.ParamOption.HotReset != common.HotResetInfer {
		hwlog.RunLog.Debugf("not open infer device hot reset function, it's %d", common.ParamOption.HotReset)
		return
	}
	ascend310PDevices, ok := groupDevice[common.Ascend310P]
	if !ok {
		hwlog.RunLog.Debugf("not found 310P devices")
		return
	}
	unHealthyCards := getUnHealthyCard(ascend310PDevices)
	for _, device := range ascend310PDevices {
		if _, ok := unHealthyCards[device.CardID]; ok {
			device.Health = v1beta1.Unhealthy
		}
	}
}

func getUnHealthyCard(ascend310PDevices []*common.NpuDevice) map[int32]int8 {
	unHealthyCards := make(map[int32]int8, len(ascend310PDevices))
	for _, device := range ascend310PDevices {
		if device.Health == v1beta1.Healthy {
			continue
		}
		unHealthyCards[device.CardID] = 0
	}
	return unHealthyCards
}

// ClassifyDevices classify diff type devices
func ClassifyDevices(allDevs []common.NpuDevice, devTypes []string) map[string][]*common.NpuDevice {
	var classifyMap = make(map[string][]*common.NpuDevice, len(devTypes))
	for _, suffix := range devTypes {
		classifyMap[suffix] = classifyDevByType(allDevs, suffix)
	}
	return classifyMap
}

func classifyDevByType(allDevs []common.NpuDevice, suffix string) []*common.NpuDevice {
	var classifyDev []*common.NpuDevice
	for index, device := range allDevs {
		if device.DevType == suffix {
			classifyDev = append(classifyDev, &allDevs[index])
		}
	}
	return classifyDev
}

func isHealthy(faultCodes []int64) string {
	faultType := common.GetFaultTypeByCode(faultCodes)
	if faultType == common.NormalNPU || faultType == common.NotHandleFault {
		return v1beta1.Healthy
	}
	return v1beta1.Unhealthy
}

// UnhealthyState state unhealthy info
func (tool *AscendTools) unhealthyState(healthyState uint32, logicID int32) error {
	phyID, err := tool.dmgr.GetPhysicIDFromLogicID(logicID)
	if err != nil {
		return fmt.Errorf("get phyID failed %#v", err)
	}
	if _, _, err := tool.dmgr.GetDeviceErrorCode(logicID); err != nil {
		return fmt.Errorf("get device error code failed %#v", err)
	}
	hwlog.RunLog.Errorf("device logicID: %d, phyID: %d, state is %d", logicID, phyID, healthyState)
	return nil
}

func (tool *AscendTools) getVGroupID(device string) (uint32, error) {
	phyID, virID, err := common.GetDeviceID(device, common.VirtualDev)
	if err != nil {
		return 0, err
	}
	logicID, err := tool.dmgr.GetLogicIDFromPhysicID(int32(phyID))
	if err != nil {
		return 0, err
	}
	virtualDevInfos, err := tool.dmgr.GetVirtualDeviceInfo(logicID)
	if err != nil {
		return 0, fmt.Errorf("query virtual device info failure: %s", err)
	}
	for _, vDevInfo := range virtualDevInfos.VDevInfo {
		if vDevInfo.VDevID == uint32(virID) {
			return vDevInfo.QueryInfo.Base.VfgID, nil
		}
	}
	return 0, fmt.Errorf("not found virutal device info, %s", device)
}

// AppendVGroupInfo append virtual group id info after device name
func (tool *AscendTools) AppendVGroupInfo(allocateDevice []string) {
	hwlog.RunLog.Debugf("allocateDevice:%v", allocateDevice)
	for i, device := range allocateDevice {
		if !common.IsVirtualDev(device) {
			continue
		}
		vGroupID, err := tool.getVGroupID(device)
		if err != nil {
			hwlog.RunLog.Warn(err)
			continue
		}
		allocateDevice[i] = fmt.Sprintf("%s%s%d", device, common.UnderLine, vGroupID)
	}
}

// CheckDeviceTypeLabel check device type label
func (tool *AscendTools) CheckDeviceTypeLabel() error {
	curNode, err := tool.client.GetNode()
	if err != nil {
		return err
	}
	deviceType, exist := curNode.Labels[common.ServerTypeLabelKey]
	if !exist {
		return fmt.Errorf("label of %s not exist", common.ServerTypeLabelKey)
	}
	deviceTypeInfos := strings.Split(deviceType, common.MiddelLine)
	if len(deviceTypeInfos) < common.ServerTypeInfoMinLen {
		return fmt.Errorf("length of device type info %d is invalid", len(deviceTypeInfos))
	}
	if deviceTypeInfos[0] != tool.name {
		return fmt.Errorf("label chip name %s is not meet real chip name %s", deviceTypeInfos[0], tool.name)
	}
	aiCore, err := strconv.Atoi(deviceTypeInfos[1])
	if err != nil {
		return fmt.Errorf("covert label ai core failed, error is %v", err)
	}
	if aiCore != int(common.ParamOption.AiCoreCount) {
		return fmt.Errorf("label ai core %d not equal real chip ai core %d", aiCore, common.ParamOption.AiCoreCount)
	}
	return nil
}

// CreateVirtualDevice create virtual device
func (tool *AscendTools) CreateVirtualDevice(phyID int32, templateName string) (string, error) {
	createInfo := npuCommon.CgoCreateVDevRes{
		VDevID:       common.DefaultIDForCreateVNPU,
		VfgID:        common.DefaultIDForCreateVNPU,
		TemplateName: templateName,
	}
	logicID, err := tool.dmgr.GetLogicIDFromPhysicID(phyID)
	if err != nil {
		return "", err
	}
	createOut, err := tool.dmgr.CreateVirtualDevice(logicID, createInfo)
	if err != nil {
		hwlog.RunLog.Error(err)
		return "", fmt.Errorf(common.NPUSegmentFailed)
	}
	hwlog.RunLog.Infof("create %s from device %d success", createInfo.TemplateName, phyID)
	vDevType, exist := common.GetTemplateName2DeviceTypeMap()[templateName]
	if !exist {
		return "", fmt.Errorf("check templatename failed, templatename is %s", templateName)
	}
	vDevName := fmt.Sprintf("%s-%s-%d-%d", tool.name, vDevType, createOut.VDevID, phyID)
	return vDevName, nil
}

// DestroyVirtualDevice destroy virtual device
func (tool *AscendTools) DestroyVirtualDevice(deviceName string) error {
	phyID, virID, err := common.GetDeviceID(deviceName, common.VirtualDev)
	if err != nil {
		return fmt.Errorf("get device id failed, %v", err)
	}
	logicID, err := tool.dmgr.GetLogicIDFromPhysicID(int32(phyID))
	if err != nil {
		return err
	}
	for i := 0; i < common.RetryUpdateCount; i++ {
		if err = tool.dmgr.DestroyVirtualDevice(logicID, uint32(virID)); err == nil {
			break
		}
		time.Sleep(time.Second)
	}
	return err
}

// GetChipAiCoreCount get chip aicore count
func (tool *AscendTools) GetChipAiCoreCount() (int32, error) {
	_, logicIDs, err := tool.dmgr.GetDeviceList()
	if err != nil {
		return 0, err
	}
	if len(logicIDs) < 1 {
		return 0, fmt.Errorf("not found logicIDs")
	}
	for _, logicID := range logicIDs {
		cgoVDevInfo, err := tool.dmgr.GetVirtualDeviceInfo(logicID)
		if err != nil && strings.Contains(err.Error(), strconv.Itoa(common.DeviceNotSupport)) {
			return common.DeviceNotSupport, nil
		}
		if err != nil {
			// if not support found aicore number, setting a default value
			hwlog.RunLog.Infof("not found aicore number by dcmi: %#v", err)
			return common.DefaultAiCoreNum, nil
		}
		return tool.getAiCoreCount(cgoVDevInfo)
	}
	return 0, fmt.Errorf("not get aicore count")
}

func (tool *AscendTools) getAiCoreCount(cgoVDevInfo npuCommon.VirtualDevInfo) (int32, error) {
	chipAICore := cgoVDevInfo.TotalResource.Computing.Aic
	if chipAICore < common.MinAICoreNum || chipAICore > common.MaxAICoreNum {
		return 0, fmt.Errorf("invalid ai core num %f", chipAICore)
	}
	return int32(chipAICore), nil
}

// writeNewFaultCode writes fault code and health to device
func (tool *AscendTools) writeNewFaultCode(deviceMap map[string][]*common.NpuDevice, runMode string) {
	initLogicIDs := common.GetAndCleanLogicID()
	devFaultInfoMap := common.GetAndCleanFaultInfo()
	for _, devices := range deviceMap {
		for idx, device := range devices {
			tool.flushFaultCodesWithInit(device, initLogicIDs, devFaultInfoMap)
			device.Health = isHealthy(device.FaultCodes)
			if runMode == common.Ascend910 {
				devices[idx].NetworkHealth = tool.getDeviceNetworkState(device.LogicID)
			}
		}
	}
	isFirstFlushFault = false
}

func (tool *AscendTools) flushFaultCodesWithInit(device *common.NpuDevice, initLogicIDs []int32,
	devFaultInfoMap map[int32][]npuCommon.DevFaultInfo) {
	if isFirstFlushFault || (common.Int32Tool.Contains(initLogicIDs, device.LogicID)) || common.SubscribeFailed ||
		(device.Health == v1beta1.Unhealthy && moreThanFiveMin(device)) {
		_, errCodes, err := tool.dmgr.GetDeviceAllErrorCode(device.LogicID)
		if err != nil {
			hwlog.RunLog.Errorf("get device fault failed logic: %d, err: %v", device.LogicID, err)
			return
		}
		common.SetFaultCodes(device, errCodes)
		logFaultModeChange(device, initLogicIDs, polling)
		return
	}
	common.SetNewFaultAndCacheOnceRecoverFault(device.LogicID, devFaultInfoMap[device.LogicID], device)
	logFaultModeChange(device, initLogicIDs, subscribe)
}

func moreThanFiveMin(device *common.NpuDevice) bool {
	if device.AlarmRaisedTime == 0 {
		return false
	}
	return time.Now().UnixMilli()-device.AlarmRaisedTime > subscribeToPollingTime
}

func logFaultModeChange(device *common.NpuDevice, initLogicIDs []int32, newMode string) {
	var oldMode string
	var ok bool
	if oldMode, ok = faultMode[device.LogicID]; !ok {
		faultMode[device.LogicID] = newMode
		return
	}
	if oldMode == newMode {
		return
	}
	faultMode[device.LogicID] = newMode
	if newMode == polling {
		var reason string
		if device.Health == v1beta1.Unhealthy && moreThanFiveMin(device) {
			reason = "fault raised more than five minutes"
		} else if common.Int32Tool.Contains(initLogicIDs, device.LogicID) {
			reason = "device reset"
		} else if common.SubscribeFailed {
			reason = "subscribe failed"
		} else if isFirstFlushFault {
			reason = "first flush fault"
		} else {
			reason = "unknown reason"
		}
		hwlog.RunLog.Infof("fault get mode downgrade. logicId: %v, cause by : %v", device.LogicID, reason)
		return
	}
	hwlog.RunLog.Infof("fault get mode upgrade. logicId: %v", device.LogicID)
}
