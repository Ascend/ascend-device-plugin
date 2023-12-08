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

// Package common a series of common function
package common

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"huawei.com/npu-exporter/v5/common-utils/hwlog"
	"huawei.com/npu-exporter/v5/devmanager/common"
	"k8s.io/api/core/v1"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

var (
	dpRegexp = map[string]*regexp.Regexp{
		"nodeName":    regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`),
		"podName":     regexp.MustCompile("^[a-z0-9]+[a-z0-9\\-]*[a-z0-9]+$"),
		"fullPodName": regexp.MustCompile("^[a-z0-9]+([a-z0-9\\-.]*)[a-z0-9]+$"),
		"vir910":      regexp.MustCompile("Ascend910-([2-6]|8|10|12|16)c"),
		"vir310p":     regexp.MustCompile("Ascend310P-(1|2|4)c"),
		"ascend910":   regexp.MustCompile(`^Ascend910-\d+`),
		"ascend310":   regexp.MustCompile(`^Ascend310-\d+`),
		"ascend310P":  regexp.MustCompile(`^Ascend310P-\d+`),
	}
)

// GetPattern return pattern map
func GetPattern() map[string]*regexp.Regexp {
	return dpRegexp
}

var (
	allDeviceInfoLock sync.Mutex
)

// LockAllDeviceInfo lock for device info status
func LockAllDeviceInfo() {
	allDeviceInfoLock.Lock()
}

// UnlockAllDeviceInfo unlock for device info status
func UnlockAllDeviceInfo() {
	allDeviceInfoLock.Unlock()
}

// SetAscendRuntimeEnv is to set ascend runtime environment
func SetAscendRuntimeEnv(devices []int, ascendRuntimeOptions string,
	resp *v1beta1.ContainerAllocateResponse) {
	if resp == nil {
		hwlog.RunLog.Error("resp is nil")
		return
	}
	if len((*resp).Envs) == 0 {
		(*resp).Envs = make(map[string]string, runtimeEnvNum)
	}
	var deviceStr []string
	for _, id := range devices {
		deviceStr = append(deviceStr, strconv.Itoa(id))
	}
	(*resp).Envs[AscendVisibleDevicesEnv] = strings.Join(deviceStr, ",")
	(*resp).Envs[ascendRuntimeOptionsEnv] = ascendRuntimeOptions
	if ParamOption.RealCardType == Ascend310B {
		(*resp).Envs[ascendAllowLinkEnv] = "True"
	}

	hwlog.RunLog.Infof("allocate resp env: %s; %s", (*resp).Envs[AscendVisibleDevicesEnv], ascendRuntimeOptions)
}

// MakeDataHash Make Data Hash
func MakeDataHash(data interface{}) string {
	var dataBuffer []byte
	if dataBuffer = MarshalData(data); len(dataBuffer) == 0 {
		return ""
	}
	h := sha256.New()
	if _, err := h.Write(dataBuffer); err != nil {
		hwlog.RunLog.Error("hash data error")
		return ""
	}
	sum := h.Sum(nil)
	return hex.EncodeToString(sum)
}

// MarshalData marshal data to bytes
func MarshalData(data interface{}) []byte {
	dataBuffer, err := json.Marshal(data)
	if err != nil {
		hwlog.RunLog.Errorf("marshal data err: %v", err)
		return nil
	}
	return dataBuffer
}

// MapDeepCopy map deep copy
func MapDeepCopy(source map[string]string) map[string]string {
	dest := make(map[string]string, len(source))
	if source == nil {
		return dest
	}
	for key, value := range source {
		dest[key] = value
	}
	return dest
}

// GetPodAnnotationByDeviceType get pod annotation by device type
func GetPodAnnotationByDeviceType(pod *v1.Pod, deviceType string) (string, error) {
	if pod == nil {
		return "", fmt.Errorf("invalid pod")
	}
	annotationTag := fmt.Sprintf("%s%s", ResourceNamePrefix, deviceType)
	annotation, exist := pod.Annotations[annotationTag]
	if !exist {
		return "", fmt.Errorf("cannot find the annotation")
	}
	if len(annotation) > PodAnnotationMaxLength {
		return "", fmt.Errorf("pod annotation size out of memory")
	}
	return annotation, nil
}

// GetDeviceFromPodAnnotation get devices from pod annotation
func GetDeviceFromPodAnnotation(pod *v1.Pod, deviceType string) ([]string, error) {
	if pod == nil {
		return nil, fmt.Errorf("param pod is nil")
	}
	annotation, err := GetPodAnnotationByDeviceType(pod, deviceType)
	if err != nil {
		return nil, err
	}
	return strings.Split(annotation, CommaSepDev), nil
}

func setDeviceByPathWhen200RC(defaultDevices *[]string) {
	setDeviceByPath(defaultDevices, HiAi200RCEventSched)
	setDeviceByPath(defaultDevices, HiAi200RCHiDvpp)
	setDeviceByPath(defaultDevices, HiAi200RCLog)
	setDeviceByPath(defaultDevices, HiAi200RCMemoryBandwidth)
	setDeviceByPath(defaultDevices, HiAi200RCSVM0)
	setDeviceByPath(defaultDevices, HiAi200RCTsAisle)
	setDeviceByPath(defaultDevices, HiAi200RCUpgrade)
}

func setDeviceByPath(defaultDevices *[]string, device string) {
	if _, err := os.Stat(device); err == nil {
		*defaultDevices = append(*defaultDevices, device)
	}
}

// GetDefaultDevices get default device, for allocate mount
func GetDefaultDevices(getFdFlag bool) ([]string, error) {
	davinciManager, err := getDavinciManagerPath()
	if err != nil {
		return nil, err
	}
	var defaultDevices []string
	defaultDevices = append(defaultDevices, davinciManager)

	setDeviceByPath(&defaultDevices, HiAIHDCDevice)
	setDeviceByPath(&defaultDevices, HiAISVMDevice)
	if getFdFlag {
		setDeviceByPathWhen200RC(&defaultDevices)
	}

	var productType string
	if len(ParamOption.ProductTypes) == 1 {
		productType = ParamOption.ProductTypes[0]
	}
	if productType == Atlas200ISoc {
		socDefaultDevices, err := set200SocDefaultDevices()
		if err != nil {
			hwlog.RunLog.Errorf("get 200I soc default devices failed, err: %v", err)
			return nil, err
		}
		defaultDevices = append(defaultDevices, socDefaultDevices...)
	}
	if ParamOption.RealCardType == Ascend310B {
		a310BDefaultDevices := set310BDefaultDevices()
		defaultDevices = append(defaultDevices, a310BDefaultDevices...)
	}
	return defaultDevices, nil
}

func getDavinciManagerPath() (string, error) {
	if ParamOption.RealCardType == Ascend310B {
		if _, err := os.Stat(HiAIManagerDeviceDocker); err == nil {
			return HiAIManagerDeviceDocker, nil
		}
		hwlog.RunLog.Warn("get davinci manager docker failed")
	}
	if _, err := os.Stat(HiAIManagerDevice); err != nil {
		return "", err
	}
	return HiAIManagerDevice, nil
}

// set200SocDefaultDevices set 200 soc defaults devices
func set200SocDefaultDevices() ([]string, error) {
	var socDefaultDevices = []string{
		Atlas200ISocVPC,
		Atlas200ISocVDEC,
		Atlas200ISocSYS,
		Atlas200ISocSpiSmbus,
		Atlas200ISocUserConfig,
		HiAi200RCTsAisle,
		HiAi200RCSVM0,
		HiAi200RCLog,
		HiAi200RCMemoryBandwidth,
		HiAi200RCUpgrade,
	}
	for _, devPath := range socDefaultDevices {
		if _, err := os.Stat(devPath); err != nil {
			return nil, err
		}
	}
	var socOptionsDevices = []string{
		HiAi200RCEventSched,
		Atlas200ISocXSMEM,
	}
	for _, devPath := range socOptionsDevices {
		if _, err := os.Stat(devPath); err != nil {
			hwlog.RunLog.Warnf("device %s not exist", devPath)
			continue
		}
		socDefaultDevices = append(socDefaultDevices, devPath)
	}
	return socDefaultDevices, nil
}

func set310BDefaultDevices() []string {
	var a310BDefaultDevices = []string{
		Atlas310BDvppCmdlist,
		Atlas310BPngd,
		Atlas310BVenc,
		HiAi200RCUpgrade,
		Atlas200ISocSYS,
		HiAi200RCSVM0,
		Atlas200ISocVDEC,
		Atlas200ISocVPC,
		HiAi200RCTsAisle,
		HiAi200RCLog,
		Atlas310BAcodec,
		Atlas310BAi,
		Atlas310BAo,
		Atlas310BVo,
		Atlas310BHdmi,
	}
	var available310BDevices []string
	for _, devPath := range a310BDefaultDevices {
		if _, err := os.Stat(devPath); err != nil {
			hwlog.RunLog.Warnf("device %s not exist", devPath)
			continue
		}
		available310BDevices = append(available310BDevices, devPath)
	}
	return available310BDevices
}

func getNPUResourceNumOfPod(pod *v1.Pod, deviceType string) int64 {
	containers := pod.Spec.Containers
	if len(containers) > MaxContainerLimit {
		hwlog.RunLog.Error("The number of container exceeds the upper limit")
		return int64(0)
	}
	var total int64
	annotationTag := fmt.Sprintf("%s%s", ResourceNamePrefix, deviceType)
	for _, container := range containers {
		val, ok := container.Resources.Limits[v1.ResourceName(annotationTag)]
		if !ok {
			continue
		}
		limitsDevNum := val.Value()
		if limitsDevNum < 0 || limitsDevNum > int64(MaxDevicesNum*MaxAICoreNum) {
			hwlog.RunLog.Errorf("apply devices number should be in the range of [0, %d]", MaxDevicesNum*MaxAICoreNum)
			return int64(0)
		}
		total += limitsDevNum
	}
	return total
}

func isAscendAssignedPod(pod *v1.Pod, deviceType string) bool {
	if IsVirtualDev(deviceType) {
		return true
	}
	annotationTag := fmt.Sprintf("%s%s", ResourceNamePrefix, deviceType)
	if _, ok := pod.ObjectMeta.Annotations[annotationTag]; !ok {
		hwlog.RunLog.Debugf("no assigned flag, pod Name: %s, pod NameSpace: %s", pod.Name, pod.Namespace)
		return false
	}
	return true
}

func isShouldDeletePod(pod *v1.Pod) bool {
	if pod.DeletionTimestamp != nil {
		return true
	}
	if len(pod.Status.ContainerStatuses) > MaxContainerLimit {
		hwlog.RunLog.Error("The number of container exceeds the upper limit")
		return true
	}
	for _, status := range pod.Status.ContainerStatuses {
		if status.State.Waiting != nil &&
			strings.Contains(status.State.Waiting.Message, "PreStartContainer check failed") {
			return true
		}
	}
	return pod.Status.Reason == "UnexpectedAdmissionError"
}

// FilterPods get pods which meet the conditions
func FilterPods(pods []v1.Pod, deviceType string, conditionFunc func(pod *v1.Pod) bool) []v1.Pod {
	var res []v1.Pod
	for _, pod := range pods {
		hwlog.RunLog.Debugf("pod: %s, %s", pod.Name, pod.Status.Phase)
		if getNPUResourceNumOfPod(&pod, deviceType) == 0 || !isAscendAssignedPod(&pod,
			deviceType) || isShouldDeletePod(&pod) {
			continue
		}
		if conditionFunc != nil && !conditionFunc(&pod) {
			continue
		}
		res = append(res, pod)
	}
	return res
}

// VerifyPathAndPermission used to verify the validity of the path and permission and return resolved absolute path
func VerifyPathAndPermission(verifyPath string, waitSecond int) (string, bool) {
	hwlog.RunLog.Debug("starting check device socket file path.")
	absVerifyPath, err := filepath.Abs(verifyPath)
	if err != nil {
		hwlog.RunLog.Error("abs current path failed")
		return "", false
	}
	pathInfo, err := os.Stat(absVerifyPath)
	if err != nil {
		for i := 0; i < waitSecond; i++ {
			time.Sleep(time.Second)
			pathInfo, err = os.Stat(absVerifyPath)
			if err == nil {
				break
			}
		}
		if err != nil {
			hwlog.RunLog.Error("file path not exist")
			return "", false
		}
	}
	realPath, err := filepath.EvalSymlinks(absVerifyPath)
	if err != nil || absVerifyPath != realPath {
		hwlog.RunLog.Error("Symlinks is not allowed")
		return "", false
	}
	stat, ok := pathInfo.Sys().(*syscall.Stat_t)
	if !ok || stat.Uid != RootUID || stat.Gid != RootGID {
		hwlog.RunLog.Error("Non-root owner group of the path")
		return "", false
	}
	return realPath, true
}

// CheckPodNameAndSpace used to check pod name or pod namespace
func CheckPodNameAndSpace(podPara string, maxLength int) error {
	if len(podPara) > maxLength {
		return fmt.Errorf("para length %d is bigger than %d", len(podPara), maxLength)
	}
	patternMap := GetPattern()
	pattern := patternMap["podName"]
	if maxLength == PodNameMaxLength {
		pattern = patternMap["fullPodName"]
	}

	if match := pattern.MatchString(podPara); !match {
		return fmt.Errorf("podPara %s is illegal", podPara)
	}
	return nil
}

// CheckDeviceName used to check device name
func CheckDeviceName(deviceName, deviceRunMode string) bool {
	patternMap := GetPattern()

	runModeRegexpMap := map[string]string{
		common.Ascend910:  RunMode910,
		common.Ascend310:  RunMode310,
		common.Ascend310P: RunMode310P,
	}

	pattern := patternMap[runModeRegexpMap[deviceRunMode]]
	if !pattern.MatchString(deviceName) {
		hwlog.RunLog.Warnf("in %s device run mode, device name %s is illegal", deviceRunMode, deviceName)
		return false
	}

	return true
}

// NewFileWatch is used to watch socket file
func NewFileWatch() (*FileWatch, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}
	return &FileWatch{FileWatcher: watcher}, nil
}

// WatchFile add file to watch
func (fw *FileWatch) WatchFile(fileName string) error {
	if _, err := os.Stat(fileName); err != nil {
		return err
	}
	return fw.FileWatcher.Add(fileName)
}

// NewSignWatcher new sign watcher
func NewSignWatcher(osSigns ...os.Signal) chan os.Signal {
	// create signs chan
	signChan := make(chan os.Signal, 1)
	for _, sign := range osSigns {
		signal.Notify(signChan, sign)
	}
	return signChan
}

// GetPodConfiguration get annotation configuration of pod
func GetPodConfiguration(phyDevMapVirtualDev map[int]int, devices map[int]string, podName,
	serverID string, deviceType string) string {
	var sortDevicesKey []int
	for deviceID := range devices {
		sortDevicesKey = append(sortDevicesKey, deviceID)
	}
	sort.Ints(sortDevicesKey)
	instance := Instance{PodName: podName, ServerID: serverID}
	for _, deviceID := range sortDevicesKey {
		if !IsVirtualDev(deviceType) {
			instance.Devices = append(instance.Devices, Device{
				DeviceID: fmt.Sprintf("%d", deviceID),
				DeviceIP: devices[deviceID],
			})
			continue
		}
		phyID, exist := phyDevMapVirtualDev[deviceID]
		if !exist {
			hwlog.RunLog.Warn("virtual device not found phyid")
			continue
		}
		instance.Devices = append(instance.Devices, Device{
			DeviceID: fmt.Sprintf("%d", phyID),
			DeviceIP: devices[deviceID],
		})
	}
	instanceByte, err := json.Marshal(instance)
	if err != nil {
		hwlog.RunLog.Errorf("Transform marshal failed, err: %v", err)
		return ""
	}
	return string(instanceByte)
}

// CheckFileUserSameWithProcess to check whether the owner of the log file is the same as the uid
func CheckFileUserSameWithProcess(loggerPath string) bool {
	curUid := os.Getuid()
	if curUid == RootUID {
		return true
	}
	pathInfo, err := os.Lstat(loggerPath)
	if err != nil {
		path := filepath.Dir(loggerPath)
		pathInfo, err = os.Lstat(path)
		if err != nil {
			fmt.Printf("get logger path stat failed, error is %v\n", err)
			return false
		}
	}
	stat, ok := pathInfo.Sys().(*syscall.Stat_t)
	if !ok {
		fmt.Printf("get logger file stat failed\n")
		return false
	}
	if int(stat.Uid) != curUid || int(stat.Gid) != curUid {
		fmt.Printf("check log file failed, owner not right\n")
		return false
	}
	return true
}

// IsContainAtlas300IDuo in ProductTypes list, is contain Atlas 300I Duo card
func IsContainAtlas300IDuo() bool {
	for _, product := range ParamOption.ProductTypes {
		if product == Atlas300IDuo {
			return true
		}
	}
	return false
}

// RecordFaultInfoList record the fault info
func RecordFaultInfoList(devFaultInfoList []*TaskDevInfo) {
	for _, devFaultInfo := range devFaultInfoList {
		hexErrorCode := strings.ToUpper(Int64Tool.ToHexString(devFaultInfo.ErrorCode))
		hwlog.RunLog.Infof("rank id: %d, log id: %d, policy: %s, error code: %s",
			devFaultInfo.RankId, devFaultInfo.LogicId, devFaultInfo.Policy, hexErrorCode)
	}
}

// Int32Join int32 join to string
func Int32Join(data []int32, sep string) string {
	strData := make([]string, 0, len(data))
	for _, val := range data {
		strData = append(strData, strconv.Itoa(int(val)))
	}
	return strings.Join(strData, sep)
}

// GetDeviceRunMode get current env device run mode
func GetDeviceRunMode() (string, error) {
	devType := ParamOption.RealCardType

	switch devType {
	case common.Ascend310, common.Ascend310B:
		return common.Ascend310, nil
	case common.Ascend910, common.Ascend910B:
		return common.Ascend910, nil
	case common.Ascend310P:
		return common.Ascend310P, nil
	default:
		hwlog.RunLog.Errorf("found an unsupported device type %s", devType)
		return "", fmt.Errorf("%v is a unsupported device type", devType)
	}
}
