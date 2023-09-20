/* Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
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
	"fmt"
	"sort"
	"strconv"
	"strings"

	"huawei.com/npu-exporter/v5/common-utils/hwlog"
	"k8s.io/api/core/v1"

	"Ascend-device-plugin/pkg/common"
)

// HotResetManager hot reset manager
type HotResetManager interface {
	GetRingNum() int
	GetDevIdList(string) []int32
	GetTaskDevFaultInfoList(string) ([]*common.TaskDevInfo, error)
	GetTaskPod(string) (v1.Pod, error)
	GetAllTaskDevFaultInfoList() map[string][]*common.TaskDevInfo
	GetDevProcessPolicy(string) string
	GetTaskProcessPolicy(string) (string, int, error)
	GetDevListInReset() map[int32]struct{}
	GetDevListByPolicyLevel([]*common.TaskDevInfo, int) (map[int32]struct{}, error)
	GetNeedResetDevList([]*common.TaskDevInfo) (map[int32]struct{}, error)
	GetTaskResetInfo([]*common.TaskDevInfo, string, string, string) (*common.TaskResetInfo, error)
	GetTaskFaultRankInfo([]*common.TaskDevInfo) (*common.TaskFaultInfo, error)
	GetFaultDev2PodMap() (map[int32]v1.Pod, error)
	GenerateTaskDevFaultInfoList(devIdList []int32, rankIndex string) ([]*common.TaskDevInfo, error)
	UpdateFaultDev2PodMap([]int32, v1.Pod) error
	UpdateGlobalDevFaultInfoCache([]*common.NpuDevice) error
	UpdateTaskDevListCache(map[string][]int32) error
	UpdateTaskDevFaultInfoCache(map[string][]*common.TaskDevInfo) error
	UpdateTaskPodCache(map[string]v1.Pod) error
	UpdateFreeTask(map[string]struct{})
	SetTaskInReset(string) error
	SetDevInReset(int32) error
	SetAllDevInReset(info *common.TaskResetInfo) error
	UnSetTaskInReset(string) error
	UnSetDevInReset(int32) error
	UnSetAllDevInReset(*common.TaskResetInfo) error
	IsCurNodeTaskInReset(string) bool
	IsExistFaultyDevInTask(string) bool
	DeepCopyDevFaultInfoList([]*common.TaskDevInfo) []*common.TaskDevInfo
}

// HotResetTools hot reset tool
type HotResetTools struct {
	ringNum             int
	allTaskDevList      map[string][]int32
	allTaskDevFaultInfo map[string][]*common.TaskDevInfo
	globalDevFaultInfo  map[int32]*common.DevFaultInfo
	taskPod             map[string]v1.Pod
	faultDev2PodMap     map[int32]v1.Pod
	resetTask           map[string]struct{}
	resetDev            map[int32]struct{}
	processPolicyTable  map[string]int
}

// NewHotResetManager create HotResetManager and init data
func NewHotResetManager(devUsage string) HotResetManager {
	var ringNumber int
	switch common.ParamOption.RealCardType {
	case common.Ascend910:
		ringNumber = common.Ascend910RingsNum
	case common.Ascend910B:
		switch devUsage {
		case common.Infer:
			ringNumber = common.Ascend910BRingsNumInfer
		case common.Train:
			ringNumber = common.Ascend910BRingsNumTrain
		default:
			return nil
		}
	default:
		return nil
	}
	return &HotResetTools{
		ringNum:         ringNumber,
		resetTask:       map[string]struct{}{},
		resetDev:        map[int32]struct{}{},
		faultDev2PodMap: map[int32]v1.Pod{},
		processPolicyTable: map[string]int{
			common.EmptyError:          common.EmptyErrorLevel,
			common.IgnoreError:         common.IgnoreErrorLevel,
			common.RestartRequestError: common.RestartRequestErrorLevel,
			common.RestartError:        common.RestartErrorLevel,
			common.ResetError:          common.ResetErrorLevel,
			common.IsolateError:        common.IsolateErrorLevel,
		},
	}
}

func getChipCountOnRing() int {
	var ring = map[string]int{
		common.Ascend910:  common.Ascend910RingsNum,
		common.Ascend910B: common.Ascend910BRingsNumTrain,
	}
	return ring[common.ParamOption.RealCardType]
}

// GetRingNum get device num in a ring
func (hrt *HotResetTools) GetRingNum() int {
	if hrt.ringNum == 0 {
		return getChipCountOnRing()
	}
	return hrt.ringNum
}

// GetTaskDevFaultInfoList return task device fault info list
func (hrt *HotResetTools) GetTaskDevFaultInfoList(taskName string) ([]*common.TaskDevInfo, error) {
	taskDevFaultInfoList, ok := hrt.allTaskDevFaultInfo[taskName]
	if !ok {
		return nil, fmt.Errorf("task %s is not in task device fault info list cache", taskName)
	}
	return taskDevFaultInfoList, nil
}

// GetTaskPod return task pod
func (hrt *HotResetTools) GetTaskPod(taskName string) (v1.Pod, error) {
	pod, ok := hrt.taskPod[taskName]
	if !ok {
		return v1.Pod{}, fmt.Errorf("task %s is not in task pod cache", taskName)
	}
	return pod, nil
}

// GetAllTaskDevFaultInfoList return all task device fault info list
func (hrt *HotResetTools) GetAllTaskDevFaultInfoList() map[string][]*common.TaskDevInfo {
	return hrt.allTaskDevFaultInfo
}

// GetDevListInReset return the logic id list of device in reset
func (hrt *HotResetTools) GetDevListInReset() map[int32]struct{} {
	return hrt.resetDev
}

// GetDevProcessPolicy return the policy of device with fault
func (hrt *HotResetTools) GetDevProcessPolicy(faultType string) string {
	switch faultType {
	case common.NormalNPU, common.NotHandleFault:
		return common.EmptyError
	case common.RestartRequest:
		return common.RestartRequestError
	case common.RestartBusiness:
		return common.RestartError
	case common.FreeRestartNPU, common.RestartNPU:
		return common.ResetError
	default:
		return common.IsolateError
	}
}

// GetTaskProcessPolicy return a task process policy
func (hrt *HotResetTools) GetTaskProcessPolicy(taskName string) (string, int, error) {
	devFaultInfoList, ok := hrt.allTaskDevFaultInfo[taskName]
	if !ok {
		return "", -1, fmt.Errorf("this task is not in the cache")
	}
	var processPolicy string
	var processPolicyLevel int
	for _, devFaultInfo := range devFaultInfoList {
		devPolicyLevel, ok := hrt.processPolicyTable[devFaultInfo.Policy]
		if !ok {
			return "", -1, fmt.Errorf("invalid policy of device fault info in task %s", taskName)
		}
		if devPolicyLevel > processPolicyLevel {
			processPolicy = devFaultInfo.Policy
			processPolicyLevel = devPolicyLevel
		}
	}
	return processPolicy, processPolicyLevel, nil
}

// GetDevIdList convert device str to device logic id list
func (hrt *HotResetTools) GetDevIdList(devStr string) []int32 {
	var phyIDs []int32
	for _, deviceName := range strings.Split(devStr, common.CommaSepDev) {
		phyID, _, err := common.GetDeviceID(deviceName, common.CommaSepDev)
		if err != nil {
			hwlog.RunLog.Errorf("get phyID failed, err: %#v", err)
			return nil
		}
		phyIDs = append(phyIDs, int32(phyID))
	}
	return phyIDs
}

// GetNeedRestartDevList return the task list by policy level
func (hrt *HotResetTools) GetDevListByPolicyLevel(devFaultInfoList []*common.TaskDevInfo,
	policyLevel int) (map[int32]struct{}, error) {
	devList := make(map[int32]struct{})
	for _, devFaultInfo := range devFaultInfoList {
		policyType, ok := hrt.processPolicyTable[devFaultInfo.Policy]
		if !ok {
			err := fmt.Errorf("invalid policy str of device %d", devFaultInfo.LogicId)
			hwlog.RunLog.Error(err)
			return nil, err
		}
		if policyType >= policyLevel {
			if _, ok := devList[devFaultInfo.LogicId]; !ok {
				devList[devFaultInfo.LogicId] = struct{}{}
			}
		}
	}
	return devList, nil
}

// GetNeedResetDevList return device logic id list to be reset
func (hrt *HotResetTools) GetNeedResetDevList(devFaultInfoList []*common.TaskDevInfo) (map[int32]struct{}, error) {
	needResetDevList := make(map[int32]struct{})
	for _, devFaultInfo := range devFaultInfoList {
		policyType, ok := hrt.processPolicyTable[devFaultInfo.Policy]
		if !ok {
			err := fmt.Errorf("invalid policy str of device %d",
				devFaultInfo.LogicId)
			hwlog.RunLog.Error(err)
			return nil, err
		}
		if policyType == common.RestartErrorLevel || policyType == common.ResetErrorLevel ||
			policyType == common.RestartRequestErrorLevel {
			resetIndex := devFaultInfo.LogicId / int32(hrt.GetRingNum())
			if _, ok := needResetDevList[devFaultInfo.LogicId]; !ok {
				needResetDevList[resetIndex*int32(hrt.GetRingNum())] = struct{}{}
			}
		}
	}
	return needResetDevList, nil
}

// GetTaskResetInfo return the detail reset info of task to process
func (hrt *HotResetTools) GetTaskResetInfo(devFaultInfoList []*common.TaskDevInfo, policy, initPolicy,
	status string) (*common.TaskResetInfo, error) {
	faultRing := make(map[int]struct{}, common.RingSum)
	var rankList []*common.TaskDevInfo
	for _, devFaultInfo := range devFaultInfoList {
		policy := hrt.processPolicyTable[devFaultInfo.Policy]
		if policy != common.RestartErrorLevel && policy != common.ResetErrorLevel &&
			policy != common.RestartRequestErrorLevel {
			continue
		}
		ringStartIndex := int(devFaultInfo.LogicId) / hrt.GetRingNum()
		faultRing[ringStartIndex] = struct{}{}
	}
	for _, devInfo := range devFaultInfoList {
		ringIndex := int(devInfo.LogicId) / hrt.GetRingNum()
		if _, ok := faultRing[ringIndex]; !ok {
			continue
		}
		newDevInfo := hrt.DeepCopyDevInfo(devInfo)
		newDevInfo.Policy = policy
		newDevInfo.InitialPolicy = initPolicy
		newDevInfo.Status = status
		rankList = append(rankList, newDevInfo)
	}
	return &common.TaskResetInfo{
		RankList: rankList,
	}, nil
}

// GetTaskFaultRankInfo return the fault rank info of task to update fault cm
func (hrt *HotResetTools) GetTaskFaultRankInfo(devFaultInfoList []*common.TaskDevInfo) (*common.TaskFaultInfo, error) {
	taskFaultInfo := &common.TaskFaultInfo{
		FaultRank: make([]int, 0),
	}
	faultRing := make(map[int]struct{}, common.RingSum)
	for _, devFaultInfo := range devFaultInfoList {
		policy := hrt.processPolicyTable[devFaultInfo.Policy]
		if policy != common.RestartErrorLevel && policy != common.ResetErrorLevel &&
			policy != common.RestartRequestErrorLevel {
			continue
		}
		ringStartIndex := int(devFaultInfo.LogicId) / hrt.GetRingNum()
		faultRing[ringStartIndex] = struct{}{}
	}
	for _, devInfo := range devFaultInfoList {
		ringIndex := int(devInfo.LogicId) / hrt.GetRingNum()
		if _, ok := faultRing[ringIndex]; !ok {
			continue
		}
		taskFaultInfo.FaultRank = append(taskFaultInfo.FaultRank, devInfo.RankId)
	}
	return taskFaultInfo, nil
}

func (hrt *HotResetTools) GetFaultDev2PodMap() (map[int32]v1.Pod, error) {
	if hrt.faultDev2PodMap == nil {
		return nil, fmt.Errorf("no valid faultDev2PodMap here")
	}
	return hrt.faultDev2PodMap, nil
}

// GenerateTaskDevFaultInfoList generate device fault info list in a task by device logic id list and rank index
func (hrt *HotResetTools) GenerateTaskDevFaultInfoList(devIdList []int32,
	rankIndex string) ([]*common.TaskDevInfo, error) {
	sort.Slice(devIdList, func(i, j int) bool {
		return devIdList[i] < devIdList[j]
	})
	rankStart, err := strconv.Atoi(rankIndex)
	if err != nil {
		hwlog.RunLog.Errorf("failed to convert rank index to int, err: %#v", err)
		return nil, err
	}
	devNum := len(devIdList)
	taskDevInfoList := make([]*common.TaskDevInfo, 0, len(devIdList))
	for _, devId := range devIdList {
		var rankId int
		switch rankIndex {
		case common.InferRankIndex:
			rankId = rankStart
		default:
			rankId = rankStart*devNum + len(taskDevInfoList)
		}
		faultInfo, ok := hrt.globalDevFaultInfo[devId]
		if !ok {
			return nil, fmt.Errorf("device %d is not in global cache", devId)
		}
		taskDevInfo := &common.TaskDevInfo{
			RankId:       rankId,
			DevFaultInfo: *faultInfo,
		}
		taskDevInfoList = append(taskDevInfoList, taskDevInfo)
	}
	return taskDevInfoList, nil
}

// UpdateFaultDevAndPodMap updates the mapping between the unhealthy device and pod
func (hrt *HotResetTools) UpdateFaultDev2PodMap(devList []int32, pod v1.Pod) error {
	if hrt.faultDev2PodMap == nil {
		return fmt.Errorf("no valid faultDev2PodMap here")
	}
	for _, device := range devList {
		// save when device is unhealthy
		if hrt.globalDevFaultInfo[device].Policy != common.EmptyError &&
			hrt.globalDevFaultInfo[device].Policy != common.IgnoreError {
			hrt.faultDev2PodMap[device] = pod
			continue
		}

		// delete when device is healthy
		if _, ok := hrt.faultDev2PodMap[device]; ok {
			delete(hrt.faultDev2PodMap, device)
		}
	}

	return nil
}

// UpdateGlobalDevFaultInfoCache update global device fault info cache
func (hrt *HotResetTools) UpdateGlobalDevFaultInfoCache(devDeviceList []*common.NpuDevice) error {
	if len(devDeviceList) == 0 {
		return fmt.Errorf("npu device list is nil")
	}
	hrt.globalDevFaultInfo = make(map[int32]*common.DevFaultInfo, len(devDeviceList))
	for _, device := range devDeviceList {
		hrt.globalDevFaultInfo[device.LogicID] = &common.DevFaultInfo{}
		hrt.globalDevFaultInfo[device.LogicID].LogicId = device.LogicID
		hrt.globalDevFaultInfo[device.LogicID].ErrorCode = device.FaultCodes
		hrt.globalDevFaultInfo[device.LogicID].Policy =
			hrt.GetDevProcessPolicy(common.GetFaultTypeByCode(device.FaultCodes))
	}
	return nil
}

// UpdateTaskDevListCache update all task device list cache
func (hrt *HotResetTools) UpdateTaskDevListCache(taskDevList map[string][]int32) error {
	if taskDevList == nil {
		return fmt.Errorf("task device list is nil")
	}
	hrt.allTaskDevList = taskDevList
	return nil
}

// UpdateTaskDevFaultInfoCache update all task device fault info cache
func (hrt *HotResetTools) UpdateTaskDevFaultInfoCache(taskDevFaultInfo map[string][]*common.TaskDevInfo) error {
	if taskDevFaultInfo == nil {
		return fmt.Errorf("taskDevFaultInfo is nil")
	}
	hrt.allTaskDevFaultInfo = taskDevFaultInfo
	return nil
}

// UpdateTaskPodCache update all task pod cache
func (hrt *HotResetTools) UpdateTaskPodCache(taskPod map[string]v1.Pod) error {
	if taskPod == nil {
		return fmt.Errorf("taskPod is nil")
	}
	hrt.taskPod = taskPod
	return nil
}

// UpdateFreeTask unset task in reset task after delete task
func (hrt *HotResetTools) UpdateFreeTask(taskListUsedDevice map[string]struct{}) {
	for taskName := range hrt.resetTask {
		if _, ok := taskListUsedDevice[taskName]; !ok {
			delete(hrt.resetTask, taskName)
		}
	}
}

// IsCurNodeTaskInReset check whether the current task is being reset on the current node
func (hrt *HotResetTools) IsCurNodeTaskInReset(taskName string) bool {
	if _, ok := hrt.resetTask[taskName]; !ok {
		return false
	}
	return true
}

func (hrt *HotResetTools) IsExistFaultyDevInTask(taskName string) bool {
	if _, ok := hrt.allTaskDevList[taskName]; !ok {
		hwlog.RunLog.Warnf("task: %s is not exist in cache", taskName)
		return false
	}
	for _, pod := range hrt.faultDev2PodMap {
		if pod.Annotations[common.ResetTaskNameKey] == taskName {
			hwlog.RunLog.Infof("faulty device exists in task: %s", taskName)
			return true
		}
	}

	return false
}

// SetTaskInReset set a task to the reset state
func (hrt *HotResetTools) SetTaskInReset(taskName string) error {
	if _, ok := hrt.resetTask[taskName]; ok {
		return fmt.Errorf("task %s is resetting", taskName)
	}
	hrt.resetTask[taskName] = struct{}{}
	return nil
}

// SetDevInReset set a device to the reset state
func (hrt *HotResetTools) SetDevInReset(devId int32) error {
	if _, ok := hrt.resetDev[devId]; ok {
		return fmt.Errorf("dev %d is resetting", devId)
	}
	hrt.resetDev[devId] = struct{}{}
	return nil
}

// SetAllDevInReset set all device in a task to the reset state
func (hrt *HotResetTools) SetAllDevInReset(resetInfo *common.TaskResetInfo) error {
	for _, devInfo := range resetInfo.RankList {
		if err := hrt.SetDevInReset(devInfo.LogicId); err != nil {
			return err
		}
	}
	return nil
}

// UnSetDevInReset unset a device in a task to leave the reset state
func (hrt *HotResetTools) UnSetDevInReset(devId int32) error {
	if _, ok := hrt.resetDev[devId]; !ok {
		return fmt.Errorf("device %d is not resetting", devId)
	}
	delete(hrt.resetDev, devId)
	return nil
}

// UnSetAllDevInReset unset all device in a task to leave the reset state
func (hrt *HotResetTools) UnSetAllDevInReset(resetInfo *common.TaskResetInfo) error {
	for _, devInfo := range resetInfo.RankList {
		if err := hrt.UnSetDevInReset(devInfo.LogicId); err != nil {
			return err
		}
	}
	return nil
}

// UnSetTaskInReset unset a task to leave the reset state
func (hrt *HotResetTools) UnSetTaskInReset(taskName string) error {
	if _, ok := hrt.resetTask[taskName]; !ok {
		return fmt.Errorf("task %s is not in reset task cache", taskName)
	}
	delete(hrt.resetTask, taskName)
	return nil
}

// DeepCopyDevInfo copy device info deeply
func (hrt *HotResetTools) DeepCopyDevInfo(devInfo *common.TaskDevInfo) *common.TaskDevInfo {
	return &common.TaskDevInfo{
		RankId:       devInfo.RankId,
		DevFaultInfo: devInfo.DevFaultInfo,
	}
}

// DeepCopyDevFaultInfoList copy device fault info list deeply
func (hrt *HotResetTools) DeepCopyDevFaultInfoList(devFaultInfoList []*common.TaskDevInfo) []*common.TaskDevInfo {
	var newDevFaultInfoList []*common.TaskDevInfo
	for _, devFaultInfo := range devFaultInfoList {
		newDevFaultInfoList = append(newDevFaultInfoList, &common.TaskDevInfo{
			RankId:       devFaultInfo.RankId,
			DevFaultInfo: devFaultInfo.DevFaultInfo,
		})
	}
	return newDevFaultInfoList
}
