/*
Copyright(C) 2020-2021. Huawei Technologies Co.,Ltd.  All rights reserved.
*/

package huawei

import (
	"fmt"
	"huawei.com/npu-exporter/hwlog"
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

func (ki *KubeInteractor) patchAnnotationOnNode(groupAllocatableDevs map[string]string, devType string) error {
	var err error
	err = wait.PollImmediate(interval*time.Second, timeout*time.Second, func() (bool, error) {
		var node *v1.Node
		node, err = ki.clientset.CoreV1().Nodes().Get(ki.nodeName, metav1.GetOptions{})

		if err != nil {
			hwlog.Errorf("failed to get node, nodeName: %s, err: %v", ki.nodeName, err)
			return false, nil
		}

		newNode := ki.updateNodeAnnotations(devType, groupAllocatableDevs, node)
		// variables are defined in advance
		// the value will be used in subsequent assignment
		newNetworkRecoverDevSets := sets.String{}
		// Ascend910
		if devType == hiAIAscend910Prefix {
			ki.prepareAnnotationData(node, newNode, groupAllocatableDevs, &newNetworkRecoverDevSets)
		}
		_, _, err = nodeutil.PatchNodeStatus(ki.clientset.CoreV1(), types.NodeName(ki.nodeName), node, newNode)
		if err != nil {
			hwlog.Errorf("failed to patch volcano npu resource: %v", err)
			return false, nil
		}
		// if update success, update the lastTimeNetworkRecoverDevices
		// Ascend910
		if devType == hiAIAscend910Prefix {
			lastTimeNetworkRecoverDevices = newNetworkRecoverDevSets
		}
		return true, nil
	})
	return err
}

func (ki *KubeInteractor) prepareAnnotationData(node, newNode *v1.Node, groupAllocatableDevs map[string]string,
	newNetworkRecoverDevSets *sets.String) {
	newLabelsRecoverDev, newAscend910 := getUnHealthDev(totalUHDevices,
		ki.convertDevListToSets(node.Annotations[huaweiUnHealthAscend910], nodeAnnotationsDeviceSep),
		ki.convertDevListToSets(node.Labels[huaweiRecoverAscend910], nodeLabelsDeviceSep),
		ki.convertDevListToSets(groupAllocatableDevs[huaweiAscend910], nodeAnnotationsDeviceSep))

	newRecoverDevSets, newNetworkUnhealthDevSets := getNewNetworkRecoverDev(
		ki.convertDevListToSets(node.Annotations[huaweiNetworkUnHealthAscend910], nodeAnnotationsDeviceSep),
		ki.convertDevListToSets(node.Labels[huaweiNetworkRecoverAscend910], nodeLabelsDeviceSep))

	newNode.Annotations[huaweiAscend910] = newAscend910
	newNode.Annotations[huaweiUnHealthAscend910] = ki.convertSetsToString(totalUHDevices, nodeAnnotationsDeviceSep)
	newNode.Annotations[huaweiNetworkUnHealthAscend910] = ki.convertSetsToString(newNetworkUnhealthDevSets,
		nodeAnnotationsDeviceSep)
	newNode.Labels[huaweiRecoverAscend910] = ki.convertSetsToString(newLabelsRecoverDev, nodeLabelsDeviceSep)
	newNode.Labels[huaweiNetworkRecoverAscend910] = ki.convertSetsToString(newRecoverDevSets, nodeLabelsDeviceSep)

	*newNetworkRecoverDevSets = newRecoverDevSets
}

func (ki *KubeInteractor) convertDevListToSets(devices string, sepType string) sets.String {
	deviceSets := sets.String{}
	var devicesList []string
	if sepType == nodeLabelsDeviceSep {
		devicesList = strings.Split(devices, ".")
	} else {
		devicesList = strings.Split(devices, ",")
	}
	for _, device := range devicesList {
		if len(device) == 0 {
			continue
		}
		deviceSets.Insert(device)
	}
	return deviceSets
}

func (ki *KubeInteractor) convertSetsToString(annotationUHDevice sets.String, sepType string) string {
	var unHealthDevs []string
	for device := range annotationUHDevice {
		unHealthDevs = append(unHealthDevs, device)
	}
	if sepType == nodeLabelsDeviceSep {
		return strings.Join(unHealthDevs, ".")
	}
	return strings.Join(unHealthDevs, ",")
}

func (ki *KubeInteractor) updateNodeAnnotations(devType string, groupAllocatableDevs map[string]string,
	node *v1.Node) *v1.Node {
	if firstTimeList {
		delete(node.Annotations, huaweiUnHealthAscend910)
		delete(node.Annotations, huaweiNetworkUnHealthAscend910)
		delete(node.Annotations, huaweiAscend910)
		delete(node.Labels, huaweiRecoverAscend910)
		delete(node.Labels, huaweiNetworkRecoverAscend910)
		firstTimeList = false
	}
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
