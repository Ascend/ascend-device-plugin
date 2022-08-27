// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package kubeclient a series of k8s function
package kubeclient

import (
	"context"
	"fmt"
	"os"
	"regexp"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/mindx/common/k8stool"
	"huawei.com/mindx/common/utils"
	"huawei.com/mindx/common/x509"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/pkg/util/node"

	"Ascend-device-plugin/pkg/common"
)

// ClientK8s include ClientK8sSet & nodeName & configmap name
type ClientK8s struct {
	Clientset      kubernetes.Interface
	NodeName       string
	DeviceInfoName string
}

// NewClientK8s create ClientK8s
func NewClientK8s(kubeConfig string) (*ClientK8s, error) {
	if kubeConfig == "" && (utils.IsExist(common.DefaultKubeConfig) || utils.IsExist(common.DefaultKubeConfigBkp)) {
		// if default path kubeConfig file is not exist means use serverAccount
		cfgInstance, err := x509.NewBKPInstance(nil, common.DefaultKubeConfig, common.DefaultKubeConfigBkp)
		if err != nil {
			return nil, err
		}
		cfgBytes, err := cfgInstance.ReadFromDisk(utils.FileMode, true)
		if err != nil || cfgBytes == nil {
			return nil, fmt.Errorf("no kubeConfig file Found")
		}
		kubeConfig = common.DefaultKubeConfig
	}
	client, err := k8stool.K8sClientFor(kubeConfig, common.Component)
	if err != nil {
		return nil, fmt.Errorf("failed to create kube client: %v", err)
	}
	nodeName, err := getNodeNameFromEnv()
	if err != nil {
		return nil, fmt.Errorf("get node name failed: %v", err)
	}

	return &ClientK8s{
		Clientset:      client,
		NodeName:       nodeName,
		DeviceInfoName: common.DeviceInfoCMNamePrefix + nodeName,
	}, nil
}

// GetNode get node
func (ki *ClientK8s) GetNode() (*v1.Node, error) {
	return ki.Clientset.CoreV1().Nodes().Get(context.Background(), ki.NodeName, metav1.GetOptions{})
}

// PatchNodeState patch node state
func (ki *ClientK8s) PatchNodeState(curNode, newNode *v1.Node) (*v1.Node, []byte, error) {
	return node.PatchNodeStatus(ki.Clientset.CoreV1(), types.NodeName(ki.NodeName), curNode, newNode)
}

// GetPod get pod by namespace and name
func (ki *ClientK8s) GetPod(pod *v1.Pod) (*v1.Pod, error) {
	return ki.Clientset.CoreV1().Pods(pod.Namespace).Get(context.Background(), pod.Name, metav1.GetOptions{})
}

// UpdatePod update pod by namespace and name
func (ki *ClientK8s) UpdatePod(pod *v1.Pod) (*v1.Pod, error) {
	return ki.Clientset.CoreV1().Pods(pod.Namespace).Update(context.Background(), pod, metav1.UpdateOptions{})
}

// GetPodList is to get pod list
func (ki *ClientK8s) GetPodList() (*v1.PodList, error) {
	selector := fields.SelectorFromSet(fields.Set{"spec.nodeName": ki.NodeName})
	return ki.Clientset.CoreV1().Pods(v1.NamespaceAll).List(context.Background(), metav1.ListOptions{
		FieldSelector: selector.String(),
	})
}

// CreateConfigMap create device info, which is cm
func (ki *ClientK8s) CreateConfigMap(cm *v1.ConfigMap) (*v1.ConfigMap, error) {
	return ki.Clientset.CoreV1().ConfigMaps(cm.ObjectMeta.Namespace).Create(context.TODO(), cm, metav1.CreateOptions{})
}

// GetConfigMap get config map
func (ki *ClientK8s) GetConfigMap() (*v1.ConfigMap, error) {
	return ki.Clientset.CoreV1().ConfigMaps(common.DeviceInfoCMNameSpace).Get(context.TODO(),
		ki.DeviceInfoName, metav1.GetOptions{})
}

// UpdateConfigMap update device info, which is cm
func (ki *ClientK8s) UpdateConfigMap(cm *v1.ConfigMap) (*v1.ConfigMap, error) {
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
	if _, err := ki.WriteDeviceInfoDataIntoCM(deviceList); err != nil {
		hwlog.RunLog.Errorf("write device info failed, error is %#v", err)
	}
}

func getNodeNameFromEnv() (string, error) {
	nodeName := os.Getenv("NODE_NAME")
	if err := checkNodeName(nodeName); err != nil {
		return "", fmt.Errorf("check node name failed: %#v", err)
	}
	return nodeName, nil
}

func checkNodeName(nodeName string) error {
	if len(nodeName) > common.KubeEnvMaxLength {
		return fmt.Errorf("node name length %d is bigger than %d", len(nodeName), common.KubeEnvMaxLength)
	}
	pattern := common.GetPattern()["nodeName"]
	if match, err := regexp.MatchString(pattern, nodeName); !match || err != nil {
		return fmt.Errorf("node name %s is illegal", nodeName)
	}
	return nil
}
