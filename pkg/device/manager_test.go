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

// Package device a series of device function
package device

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/npu-exporter/devmanager"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/kubeclient"
	"Ascend-device-plugin/pkg/server"
)

func TestNewHwDevManager(t *testing.T) {
	convey.Convey("test NewHwDevManager", t, func() {
		convey.Convey("init HwDevManager", func() {
			common.ParamOption.UseVolcanoType = true
			res := NewHwDevManager(&devmanager.DeviceManagerMock{}, &kubeclient.ClientK8s{})
			convey.So(res, convey.ShouldNotBeNil)
		})
		convey.Convey("init HwDevManager get device type failed", func() {
			mockGetDevType := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)), "GetDevType",
				func(_ *devmanager.DeviceManagerMock) string {
					return "errorType"
				})
			defer mockGetDevType.Reset()
			res := NewHwDevManager(&devmanager.DeviceManagerMock{}, &kubeclient.ClientK8s{})
			convey.So(res, convey.ShouldBeNil)
		})
	})
}

func TestStartAllServer(t *testing.T) {
	convey.Convey("test startAllServer", t, func() {
		mockStart := gomonkey.ApplyMethod(reflect.TypeOf(new(server.PluginServer)), "Start",
			func(_ *server.PluginServer, socketWatcher *common.FileWatch) error {
				return fmt.Errorf("error")
			})
		defer mockStart.Reset()
		mockStart2 := gomonkey.ApplyMethod(reflect.TypeOf(new(server.PodResource)), "Start",
			func(_ *server.PodResource, socketWatcher *common.FileWatch) error {
				return fmt.Errorf("error")
			})
		defer mockStart.Reset()
		defer mockStart2.Reset()
		hdm := NewHwDevManager(&devmanager.DeviceManagerMock{}, &kubeclient.ClientK8s{})
		res := hdm.startAllServer(&common.FileWatch{})
		convey.So(res, convey.ShouldBeFalse)
	})
}

// TestUpdatePodAnnotation1 for test the interface updatePodAnnotation, part 1
func TestUpdatePodAnnotation1(t *testing.T) {
	common.ParamOption.UseVolcanoType = true
	convey.Convey("test updatePodAnnotation", t, func() {
		convey.Convey("pod source not exist", func() {
			hdm := NewHwDevManager(&devmanager.DeviceManagerMock{}, &kubeclient.ClientK8s{})
			delete(hdm.ServerMap, common.PodResourceSeverKey)
			err := hdm.updatePodAnnotation()
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("type assertion failed", func() {
			hdm := NewHwDevManager(&devmanager.DeviceManagerMock{}, &kubeclient.ClientK8s{})
			hdm.ServerMap[common.PodResourceSeverKey] = &server.PluginServer{}
			err := hdm.updatePodAnnotation()
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("GetPodResource failed", func() {
			mockGetPodResource := gomonkey.ApplyMethod(reflect.TypeOf(new(server.PodResource)), "GetPodResource",
				func(_ *server.PodResource) (map[string]server.PodDevice, error) { return nil, fmt.Errorf("err") })
			defer mockGetPodResource.Reset()
			hdm := NewHwDevManager(&devmanager.DeviceManagerMock{}, &kubeclient.ClientK8s{})
			err := hdm.updatePodAnnotation()
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("GetPodList failed", func() {
			mockGetPodResource := gomonkey.ApplyMethod(reflect.TypeOf(new(server.PodResource)), "GetPodResource",
				func(_ *server.PodResource) (map[string]server.PodDevice, error) { return nil, nil })
			defer mockGetPodResource.Reset()
			mockGetPodList := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)), "GetPodList",
				func(_ *kubeclient.ClientK8s) (*v1.PodList, error) { return nil, fmt.Errorf("err") })
			defer mockGetPodList.Reset()
			hdm := NewHwDevManager(&devmanager.DeviceManagerMock{}, &kubeclient.ClientK8s{})
			err := hdm.updatePodAnnotation()
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("GetNodeServerID failed", func() {
			mockGetPodResource := gomonkey.ApplyMethod(reflect.TypeOf(new(server.PodResource)), "GetPodResource",
				func(_ *server.PodResource) (map[string]server.PodDevice, error) { return nil, nil })
			defer mockGetPodResource.Reset()
			mockGetPodList := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)), "GetPodList",
				func(_ *kubeclient.ClientK8s) (*v1.PodList, error) { return nil, nil })
			defer mockGetPodList.Reset()
			mockGetNodeServerID := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)), "GetNodeServerID",
				func(_ *kubeclient.ClientK8s) (string, error) { return "", fmt.Errorf("err") })
			defer mockGetNodeServerID.Reset()
			hdm := NewHwDevManager(&devmanager.DeviceManagerMock{}, &kubeclient.ClientK8s{})
			err := hdm.updatePodAnnotation()
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestUpdatePodAnnotation2 for test the interface updatePodAnnotation, part 2
func TestUpdatePodAnnotation2(t *testing.T) {
	common.ParamOption.UseVolcanoType = true
	mockGetPodResource := gomonkey.ApplyMethod(reflect.TypeOf(new(server.PodResource)), "GetPodResource",
		func(_ *server.PodResource) (map[string]server.PodDevice, error) { return nil, nil })
	defer mockGetPodResource.Reset()
	mockGetPodList := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)), "GetPodList",
		func(_ *kubeclient.ClientK8s) (*v1.PodList, error) { return nil, nil })
	defer mockGetPodList.Reset()
	mockGetNodeServerID := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)), "GetNodeServerID",
		func(_ *kubeclient.ClientK8s) (string, error) { return "", nil })
	defer mockGetNodeServerID.Reset()
	convey.Convey("test updatePodAnnotation", t, func() {
		convey.Convey("not found plugin server", func() {
			hdm := NewHwDevManager(&devmanager.DeviceManagerMock{}, &kubeclient.ClientK8s{})
			delete(hdm.ServerMap, common.Ascend910)
			err := hdm.updatePodAnnotation()
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("serverMap convert failed", func() {
			hdm := NewHwDevManager(&devmanager.DeviceManagerMock{}, &kubeclient.ClientK8s{})
			hdm.ServerMap[common.Ascend910] = nil
			err := hdm.updatePodAnnotation()
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("updateSpecTypePodAnnotation failed", func() {
			mockFilterPods := gomonkey.ApplyFunc(common.FilterPods, func(pods *v1.PodList,
				blackList map[v1.PodPhase]int, deviceType string, conditionFunc func(pod *v1.Pod) bool) ([]v1.Pod,
				error) {
				return nil, fmt.Errorf("err")
			})
			defer mockFilterPods.Reset()
			hdm := NewHwDevManager(&devmanager.DeviceManagerMock{}, &kubeclient.ClientK8s{})
			err := hdm.updatePodAnnotation()
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestUpdateSpecTypePodAnnotation for test the interface updateSpecTypePodAnnotation
func TestUpdateSpecTypePodAnnotation(t *testing.T) {
	var mockPods []v1.Pod
	mockFilterPods := gomonkey.ApplyFunc(common.FilterPods, func(pods *v1.PodList, blackList map[v1.PodPhase]int,
		deviceType string, conditionFunc func(pod *v1.Pod) bool) ([]v1.Pod, error) {
		return mockPods, nil
	})
	defer mockFilterPods.Reset()
	convey.Convey("real annotation exist", t, func() {
		mockPods = []v1.Pod{{ObjectMeta: metav1.ObjectMeta{
			Annotations: map[string]string{common.PodRealAlloc: "Ascend910-0"}}}}
		hdm := NewHwDevManager(&devmanager.DeviceManagerMock{}, &kubeclient.ClientK8s{})
		err := hdm.updateSpecTypePodAnnotation(nil, common.Ascend910, "", nil, nil)
		convey.So(err, convey.ShouldBeNil)
	})
	podName := "pod-name"
	mockPods = []v1.Pod{{ObjectMeta: metav1.ObjectMeta{Name: podName}}}
	hdm := NewHwDevManager(&devmanager.DeviceManagerMock{}, &kubeclient.ClientK8s{})
	podDevice := map[string]server.PodDevice{}
	convey.Convey("pod not in pod device", t, func() {
		err := hdm.updateSpecTypePodAnnotation(nil, common.Ascend910, "", podDevice, nil)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("resource name not equal", t, func() {
		podDevice[podName] = server.PodDevice{ResourceName: common.Ascend310, DeviceIds: []string{"Ascend310-0"}}
		err := hdm.updateSpecTypePodAnnotation(nil, common.Ascend910, "", podDevice, nil)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("GetRealAllocateDevices failed", t, func() {
		mockGetNodeServerID := gomonkey.ApplyMethod(reflect.TypeOf(new(server.PluginServer)), "GetRealAllocateDevices",
			func(_ *server.PluginServer, kltAllocate []string) ([]string, error) { return nil, fmt.Errorf("error") })
		defer mockGetNodeServerID.Reset()
		podDevice[podName] = server.PodDevice{ResourceName: common.ResourceNamePrefix + common.Ascend910,
			DeviceIds: []string{"Ascend910-0"}}
		err := hdm.updateSpecTypePodAnnotation(nil, common.Ascend910, "", podDevice, nil)
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("AddPodAnnotation failed", t, func() {
		mockGetNodeServerID := gomonkey.ApplyMethod(reflect.TypeOf(new(server.PluginServer)), "GetRealAllocateDevices",
			func(_ *server.PluginServer, kltAllocate []string) ([]string, error) { return nil, nil })
		defer mockGetNodeServerID.Reset()
		mockAddPodAnnotation := gomonkey.ApplyMethod(reflect.TypeOf(new(AscendTools)), "AddPodAnnotation",
			func(_ *AscendTools, pod *v1.Pod, kltRequestDevices, dpResponseDevices []string,
				deviceType, serverID string) error {
				return fmt.Errorf("error")
			})
		defer mockAddPodAnnotation.Reset()
		err := hdm.updateSpecTypePodAnnotation(nil, common.Ascend910, "", podDevice, nil)
		convey.So(err, convey.ShouldBeNil)
	})
}
