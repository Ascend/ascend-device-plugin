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

// Package kubeclient a series of k8s function
package kubeclient

import (
	"fmt"
	"net"
	"strings"
	"time"

	"huawei.com/npu-exporter/v5/common-utils/hwlog"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	"Ascend-device-plugin/pkg/common"
)

var tryUpdatePodWaitTime = 200 * time.Millisecond
var deviceInfoFlushTime = int64(60 * 60)

// TryUpdatePodAnnotation is to try updating pod annotation
func (ki *ClientK8s) TryUpdatePodAnnotation(pod *v1.Pod, annotation map[string]string) error {
	if annotation == nil {
		return fmt.Errorf("invalid annotation")
	}
	for i := 0; i < common.RetryUpdateCount; i++ {
		if pod.Name != "" {
			for k, v := range annotation {
				pod.Annotations[k] = v
			}
			_, err := ki.UpdatePod(pod)
			if err == nil {
				return nil
			}
			hwlog.RunLog.Debugf("update pod annotation failed, times: %d, error is %#v", i+1, err)
		}
		time.Sleep(tryUpdatePodWaitTime)
		podNew := ki.GetPodCache(pod.Namespace, pod.Name)
		pod = &podNew
	}
	return fmt.Errorf("update pod annotation failed, exceeded max number of retries")
}

// TryUpdatePodCacheAnnotation is to try updating pod annotation in both api server and cache
func (ki *ClientK8s) TryUpdatePodCacheAnnotation(pod *v1.Pod, annotation map[string]string) error {
	if err := ki.TryUpdatePodAnnotation(pod, annotation); err != nil {
		hwlog.RunLog.Errorf("update pod annotation in api server failed, err: %v", err)
		return err
	}
	// update cache
	lock.Lock()
	defer lock.Unlock()
	for i, podInCache := range podList {
		if podInCache.Namespace == pod.Namespace && podInCache.Name == pod.Name {
			for k, v := range annotation {
				podList[i].Annotations[k] = v
			}
			hwlog.RunLog.Debugf("update annotation in pod cache success, name: %s, namespace: %s", pod.Name, pod.Namespace)
			return nil
		}
	}
	hwlog.RunLog.Warnf("no pod found in cache when update annotation, name: %s, namespace: %s", pod.Name, pod.Namespace)
	return nil
}

func (ki *ClientK8s) createOrUpdateDeviceCM(cm *v1.ConfigMap) error {
	// use update first
	if _, err := ki.UpdateConfigMap(cm); errors.IsNotFound(err) {
		if _, err := ki.CreateConfigMap(cm); err != nil {
			return fmt.Errorf("unable to create configmap, %#v", err)
		}
		return nil
	} else {
		return err
	}
}

// WriteDeviceInfoDataIntoCM write deviceinfo into config map
func (ki *ClientK8s) WriteDeviceInfoDataIntoCM(deviceInfo map[string]string) (*common.NodeDeviceInfoCache, error) {

	var nodeDeviceData = common.NodeDeviceInfoCache{
		DeviceInfo: common.NodeDeviceInfo{
			DeviceList: deviceInfo,
			UpdateTime: time.Now().Unix(),
		},
	}
	nodeDeviceData.CheckCode = common.MakeDataHash(nodeDeviceData.DeviceInfo)

	var data []byte
	if data = common.MarshalData(nodeDeviceData); len(data) == 0 {
		return nil, fmt.Errorf("marshal nodeDeviceData failed")
	}
	deviceInfoCM := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ki.DeviceInfoName,
			Namespace: common.DeviceInfoCMNameSpace,
		},
		Data: map[string]string{common.DeviceInfoCMDataKey: string(data)},
	}

	hwlog.RunLog.Debugf("write device info cache into cm: %s/%s.", deviceInfoCM.Namespace, deviceInfoCM.Name)
	if err := ki.createOrUpdateDeviceCM(deviceInfoCM); err != nil {
		return nil, err
	}
	return &nodeDeviceData, nil
}

func isNotChangeOrLessOneHour(cache *common.NodeDeviceInfoCache, deviceMap map[string]string) bool {
	if cache == nil || len(cache.DeviceInfo.DeviceList) != len(deviceMap) {
		return false
	}
	for key, oldDeviceInfo := range cache.DeviceInfo.DeviceList {
		if deviceMap[key] != oldDeviceInfo {
			return false
		}
	}
	return time.Now().Unix()-cache.DeviceInfo.UpdateTime < deviceInfoFlushTime
}

// WriteResetInfoDataIntoCM write reset info into config map
func (ki *ClientK8s) WriteResetInfoDataIntoCM(taskName string, namespace string,
	taskInfo *common.TaskResetInfo) (*v1.ConfigMap, error) {
	oldCM, err := ki.GetConfigMap(common.ResetInfoCMNamePrefix+taskName, namespace)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get reset cm of task %s, err: %#v", taskName, err)
		return nil, err
	}

	oldResetInfoData, ok := oldCM.Data[common.ResetInfoCMDataKey]
	if !ok {
		return nil, fmt.Errorf("invalid reset info data")
	}
	if strings.Contains(oldResetInfoData, common.IsolateError) && len(taskInfo.RankList) != 0 {
		return nil, fmt.Errorf("task should be rescheduled")
	}

	newTaskInfo := setNewTaskInfoWithHexString(taskInfo)
	newTaskInfo.UpdateTime = time.Now().Unix()
	checkCode := common.MakeDataHash(newTaskInfo)
	var data []byte
	if data = common.MarshalData(newTaskInfo); len(data) == 0 {
		return nil, fmt.Errorf("marshal task reset data failed")
	}
	resetInfoCM := &v1.ConfigMap{
		TypeMeta:   oldCM.TypeMeta,
		ObjectMeta: oldCM.ObjectMeta,
		Data: map[string]string{
			common.ResetInfoCMDataKey:      string(data),
			common.ResetInfoCMCheckCodeKey: checkCode,
		},
	}

	hwlog.RunLog.Debugf("write reset info cache into cm: %s/%s.", resetInfoCM.Namespace, resetInfoCM.Name)
	return ki.UpdateConfigMap(resetInfoCM)
}

func setNewTaskInfoWithHexString(taskInfo *common.TaskResetInfo) *common.TaskResetInfo {
	var newTaskInfo common.TaskResetInfo
	for _, deviceInfo := range taskInfo.RankList {
		newDeviceInfo := *deviceInfo
		newDeviceInfo.ErrorCodeHex = strings.ToUpper(common.Int64Tool.ToHexString(newDeviceInfo.ErrorCode))
		newDeviceInfo.ErrorCode = []int64{}
		newTaskInfo.RankList = append(newTaskInfo.RankList, &newDeviceInfo)
	}
	return &newTaskInfo
}

// WriteFaultInfoDataIntoCM write fault info into config map
func (ki *ClientK8s) WriteFaultInfoDataIntoCM(taskName string, namespace string,
	faultInfo *common.TaskFaultInfo) (*v1.ConfigMap, error) {
	oldCM, err := ki.GetConfigMap(common.FaultInfoCMNamePrefix+taskName, namespace)
	if err != nil {
		if errors.IsNotFound(err) {
			hwlog.RunLog.Infof("fault config map in task %s is not found", taskName)
			return nil, nil
		}
		hwlog.RunLog.Errorf("failed to get fault cm of task %s, err: %#v", taskName, err)
		return nil, err
	}
	taskFaultInfo := &common.TaskFaultInfoCache{
		FaultInfo: faultInfo,
	}
	taskFaultInfo.FaultInfo.UpdateTime = time.Now().Unix()
	checkCode := common.MakeDataHash(taskFaultInfo.FaultInfo)
	var data []byte
	if data = common.MarshalData(taskFaultInfo.FaultInfo); len(data) == 0 {
		return nil, fmt.Errorf("marshal task reset data failed")
	}
	faultInfoCM := &v1.ConfigMap{
		TypeMeta:   oldCM.TypeMeta,
		ObjectMeta: oldCM.ObjectMeta,
		Data: map[string]string{
			common.FaultInfoCMDataKey:      string(data),
			common.FaultInfoCMCheckCodeKey: checkCode,
		},
	}

	hwlog.RunLog.Debugf("write fault info cache into cm: %s/%s.", faultInfoCM.Namespace, faultInfoCM.Name)
	return ki.UpdateConfigMap(faultInfoCM)
}

// AnnotationReset reset annotation and device info
func (ki *ClientK8s) AnnotationReset() error {
	curNode, err := ki.GetNode()
	if err != nil {
		hwlog.RunLog.Errorf("failed to get node, nodeName: %s, err: %#v", ki.NodeName, err)
		return err
	}
	if curNode == nil {
		hwlog.RunLog.Error("invalid node")
		return fmt.Errorf("invalid node")
	}
	newNode := curNode.DeepCopy()
	ki.resetNodeAnnotations(newNode)
	ki.ResetDeviceInfo()
	for i := 0; i < common.RetryUpdateCount; i++ {
		if _, _, err = ki.PatchNodeState(curNode, newNode); err == nil {
			hwlog.RunLog.Infof("reset annotation success")
			return nil
		}
		hwlog.RunLog.Errorf("failed to patch volcano npu resource, times:%d", i+1)
		time.Sleep(time.Second)
		continue
	}
	hwlog.RunLog.Errorf("failed to patch volcano npu resource: %#v", err)
	return err
}

// GetPodsUsedNpu get npu by status
func (ki *ClientK8s) GetPodsUsedNpu(devType string) sets.String {
	podList := ki.GetActivePodListCache()
	var useNpu []string
	for _, pod := range podList {
		annotationTag := fmt.Sprintf("%s%s", common.ResourceNamePrefix, devType)
		tmpNpu, ok := pod.Annotations[annotationTag]
		if !ok || len(tmpNpu) == 0 || len(tmpNpu) > common.PodAnnotationMaxLength {
			continue
		}
		tmpNpuList := strings.Split(tmpNpu, common.CommaSepDev)
		if len(tmpNpuList) == 0 || len(tmpNpuList) > common.MaxDevicesNum {
			hwlog.RunLog.Warnf("invalid annotation, len is %d", len(tmpNpu))
			continue
		}
		useNpu = append(useNpu, tmpNpuList...)
		hwlog.RunLog.Debugf("pod Name: %s, getNPUByStatus vol : %#v", pod.Name, tmpNpu)
	}
	hwlog.RunLog.Debugf("nodeName: %s, useNpus: %#v", ki.NodeName, useNpu)
	return sets.NewString(useNpu...)
}

// GetNodeServerID Get Node Server ID
func (ki *ClientK8s) GetNodeServerID() (string, error) {
	node, err := ki.GetNode()
	if err != nil {
		return "", err
	}
	if len(node.Status.Addresses) > common.MaxPodLimit {
		hwlog.RunLog.Error("the number of node status in exceeds the upper limit")
		return "", fmt.Errorf("the number of node status in exceeds the upper limit")
	}
	var serverID string
	for _, addresses := range node.Status.Addresses {
		if addresses.Type == v1.NodeInternalIP && net.ParseIP(addresses.Address) != nil {
			serverID = addresses.Address
			break
		}
	}
	return serverID, nil
}
