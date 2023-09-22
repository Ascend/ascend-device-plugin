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
	return podList
}

// GetActivePodListCache is to get active pod list with cache
func (ki *ClientK8s) GetActivePodListCache() []v1.Pod {
	if len(podList) == 0 {
		return []v1.Pod{}
	}
	var newPodList []v1.Pod
	lock.Lock()
	defer lock.Unlock()
	for _, pod := range podList {
		if err := common.CheckPodNameAndSpace(pod.GetName(), common.PodNameMaxLength); err != nil {
			hwlog.RunLog.Warnf("pod name syntax illegal, err: %#v", err)
			continue
		}
		if err := common.CheckPodNameAndSpace(pod.GetNamespace(), common.PodNameSpaceMaxLength); err != nil {
			hwlog.RunLog.Warnf("pod namespace syntax illegal, err: %#v", err)
			continue
		}
		if pod.Status.Phase == v1.PodFailed || pod.Status.Phase == v1.PodSucceeded {
			continue
		}
		newPodList = append(newPodList, pod)
	}
	return newPodList
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
