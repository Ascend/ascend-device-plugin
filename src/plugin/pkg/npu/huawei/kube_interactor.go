/*
Copyright(C) 2020-2022. Huawei Technologies Co.,Ltd.  All rights reserved.
*/

package huawei

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"huawei.com/npu-exporter/hwlog"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubernetes/pkg/util/node"

	"Ascend-device-plugin/src/plugin/pkg/npu/common"
)

const (
	// nodeLabelsDeviceSep if the separator between devices on labels
	nodeLabelsDeviceSep = "dot"

	// nodeAnnotationsDeviceSep if the separator between devices on annotation
	nodeAnnotationsDeviceSep = "comma"

	// labelDeviceLen like Ascend910-0 split length is 2
	labelDeviceLen = 2
)

// KubeInteractor include kubeclientSet & nodeName
type KubeInteractor struct {
	clientset kubernetes.Interface
	nodeName  string
}

// NewKubeInteractor create KubeInteractor
func NewKubeInteractor() (*KubeInteractor, error) {
	client, err := common.NewKubeClient(kubeConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kube client: %v", err)
	}

	return &KubeInteractor{
		clientset: client,
		nodeName:  common.NodeName,
	}, nil
}

func (ki *KubeInteractor) annotationReset() {
	curNode, err := ki.clientset.CoreV1().Nodes().Get(context.Background(), ki.nodeName, metav1.GetOptions{})
	if err != nil {
		hwlog.RunLog.Errorf("failed to get node, nodeName: %s, err: %v", ki.nodeName, err)
		return
	}
	newNode := curNode.DeepCopy()
	ki.resetNodeAnnotations(newNode)
	hwlog.RunLog.Infof("newNode.Annotations: %v", newNode.Annotations)
	updatedNode, _, err := node.PatchNodeStatus(ki.clientset.CoreV1(), types.NodeName(ki.nodeName), curNode, newNode)
	if err != nil {
		hwlog.RunLog.Errorf("failed to patch volcano npu resource: %v", err)
		return
	}
	hwlog.RunLog.Infof("updatedNode.Annotations: %v", updatedNode.Annotations)
}

func (ki *KubeInteractor) patchAnnotationOnNode(groupAllocatableDevs map[string]string,
	isAlloc, isVir bool, devType, phyCoreCount string) error {
	var err error
	err = wait.PollImmediate(interval*time.Second, timeout*time.Second, func() (bool, error) {
		curNode, err := ki.clientset.CoreV1().Nodes().Get(context.Background(), ki.nodeName, metav1.GetOptions{})
		if err != nil {
			hwlog.RunLog.Errorf("failed to get node, nodeName: %s, err: %v", ki.nodeName, err)
			return false, nil
		}
		newNode := curNode.DeepCopy()
		if isAlloc {
			annotationTag := fmt.Sprintf("%s%s", resourceNamePrefix, devType)
			ki.singleDevAnnotationUpdate(annotationTag, groupAllocatableDevs, newNode)
		} else {
			ki.multiDevAnnotationUpdate(groupAllocatableDevs, curNode, newNode)
		}
		ki.addChipCoreToAnnotation(devType, phyCoreCount, newNode)
		// variables are defined in advance, the value will be used in subsequent assignment
		newNetworkRecoverDevSets := sets.String{}
		// for 910 failure rescheduling
		if devType == hiAIAscend910Prefix && !isVir {
			ki.update910Annotation(curNode, newNode, groupAllocatableDevs, &newNetworkRecoverDevSets)
		}
		if devType == hiAIAscend710Prefix && !isVir {
			ki.update710Annotation(curNode, newNode, groupAllocatableDevs[huaweiAscend710])
		}
		hwlog.RunLog.Infof("newNode.Annotations: %v", newNode.Annotations)
		updatedNode, _, err := node.PatchNodeStatus(ki.clientset.CoreV1(), types.NodeName(ki.nodeName), curNode,
			newNode)
		if err != nil {
			hwlog.RunLog.Errorf("failed to patch volcano npu resource: %v", err)
			return false, nil
		}
		hwlog.RunLog.Infof("updatedNode.Annotations: %v", updatedNode.Annotations)
		ki.atomicListenAnnotation(updatedNode.Annotations)
		// if update success, update the lastTimeNetworkRecoverDevices
		// Ascend910
		if (devType == hiAIAscend910Prefix) && !isVir {
			lastTimeNetworkRecoverDevices = newNetworkRecoverDevSets
		}
		return true, nil
	})
	return err
}

func (ki *KubeInteractor) atomicListenAnnotation(annotation map[string]string) {
	GetAnnotationObj().WaitUpdateAnnotation = annotation
	GetAnnotationObj().IsPatchSuccess.Store(true)
}

func (ki *KubeInteractor) addChipCoreToAnnotation(devType, phyCoreCount string, newNode *v1.Node) {
	if phyCoreCount == patchSpec {
		return
	}
	if strings.Contains(devType, hiAIAscend910Prefix) {
		newNode.Annotations[huaweiAscend910Spec] = phyCoreCount
	}
	if strings.Contains(devType, hiAIAscend710Prefix) {
		newNode.Annotations[huaweiAscend710Spec] = phyCoreCount
	}
}

func (ki *KubeInteractor) update910Annotation(node, newNode *v1.Node, groupAllocatableDevs map[string]string,
	newNetworkRecoverDevSets *sets.String) {

	// format recover label data
	formatedLabelRecover := changeToLongFormat(ki.convertDevListToSets(node.Labels[huaweiRecoverAscend910],
		nodeLabelsDeviceSep, common.RunMode910))
	newLabelsRecoverDev, newAscend910 := getUnHealthDev(totalUHDevices,
		ki.convertDevListToSets(node.Annotations[huaweiUnHealthAscend910],
			nodeAnnotationsDeviceSep, common.RunMode910),
		formatedLabelRecover,
		ki.convertDevListToSets(groupAllocatableDevs[huaweiAscend910],
			nodeAnnotationsDeviceSep, common.RunMode910))

	// format network recover label data
	formatedLabelNetworkRecover := changeToLongFormat(ki.convertDevListToSets(node.
		Labels[huaweiNetworkRecoverAscend910], nodeLabelsDeviceSep, common.RunMode910))
	newRecoverDevSets, newNetworkUnhealthDevSets := getNewNetworkRecoverDev(
		ki.convertDevListToSets(node.Annotations[huaweiNetworkUnHealthAscend910],
			nodeAnnotationsDeviceSep, common.RunMode910),
		formatedLabelNetworkRecover)

	// change to short format
	shortNewLabelsRecoverDev := changeToShortFormat(newLabelsRecoverDev)
	shortNewRecoverDevSets := changeToShortFormat(newRecoverDevSets)

	newNode.Annotations[huaweiAscend910] = newAscend910
	newNode.Annotations[huaweiUnHealthAscend910] = ki.convertSetsToString(totalUHDevices, nodeAnnotationsDeviceSep)
	newNode.Annotations[huaweiNetworkUnHealthAscend910] = ki.convertSetsToString(newNetworkUnhealthDevSets,
		nodeAnnotationsDeviceSep)
	newNode.Labels[huaweiRecoverAscend910] = ki.convertSetsToString(shortNewLabelsRecoverDev, nodeLabelsDeviceSep)
	newNode.Labels[huaweiNetworkRecoverAscend910] = ki.convertSetsToString(shortNewRecoverDevSets, nodeLabelsDeviceSep)

	*newNetworkRecoverDevSets = newRecoverDevSets
}

func (ki *KubeInteractor) update710Annotation(node, newNode *v1.Node, newAscend710 string) {
	_, ascend710 := getUnHealthDev(totalUHDevices,
		ki.convertDevListToSets(node.Annotations[huaweiUnHealthAscend710],
			nodeAnnotationsDeviceSep, common.RunMode710), nil,
		ki.convertDevListToSets(newAscend710, nodeAnnotationsDeviceSep, common.RunMode710))

	newNode.Annotations[huaweiAscend710] = ascend710
	newNode.Annotations[huaweiUnHealthAscend710] = ki.convertSetsToString(totalUHDevices, nodeAnnotationsDeviceSep)
}

// get elements one by one from the sets and mark the physical id "x" to "Ascend910-x"
func changeToLongFormat(chips sets.String) sets.String {
	if chips.Len() == 0 {
		return sets.String{}
	}

	newSets := sets.String{}
	for devID := range chips {
		tmpName := fmt.Sprintf("%s-%s", hiAIAscend910Prefix, devID)
		newSets.Insert(tmpName)
	}

	return newSets
}

// get elements one by one from the sets and change the element "Ascend910-x" to "x"
func changeToShortFormat(chips sets.String) sets.String {
	if chips.Len() == 0 {
		return sets.String{}
	}

	newSets := sets.String{}
	for devName := range chips {
		if len(devName) > 1 {
			idSplit := strings.Split(devName, "-")
			if len(idSplit) != labelDeviceLen {
				continue
			}
			devID := idSplit[len(idSplit)-1]
			if _, err := strconv.ParseInt(devID, baseDec, bitSize); err != nil {
				continue
			}
			newSets.Insert(devID)
		}
	}

	return newSets
}

func (ki *KubeInteractor) convertDevListToSets(devices, sepType, runMode string) sets.String {
	deviceSets := sets.String{}
	if devices == "" {
		return deviceSets
	}
	if sepType == nodeLabelsDeviceSep {
		// for label
		// check device format, must 0.1.2 and more
		for _, device := range strings.Split(devices, ".") {
			if _, err := strconv.ParseInt(device, baseDec, bitSize); err != nil {
				hwlog.RunLog.Warnf("current device id invalid, err: %v", err)
				continue
			}
			deviceSets.Insert(device)
		}
	}
	// for annotation
	// check device format, must Ascend910-0,Ascend910-1 and more
	pattern := `^Ascend910-\d+`
	if runMode == common.RunMode710 {
		pattern = `^Ascend710-\d+`
	}
	reg := regexp.MustCompile(pattern)
	for _, device := range strings.Split(devices, ",") {
		if !reg.MatchString(device) {
			hwlog.RunLog.Warnf("current device format error")
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

func (ki *KubeInteractor) multiDevAnnotationUpdate(groupAllocatableDevs map[string]string,
	node, newNode *v1.Node) {
	for annotationTag, deviceNames := range groupAllocatableDevs {
		annotation, isNil := node.Annotations[annotationTag]
		setDevs := ki.convertStringToSet(deviceNames)
		if !checkNeedUpdate(isNil, annotation, setDevs) {
			continue
		}
		newNode.Annotations[annotationTag] = deviceNames
	}
}

func (ki *KubeInteractor) singleDevAnnotationUpdate(annotationTag string, groupAllocatableDevs map[string]string,
	newNode *v1.Node) {
	newNode.Annotations[annotationTag] = groupAllocatableDevs[annotationTag]
}

func (ki *KubeInteractor) resetNodeAnnotations(node *v1.Node) {
	annotationList := []string{huaweiUnHealthAscend910, huaweiNetworkUnHealthAscend910, huaweiAscend910,
		huaweiAscend710, resourceNamePrefix + pwr2CSuffix, resourceNamePrefix + pwr4CSuffix,
		resourceNamePrefix + pwr8CSuffix, resourceNamePrefix + pwr16CSuffix, resourceNamePrefix + chip710Core1C,
		resourceNamePrefix + chip710Core2C, resourceNamePrefix + chip710Core4C, huaweiRecoverAscend910,
		huaweiNetworkRecoverAscend910}
	for _, k := range annotationList {
		if _, exist := node.Status.Allocatable[v1.ResourceName(k)]; !exist {
			delete(node.Annotations, k)
			continue
		}
		node.Annotations[k] = ""
	}
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

func (ki *KubeInteractor) patchNode(patchFunc func(*v1.Node) []byte) error {
	node, err := ki.clientset.CoreV1().Nodes().Get(context.TODO(), ki.nodeName, metav1.GetOptions{})
	if err != nil {
		hwlog.RunLog.Warnf("get node error, %v", err)
		return err
	}
	pbyte := patchFunc(node)
	_, err = ki.clientset.CoreV1().Nodes().Patch(context.TODO(), ki.nodeName, types.MergePatchType, pbyte,
		metav1.PatchOptions{})
	if err != nil {
		hwlog.RunLog.Warnf("path node error, %v", err)
		return err
	}
	return nil
}
