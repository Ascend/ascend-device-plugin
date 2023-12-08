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

// Package kubeclient a series of k8s function
package kubeclient

import (
	"sync"

	"huawei.com/npu-exporter/v5/common-utils/hwlog"
	"k8s.io/api/core/v1"

	"Ascend-device-plugin/pkg/common"
)

const podDeleteOperator = "delete"
const podAddOperator = "add"
const podUpdateOperator = "update"

var podList []v1.Pod
var lock sync.Mutex
var nodeServerIp string
var nodeDeviceInfoCache *common.NodeDeviceInfoCache

// UpdatePodList update pod list by informer
func UpdatePodList(oldObj, newObj interface{}, operator string) {
	newPod, ok := newObj.(*v1.Pod)
	if !ok {
		return
	}
	lock.Lock()
	defer lock.Unlock()
	switch operator {
	case podAddOperator:
		podList = append(podList, *newPod)
	case podDeleteOperator:
		for i, localPod := range podList {
			if localPod.Namespace == newPod.Namespace && localPod.Name == newPod.Name {
				podList = append(podList[:i], podList[i+1:]...)
				return
			}
		}
		hwlog.RunLog.Infof("pod delete failed, pod %s already delete", newPod.Name)
	case podUpdateOperator:
		oldPod, ok := oldObj.(*v1.Pod)
		if !ok {
			return
		}
		for i, localPod := range podList {
			if localPod.UID == oldPod.UID {
				podList[i] = *newPod
				return
			}
		}
		podList = append(podList, *newPod)
		hwlog.RunLog.Infof("pod %s update failed, use add", newPod.Name)
	}
}

// GetAllPodListCache get pod list by field selector with cache,
func (ki *ClientK8s) GetAllPodListCache() []v1.Pod {
	if ki.IsApiErr {
		newV1PodList, err := ki.GetAllPodList()
		if err != nil {
			hwlog.RunLog.Errorf("get pod list from api-server failed: %v", err)
			return podList
		}
		podList = newV1PodList.Items
		ki.IsApiErr = false
		hwlog.RunLog.Info("get new pod list success")
	}

	return podList
}

// GetActivePodListCache is to get active pod list with cache
func (ki *ClientK8s) GetActivePodListCache() []v1.Pod {
	if len(podList) == 0 {
		return []v1.Pod{}
	}
	newPodList := make([]v1.Pod, 0, common.GeneralMapSize)
	lock.Lock()
	defer lock.Unlock()

	if ki.IsApiErr {
		newV1PodList, err := ki.GetAllPodList()
		if err != nil {
			hwlog.RunLog.Errorf("get pod list from api-server failed: %v", err)
		} else {
			podList = newV1PodList.Items
			ki.IsApiErr = false
			hwlog.RunLog.Info("get new pod list success")
		}
	}

	for _, pod := range podList {
		if err := common.CheckPodNameAndSpace(pod.GetName(), common.PodNameMaxLength); err != nil {
			hwlog.RunLog.Warnf("pod name syntax illegal, err: %v", err)
			continue
		}
		if err := common.CheckPodNameAndSpace(pod.GetNamespace(), common.PodNameSpaceMaxLength); err != nil {
			hwlog.RunLog.Warnf("pod namespace syntax illegal, err: %v", err)
			continue
		}
		if pod.Status.Phase == v1.PodFailed || pod.Status.Phase == v1.PodSucceeded {
			continue
		}
		newPodList = append(newPodList, pod)
	}
	return newPodList
}

// GetPodCache get pod by namespace and name with cache
func (ki *ClientK8s) GetPodCache(namespace, name string) v1.Pod {
	if len(podList) == 0 {
		return v1.Pod{}
	}
	lock.Lock()
	defer lock.Unlock()
	for _, pod := range podList {
		if pod.Namespace == namespace && pod.Name == name {
			return pod
		}
	}
	return v1.Pod{}
}

// GetNodeServerIDCache Get Node Server ID with cache
func (ki *ClientK8s) GetNodeServerIDCache() (string, error) {
	if nodeServerIp != "" {
		return nodeServerIp, nil
	}
	serverID, err := ki.GetNodeServerID()
	if err != nil {
		return "", err
	}
	nodeServerIp = serverID
	return serverID, nil
}

// GetDeviceInfoCMCache get device info configMap with cache
func (ki *ClientK8s) GetDeviceInfoCMCache() *common.NodeDeviceInfoCache {
	return nodeDeviceInfoCache
}

// WriteDeviceInfoDataIntoCMCache write deviceinfo into config map with cache
func (ki *ClientK8s) WriteDeviceInfoDataIntoCMCache(deviceInfo map[string]string, manuallySeparateNPU string) error {
	newNodeDeviceInfoCache, err := ki.WriteDeviceInfoDataIntoCM(deviceInfo, manuallySeparateNPU)
	if err != nil {
		return err
	}
	nodeDeviceInfoCache = newNodeDeviceInfoCache
	return nil
}

// SetNodeDeviceInfoCache set device info cache
func (ki *ClientK8s) SetNodeDeviceInfoCache(deviceInfoCache *common.NodeDeviceInfoCache) {
	nodeDeviceInfoCache = deviceInfoCache
}
