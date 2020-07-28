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
	"fmt"
	"go.uber.org/zap"
	"os"
	"strings"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	nodeutil "k8s.io/kubernetes/pkg/util/node"
)

// KubeInteractor include kubeclientSet & nodeName
type KubeInteractor struct {
	clientset *kubernetes.Clientset
	nodeName  string
}

// NewKubeClient get client from KUBECONFIG  or not
func NewKubeClient() (*kubernetes.Clientset, error) {
	clientCfg, err := clientcmd.BuildConfigFromFlags("", os.Getenv("KUBECONFIG"))
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

	return &KubeInteractor{
		clientset: client,
		nodeName:  os.Getenv("NODE_NAME"),
	}, nil
}

func (ki *KubeInteractor) getPendingPodsOnNode() ([]v1.Pod, error) {
	var (
		res []v1.Pod
		pl  *v1.PodList
		err error
	)
	selector := fields.SelectorFromSet(fields.Set{"spec.nodeName": ki.nodeName, "status.phase": string(v1.PodPending)})
	err = wait.PollImmediate(interval*time.Second, timeout*time.Second, func() (bool, error) {
		pl, err = ki.clientset.CoreV1().Pods(v1.NamespaceAll).List(metav1.ListOptions{
			FieldSelector: selector.String(),
		})
		if err != nil {
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return nil, fmt.Errorf("kube interactor timedout: %v", err)
	}

	for _, pod := range pl.Items {
		res = append(res, pod)
	}

	return res, nil
}

func (ki *KubeInteractor) patchAnnotationOnNode(allocateDevices sets.String) error {
	var err error
	err = wait.PollImmediate(interval*time.Second, timeout*time.Second, func() (bool, error) {
		var node *v1.Node
		node, err = ki.clientset.CoreV1().Nodes().Get(ki.nodeName, metav1.GetOptions{})

		if err != nil {
			logger.Error("failed to get node: ", zap.String("nodeName", ki.nodeName), zap.Error(err))
			return false, nil
		}
		var str string
		for k := range allocateDevices {
			str += k + ","
		}
		str = strings.TrimSuffix(str, ",")
		annotation, isNil := node.Annotations[huaweiAscend910]
		if checkNeedUpdate(isNil, annotation, allocateDevices) {
			newNode := node.DeepCopy()
			newNode.Annotations[huaweiAscend910] = str
			_, _, err = nodeutil.PatchNodeStatus(ki.clientset.CoreV1(), types.NodeName(ki.nodeName), node, newNode)
			if err != nil {
				logger.Error("failed to patch volcano npu resource: %v", zap.Error(err))
				return false, nil
			}
		}
		return true, nil
	})
	return err
}
func checkNeedUpdate(isNil bool, annotation string, allocate sets.String) bool {
	return !isNil || len(strings.Split(annotation, ",")) != allocate.Len() || strings.TrimSpace(annotation) == ""
}
