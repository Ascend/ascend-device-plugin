/* Copyright(C) 2022. Huawei Technologies Co.,Ltd. All rights reserved.
+   Licensed under the Apache License, Version 2.0 (the "License");
+   you may not use this file except in compliance with the License.
+   You may obtain a copy of the License at
+
+   http://www.apache.org/licenses/LICENSE-2.0
+
+   Unless required by applicable law or agreed to in writing, software
+   distributed under the License is distributed on an "AS IS" BASIS,
+   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
+   See the License for the specific language governing permissions and
+   limitations under the License.
+*/

// Package server holds the implementation of registration to kubelet, k8s pod resource interface.
package server

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/npu-exporter/v5/devmanager"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/device"
	"Ascend-device-plugin/pkg/kubeclient"
)

const (
	serverNum  = 2
	rqtTaskNum = 4
)

// TestTestNewHwDevManager for testTestNewHwDevManager
func TestNewHwDevManager(t *testing.T) {
	convey.Convey("test NewHwDevManager", t, func() {
		mockGetDevType := gomonkey.ApplyMethod(reflect.TypeOf(new(HwDevManager)), "UpdateServerType",
			func(_ *HwDevManager) error {
				return nil
			})
		defer mockGetDevType.Reset()
		convey.Convey("init HwDevManager", func() {
			common.ParamOption.UseVolcanoType = true
			res := NewHwDevManager(&devmanager.DeviceManagerMock{})
			convey.So(res, convey.ShouldNotBeNil)
		})
		convey.Convey("init HwDevManager get device type failed", func() {
			mockGetDevType := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)), "GetDevType",
				func(_ *devmanager.DeviceManagerMock) string {
					return "errorType"
				})
			defer mockGetDevType.Reset()
			res := NewHwDevManager(&devmanager.DeviceManagerMock{})
			convey.So(res, convey.ShouldBeNil)
		})
		convey.Convey("test NewHwDevManager, product type is not supported", func() {
			common.ParamOption.PresetVDevice = false
			common.ParamOption.ProductTypes = []string{common.Atlas300IDuo}
			res := NewHwDevManager(&devmanager.DeviceManagerMock{})
			convey.So(res, convey.ShouldBeNil)
		})
	})
}

// TestStartAllServer for testStartAllServer
func TestStartAllServer(t *testing.T) {
	mockGetDevType := gomonkey.ApplyMethod(reflect.TypeOf(new(HwDevManager)), "UpdateServerType",
		func(_ *HwDevManager) error {
			return nil
		})
	defer mockGetDevType.Reset()
	convey.Convey("test startAllServer", t, func() {
		mockStart := gomonkey.ApplyMethod(reflect.TypeOf(new(PluginServer)), "Start",
			func(_ *PluginServer, socketWatcher *common.FileWatch) error {
				return fmt.Errorf("error")
			})
		defer mockStart.Reset()
		hdm := NewHwDevManager(&devmanager.DeviceManagerMock{})
		res := hdm.startAllServer(&common.FileWatch{})
		convey.So(res, convey.ShouldBeFalse)
	})
}

// TestUpdatePodAnnotation for testUpdatePodAnnotation
func TestUpdatePodAnnotation(t *testing.T) {
	node := getMockNode(common.Ascend310P)
	podDeviceInfo := []PodDeviceInfo{
		{
			Pod:        getMockPod(),
			KltDevice:  []string{},
			RealDevice: []string{},
		},
		{
			Pod:        getMockPod(),
			KltDevice:  []string{""},
			RealDevice: []string{""},
		},
	}
	mockGetDevType := gomonkey.ApplyMethod(reflect.TypeOf(new(HwDevManager)), "UpdateServerType",
		func(_ *HwDevManager) error {
			return nil
		})
	defer mockGetDevType.Reset()
	convey.Convey("test updatePodAnnotation", t, func() {
		convey.Convey("updatePodAnnotation success", func() {
			mockNode := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)), "GetNode",
				func(_ *kubeclient.ClientK8s) (*v1.Node, error) {
					return node, nil
				})
			mockPodDeviceInfo := gomonkey.ApplyMethod(reflect.TypeOf(new(PluginServer)), "GetKltAndRealAllocateDev",
				func(_ *PluginServer, _ []v1.Pod) ([]PodDeviceInfo, error) {
					return podDeviceInfo, nil
				})
			mockManager := gomonkey.ApplyMethod(reflect.TypeOf(new(device.AscendTools)), "AddPodAnnotation",
				func(_ *device.AscendTools, _ *v1.Pod, _ []string, _ []string, _ string, _ string) error {
					return nil
				})
			mockPodList := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)), "GetActivePodList",
				func(_ *kubeclient.ClientK8s) ([]v1.Pod, error) {
					return []v1.Pod{}, nil
				})
			defer mockPodList.Reset()
			defer mockManager.Reset()
			defer mockNode.Reset()
			defer mockPodDeviceInfo.Reset()
			hdm := NewHwDevManager(&devmanager.DeviceManagerMock{})
			err := hdm.updatePodAnnotation()
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestUpdateDevice for testUpdateDevice
func TestUpdateDevice(t *testing.T) {
	mockGetDevType := gomonkey.ApplyMethod(reflect.TypeOf(new(HwDevManager)), "UpdateServerType",
		func(_ *HwDevManager) error {
			return nil
		})
	defer mockGetDevType.Reset()
	convey.Convey("test UpdateDevice", t, func() {
		convey.Convey("UpdateDevice success", func() {
			mockCheckLabel := gomonkey.ApplyMethod(reflect.TypeOf(new(device.AscendTools)),
				"CheckDeviceTypeLabel",
				func(_ *device.AscendTools) error {
					return nil
				})
			mockDestroy := gomonkey.ApplyMethod(reflect.TypeOf(new(PluginServer)), "DestroyNotUsedVNPU",
				func(_ *PluginServer) error {
					return nil
				})
			defer mockDestroy.Reset()
			defer mockCheckLabel.Reset()
			common.ParamOption.PresetVDevice = false
			hdm := NewHwDevManager(&devmanager.DeviceManagerMock{})
			hdm.ServerMap[common.AiCoreResourceName] = NewPluginServer(common.Ascend310P, nil, nil, nil)
			err := hdm.updateAllInfo()
			convey.So(err, convey.ShouldBeNil)
			common.ParamOption.PresetVDevice = true
		})
	})
}

// TestNotifyToK8s for testNotifyToK8s
func TestNotifyToK8s(t *testing.T) {
	mockGetDevType := gomonkey.ApplyMethod(reflect.TypeOf(new(HwDevManager)), "UpdateServerType",
		func(_ *HwDevManager) error {
			return nil
		})
	defer mockGetDevType.Reset()
	convey.Convey("test NotifyToK8s", t, func() {
		convey.Convey("NotifyToK8s success", func() {
			mockChange := gomonkey.ApplyMethod(reflect.TypeOf(new(device.AscendTools)), "IsDeviceStatusChange",
				func(_ *device.AscendTools, _ map[string][]*common.NpuDevice, _ []*common.NpuDevice, _ string) map[string]bool {
					return map[string]bool{common.Ascend310P: true, common.Ascend310: false}
				})
			defer mockChange.Reset()
			hdm := NewHwDevManager(&devmanager.DeviceManagerMock{})
			hdm.ServerMap[common.AiCoreResourceName] = NewPluginServer(common.Ascend310P, nil, nil, nil)
			hdm.notifyToK8s()
			convey.So(len(hdm.ServerMap), convey.ShouldEqual, serverNum)
		})
	})
}

func getMockPod() v1.Pod {
	limitValue := v1.ResourceList{
		common.HuaweiAscend910: *resource.NewQuantity(rqtTaskNum, resource.BinarySI),
	}
	annotation := map[string]string{
		common.HuaweiAscend910: "0-vir01",
	}
	return v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "mindx-dls-npu-1p-default-2p-0",
			Namespace:   "btg-test",
			Annotations: annotation,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{Resources: v1.ResourceRequirements{
					Limits: limitValue,
				}},
			},
		},
		Status: v1.PodStatus{
			Reason: "UnexpectedAdmissionError1",
			ContainerStatuses: []v1.ContainerStatus{
				{State: v1.ContainerState{
					Waiting: &v1.ContainerStateWaiting{},
				}},
			},
		},
	}
}

func getMockNode(ascendType string) *v1.Node {
	return &v1.Node{
		Status: v1.NodeStatus{
			Allocatable: v1.ResourceList{
				v1.ResourceName(ascendType): resource.Quantity{},
			},
			Addresses: getAddresses(),
		},
	}
}

func getAddresses() []v1.NodeAddress {
	return []v1.NodeAddress{
		{
			Type:    v1.NodeHostName,
			Address: common.DefaultDeviceIP,
		},
	}
}
