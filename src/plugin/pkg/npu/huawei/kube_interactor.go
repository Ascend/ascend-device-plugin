/*
Copyright 2020 The Volcano Authors.

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

package huawei

import (
	"Ascend-device-plugin/src/plugin/pkg/npu/hwlog"
	"fmt"
	"os"
	"regexp"
	"strings"
	"syscall"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	nodeutil "k8s.io/kubernetes/pkg/util/node"
)

const (
	kubeEnvMaxLength = 253

	// nodeLabelsDeviceSep if the separator between devices on labels
	nodeLabelsDeviceSep = "dot"

	// nodeAnnotationsDeviceSep if the separator between devices on annotation
	nodeAnnotationsDeviceSep = "comma"
)

// KubeInteractor include kubeclientSet & nodeName
type KubeInteractor struct {
	clientset kubernetes.Interface
	nodeName  string
}

// NewKubeClient get client from KUBECONFIG  or not
func NewKubeClient() (*kubernetes.Clientset, error) {
	kubeConfig := os.Getenv("KUBECONFIG")
	if err := checkKubeConfig(kubeConfig); err != nil {
		return nil, fmt.Errorf("check kube config failed: %v", err)
	}

	clientCfg, err := clientcmd.BuildConfigFromFlags("", kubeConfig)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(clientCfg)
	if err != nil {
		return nil, err
	}

	return clientset, nil
}

// NewKubeInteractor create KubeInteractor
func NewKubeInteractor() (*KubeInteractor, error) {
	client, err := NewKubeClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create kube client: %v", err)
	}

	nodeName := os.Getenv("NODE_NAME")
	if err := checkNodeName(nodeName); err != nil {
		return nil, fmt.Errorf("check node name failed: %v", err)
	}

	return &KubeInteractor{
		clientset: client,
		nodeName:  nodeName,
	}, nil
}

func checkKubeConfig(kubeConfig string) error {
	if len(kubeConfig) > kubeEnvMaxLength {
		return fmt.Errorf("kube config length %d is bigger than %d", len(kubeConfig), kubeEnvMaxLength)
	}
	kubeConfigPathInfo, err := os.Stat(kubeConfig)
	if err != nil || os.IsNotExist(err) {
		return nil
	}
	stat, ok := kubeConfigPathInfo.Sys().(*syscall.Stat_t)
	if !ok || stat.Uid != rootUID || stat.Gid != rootGID {
		return fmt.Errorf("non-root owner group of the path")
	}
	return nil
}

func checkNodeName(nodeName string) error {
	if len(nodeName) > kubeEnvMaxLength {
		return fmt.Errorf("node name length %d is bigger than %d", len(nodeName), kubeEnvMaxLength)
	}
	pattern := "^[a-z0-9A-Z]+([a-z0-9A-Z\\-]*)[a-z0-9A-Z]+$"
	reg := regexp.MustCompile(pattern)
	if !reg.MatchString(nodeName) {
		return fmt.Errorf("node name %s is illegal", nodeName)
	}
	return nil
}

func (ki *KubeInteractor) patchAnnotationOnNode(allocatableDevices sets.String, devType string) error {
	var err error
	err = wait.PollImmediate(interval*time.Second, timeout*time.Second, func() (bool, error) {
		var node *v1.Node
		node, err = ki.clientset.CoreV1().Nodes().Get(ki.nodeName, metav1.GetOptions{})

		if err != nil {
			hwlog.Errorf("failed to get node, nodeName: %s, err: %v", ki.nodeName, err)
			return false, nil
		}

		groupAllocatableDevs := ki.groupDevByPower(allocatableDevices)
		newNode := ki.updateNodeAnnotations(devType, groupAllocatableDevs, node)
		_, _, err = nodeutil.PatchNodeStatus(ki.clientset.CoreV1(), types.NodeName(ki.nodeName), node, newNode)
		if err != nil {
			hwlog.Errorf("failed to patch volcano npu resource: %v", err)
			return false, nil
		}
		return true, nil
	})
	return err
}

func (ki *KubeInteractor) updateNodeAnnotations(devType string, groupAllocatableDevs map[string]string,
	node *v1.Node) *v1.Node {
	newNode := node.DeepCopy()
	if devType != "" {
		annotationTag := fmt.Sprintf("%s%s", resourceNamePrefix, devType)
		newNode.Annotations[annotationTag] = groupAllocatableDevs[annotationTag]
		return newNode
	}

	for annotationTag, deviceNames := range groupAllocatableDevs {
		annotation, isNil := node.Annotations[annotationTag]
		setDevs := ki.convertStringToSet(deviceNames)
		if !checkNeedUpdate(isNil, annotation, setDevs) {
			continue
		}
		newNode.Annotations[annotationTag] = deviceNames
	}
	return newNode
}

func (ki *KubeInteractor) groupDevByPower(allocatableDevices sets.String) map[string]string {
	var pwrSuffix = []string{hiAIAscend910Prefix, pwr2CSuffix, pwr4CSuffix, pwr8CSuffix, pwr16CSuffix}
	var groupAllocatableDevs = make(map[string]string, len(pwrSuffix))
	for _, suffix := range pwrSuffix {
		powerAnnotation := ki.filterTagPowerDevice(allocatableDevices, suffix)
		annotationTag := fmt.Sprintf("%s%s", resourceNamePrefix, suffix)
		groupAllocatableDevs[annotationTag] = powerAnnotation
	}
	return groupAllocatableDevs
}

func (ki *KubeInteractor) filterTagPowerDevice(allocatableDevices sets.String, suffix string) string {
	var powerAnnotation []string
	for deviceName := range allocatableDevices {
		switch suffix {
		case hiAIAscend910Prefix:
			if !IsVirtualDev(deviceName) {
				powerAnnotation = append(powerAnnotation, deviceName)
			}
		default:
			if strings.Contains(deviceName, suffix) {
				powerAnnotation = append(powerAnnotation, deviceName)
			}
		}
	}
	return strings.Join(powerAnnotation, ",")
}

func (ki *KubeInteractor) convertStringToSet(deviceNames string) sets.String {
	setDevs := sets.NewString()
	for _, deviceName := range strings.Split(deviceNames, ",") {
		setDevs.Insert(deviceName)
	}
	return setDevs
}

func checkNeedUpdate(isNil bool, annotation string, allocatableDevices sets.String) bool {
	return !isNil || !judgeSameAscend(annotation, allocatableDevices) || strings.TrimSpace(annotation) == ""
}

func judgeSameAscend(annotation string, allocatableDevices sets.String) bool {
	annotationSet := sets.String{}
	for _, device := range strings.Split(annotation, ",") {
		annotationSet.Insert(device)
	}
	return annotationSet.Equal(allocatableDevices)
}
