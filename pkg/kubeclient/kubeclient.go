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
	"context"
	"fmt"
	"os"

	"huawei.com/npu-exporter/v5/common-utils/hwlog"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/component-helpers/node/util"

	"Ascend-device-plugin/pkg/common"
)

// ClientK8s include ClientK8sSet & nodeName & configmap name
type ClientK8s struct {
	Clientset      kubernetes.Interface
	NodeName       string
	DeviceInfoName string
}

// NewClientK8s create k8s client
func NewClientK8s() (*ClientK8s, error) {
	clientCfg, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		hwlog.RunLog.Errorf("build client config err: %v", err)
		return nil, err
	}

	client, err := kubernetes.NewForConfig(clientCfg)
	if err != nil {
		hwlog.RunLog.Errorf("get client err: %v", err)
		return nil, err
	}
	nodeName, err := getNodeNameFromEnv()
	if err != nil {
		return nil, err
	}

	return &ClientK8s{
		Clientset:      client,
		NodeName:       nodeName,
		DeviceInfoName: common.DeviceInfoCMNamePrefix + nodeName,
	}, nil
}

// GetNode get node
func (ki *ClientK8s) GetNode() (*v1.Node, error) {
	return ki.Clientset.CoreV1().Nodes().Get(context.Background(), ki.NodeName, metav1.GetOptions{
		ResourceVersion: "0",
	})
}

// PatchNodeState patch node state
func (ki *ClientK8s) PatchNodeState(curNode, newNode *v1.Node) (*v1.Node, []byte, error) {
	return util.PatchNodeStatus(ki.Clientset.CoreV1(), types.NodeName(ki.NodeName), curNode, newNode)
}

// GetPod get pod by namespace and name
func (ki *ClientK8s) GetPod(pod *v1.Pod) (*v1.Pod, error) {
	if pod == nil {
		return nil, fmt.Errorf("param pod is nil")
	}
	return ki.Clientset.CoreV1().Pods(pod.Namespace).Get(context.Background(), pod.Name, metav1.GetOptions{
		ResourceVersion: "0",
	})
}

// UpdatePod update pod by namespace and name
func (ki *ClientK8s) UpdatePod(pod *v1.Pod) (*v1.Pod, error) {
	return ki.Clientset.CoreV1().Pods(pod.Namespace).Update(context.Background(), pod, metav1.UpdateOptions{})
}

// GetActivePodList is to get active pod list
func (ki *ClientK8s) GetActivePodList() ([]v1.Pod, error) {
	fieldSelector, err := fields.ParseSelector("spec.nodeName=" + ki.NodeName + "," +
		"status.phase!=" + string(v1.PodSucceeded) + ",status.phase!=" + string(v1.PodFailed))
	if err != nil {
		return nil, err
	}
	podList, err := ki.getPodListByCondition(fieldSelector)
	if err != nil {
		return nil, err
	}
	return checkPodList(podList)
}

// GetAllPodList get pod list by field selector
func (ki *ClientK8s) GetAllPodList() (*v1.PodList, error) {
	selector := fields.SelectorFromSet(fields.Set{"spec.nodeName": ki.NodeName})
	podList, err := ki.getPodListByCondition(selector)
	if err != nil {
		hwlog.RunLog.Errorf("get pod list failed, err: %v", err)
		return nil, err
	}
	if len(podList.Items) >= common.MaxPodLimit {
		hwlog.RunLog.Error("The number of pods exceeds the upper limit")
		return nil, fmt.Errorf("pod list count invalid")
	}
	return podList, nil
}

// getPodListByCondition get pod list by field selector
func (ki *ClientK8s) getPodListByCondition(selector fields.Selector) (*v1.PodList, error) {
	return ki.Clientset.CoreV1().Pods(v1.NamespaceAll).List(context.Background(), metav1.ListOptions{
		FieldSelector:   selector.String(),
		ResourceVersion: "0",
	})
}

// checkPodList check each pod and return podList
func checkPodList(podList *v1.PodList) ([]v1.Pod, error) {
	if podList == nil {
		return nil, fmt.Errorf("pod list is invalid")
	}
	if len(podList.Items) >= common.MaxPodLimit {
		return nil, fmt.Errorf("the number of pods exceeds the upper limit")
	}
	var pods []v1.Pod
	for _, pod := range podList.Items {
		if err := common.CheckPodNameAndSpace(pod.Name, common.PodNameMaxLength); err != nil {
			hwlog.RunLog.Warnf("pod name syntax illegal, err: %#v", err)
			continue
		}
		if err := common.CheckPodNameAndSpace(pod.Namespace, common.PodNameSpaceMaxLength); err != nil {
			hwlog.RunLog.Warnf("pod namespace syntax illegal, err: %#v", err)
			continue
		}
		pods = append(pods, pod)
	}
	return pods, nil
}

// CreateConfigMap create device info, which is cm
func (ki *ClientK8s) CreateConfigMap(cm *v1.ConfigMap) (*v1.ConfigMap, error) {
	if cm == nil {
		return nil, fmt.Errorf("param cm is nil")
	}
	return ki.Clientset.CoreV1().ConfigMaps(cm.ObjectMeta.Namespace).Create(context.TODO(), cm, metav1.CreateOptions{})
}

// GetConfigMap get config map by name and namespace
func (ki *ClientK8s) GetConfigMap(cmName, cmNameSpace string) (*v1.ConfigMap, error) {
	return ki.Clientset.CoreV1().ConfigMaps(cmNameSpace).Get(context.TODO(), cmName, metav1.GetOptions{
		ResourceVersion: "0",
	})
}

// UpdateConfigMap update device info, which is cm
func (ki *ClientK8s) UpdateConfigMap(cm *v1.ConfigMap) (*v1.ConfigMap, error) {
	if cm == nil {
		return nil, fmt.Errorf("param cm is nil")
	}
	return ki.Clientset.CoreV1().ConfigMaps(cm.ObjectMeta.Namespace).Update(context.TODO(), cm, metav1.UpdateOptions{})
}

func (ki *ClientK8s) resetNodeAnnotations(node *v1.Node) {
	for k := range common.GetAllDeviceInfoTypeList() {
		delete(node.Annotations, k)
	}

	if common.ParamOption.AutoStowingDevs {
		delete(node.Labels, common.HuaweiRecoverAscend910)
		delete(node.Labels, common.HuaweiNetworkRecoverAscend910)
	}
}

// ResetDeviceInfo reset device info
func (ki *ClientK8s) ResetDeviceInfo() {
	deviceList := make(map[string]string, 1)
	if err := ki.WriteDeviceInfoDataIntoCMCache(deviceList, ""); err != nil {
		hwlog.RunLog.Errorf("write device info failed, error is %v", err)
	}
}

// ClearResetInfo clear reset info
func (ki *ClientK8s) ClearResetInfo(taskName, namespace string) error {
	taskInfo := &common.TaskResetInfo{
		RankList: make([]*common.TaskDevInfo, 0),
	}
	if _, err := ki.WriteResetInfoDataIntoCM(taskName, namespace, taskInfo); err != nil {
		hwlog.RunLog.Errorf("failed to clear reset info, err: %v", err)
		return err
	}
	return nil
}

func getNodeNameFromEnv() (string, error) {
	nodeName := os.Getenv("NODE_NAME")
	if err := checkNodeName(nodeName); err != nil {
		return "", fmt.Errorf("check node name failed: %#v", err)
	}
	return nodeName, nil
}

func checkNodeName(nodeName string) error {
	if len(nodeName) == 0 {
		return fmt.Errorf("the env variable whose key is NODE_NAME must be set")
	}
	if len(nodeName) > common.KubeEnvMaxLength {
		return fmt.Errorf("node name length %d is bigger than %d", len(nodeName), common.KubeEnvMaxLength)
	}
	pattern := common.GetPattern()["nodeName"]
	if match := pattern.MatchString(nodeName); !match {
		return fmt.Errorf("node name %s is illegal", nodeName)
	}
	return nil
}
