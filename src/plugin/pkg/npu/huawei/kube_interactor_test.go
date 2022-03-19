/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2020-2022. All rights reserved.
 */

package huawei

import (
	"Ascend-device-plugin/src/plugin/pkg/npu/huawei/mock_kubernetes"
	"Ascend-device-plugin/src/plugin/pkg/npu/huawei/mock_v1"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/context"
	"huawei.com/npu-exporter/hwlog"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
)

const nodeRunTime = 2

// TestPatchAnnotationOnNode for patch Annonation
func TestPatchAnnotationOnNode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	node1 := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{Annotations: make(map[string]string), Labels: make(map[string]string)},
	}
	node1.Annotations[huaweiAscend910] = "Ascend910-1,Ascend910-2"
	mockK8s := mock_kubernetes.NewMockInterface(ctrl)
	mockV1 := mock_v1.NewMockCoreV1Interface(ctrl)
	mockNode := mock_v1.NewMockNodeInterface(ctrl)
	mockNode.EXPECT().Get(context.Background(), gomock.Any(), metav1.GetOptions{}).Return(node1, nil)
	mockNode.EXPECT().Patch(context.Background(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(node1, nil)
	mockV1.EXPECT().Nodes().Return(mockNode).Times(nodeRunTime)
	mockK8s.EXPECT().CoreV1().Return(mockV1).Times(nodeRunTime)
	freeDevices := sets.NewString()
	freeDevices.Insert("Ascend910-1")
	freeDevices.Insert("Ascend910-5")
	fakeKubeInteractor := &KubeInteractor{
		clientset: mockK8s,
		nodeName:  "NODE_NAME",
	}

	groupAllocatableDevs := NewHwAscend910Manager().GetAnnotationMap(freeDevices, "Ascend910")
	err := fakeKubeInteractor.patchAnnotationOnNode(groupAllocatableDevs)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("TestPatchAnnotationOnNode Run Pass")
}

// TestChangeLabelFormat for test label format
func TestChangeLabelFormat(t *testing.T) {
	convey.Convey("format change", t, func() {
		convey.Convey("empty sets", func() {
			emptySets := changeToShortFormat(sets.String{})
			emptySets2 := changeToLongFormat(sets.String{})
			convey.So(emptySets, convey.ShouldBeEmpty)
			convey.So(emptySets2, convey.ShouldBeEmpty)
		})
		convey.Convey("long format", func() {
			shortSets := sets.String{}
			shortSets.Insert("1")
			longSets := changeToLongFormat(shortSets)
			convey.So(longSets, convey.ShouldEqual, sets.String{"Ascend910-1": sets.Empty{}})
		})
		convey.Convey("short format", func() {
			longSets := sets.String{}
			longSets.Insert("Ascend910-1")
			shortSets := changeToShortFormat(longSets)
			convey.So(shortSets, convey.ShouldEqual, sets.String{"1": sets.Empty{}})
		})
	})
}

func init() {
	stopCh := make(chan struct{})
	defer close(stopCh)
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, stopCh)
}
