// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

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
	"strings"
	"syscall"

	"github.com/fsnotify/fsnotify"
	"huawei.com/npu-exporter/hwlog"
	"k8s.io/api/core/v1"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

// GetPattern return pattern map
func GetPattern() map[string]string {
	return map[string]string{
		"nodeName":    `^[a-z0-9]([-a-z0-9]*[a-z0-9])?(\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*$`,
		"podName":     "^[a-z0-9]+[a-z0-9\\-]*[a-z0-9]+$",
		"fullPodName": "^[a-z0-9]+([a-z0-9\\-.]*)[a-z0-9]+$",
		"vir910":      "Ascend910-(2|4|8|16)c",
		"vir310p":     "Ascend310P-(1|2|4)c",
	}
}

// GetPodPhaseBlackList get black list of pod phase
func GetPodPhaseBlackList() map[v1.PodPhase]int {
	return map[v1.PodPhase]int{v1.PodFailed: 0, v1.PodSucceeded: 0}
}

// SetAscendRuntimeEnv is to set ascend runtime environment
func SetAscendRuntimeEnv(devices []string, ascendRuntimeOptions string,
	resp *v1beta1.ContainerAllocateResponse) {
	var ascendVisibleDevices []string
	if len((*resp).Envs) == 0 {
		(*resp).Envs = make(map[string]string, runtimeEnvNum)
	}
	(*resp).Envs[ascendVisibleDevicesEnv] = strings.Join(devices, ",")
	(*resp).Envs[ascendRuntimeOptionsEnv] = ascendRuntimeOptions

	hwlog.RunLog.Infof("allocate resp env: %s; %s", strings.Join(ascendVisibleDevices, ","), ascendRuntimeOptions)
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
		hwlog.RunLog.Errorf("marshal data err: %#v", err)
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

// GetDeviceFromPodAnnotation get devices from pod annotation
func GetDeviceFromPodAnnotation(pod *v1.Pod, deviceType string) ([]string, error) {
	if pod == nil {
		return nil, fmt.Errorf("invalid pod")
	}
	annotationTag := fmt.Sprintf("%s%s", ResourceNamePrefix, deviceType)
	annotation, exist := pod.Annotations[annotationTag]
	if !exist {
		return nil, fmt.Errorf("cannot find the annotation")
	}
	return strings.Split(annotation, ","), nil
}

func setDeviceByPathWhen200RC(defaultDevices []string) {
	setDeviceByPath(&defaultDevices, HiAi200RCEventSched)
	setDeviceByPath(&defaultDevices, HiAi200RCHiDvpp)
	setDeviceByPath(&defaultDevices, HiAi200RCLog)
	setDeviceByPath(&defaultDevices, HiAi200RCMemoryBandwidth)
	setDeviceByPath(&defaultDevices, HiAi200RCSVM0)
	setDeviceByPath(&defaultDevices, HiAi200RCTsAisle)
	setDeviceByPath(&defaultDevices, HiAi200RCUpgrade)
}

func setDeviceByPath(defaultDevices *[]string, device string) {
	if _, err := os.Stat(device); err == nil {
		*defaultDevices = append(*defaultDevices, device)
	}
}

// GetDefaultDevices get default device, for allocate mount
func GetDefaultDevices(getFdFlag bool) ([]string, error) {
	// hiAIManagerDevice is required
	if _, err := os.Stat(HiAIManagerDevice); err != nil {
		return nil, err
	}
	var defaultDevices []string
	defaultDevices = append(defaultDevices, HiAIManagerDevice)

	setDeviceByPath(&defaultDevices, HiAIHDCDevice)
	setDeviceByPath(&defaultDevices, HiAISVMDevice)
	if getFdFlag {
		setDeviceByPathWhen200RC(defaultDevices)
	}
	return defaultDevices, nil
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
		if limitsDevNum < 0 || limitsDevNum > int64(MaxDevicesNum) {
			hwlog.RunLog.Errorf("apply devices number should be in the range of [0, %d]", MaxDevicesNum)
			return int64(0)
		}
		total += limitsDevNum
	}
	return total
}

func isAscendAssignedPod(pod *v1.Pod, deviceType string) bool {
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
func FilterPods(pods *v1.PodList, blackList map[v1.PodPhase]int, deviceType string,
	conditionFunc func(pod *v1.Pod) bool) ([]v1.Pod, error) {
	var res []v1.Pod
	if pods == nil {
		return res, fmt.Errorf("input pods variable is nil")
	}
	if len(pods.Items) >= MaxPodLimit {
		return res, fmt.Errorf("filter the number of pods exceeds the upper limit")
	}
	for _, pod := range pods.Items {
		hwlog.RunLog.Debugf("pod: %v, %v", pod.Name, pod.Status.Phase)
		if err := checkPodNameAndSpace(pod.Name, PodNameMaxLength); err != nil {
			hwlog.RunLog.Errorf("pod name syntax illegal, err: %v", err)
			continue
		}
		if err := checkPodNameAndSpace(pod.Namespace, PodNameSpaceMaxLength); err != nil {
			hwlog.RunLog.Errorf("pod namespace syntax illegal, err: %v", err)
			continue
		}
		if _, exist := blackList[pod.Status.Phase]; exist {
			continue
		}
		if conditionFunc != nil && !conditionFunc(&pod) {
			continue
		}
		if getNPUResourceNumOfPod(&pod, deviceType) > 0 && isAscendAssignedPod(&pod,
			deviceType) && !isShouldDeletePod(&pod) {
			res = append(res, pod)
		}
	}
	return res, nil
}

// VerifyPathAndPermission used to verify the validity of the path and permission and return resolved absolute path
func VerifyPathAndPermission(verifyPath string) (string, bool) {
	hwlog.RunLog.Debug("starting check device socket file path.")
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

func checkPodNameAndSpace(podPara string, maxLength int) error {
	if len(podPara) > maxLength {
		return fmt.Errorf("para length %d is bigger than %d", len(podPara), maxLength)
	}
	patternMap := GetPattern()
	pattern := patternMap["podName"]
	if maxLength == PodNameMaxLength {
		pattern = patternMap["fullPodName"]
	}

	if match, err := regexp.MatchString(pattern, podPara); !match || err != nil {
		return fmt.Errorf("podPara is illegal")
	}
	return nil
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
