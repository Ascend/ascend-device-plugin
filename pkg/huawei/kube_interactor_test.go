/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2020-2022. All rights reserved.
 */
// Package huawei kube interactor
package huawei

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/npu-exporter/hwlog"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	"Ascend-device-plugin/pkg/common"
)

// Test310PPatchAnnotationOnNode for patch Annonation
func Test310PPatchAnnotationOnNode(t *testing.T) {
	node := getTestNode(huaweiAscend310P, "Ascend310P-0")
	mockNodeCtx := gomonkey.ApplyFunc(getNodeWithBackgroundCtx, func(_ *KubeInteractor) (*v1.Node, error) {
		return node, nil
	})
	mockState := gomonkey.ApplyFunc(patchNodeState, func(_ *KubeInteractor, _, _ *v1.Node) (*v1.Node, []byte, error) {
		return node, nil, nil
	})
	defer func() {
		mockNodeCtx.Reset()
		mockState.Reset()
	}()
	fakeKubeInteractor := &KubeInteractor{
		clientset: nil,
		nodeName:  "NODE_NAME",
	}
	if err := fakeKubeInteractor.patchAnnotationOnNode(getGroupAllocatableDevs("Ascend310P-1"), false,
		hiAIAscend310PPrefix); err != nil {
		t.Fatal(err)
	}
	t.Logf("Test310PPatchAnnotationOnNode Run Pass")
}

// Test910PatchAnnotationOnNode for patch Annonation
func Test910PatchAnnotationOnNode(t *testing.T) {
	node := getTestNode(huaweiAscend910, "Ascend910-0")
	mockNodeCtx := gomonkey.ApplyFunc(getNodeWithBackgroundCtx, func(_ *KubeInteractor) (*v1.Node, error) {
		return node, nil
	})
	mockState := gomonkey.ApplyFunc(patchNodeState, func(_ *KubeInteractor, _, _ *v1.Node) (*v1.Node, []byte, error) {
		return node, nil, nil
	})
	defer func() {
		mockNodeCtx.Reset()
		mockState.Reset()
	}()
	fakeKubeInteractor := &KubeInteractor{
		clientset: nil,
		nodeName:  "NODE_NAME",
	}
	if err := fakeKubeInteractor.patchAnnotationOnNode(getGroupAllocatableDevs("Ascend910-1"), false,
		hiAIAscend910Prefix); err != nil {
		t.Fatal(err)
	}
	t.Logf("Test910PatchAnnotationOnNode Run Pass")
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

// TestAnnotationReset test annotation reset
func TestAnnotationReset(t *testing.T) {
	hdm := setParams(false, common.RunMode310P)
	if err := hdm.GetNPUs(); err != nil {
		t.Fatal(err)
	}
	node := getTestNode(huaweiAscend310P, "Ascend310P-0")
	mockNodeCtx := gomonkey.ApplyFunc(getNodeWithBackgroundCtx, func(_ *KubeInteractor) (*v1.Node, error) {
		return node, nil
	})
	mockState := gomonkey.ApplyFunc(patchNodeState, func(_ *KubeInteractor, _, _ *v1.Node) (*v1.Node, []byte, error) {
		return node, nil, nil
	})
	devices := map[string]*common.NpuDevice{"Ascend310P": &common.NpuDevice{ID: "0", Health: "Healthy"}}
	hps := &HwPluginServe{devices: devices, hdm: hdm, devType: hiAIAscend310PPrefix}
	hps.kubeInteractor.annotationReset()
	mockNodeCtx.Reset()
	mockState.Reset()
	if node.Annotations[huaweiAscend310P] != "Ascend310P-0" {
		t.Fatal("TestAnnotationReset Run Failed")
	}
	t.Logf("TestAnnotationReset Run Pass")
}

func getGroupAllocatableDevs(ascendValue string) map[string]string {
	freeDevices := sets.NewString()
	freeDevices.Insert(ascendValue)
	return NewHwAscend910Manager().GetAnnotationMap(freeDevices, []string{hiAIAscend310PPrefix})
}

func init() {
	stopCh := make(chan struct{})
	defer close(stopCh)
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, stopCh)
}
