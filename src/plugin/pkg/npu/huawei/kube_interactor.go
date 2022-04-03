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
	nodeutil "k8s.io/kubernetes/pkg/util/node"

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

func (ki *KubeInteractor) patchAnnotationOnNode(groupAllocatableDevs map[string]string) error {
	var err error
	err = wait.PollImmediate(interval*time.Second, timeout*time.Second, func() (bool, error) {
		var node *v1.Node
		node, err = ki.clientset.CoreV1().Nodes().Get(context.Background(), ki.nodeName, metav1.GetOptions{})

		if err != nil {
			hwlog.RunLog.Errorf("failed to get node, nodeName: %s, err: %v", ki.nodeName, err)
			return false, nil
		}
		if firstTimeList {
			ki.resetNodeAnnotations(node)
		}
		newNode := node.DeepCopy()
		devType, isSingleRoutine := ki.isSingleDevType(groupAllocatableDevs)
		if isSingleRoutine {
			annotationTag := fmt.Sprintf("%s%s", resourceNamePrefix, devType)
			ki.singleDevAnnotationUpdate(annotationTag, groupAllocatableDevs, node, newNode)
		} else {
			ki.multiDevAnnotationUpdate(groupAllocatableDevs, node, newNode)
		}
		ki.addChipCoreToAnnotation(devType, newNode)
		// variables are defined in advance
		// the value will be used in subsequent assignment
		newNetworkRecoverDevSets := sets.String{}
		if devType == hiAIAscend910Prefix {
			ki.update910Annotation(node, newNode, groupAllocatableDevs, &newNetworkRecoverDevSets)
		}
		if devType == hiAIAscend710Prefix {
			ki.update710Annotation(node, newNode, groupAllocatableDevs[huaweiAscend710])
		}
		updatedNode, _, err := nodeutil.PatchNodeStatus(ki.clientset.CoreV1(), types.NodeName(ki.nodeName), node, newNode)
		if err != nil {
			hwlog.RunLog.Errorf("failed to patch volcano npu resource: %v", err)
			return false, nil
		}
		ki.atomicListenAnnotation(devType, updatedNode.Annotations)
		// if update success, update the lastTimeNetworkRecoverDevices
		// Ascend910
		if devType == hiAIAscend910Prefix {
			lastTimeNetworkRecoverDevices = newNetworkRecoverDevSets
		}
		return true, nil
	})
	return err
}

func (ki *KubeInteractor) atomicListenAnnotation(devType string, annotation map[string]string) {
	if devType == hiAIAscend310Prefix {
		return
	}
	if len(annotation) == 0 {
		return
	}
	GetAnnotationObj().WaitUpdateAnnotation = annotation
	GetAnnotationObj().IsPatchSuccess = true
}

func (ki *KubeInteractor) addChipCoreToAnnotation(devType string, newNode *v1.Node) {
	if strings.Contains(devType, hiAIAscend910Prefix) {
		newNode.Annotations[huaweiAscend910Spec] = strings.Join(Dev910PhyCoreCount, ",")
	}
	if strings.Contains(devType, hiAIAscend710Prefix) {
		newNode.Annotations[huaweiAscend710Spec] = strings.Join(Dev710PhyCoreCount, ",")
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
	node, newNode *v1.Node) {
	for tag := range groupAllocatableDevs {
		annotation, isNil := node.Annotations[tag]
		if annotationTag != tag && isNil && len(annotation) > 0 {
			newNode.Annotations[tag] = ""
		}
	}
	newNode.Annotations[annotationTag] = groupAllocatableDevs[annotationTag]
}

func (ki *KubeInteractor) isSingleDevType(groupAllocatableDevs map[string]string) (string, bool) {
	devType := hiAIAscend310Prefix
	// For Ascend310
	if len(groupAllocatableDevs) == 1 {
		return devType, true
	}
	// For Ascend910/Ascend710
	devTypeNum := 0
	for devPrefix, deviceNames := range groupAllocatableDevs {
		if len(deviceNames) != 0 {
			devType = strings.Replace(devPrefix, resourceNamePrefix, "", -1)
			devTypeNum++
		}
	}
	return devType, devTypeNum == 1
}

func (ki *KubeInteractor) resetNodeAnnotations(node *v1.Node) {
	delete(node.Annotations, huaweiUnHealthAscend910)
	delete(node.Annotations, huaweiNetworkUnHealthAscend910)
	delete(node.Annotations, huaweiAscend910)
	delete(node.Annotations, resourceNamePrefix+hiAIAscend710Prefix)
	delete(node.Annotations, resourceNamePrefix+pwr2CSuffix)
	delete(node.Annotations, resourceNamePrefix+pwr4CSuffix)
	delete(node.Annotations, resourceNamePrefix+pwr8CSuffix)
	delete(node.Annotations, resourceNamePrefix+pwr16CSuffix)
	delete(node.Annotations, resourceNamePrefix+chip710Core1C)
	delete(node.Annotations, resourceNamePrefix+chip710Core2C)
	delete(node.Annotations, resourceNamePrefix+chip710Core4C)
	delete(node.Labels, huaweiRecoverAscend910)
	delete(node.Labels, huaweiNetworkRecoverAscend910)
	firstTimeList = false
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
