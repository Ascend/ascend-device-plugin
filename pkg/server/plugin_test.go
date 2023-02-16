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

// Package server holds the implementation of registration to kubelet, k8s device plugin interface and grpc service.
package server

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/npu-exporter/v3/common-utils/hwlog"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/device"
	"Ascend-device-plugin/pkg/kubeclient"
)

var (
	devices = []*common.NpuDevice{
		{DevType: common.Ascend910, DeviceName: "Ascend910-0", Health: "Healthy"},
		{DevType: common.Ascend910, DeviceName: "Ascend910-1", Health: "Healthy"},
		{DevType: common.Ascend910, DeviceName: "Ascend910-2", Health: "Healthy"},
		{DevType: common.Ascend910, DeviceName: "Ascend910-3", Health: "Healthy"},
		{DevType: common.Ascend910, DeviceName: "Ascend910-4", Health: "Healthy"},
		{DevType: common.Ascend910, DeviceName: "Ascend910-5", Health: "Healthy"},
		{DevType: common.Ascend910, DeviceName: "Ascend910-6", Health: "Healthy"},
		{DevType: common.Ascend910, DeviceName: "Ascend910-7", Health: "Healthy"},
	}
	mockPods = []v1.Pod{
		{ObjectMeta: metav1.ObjectMeta{Name: "test1", Namespace: "test1"}},
		{ObjectMeta: metav1.ObjectMeta{Name: "test2", Namespace: "test2",
			Annotations: map[string]string{common.PodPredicateTime: "abcdef"}}},
		{ObjectMeta: metav1.ObjectMeta{Name: "test3", Namespace: "test3", Annotations: map[string]string{common.
			PodPredicateTime: "1", common.HuaweiAscend910: "Ascend910-1"}}},
		{ObjectMeta: metav1.ObjectMeta{Name: "test4", Namespace: "test4", Annotations: map[string]string{common.
			PodPredicateTime: "4", common.HuaweiAscend910: "Ascend910-2"}}},
		{ObjectMeta: metav1.ObjectMeta{Name: "test5", Namespace: "test5", Annotations: map[string]string{common.
			PodPredicateTime: "5", common.ResourceNamePrefix + common.Ascend910c2: "Ascend910-2c-180-3"}}},
	}
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
	common.ParamOption.PresetVDevice = true
}

// TestListAndWatch for test the interface ListAndWatch
func TestListAndWatch(t *testing.T) {
	ps := NewPluginServer(common.Ascend910, nil, nil, nil)
	convey.Convey("test ListAndWatch", t, func() {
		mockSend := gomonkey.ApplyFunc(sendToKubelet, func(stream v1beta1.DevicePlugin_ListAndWatchServer,
			resp *v1beta1.ListAndWatchResponse) error {
			return nil
		})
		convey.Convey("Notify false", func() {
			ret := ps.Notify(devices)
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("Notify true", func() {
			go ps.ListAndWatch(&v1beta1.Empty{}, nil)
			time.Sleep(time.Second)
			ret := ps.Notify(devices)
			convey.So(ret, convey.ShouldBeTrue)
			convey.So(len(ps.cachedDevices), convey.ShouldEqual, len(devices))
			for i, id := range devices {
				convey.So(id.DeviceName, convey.ShouldEqual, ps.cachedDevices[i].DeviceName)
				convey.So(id.Health, convey.ShouldEqual, ps.cachedDevices[i].Health)
			}
			ps.stopListAndWatch()
		})
		mockSend.Reset()
	})
}

// TestUpdateAllocMap for test the updateAllocMap
func TestUpdateAllocMap(t *testing.T) {
	ps := NewPluginServer(common.Ascend910, devices, nil, nil)
	convey.Convey("length no equal", t, func() {
		realAlloc := []string{"Ascend910-0", "Ascend910-2", "Ascend910-1"}
		kltAlloc := []string{"Ascend910-2", "Ascend910-7", "Ascend910-0", "Ascend910-1"}
		ps.updateAllocMap(realAlloc, kltAlloc)
		convey.So(len(ps.klt2RealDevMap), convey.ShouldEqual, 0)
	})
	convey.Convey("update map", t, func() {
		realAlloc := []string{"Ascend910-0", "Ascend910-2", "Ascend910-1", "Ascend910-3"}
		kltAlloc := []string{"Ascend910-2", "Ascend910-7", "Ascend910-0", "Ascend910-1"}
		ps.updateAllocMap(realAlloc, kltAlloc)
		convey.So(len(ps.klt2RealDevMap), convey.ShouldEqual, len(realAlloc))
		for i, id := range kltAlloc {
			v, exist := ps.klt2RealDevMap[id]
			convey.So(exist, convey.ShouldBeTrue)
			convey.So(v, convey.ShouldEqual, realAlloc[i])
		}
	})
	convey.Convey("update duplicate device", t, func() {
		lastLength := len(ps.klt2RealDevMap)
		realAlloc := []string{"Ascend910-4"}
		kltAlloc := []string{"Ascend910-2"}
		ps.updateAllocMap(realAlloc, kltAlloc)
		convey.So(len(ps.klt2RealDevMap), convey.ShouldEqual, lastLength)
		for i, id := range kltAlloc {
			v, exist := ps.klt2RealDevMap[id]
			convey.So(exist, convey.ShouldBeTrue)
			convey.So(v, convey.ShouldEqual, realAlloc[i])
		}
	})
}

// TestGenerateAllDeviceMap for test the generateAllDeviceMap
func TestGenerateAllDeviceMap(t *testing.T) {
	ps := NewPluginServer(common.Ascend910, devices, nil, nil)
	convey.Convey("length no equal", t, func() {
		ps.deepCopyDevice(devices)
		realAlloc := []string{"Ascend910-0", "Ascend910-2", "Ascend910-1", "Ascend910-3"}
		kltAlloc := []string{"Ascend910-2", "Ascend910-7", "Ascend910-0", "Ascend910-1"}
		ps.updateAllocMap(realAlloc, kltAlloc)
		expectMap := map[string]string{
			"Ascend910-4": "Ascend910-3", "Ascend910-5": "Ascend910-4", "Ascend910-6": "Ascend910-5",
			"Ascend910-7": "Ascend910-6",
		}
		actualMap := ps.generateAllDeviceMap()
		convey.So(len(ps.klt2RealDevMap), convey.ShouldEqual, len(expectMap))
		for k, v := range expectMap {
			id, exist := actualMap[k]
			convey.So(exist, convey.ShouldBeTrue)
			convey.So(id, convey.ShouldEqual, v)
		}
	})
}

// TestResponseToKubelet for test the responseToKubelet
func TestResponseToKubelet(t *testing.T) {
	ps := NewPluginServer(common.Ascend910, devices, nil, nil)
	convey.Convey("use volcano", t, func() {
		common.ParamOption.UseVolcanoType = true
		ps.deepCopyDevice(devices)
		ps.klt2RealDevMap = map[string]string{
			"Ascend910-4": "Ascend910-3", "Ascend910-5": "Ascend910-4", "Ascend910-6": "Ascend910-5",
			"Ascend910-7": "Ascend910-6", "Ascend910-0": "Ascend910-7", "Ascend910-1": "Ascend910-2",
			"Ascend910-2": "Ascend910-1", "Ascend910-3": "Ascend910-0",
		}
		resp := ps.responseToKubelet()
		convey.So(resp, convey.ShouldNotBeNil)
		convey.So(len(resp.Devices), convey.ShouldEqual, len(ps.cachedDevices))
		for i, id := range ps.cachedDevices {
			convey.So(id.DeviceName, convey.ShouldEqual, ps.klt2RealDevMap[resp.Devices[i].ID])
			convey.So(id.Health, convey.ShouldEqual, ps.cachedDevices[i].Health)
		}
	})
}

// TestAllocateRequestPhysicalDevice for test the Allocate request physical device
func TestAllocateRequestPhysicalDevice(t *testing.T) {
	ps := NewPluginServer(common.Ascend910, devices, nil, nil)
	common.ParamOption.UseVolcanoType = false
	var requests v1beta1.AllocateRequest
	convey.Convey("invalid request", t, func() {
		convey.Convey("input nil", func() {
			_, err := ps.Allocate(context.Background(), nil)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("container num exceeds the upper limit", func() {
			requests.ContainerRequests = make([]*v1beta1.ContainerAllocateRequest, common.MaxContainerLimit+1)
			_, err := ps.Allocate(context.Background(), &requests)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("devices num exceeds the upper limit", func() {
			requests.ContainerRequests = []*v1beta1.ContainerAllocateRequest{{DevicesIDs: make([]string,
				common.MaxDevicesNum+1)}}
			_, err := ps.Allocate(context.Background(), &requests)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("request physical device not exist", func() {
			ps.deepCopyDevice(devices)
			requests.ContainerRequests = []*v1beta1.ContainerAllocateRequest{{DevicesIDs: []string{"Ascend910-8"}}}
			_, err := ps.Allocate(context.Background(), &requests)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("request physical device exist", func() {
			ps.deepCopyDevice(devices)
			deviceID := "1"
			requests.ContainerRequests = []*v1beta1.
				ContainerAllocateRequest{{DevicesIDs: []string{"Ascend910-" + deviceID}}}
			resp, err := ps.Allocate(context.Background(), &requests)
			convey.So(err, convey.ShouldBeNil)
			convey.So(resp, convey.ShouldNotBeNil)
			convey.So(len(resp.ContainerResponses), convey.ShouldEqual, 1)
			convey.So(resp.ContainerResponses[0].Envs["ASCEND_VISIBLE_DEVICES"], convey.ShouldEqual, deviceID)
			convey.So(resp.ContainerResponses[0].Envs["ASCEND_RUNTIME_OPTIONS"], convey.ShouldBeEmpty)
		})
	})
}

// TestAllocateRequestVirtualDevice for test the Allocate request virtual device
func TestAllocateRequestVirtualDevice(t *testing.T) {
	common.ParamOption.UseVolcanoType = false
	ps := NewPluginServer(common.Ascend910c2, devices, nil, nil)
	var requests v1beta1.AllocateRequest
	convey.Convey("invalid request", t, func() {
		convey.Convey("request more than 1 virtual device", func() {
			ps.cachedDevices = []common.NpuDevice{{DevType: common.Ascend910c2, DeviceName: "Ascend910-2c-100-0"}}
			requests.ContainerRequests = []*v1beta1.
				ContainerAllocateRequest{{DevicesIDs: []string{"Ascend910-2c-100-0", "Ascend910-2c-100-1"}}}
			_, err := ps.Allocate(context.Background(), &requests)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("request virtual device not exist", func() {
			ps.cachedDevices = []common.NpuDevice{{DevType: common.Ascend910c2, DeviceName: "Ascend910-2c-100-0"}}
			requests.ContainerRequests = []*v1beta1.
				ContainerAllocateRequest{{DevicesIDs: []string{"Ascend910-2c-100-1"}}}
			_, err := ps.Allocate(context.Background(), &requests)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("request virtual device exist", func() {
			deviceID := "100"
			ps := NewPluginServer(common.Ascend910c2, devices, nil, nil)
			ps.cachedDevices = []common.NpuDevice{{DevType: common.Ascend910c2,
				DeviceName: "Ascend910-2c-" + deviceID + "-0"}}
			requests.ContainerRequests = []*v1beta1.
				ContainerAllocateRequest{{DevicesIDs: []string{"Ascend910-2c-" + deviceID + "-0"}}}
			resp, err := ps.Allocate(context.Background(), &requests)
			convey.So(err, convey.ShouldBeNil)
			convey.So(resp, convey.ShouldNotBeNil)
			convey.So(len(resp.ContainerResponses), convey.ShouldEqual, 1)
			convey.So(resp.ContainerResponses[0].Envs["ASCEND_VISIBLE_DEVICES"], convey.ShouldEqual, deviceID)
			convey.So(resp.ContainerResponses[0].Envs["ASCEND_RUNTIME_OPTIONS"], convey.ShouldEqual, common.VirtualDev)
		})
	})
}

// TestAllocateWithVolcano1 for test the Allocate request physical device with volcano, not get valid oldest pod
func TestAllocateWithVolcano1(t *testing.T) {
	ps := NewPluginServer(common.Ascend910, devices, nil, device.NewHwAscend910Manager())
	common.ParamOption.UseVolcanoType = true
	var requests v1beta1.AllocateRequest
	requests.ContainerRequests = []*v1beta1.ContainerAllocateRequest{{DevicesIDs: []string{"Ascend910-0"}}}
	convey.Convey("with volcano", t, func() {
		convey.Convey("GetPodList failed", func() {
			mock := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)), "GetActivePodList",
				func(_ *kubeclient.ClientK8s) ([]v1.Pod, error) { return nil, fmt.Errorf("err") })
			defer mock.Reset()
			_, err := ps.Allocate(context.Background(), &requests)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("oldestPod is nil", func() {
			mockGetPodList := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)), "GetActivePodList",
				func(_ *kubeclient.ClientK8s) ([]v1.Pod, error) { return mockPods, nil })
			defer mockGetPodList.Reset()
			mockFilter := gomonkey.ApplyFunc(common.FilterPods, func(pods []v1.Pod, deviceType string,
				conditionFunc func(pod *v1.Pod) bool) []v1.Pod {
				return nil
			})
			defer mockFilter.Reset()
			_, err := ps.Allocate(context.Background(), &requests)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestAllocateWithVolcano2 for test the Allocate request physical device with volcano, get oldest pod
func TestAllocateWithVolcano2(t *testing.T) {
	ps := NewPluginServer(common.Ascend910, devices, []string{common.HiAIManagerDevice},
		device.NewHwAscend910Manager())
	common.ParamOption.UseVolcanoType = true
	var requests v1beta1.AllocateRequest
	requests.ContainerRequests = []*v1beta1.ContainerAllocateRequest{{DevicesIDs: []string{"Ascend910-0"}}}
	convey.Convey("test AllocateWithVolcano", t, func() {
		mockGetPodList := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)), "GetActivePodList",
			func(_ *kubeclient.ClientK8s) ([]v1.Pod, error) { return mockPods, nil })
		defer mockGetPodList.Reset()
		mockFilter := gomonkey.ApplyFunc(common.FilterPods, func(pods []v1.Pod, deviceType string,
			conditionFunc func(pod *v1.Pod) bool) []v1.Pod {
			return []v1.Pod{{ObjectMeta: metav1.ObjectMeta{Name: "test",
				Annotations: map[string]string{common.PodPredicateTime: "5",
					common.HuaweiAscend910: "Ascend910-0"}}}}
		})
		defer mockFilter.Reset()
		convey.Convey("TryUpdatePodAnnotation failed", func() {
			mockTryUpdatePodAnnotation := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)),
				"TryUpdatePodAnnotation", func(_ *kubeclient.ClientK8s, _ *v1.Pod, _ map[string]string) error {
					return fmt.Errorf("err")
				})
			defer mockTryUpdatePodAnnotation.Reset()
			_, err := ps.Allocate(context.Background(), &requests)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("common.GetDeviceFromPodAnnotation failed", func() {
			mockTryUpdatePodAnnotation := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)),
				"TryUpdatePodAnnotation", func(_ *kubeclient.ClientK8s, _ *v1.Pod, _ map[string]string) error {
					return nil
				})
			defer mockTryUpdatePodAnnotation.Reset()
			mockFilter := gomonkey.ApplyFunc(common.FilterPods, func(pods []v1.Pod, deviceType string,
				conditionFunc func(pod *v1.Pod) bool) []v1.Pod {
				return []v1.Pod{{ObjectMeta: metav1.ObjectMeta{Name: "test",
					Annotations: map[string]string{common.PodPredicateTime: "5",
						common.ResourceNamePrefix + common.Ascend910c2: "Ascend910-2c-180-3"}}}}
			})
			defer mockFilter.Reset()
			_, err := ps.Allocate(context.Background(), &requests)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestAllocateWithVolcano3 for test the Allocate request physical device with volcano, part 3
func TestAllocateWithVolcano3(t *testing.T) {
	ps := NewPluginServer(common.Ascend910, devices, []string{common.HiAIManagerDevice},
		device.NewHwAscend910Manager())
	common.ParamOption.UseVolcanoType = true
	var requests v1beta1.AllocateRequest
	requests.ContainerRequests = []*v1beta1.ContainerAllocateRequest{{DevicesIDs: []string{"Ascend910-0"}}}
	convey.Convey("test AllocateWithVolcano", t, func() {
		mockGetPodList := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)), "GetActivePodList",
			func(_ *kubeclient.ClientK8s) ([]v1.Pod, error) { return mockPods, nil })
		defer mockGetPodList.Reset()
		mockTryUpdatePodAnnotation := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)),
			"TryUpdatePodAnnotation", func(_ *kubeclient.ClientK8s, _ *v1.Pod, _ map[string]string) error {
				return nil
			})
		defer mockTryUpdatePodAnnotation.Reset()
		convey.Convey("with volcano GetDeviceListID failed", func() {
			mockFilter := gomonkey.ApplyFunc(common.FilterPods, func(pods []v1.Pod, deviceType string,
				conditionFunc func(pod *v1.Pod) bool) []v1.Pod {
				return []v1.Pod{{ObjectMeta: metav1.ObjectMeta{Name: "test",
					Annotations: map[string]string{common.PodPredicateTime: "5",
						common.HuaweiAscend910: "Ascend910"}}}}
			})
			defer mockFilter.Reset()
			_, err := ps.Allocate(context.Background(), &requests)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("with volcano run ok", func() {
			mockFilter := gomonkey.ApplyFunc(common.FilterPods, func(pods []v1.Pod, deviceType string,
				conditionFunc func(pod *v1.Pod) bool) []v1.Pod {
				return mockPods
			})
			defer mockFilter.Reset()
			resp, err := ps.Allocate(context.Background(), &requests)
			convey.So(err, convey.ShouldBeNil)
			convey.So(resp, convey.ShouldNotBeNil)
			convey.So(len(resp.ContainerResponses), convey.ShouldEqual, 1)
			convey.So(resp.ContainerResponses[0].Envs["ASCEND_VISIBLE_DEVICES"], convey.ShouldEqual, "1")
			_, err = ps.GetRealAllocateDevices([]string{"Ascend910-2"})
			convey.So(err, convey.ShouldNotBeNil)
			realAllocate, err := ps.GetRealAllocateDevices([]string{"Ascend910-0"})
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(realAllocate), convey.ShouldEqual, 1)
			convey.So(realAllocate[0], convey.ShouldEqual, "Ascend910-1")
		})
	})
}

// TestGetUnhealthyAICore for testGetUnhealthyAICore
func TestGetUnhealthyAICore(t *testing.T) {
	ps := NewPluginServer(common.Ascend910, devices, []string{common.HiAIManagerDevice},
		device.NewHwAscend910Manager())
	ps.klt2RealDevMap["Ascend910-0"] = "Ascend910-0"
	common.ParamOption.AiCoreCount = common.MinAICoreNum
	convey.Convey("test GetUnhealthyAICore", t, func() {
		convey.Convey("GetUnhealthyAICore success", func() {
			unhealthyDev := ps.getUnhealthyAICore()
			convey.So(len(unhealthyDev), convey.ShouldEqual, 0)
		})
	})
}

// TestDestroyNotUsedVNPU for testDestroyNotUsedVNPU
func TestDestroyNotUsedVNPU(t *testing.T) {
	ps := NewPluginServer(common.Ascend910, devices, []string{common.HiAIManagerDevice},
		device.NewHwAscend910Manager())
	ps.klt2RealDevMap["Ascend910-0"] = "Ascend910-0"
	common.ParamOption.AiCoreCount = common.MinAICoreNum
	mockGetNPUs := gomonkey.ApplyMethod(reflect.TypeOf(new(device.HwAscend910Manager)), "GetNPUs",
		func(_ *device.HwAscend910Manager) (common.NpuAllInfo, error) {
			return common.NpuAllInfo{}, nil
		})
	mockDestroy := gomonkey.ApplyMethod(reflect.TypeOf(new(device.AscendTools)), "DestroyVirtualDevice",
		func(_ *device.AscendTools, _ string) error {
			return nil
		})
	mockAllocateDev := gomonkey.ApplyMethod(reflect.TypeOf(new(PluginServer)), "GetKltAndRealAllocateDev",
		func(_ *PluginServer) ([]PodDeviceInfo, error) {
			return []PodDeviceInfo{}, nil
		})
	defer mockDestroy.Reset()
	defer mockAllocateDev.Reset()
	defer mockGetNPUs.Reset()
	convey.Convey("test DestroyNotUsedVNPU", t, func() {
		convey.Convey("DestroyNotUsedVNPU success", func() {
			err := ps.DestroyNotUsedVNPU()
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestDoWithVolcanoSchedule for testDoWithVolcanoSchedule
func TestDoWithVolcanoSchedule(t *testing.T) {
	ps := NewPluginServer(common.Ascend910, devices, []string{common.HiAIManagerDevice},
		device.NewHwAscend910Manager())
	devicesIDs := []string{""}
	podList := getMockPodList()
	common.ParamOption.PresetVDevice = false
	mockActivePodList := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)),
		"GetActivePodList", func(_ *kubeclient.ClientK8s) ([]v1.Pod, error) {
			return podList, nil
		})
	mockUpdatePod := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)),
		"TryUpdatePodAnnotation", func(_ *kubeclient.ClientK8s, pod *v1.Pod,
			annotation map[string]string) error {
			return nil
		})
	mockDestroy := gomonkey.ApplyMethod(reflect.TypeOf(new(PluginServer)), "DestroyNotUsedVNPU",
		func(_ *PluginServer) error {
			return nil
		})
	mockCreate := gomonkey.ApplyMethod(reflect.TypeOf(new(device.AscendTools)),
		"CreateVirtualDevice", func(_ *device.AscendTools, phyID int32, templateName string) (string, error) {
			return "Ascend910-2c-100-0", nil
		})
	defer mockCreate.Reset()
	defer mockDestroy.Reset()
	defer mockUpdatePod.Reset()
	defer mockActivePodList.Reset()
	convey.Convey("test DoWithVolcanoSchedule", t, func() {
		convey.Convey("DoWithVolcanoSchedule success", func() {
			_, err := ps.useVolcano(devicesIDs)
			convey.So(err, convey.ShouldBeNil)
		})
	})
	common.ParamOption.PresetVDevice = true
}

func getMockPodList() []v1.Pod {
	return []v1.Pod{
		getMockPod(),
	}
}
