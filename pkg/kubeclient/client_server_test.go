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

// Package kubeclient a series of k8s function ut
package kubeclient

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/npu-exporter/common-utils/hwlog"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	"Ascend-device-plugin/pkg/common"
)

const (
	nodeNameKey     = "NODE_NAME"
	nodeNameValue   = "master"
	invalidNodeName = "abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz" +
		"abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmn" +
		"opqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz"

	npuChip310PhyID0  = "Ascend310-0"
	npuChip910PhyID0  = "Ascend910-0"
	npuChip310PPhyID0 = "Ascend310P-0"
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

func initK8S() (*ClientK8s, error) {
	return &ClientK8s{}, nil
}

// TestAnnotationReset test device info reset
func TestAnnotationReset(t *testing.T) {
	utKubeClient, err := initK8S()
	if err != nil {
		t.Fatal("TestAnnotationReset init kubernetes failed")
	}
	common.ParamOption.AutoStowingDevs = true
	convey.Convey("annotation reset with no error", t, func() {
		mockWrite, mockPatchNode, mockNode := annotationResetMock(nil, nil, nil)
		defer resetMock(mockWrite, mockPatchNode, mockNode)
		err := utKubeClient.AnnotationReset()
		convey.So(err, convey.ShouldEqual, nil)
	})
	convey.Convey("annotation reset with error", t, func() {
		mockWrite, mockPatchNode, mockNode := annotationResetMock(fmt.Errorf("can not found device info cm"),
			fmt.Errorf("patch node state failed"), nil)
		defer resetMock(mockWrite, mockPatchNode, mockNode)
		err := utKubeClient.AnnotationReset()
		convey.So(err.Error(), convey.ShouldEqual, "patch node state failed")
	})
	convey.Convey("annotation reset with get node failed", t, func() {
		mockWrite, mockPatchNode, mockNode := annotationResetMock(nil, nil, fmt.Errorf("get node failed"))
		defer resetMock(mockWrite, mockPatchNode, mockNode)
		err := utKubeClient.AnnotationReset()
		convey.So(err.Error(), convey.ShouldEqual, "get node failed")
	})
}

// TestGetNodeServerID test get node server id
func TestGetNodeServerID(t *testing.T) {
	utKubeClient, err := initK8S()
	if err != nil {
		t.Fatal("TestGetNodeServerID init kubernetes failed")
	}
	node := getMockNode(common.HuaweiAscend910, npuChip910PhyID0)
	convey.Convey("get node server id without get node", t, func() {
		mockNode := gomonkey.ApplyMethod(reflect.TypeOf(new(ClientK8s)), "GetNode",
			func(_ *ClientK8s) (*v1.Node, error) {
				return nil, fmt.Errorf("failed to get node")
			})
		defer mockNode.Reset()
		_, err := utKubeClient.GetNodeServerID()
		convey.So(err.Error(), convey.ShouldEqual, "failed to get node")
	})
	convey.Convey("get node server id", t, func() {
		mockNode := gomonkey.ApplyMethod(reflect.TypeOf(new(ClientK8s)), "GetNode",
			func(_ *ClientK8s) (*v1.Node, error) {
				return node, nil
			})
		defer mockNode.Reset()
		serverID, err := utKubeClient.GetNodeServerID()
		convey.So(err, convey.ShouldEqual, nil)
		convey.So(serverID, convey.ShouldEqual, common.DefaultDeviceIP)
	})
}

// TestGetPodsUsedNpu test used npu devices on pod
func TestGetPodsUsedNpu(t *testing.T) {
	utKubeClient, err := initK8S()
	if err != nil {
		t.Fatal("TestGetPodsUsedNpu init kubernetes failed")
	}
	podList := getMockPodList(common.HuaweiAscend310, npuChip310PhyID0)
	convey.Convey("get used npu on pods without get pod list", t, func() {
		mockPodList := gomonkey.ApplyMethod(reflect.TypeOf(new(ClientK8s)), "GetActivePodList",
			func(_ *ClientK8s) ([]v1.Pod, error) {
				return nil, fmt.Errorf("failed to get pod list")
			})
		defer mockPodList.Reset()
		useNpu := utKubeClient.GetPodsUsedNpu(common.Ascend310)
		convey.So(useNpu, convey.ShouldEqual, sets.String{})
	})
	convey.Convey("get used npu on pods", t, func() {
		mockPodList := gomonkey.ApplyMethod(reflect.TypeOf(new(ClientK8s)), "GetActivePodList",
			func(_ *ClientK8s) ([]v1.Pod, error) {
				return podList, nil
			})
		defer mockPodList.Reset()
		useNpu := utKubeClient.GetPodsUsedNpu(common.Ascend310)
		convey.So(strings.Join(useNpu.List(), ","), convey.ShouldEqual, npuChip310PhyID0)
	})
}

// TestWriteDeviceInfoDataIntoCM get cm write operation
func TestWriteDeviceInfoDataIntoCM(t *testing.T) {
	utKubeClient, err := initK8S()
	if err != nil {
		t.Fatal("TestWriteDeviceInfoDataIntoCM init kubernetes failed")
	}
	updateCM := getMockCreateCM(common.HuaweiAscend310P, npuChip310PPhyID0)
	mockCreateCM, mockUpdateCM, mockErr := mockCMOpr(updateCM)
	defer resetMock(mockErr, mockCreateCM, mockUpdateCM)
	convey.Convey("write device info (cm) when get cm failed", t, func() {
		mockGetCM := gomonkey.ApplyMethod(reflect.TypeOf(new(ClientK8s)), "GetConfigMap",
			func(_ *ClientK8s) (*v1.ConfigMap, error) {
				return nil, fmt.Errorf("test function errors")
			})
		defer mockGetCM.Reset()
		_, err := utKubeClient.WriteDeviceInfoDataIntoCM(getDeviceInfo(common.HuaweiAscend310P, npuChip310PPhyID0))
		convey.So(err, convey.ShouldEqual, nil)
	})
	convey.Convey("get write device info (cm) when get cm success", t, func() {
		mockGetCM := gomonkey.ApplyMethod(reflect.TypeOf(new(ClientK8s)), "GetConfigMap",
			func(_ *ClientK8s) (*v1.ConfigMap, error) {
				return updateCM, nil
			})
		defer mockGetCM.Reset()
		_, err := utKubeClient.WriteDeviceInfoDataIntoCM(getDeviceInfo(common.HuaweiAscend310P, npuChip310PPhyID0))
		convey.So(err, convey.ShouldEqual, nil)
	})
}

// TestTryUpdatePodAnnotation try update pod annotation
func TestTryUpdatePodAnnotation(t *testing.T) {
	utKubeClient, err := initK8S()
	if err != nil {
		t.Fatal("TestTryUpdatePodAnnotation init kubernetes failed")
	}
	testPod := getMockPod(common.HuaweiAscend910, npuChip910PhyID0)
	mockUpdatePod := gomonkey.ApplyMethod(reflect.TypeOf(new(ClientK8s)), "UpdatePod",
		func(_ *ClientK8s, _ *v1.Pod) (*v1.Pod, error) {
			return nil, fmt.Errorf("test function errors")
		})
	defer mockUpdatePod.Reset()
	convey.Convey("try update pod annotation when get pod failed", t, func() {
		mockGetPod := mockGetPodOpr(nil, fmt.Errorf("get pod failed"))
		defer mockGetPod.Reset()
		err := utKubeClient.TryUpdatePodAnnotation(testPod, nil)
		convey.So(err.Error(), convey.ShouldEqual, "update pod annotation failed, exceeded max number of retries")
	})
	convey.Convey("try update pod annotation when get pod is nil", t, func() {
		mockGetPod := mockGetPodOpr(nil, nil)
		defer mockGetPod.Reset()
		err := utKubeClient.TryUpdatePodAnnotation(testPod, nil)
		convey.So(err.Error(), convey.ShouldEqual, "update pod annotation failed, exceeded max number of retries")
	})
	convey.Convey("try update pod annotation when get pod is nil", t, func() {
		mockGetPod := mockGetPodOpr(testPod, nil)
		defer mockGetPod.Reset()
		err := utKubeClient.TryUpdatePodAnnotation(testPod,
			getDeviceInfo(common.HuaweiAscend310P, npuChip310PPhyID0))
		convey.So(err.Error(), convey.ShouldEqual, "update pod annotation failed, exceeded max number of retries")
	})
}

func getMockCreateCM(ascendType, ascendValue string) *v1.ConfigMap {
	return &v1.ConfigMap{
		Data: map[string]string{
			ascendType: ascendValue,
		},
	}
}

func getDeviceInfo(ascendType, ascendValue string) map[string]string {
	return map[string]string{
		ascendType: ascendValue,
	}
}

func getMockPod(ascendType, ascendValue string) *v1.Pod {
	annotations := make(map[string]string, 1)
	annotations[ascendType] = ascendValue
	annotations["predicate-time"] = "1626785193048251590"
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "mindx-dls-npu-1p-default-2p-0",
			Namespace:   "btg-test",
			Annotations: annotations,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{Resources: v1.ResourceRequirements{
					Limits: v1.ResourceList{
						common.HuaweiAscend910: resource.Quantity{},
					},
				}},
			},
		},
		Status: v1.PodStatus{
			Reason: "UnexpectedAdmissionError",
			ContainerStatuses: []v1.ContainerStatus{
				{State: v1.ContainerState{
					Waiting: &v1.ContainerStateWaiting{},
				}},
			},
		},
	}
}

func getMockNode(ascendType, ascendValue string) *v1.Node {
	annotations := make(map[string]string, 1)
	annotations[ascendType] = ascendValue
	labels := make(map[string]string, 1)
	labels[common.HuaweiRecoverAscend910] = "0"
	return &v1.Node{
		Status: v1.NodeStatus{
			Allocatable: v1.ResourceList{
				v1.ResourceName(ascendType): resource.Quantity{},
			},
			Addresses: getAddresses(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Annotations: annotations,
			Labels:      labels,
		},
	}
}

func getAddresses() []v1.NodeAddress {
	return []v1.NodeAddress{
		{
			Type:    v1.NodeHostName,
			Address: common.DefaultDeviceIP,
		},
		{
			Type:    v1.NodeInternalIP,
			Address: common.DefaultDeviceIP,
		},
	}
}

func getMockPodList(devType, ascendValue string) []v1.Pod {
	annotations := make(map[string]string, 1)
	annotations[devType] = ascendValue
	annotations[common.PodPredicateTime] = strconv.FormatUint(math.MaxUint64, common.BaseDec)
	containers := getContainers(devType)
	return []v1.Pod{
		getPodUTOne(annotations, containers),
		getPodUTTwo(annotations),
		getPodUTThree(),
	}
}

func getPodUTOne(annotations map[string]string, containers []v1.Container) v1.Pod {
	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "mindx-ut-1",
			Namespace:   "btg-test1",
			Annotations: annotations,
		},
		Status: v1.PodStatus{
			Phase: v1.PodPending,
		},
		Spec: v1.PodSpec{
			Containers: containers,
		},
	}
}

func getPodUTTwo(annotations map[string]string) v1.Pod {
	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "mindx-ut-2",
			Namespace:   "btg-test2",
			Annotations: annotations,
		},
		Status: v1.PodStatus{
			Phase: v1.PodSucceeded,
		},
	}
}

func getPodUTThree() v1.Pod {
	annotations := make(map[string]string, 1)
	annotations[common.HuaweiAscend310] = ""
	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "mindx-ut-3",
			Namespace:   "btg-test3",
			Annotations: annotations,
		},
		Status: v1.PodStatus{
			Phase: v1.PodPending,
		},
	}
}

func getContainers(devType string) []v1.Container {
	limits := resource.NewQuantity(1, resource.DecimalExponent)
	container := v1.Container{
		Resources: v1.ResourceRequirements{
			Limits: v1.ResourceList{
				v1.ResourceName(devType): *limits,
			},
		},
	}
	return []v1.Container{
		container,
	}
}

func mockCMOpr(updateCM *v1.ConfigMap) (*gomonkey.Patches, *gomonkey.Patches, *gomonkey.Patches) {
	mockCreateCM := gomonkey.ApplyMethod(reflect.TypeOf(new(ClientK8s)), "CreateConfigMap",
		func(_ *ClientK8s, _ *v1.ConfigMap) (*v1.ConfigMap, error) {
			return nil, fmt.Errorf("already exists")
		})
	mockUpdateCM := gomonkey.ApplyMethod(reflect.TypeOf(new(ClientK8s)), "UpdateConfigMap",
		func(_ *ClientK8s, _ *v1.ConfigMap) (*v1.ConfigMap, error) {
			return updateCM, nil
		})
	mockErr := gomonkey.ApplyMethod(reflect.TypeOf(new(ClientK8s)), "IsCMExist",
		func(_ *ClientK8s, _ error) bool {
			return true
		})
	return mockCreateCM, mockUpdateCM, mockErr
}

func mockGetPodOpr(mockPod *v1.Pod, err error) *gomonkey.Patches {
	mockGetPod := gomonkey.ApplyMethod(reflect.TypeOf(new(ClientK8s)), "GetPod",
		func(_ *ClientK8s, _ *v1.Pod) (*v1.Pod, error) {
			return mockPod, err
		})
	return mockGetPod
}

func resetMock(resetMockList ...*gomonkey.Patches) {
	for _, resetMock := range resetMockList {
		resetMock.Reset()
	}
}

func annotationResetMock(devErr, stateErr, nodeErr error) (*gomonkey.Patches, *gomonkey.Patches, *gomonkey.Patches) {
	node := getMockNode(common.HuaweiAscend910, npuChip910PhyID0)
	mockWrite := gomonkey.ApplyMethod(reflect.TypeOf(new(ClientK8s)), "WriteDeviceInfoDataIntoCM",
		func(_ *ClientK8s, _ map[string]string) (*v1.ConfigMap, error) {
			return nil, devErr
		})
	mockPatchNode := gomonkey.ApplyMethod(reflect.TypeOf(new(ClientK8s)), "PatchNodeState",
		func(_ *ClientK8s, _ *v1.Node, _ *v1.Node) (*v1.Node, []byte, error) {
			return nil, nil, stateErr
		})
	mockNode := gomonkey.ApplyMethod(reflect.TypeOf(new(ClientK8s)), "GetNode",
		func(_ *ClientK8s) (*v1.Node, error) {
			return node, nodeErr
		})
	return mockWrite, mockPatchNode, mockNode
}
