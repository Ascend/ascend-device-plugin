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
	"fmt"
	"time"

	"Ascend-device-plugin/pkg/common"
	"huawei.com/npu-exporter/v5/common-utils/hwlog"
	"k8s.io/api/core/v1"
)

var podUpdateSecond int64
var podList *v1.PodList

// GetAllPodListNoCache get pod list by field selector latest and flush cache
func (ki *ClientK8s) GetAllPodListNoCache() (*v1.PodList, error) {
	if common.ParamOption.CacheExpirePeriod < 1 {
		return ki.GetAllPodList()
	}
	newPodList, err := ki.GetAllPodList()
	if err != nil {
		return nil, err
	}
	// listAndWatch、Allocate、ListenDevice three thread minimum possible both set podList
	// if set lock will reduce the performance. set podList before set podUpdateSecond, it can work normally
	podList = newPodList
	podUpdateSecond = time.Now().Unix()
	return podList, nil
}

// GetAllPodListCache get pod list by field selector with cache
func (ki *ClientK8s) GetAllPodListCache() (*v1.PodList, error) {
	if common.ParamOption.CacheExpirePeriod < 1 {
		return ki.GetAllPodList()
	}
	if podUpdateSecond+common.ParamOption.CacheExpirePeriod > time.Now().Unix() {
		return podList, nil
	}
	newPodList, err := ki.GetAllPodList()
	if err != nil {
		return nil, err
	}
	// listAndWatch、Allocate、ListenDevice three thread minimum possible both set podList
	// if set lock will reduce the performance. set podList before set podUpdateSecond, it can work normally
	podList = newPodList
	podUpdateSecond = time.Now().Unix()
	return podList, nil
}

// GetActivePodListNoCache is to get active pod list latest and flush cache
func (ki *ClientK8s) GetActivePodListNoCache() ([]v1.Pod, error) {
	if common.ParamOption.CacheExpirePeriod < 1 {
		return ki.GetActivePodList()
	}
	newPodList, err := ki.GetAllPodListNoCache()
	if err != nil {
		return nil, err
	}
	if newPodList == nil {
		return nil, fmt.Errorf("pod list is invalid")
	}
	return filterStatus(newPodList.Items), nil
}

// GetActivePodListCache is to get active pod list with cache
func (ki *ClientK8s) GetActivePodListCache() ([]v1.Pod, error) {
	if common.ParamOption.CacheExpirePeriod < 1 {
		return ki.GetActivePodList()
	}
	newPodList, err := ki.GetAllPodListCache()
	if err != nil {
		return nil, err
	}
	if newPodList == nil {
		return nil, fmt.Errorf("pod list is invalid")
	}
	return filterStatus(newPodList.Items), nil
}

func filterStatus(pods []v1.Pod) []v1.Pod {
	if pods == nil {
		return pods
	}
	if len(pods) >= common.MaxPodLimit {
		return pods
	}
	var newPods []v1.Pod
	for _, pod := range pods {
		if err := common.CheckPodNameAndSpace(pod.Name, common.PodNameMaxLength); err != nil {
			hwlog.RunLog.Warnf("pod name syntax illegal, err: %#v", err)
			continue
		}
		if err := common.CheckPodNameAndSpace(pod.Namespace, common.PodNameSpaceMaxLength); err != nil {
			hwlog.RunLog.Warnf("pod namespace syntax illegal, err: %#v", err)
			continue
		}
		if pod.Status.Phase == v1.PodFailed || pod.Status.Phase == v1.PodSucceeded {
			hwlog.RunLog.Debugf("pod status: %v is not active", pod.Status.Phase)
			continue
		}
		newPods = append(newPods, pod)
	}
	return newPods
}
