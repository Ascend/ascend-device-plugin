/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2020-2021. All rights reserved.
 */

// Package huawei implements the query and allocation of the device and the function of the log.
package huawei

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"

	"huawei.com/npu-exporter/devmanager"
	npuCommon "huawei.com/npu-exporter/devmanager/common"
	"huawei.com/npu-exporter/hwlog"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"k8s.io/kubernetes/pkg/util/node"

	"Ascend-device-plugin/src/plugin/pkg/npu/common"
)

const (
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
	dmgr                devmanager.DeviceInterface
	phyDevMapVirtualDev map[int32]string
	name                string
	unHealthyKey        string
}

// GetDefaultDevices get default devices
func GetDefaultDevices(defaultDevices *[]string) error {
	// hiAIManagerDevice is required
	if _, err := os.Stat(common.HiAIManagerDevice); err != nil {
		return err
	}
	*defaultDevices = append(*defaultDevices, common.HiAIManagerDevice)

	setDeviceByPath(defaultDevices, common.HiAIHDCDevice)
	setDeviceByPath(defaultDevices, common.HiAISVMDevice)
	if GetFdFlag {
		setDeviceByPathWhen200RC(defaultDevices)
	}
	return nil
}

func setDeviceByPathWhen200RC(defaultDevices *[]string) {
	setDeviceByPath(defaultDevices, common.HiAi200RCEventSched)
	setDeviceByPath(defaultDevices, common.HiAi200RCHiDvpp)
	setDeviceByPath(defaultDevices, common.HiAi200RCLog)
	setDeviceByPath(defaultDevices, common.HiAi200RCMemoryBandwidth)
	setDeviceByPath(defaultDevices, common.HiAi200RCSVM0)
	setDeviceByPath(defaultDevices, common.HiAi200RCTsAisle)
	setDeviceByPath(defaultDevices, common.HiAi200RCUpgrade)
}

func setDeviceByPath(defaultDevices *[]string, device string) {
	if _, err := os.Stat(device); err == nil {
		*defaultDevices = append(*defaultDevices, device)
	}
}

// GetPhyIDByName get physical id from device name
func GetPhyIDByName(DeviceName string) (uint32, error) {
	var phyID uint32
	deviceID, _, err := common.GetDeviceID(DeviceName, physicalDev)
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
		return phyID, fmt.Errorf("getDevice phyid %d overflow", phyID)
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

// getNewNetworkRecoverDev , return new devices to be restored and network unhealthy device in this times
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

// UnhealthyState state unhealth info
func UnhealthyState(healthyState uint32, logicID int32, healthyType string, dmgr devmanager.DeviceInterface) error {
	phyID, err := dmgr.GetPhysicIDFromLogicID(logicID)
	if err != nil {
		return fmt.Errorf("get phyID failed %v", err)
	}
	if _, _, errs := dmgr.GetDeviceErrorCode(logicID); errs != nil {
		return errs
	}
	// if logFlag is true,print device error message
	if logFlag {
		hwlog.RunLog.Errorf("device is unHealthy, "+
			"logicID: %d, phyID: %d, %s: %d", logicID, phyID, healthyType, healthyState)
	}
	return nil
}

// VerifyPath used to verify the validity of the path
func VerifyPath(verifyPath string) (string, bool) {
	hwlog.RunLog.Infof("starting check device socket file path.")
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
func (adc *ascendCommonFunction) AssembleNpuDeviceStruct(deviType, devID string) common.NpuDevice {
	hwlog.RunLog.Infof("Found Huawei Ascend, deviceType: %s, deviceID: %s", deviType, devID)
	return common.NpuDevice{
		DevType:       deviType,
		PciID:         "",
		ID:            devID,
		Health:        v1beta1.Healthy,
		NetworkHealth: v1beta1.Healthy,
	}
}

// GetDevPath is used to get device path
func (adc *ascendCommonFunction) GetDevPath(id, ascendRuntimeOptions string) (string, string) {
	containerPath := fmt.Sprintf("%s%s", "/dev/davinci", id)
	hostPath := containerPath
	if ascendRuntimeOptions == common.VirtualDev {
		hostPath = fmt.Sprintf("%s%s", "/dev/vdavinci", id)
	}
	return containerPath, hostPath
}

// GetDevState get device state
func (adc *ascendCommonFunction) GetDevState(DeviceName string, dmgr devmanager.DeviceInterface) string {
	phyID, err := GetPhyIDByName(DeviceName)
	if err != nil {
		if logFlag {
			hwlog.RunLog.Errorf("get device phyID failed, deviceId: %s, err: %s", DeviceName, err.Error())
		}
		return v1beta1.Unhealthy
	}

	logicID, err := dmgr.GetLogicIDFromPhysicID(int32(phyID))
	if err != nil {
		if logFlag {
			hwlog.RunLog.Errorf("get device logicID failed, deviceId: %s, err: %s", DeviceName, err.Error())
		}
		return v1beta1.Unhealthy
	}
	healthState, err := dmgr.GetDeviceHealth(logicID)
	if err != nil {
		if logFlag {
			hwlog.RunLog.Errorf("get device healthy state failed, deviceId: %d, err: %s", int32(logicID), err.Error())
		}
		return v1beta1.Unhealthy
	}
	switch healthState {
	case normalState, generalAlarm:
		return v1beta1.Healthy
	default:
		err = UnhealthyState(healthState, logicID, "healthState", dmgr)
		if err != nil {
			hwlog.RunLog.Errorf("UnhealthyState, err: %v", err)
		}
		return v1beta1.Unhealthy
	}
}

// GetNPUs Discovers all HUAWEI Ascend310 devices by call devmanager interface
func (adc *ascendCommonFunction) GetNPUs(allDevices *[]common.NpuDevice, allDeviceTypes *[]string,
	deviType string) error {
	hwlog.RunLog.Infof("--->< deviType: %s", deviType)

	devNum, devList, err := adc.dmgr.GetDeviceList()
	if err != nil {
		return err
	}
	if devNum > hiAIMaxCardNum*hiAIMaxDevNumInCard {
		return fmt.Errorf("invalid device num: %d", devNum)
	}
	for i := int32(0); i < devNum; i++ {
		phyID, err := adc.dmgr.GetPhysicIDFromLogicID(devList[i])
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
func (adc *ascendCommonFunction) SetDmgr(dmgr devmanager.DeviceInterface) {
	adc.dmgr = dmgr
}

// GetDmgr to get dmgr
func (adc *ascendCommonFunction) GetDmgr() devmanager.DeviceInterface {
	return adc.dmgr
}

// GetPhyDevMapVirtualDev get phy devices and virtual devices mapping
func (adc *ascendCommonFunction) GetPhyDevMapVirtualDev() map[int32]string {
	return adc.phyDevMapVirtualDev
}

// DoWithVolcanoListAndWatch ascend310P do nothing
func (adc *ascendCommonFunction) DoWithVolcanoListAndWatch(hps *HwPluginServe) {
	adc.reloadHealthDevice(hps)
	usedDevices := sets.NewString()
	getNodeNpuUsed(&usedDevices, hps)
	freeDevices := hps.healthDevice.Difference(usedDevices)
	annoMap := adc.GetAnnotationMap(freeDevices, nil)
	annoMap[adc.unHealthyKey] = filterTagPowerDevice(hps.unHealthDevice, adc.name)
	if err := hps.kubeInteractor.patchNode(func(_ *v1.Node) []byte {
		as := adc.getAntStu(annoMap)
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

func (adc *ascendCommonFunction) getAntStu(annoMap map[string]string) antStu {
	return antStu{
		Metadata: Metadata{
			Annotation: annoMap,
		},
	}
}

// GetDeviceNetworkState check Ascend910 only
func (adc *ascendCommonFunction) GetDeviceNetworkState(_ int32, _ *common.NpuDevice) (string, error) {
	return "", nil
}

func (adc *ascendCommonFunction) reloadHealthDevice(hps *HwPluginServe) {
	hps.healthDevice = sets.String{}
	hps.unHealthDevice = sets.String{}
	for _, device := range hps.devices {
		if device.Health == v1beta1.Healthy {
			hps.healthDevice.Insert(device.ID)
			continue
		}
		hps.unHealthDevice.Insert(device.ID)
	}
}

// GetAnnotationMap Get AnnotationMap
func (adc *ascendCommonFunction) GetAnnotationMap(allocatableDevices sets.String, _ []string) map[string]string {
	var antMap = make(map[string]string, initMapCap)
	chipAnnotation := filterTagPowerDevice(allocatableDevices, adc.name)
	annotationTag := fmt.Sprintf("%s%s", resourceNamePrefix, adc.name)
	antMap[annotationTag] = chipAnnotation
	return antMap
}

func (adc *ascendCommonFunction) getVirtualDevice(logicID int32) (npuCommon.VirtualDevInfo, error) {
	vDevInfos, err := adc.dmgr.GetVirtualDeviceInfo(logicID)
	if err != nil {
		return npuCommon.VirtualDevInfo{}, fmt.Errorf("query virtual device info failure: %s", err)
	}
	return vDevInfos, nil
}

func (adc *ascendCommonFunction) removeDuplicate(allDeviceTypes *[]string) []string {
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

func (adc *ascendCommonFunction) assemblePhyDevices(phyID int32, runMode string) ([]common.NpuDevice, []string) {
	var devices []common.NpuDevice
	var deviTypes []string
	devID := fmt.Sprintf("%s-%d", runMode, phyID)
	device := adc.AssembleNpuDeviceStruct(runMode, devID)
	devices = append(devices, device)
	deviTypes = append(deviTypes, runMode)
	return devices, deviTypes
}

func (adc *ascendCommonFunction) assembleSpecVirtualDevice(runMode string, phyID int32,
	vDevInfo npuCommon.CgoVDevQueryStru) (string, string, error) {
	coreNum := int32(vDevInfo.QueryInfo.Computing.Aic)
	if coreNum <= 0 {
		return "", "", fmt.Errorf("invalid vdev info, ai core is 0")
	}
	vDeviType, exist := getDevTypeByTemplateName(runMode, vDevInfo.QueryInfo.Name)
	if !exist {
		return "", "", fmt.Errorf("check templatename failed, templatename is %s", vDevInfo.QueryInfo.Name)
	}
	devID := fmt.Sprintf("%s-%d-%d", vDeviType, vDevInfo.VDevID, phyID)
	return vDeviType, devID, nil
}

func (adc *ascendCommonFunction) assembleVirtualDevices(phyID int32, vDevInfos npuCommon.VirtualDevInfo,
	runMode string) ([]common.NpuDevice, []string, []string) {
	var devices []common.NpuDevice
	var vDeviTypes []string
	var vDevID []string
	for _, subVDevInfo := range vDevInfos.VDevInfo {
		vDeviType, devID, err := adc.assembleSpecVirtualDevice(runMode, phyID, subVDevInfo)
		if err != nil {
			hwlog.RunLog.Error(err)
			continue
		}
		device := adc.AssembleNpuDeviceStruct(vDeviType, devID)
		devices = append(devices, device)
		vDeviTypes = append(vDeviTypes, vDeviType)
		vDevID = append(vDevID, fmt.Sprintf("%d", subVDevInfo.VDevID))
	}
	return devices, vDeviTypes, vDevID
}

func (adc *ascendCommonFunction) setUnHealthyDev(devType string, device *common.NpuDevice) {
	if !common.IsVirtualDev(device.ID) {
		totalUHDevices.Insert(device.ID)
		return
	}
	phyID, _, err := common.GetDeviceID(device.ID, common.VirtualDev)
	if err != nil {
		hwlog.RunLog.Errorf("getDeviceID err: %v", err)
		return
	}
	dev := fmt.Sprintf("%s-%s", devType, phyID)
	if !totalUHDevices.Has(dev) {
		totalUHDevices.Insert(dev)
	}
}

func (adc *ascendCommonFunction) resetStateSet() {
	totalDevices = totalDevices.Intersection(sets.String{})
	totalUHDevices = totalDevices.Intersection(sets.String{})
	totalNetworkUnhealthDevices = totalNetworkUnhealthDevices.Intersection(sets.String{})
	stateThreadNum = 0
}

func getNodeWithBackgroundCtx(ki *KubeInteractor) (*v1.Node, error) {
	return ki.clientset.CoreV1().Nodes().Get(context.Background(), ki.nodeName, metav1.GetOptions{})
}

func getNodeWithTodoCtx(ki *KubeInteractor) (*v1.Node, error) {
	return ki.clientset.CoreV1().Nodes().Get(context.TODO(), ki.nodeName, metav1.GetOptions{})
}

func patchNodeWithTodoCtx(ki *KubeInteractor, pByte []byte) (*v1.Node, error) {
	return ki.clientset.CoreV1().Nodes().Patch(context.TODO(), ki.nodeName, types.MergePatchType, pByte,
		metav1.PatchOptions{})
}

func patchNodeState(ki *KubeInteractor, curNode, newNode *v1.Node) (*v1.Node, []byte, error) {
	return node.PatchNodeStatus(ki.clientset.CoreV1(), types.NodeName(ki.nodeName), curNode, newNode)
}

func getPodList(ki *KubeInteractor) (*v1.PodList, error) {
	selector := fields.SelectorFromSet(fields.Set{"spec.nodeName": ki.nodeName})
	return ki.clientset.CoreV1().Pods(v1.NamespaceAll).List(context.Background(), metav1.ListOptions{
		FieldSelector: selector.String(),
	})
}
