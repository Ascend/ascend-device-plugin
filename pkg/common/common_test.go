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

// Package common a series of common function
package common

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"syscall"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/fsnotify/fsnotify"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/hwlog"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

// TestGetPodPhaseBlackList for test GetPodPhaseBlackList
func TestGetPodPhaseBlackList(t *testing.T) {
	convey.Convey("test GetPodPhaseBlackList", t, func() {
		ret := GetPodPhaseBlackList()
		convey.So(ret, convey.ShouldNotBeNil)
	})
}

// TestSetAscendRuntimeEnv for test SetAscendRuntimeEnv
func TestSetAscendRuntimeEnv(t *testing.T) {
	convey.Convey("test SetAscendRuntimeEnv", t, func() {
		id := "100"
		devices := []string{id}
		resp := v1beta1.ContainerAllocateResponse{}
		SetAscendRuntimeEnv(devices, "", &resp)
		convey.So(resp.Envs[ascendVisibleDevicesEnv], convey.ShouldEqual, id)
	})
}

// TestMakeDataHash for test MakeDataHash
func TestMakeDataHash(t *testing.T) {
	convey.Convey("test MakeDataHash", t, func() {
		convey.Convey("h.Write success", func() {
			DeviceInfo := NodeDeviceInfo{DeviceList: map[string]string{HuaweiUnHealthAscend910: "Ascend910-0"}}
			ret := MakeDataHash(DeviceInfo)
			convey.So(ret, convey.ShouldNotBeEmpty)
		})
		convey.Convey("json.Marshal failed", func() {
			mockMarshal := gomonkey.ApplyFunc(json.Marshal, func(v interface{}) ([]byte, error) {
				return nil, fmt.Errorf("err")
			})
			defer mockMarshal.Reset()
			DeviceInfo := NodeDeviceInfo{DeviceList: map[string]string{HuaweiUnHealthAscend910: "Ascend910-0"}}
			ret := MakeDataHash(DeviceInfo)
			convey.So(ret, convey.ShouldBeEmpty)
		})
	})
}

// TestMapDeepCopy for test MapDeepCopy
func TestMapDeepCopy(t *testing.T) {
	convey.Convey("test MapDeepCopy", t, func() {
		convey.Convey("input nil", func() {
			ret := MapDeepCopy(nil)
			convey.So(ret, convey.ShouldNotBeNil)
		})
		convey.Convey("h.Write success", func() {
			devices := map[string]string{"100": DefaultDeviceIP}
			ret := MapDeepCopy(devices)
			convey.So(len(ret), convey.ShouldEqual, len(devices))
		})
	})
}

// TestGetDeviceFromPodAnnotation for test GetDeviceFromPodAnnotation
func TestGetDeviceFromPodAnnotation(t *testing.T) {
	convey.Convey("test GetDeviceFromPodAnnotation", t, func() {
		convey.Convey("input invalid pod", func() {
			_, err := GetDeviceFromPodAnnotation(nil, Ascend910)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("annotationTag not exist", func() {
			_, err := GetDeviceFromPodAnnotation(&v1.Pod{}, Ascend910)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("annotationTag exist", func() {
			pod := v1.Pod{}
			pod.Annotations = map[string]string{ResourceNamePrefix + Ascend910: "Ascend910-0"}
			_, err := GetDeviceFromPodAnnotation(&pod, Ascend910)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func createFile(filePath string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := f.Chmod(SocketChmod); err != nil {
		return err
	}
	return nil
}

// TestGetDefaultDevices for GetDefaultDevices
func TestGetDefaultDevices(t *testing.T) {
	convey.Convey("pods is nil", t, func() {
		mockStat := gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
			return nil, fmt.Errorf("err")
		})
		defer mockStat.Reset()
		_, err := GetDefaultDevices(true)
		convey.So(err, convey.ShouldNotBeNil)
	})
	if _, err := os.Stat(HiAIHDCDevice); err != nil {
		if err = createFile(HiAIHDCDevice); err != nil {
			t.Fatal("TestGetDefaultDevices Run Failed")
		}
	}

	if _, err := os.Stat(HiAIManagerDevice); err != nil {
		if err = createFile(HiAIManagerDevice); err != nil {
			t.Fatal("TestGetDefaultDevices Run Failed")
		}
	}

	if _, err := os.Stat(HiAISVMDevice); err != nil {
		if err = createFile(HiAISVMDevice); err != nil {
			t.Fatal("TestGetDefaultDevices Run Failed")
		}
	}

	defaultDevices, err := GetDefaultDevices(true)
	if err != nil {
		t.Errorf("TestGetDefaultDevices Run Failed")
	}
	defaultMap := make(map[string]string)
	defaultMap[HiAIHDCDevice] = ""
	defaultMap[HiAIManagerDevice] = ""
	defaultMap[HiAISVMDevice] = ""
	defaultMap[HiAi200RCEventSched] = ""
	defaultMap[HiAi200RCHiDvpp] = ""
	defaultMap[HiAi200RCLog] = ""
	defaultMap[HiAi200RCMemoryBandwidth] = ""
	defaultMap[HiAi200RCSVM0] = ""
	defaultMap[HiAi200RCTsAisle] = ""
	defaultMap[HiAi200RCUpgrade] = ""

	for _, str := range defaultDevices {
		if _, ok := defaultMap[str]; !ok {
			t.Errorf("TestGetDefaultDevices Run Failed")
		}
	}
	t.Logf("TestGetDefaultDevices Run Pass")
}

// TestFilterPods1 for test FilterPods part 1
func TestFilterPods1(t *testing.T) {
	convey.Convey("test FilterPods", t, func() {
		convey.Convey("pods is nil", func() {
			_, err := FilterPods(nil, nil, Ascend910, nil)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("pod number exceeds the upper limit", func() {
			pods := v1.PodList{Items: make([]v1.Pod, MaxPodLimit+1)}
			_, err := FilterPods(&pods, nil, Ascend910, nil)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("check pod name too long", func() {
			podName := make([]byte, PodNameMaxLength+1)
			pods := v1.PodList{Items: []v1.Pod{{ObjectMeta: metav1.ObjectMeta{Name: string(podName)}}}}
			ret, err := FilterPods(&pods, nil, Ascend910, nil)
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(ret), convey.ShouldEqual, 0)
		})
		convey.Convey("check pod name not match", func() {
			podName := make([]byte, PodNameMaxLength)
			podName[0] = '$'
			pods := v1.PodList{Items: []v1.Pod{{ObjectMeta: metav1.ObjectMeta{Name: string(podName)}}}}
			ret, err := FilterPods(&pods, nil, Ascend910, nil)
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(ret), convey.ShouldEqual, 0)
		})
		convey.Convey("check pod namespace too long", func() {
			podNamespace := make([]byte, PodNameMaxLength+1)
			pods := v1.PodList{Items: []v1.Pod{{ObjectMeta: metav1.ObjectMeta{Name: "pod-name",
				Namespace: string(podNamespace)}}}}
			ret, err := FilterPods(&pods, nil, Ascend910, nil)
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(ret), convey.ShouldEqual, 0)
		})
		convey.Convey("check pod namespace not match", func() {
			podNamespace := make([]byte, PodNameSpaceMaxLength)
			podNamespace[0] = '$'
			pods := v1.PodList{Items: []v1.Pod{{ObjectMeta: metav1.ObjectMeta{Name: "pod-name",
				Namespace: string(podNamespace)}}}}
			ret, err := FilterPods(&pods, nil, Ascend910, nil)
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(ret), convey.ShouldEqual, 0)
		})
	})
}

// TestFilterPods2 for test FilterPods part 2
func TestFilterPods2(t *testing.T) {
	convey.Convey("test FilterPods", t, func() {
		convey.Convey("container number exceeds the upper limit", func() {
			pods := v1.PodList{Items: []v1.Pod{{ObjectMeta: metav1.ObjectMeta{Name: "pod-name",
				Namespace: "pod-namespace"}, Status: v1.PodStatus{Phase: v1.PodRunning},
				Spec: v1.PodSpec{Containers: make([]v1.Container, MaxContainerLimit+1)}}}}
			ret, err := FilterPods(&pods, nil, Ascend910, nil)
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(ret), convey.ShouldEqual, 0)
		})
		convey.Convey("Limits is invalid", func() {
			pods := v1.PodList{Items: []v1.Pod{{ObjectMeta: metav1.ObjectMeta{Name: "pod-name",
				Namespace: "pod-namespace"}, Status: v1.PodStatus{Phase: v1.PodRunning},
				Spec: v1.PodSpec{Containers: []v1.Container{{Resources: v1.ResourceRequirements{
					Limits: v1.ResourceList{
						ResourceNamePrefix + Ascend910: *resource.NewQuantity(MaxDevicesNum+1, "")}}}}}}}}
			ret, err := FilterPods(&pods, nil, Ascend910, nil)
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(ret), convey.ShouldEqual, 0)
		})
		convey.Convey("pod Annotations has no ascend tag", func() {
			pods := v1.PodList{Items: []v1.Pod{{ObjectMeta: metav1.ObjectMeta{Name: "pod-name",
				Namespace: "pod-namespace"}, Status: v1.PodStatus{Phase: v1.PodRunning},
				Spec: v1.PodSpec{Containers: []v1.Container{{Resources: v1.ResourceRequirements{
					Limits: v1.ResourceList{
						ResourceNamePrefix + Ascend910: *resource.NewQuantity(1, "")}}}}}}}}
			ret, err := FilterPods(&pods, nil, Ascend910, nil)
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(ret), convey.ShouldEqual, 0)
		})
		convey.Convey("deletionTimestamp is no nil", func() {
			pods := v1.PodList{Items: []v1.Pod{{
				ObjectMeta: metav1.ObjectMeta{Name: "pod-name", Namespace: "pod-namespace",
					Annotations:       map[string]string{ResourceNamePrefix + Ascend910: Ascend910 + "-0"},
					DeletionTimestamp: &metav1.Time{}},
				Status: v1.PodStatus{Phase: v1.PodRunning}, Spec: v1.PodSpec{Containers: []v1.
					Container{{Resources: v1.ResourceRequirements{Limits: v1.ResourceList{
					ResourceNamePrefix + Ascend910: *resource.NewQuantity(1, "")}}}}}}}}
			ret, err := FilterPods(&pods, nil, Ascend910, nil)
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(ret), convey.ShouldEqual, 0)
		})
	})
}

// TestFilterPods3 for test FilterPods part 3
func TestFilterPods3(t *testing.T) {
	convey.Convey("test FilterPods", t, func() {
		convey.Convey("containerStatus number exceeds the upper limit", func() {
			pods := v1.PodList{Items: []v1.Pod{{
				ObjectMeta: metav1.ObjectMeta{Name: "pod-name", Namespace: "pod-namespace",
					Annotations: map[string]string{ResourceNamePrefix + Ascend910: Ascend910 + "-0"}},
				Status: v1.PodStatus{
					Phase: v1.PodRunning, ContainerStatuses: make([]v1.ContainerStatus, MaxContainerLimit+1)},
				Spec: v1.PodSpec{Containers: []v1.Container{{Resources: v1.ResourceRequirements{
					Limits: v1.ResourceList{
						ResourceNamePrefix + Ascend910: *resource.NewQuantity(1, "")}}}}}}}}
			ret, err := FilterPods(&pods, nil, Ascend910, nil)
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(ret), convey.ShouldEqual, 0)
		})
		convey.Convey("PreStartContainer check failed", func() {
			pods := v1.PodList{Items: []v1.Pod{{
				ObjectMeta: metav1.ObjectMeta{Name: "pod-name", Namespace: "pod-namespace",
					Annotations: map[string]string{ResourceNamePrefix + Ascend910: Ascend910 + "-0"}},
				Status: v1.PodStatus{
					Phase: v1.PodRunning, ContainerStatuses: []v1.ContainerStatus{{
						State: v1.ContainerState{Waiting: &v1.ContainerStateWaiting{
							Message: "PreStartContainer check failed"}}}}},
				Spec: v1.PodSpec{Containers: []v1.Container{{Resources: v1.ResourceRequirements{
					Limits: v1.ResourceList{
						ResourceNamePrefix + Ascend910: *resource.NewQuantity(1, "")}}}}}}}}
			ret, err := FilterPods(&pods, nil, Ascend910, nil)
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(ret), convey.ShouldEqual, 0)
		})
		convey.Convey("resource is UnexpectedAdmissionError", func() {
			pods := v1.PodList{Items: []v1.Pod{{
				ObjectMeta: metav1.ObjectMeta{Name: "pod-name", Namespace: "pod-namespace",
					Annotations: map[string]string{ResourceNamePrefix + Ascend910: Ascend910 + "-0"}},
				Status: v1.PodStatus{Phase: v1.PodRunning, Reason: "UnexpectedAdmissionError"},
				Spec: v1.PodSpec{Containers: []v1.Container{{Resources: v1.ResourceRequirements{
					Limits: v1.ResourceList{
						ResourceNamePrefix + Ascend910: *resource.NewQuantity(1, "")}}}}}}}}
			ret, err := FilterPods(&pods, nil, Ascend910, nil)
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(ret), convey.ShouldEqual, 0)
		})
	})
}

// TestFilterPods4 for test FilterPods part 4
func TestFilterPods4(t *testing.T) {
	convey.Convey("test FilterPods", t, func() {
		convey.Convey("pod not meet condition", func() {
			conditionFunc := func(pod *v1.Pod) bool { return pod.Status.Phase == v1.PodRunning }
			pods := v1.PodList{Items: []v1.Pod{{ObjectMeta: metav1.ObjectMeta{Name: "pod-name",
				Namespace: "pod-namespace"}, Status: v1.PodStatus{Phase: v1.PodSucceeded}}}}
			ret, err := FilterPods(&pods, nil, Ascend910, conditionFunc)
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(ret), convey.ShouldEqual, 0)
		})
		convey.Convey("get valid pod", func() {
			pods := v1.PodList{Items: []v1.Pod{{
				ObjectMeta: metav1.ObjectMeta{Name: "pod-name", Namespace: "pod-namespace",
					Annotations: map[string]string{ResourceNamePrefix + Ascend910: Ascend910 + "-0"}},
				Status: v1.PodStatus{Phase: v1.PodRunning, Reason: ""},
				Spec: v1.PodSpec{Containers: []v1.Container{{Resources: v1.ResourceRequirements{
					Limits: v1.ResourceList{
						ResourceNamePrefix + Ascend910: *resource.NewQuantity(1, "")}}}}}}}}
			ret, err := FilterPods(&pods, nil, Ascend910, nil)
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(ret), convey.ShouldEqual, 1)
		})
		convey.Convey("pod in black list", func() {
			pods := v1.PodList{Items: []v1.Pod{{ObjectMeta: metav1.ObjectMeta{Name: "pod-name",
				Namespace: "pod-namespace"}, Status: v1.PodStatus{Phase: v1.PodSucceeded}}}}
			ret, err := FilterPods(&pods, GetPodPhaseBlackList(), Ascend910, nil)
			convey.So(err, convey.ShouldBeNil)
			convey.So(len(ret), convey.ShouldEqual, 0)
		})
	})
}

// TestVerifyPath for VerifyPath
func TestVerifyPath(t *testing.T) {
	convey.Convey("TestVerifyPath", t, func() {
		convey.Convey("filepath.Abs failed", func() {
			mock := gomonkey.ApplyFunc(filepath.Abs, func(path string) (string, error) {
				return "", fmt.Errorf("err")
			})
			defer mock.Reset()
			_, ret := VerifyPathAndPermission("")
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("os.Stat failed", func() {
			mock := gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
				return nil, fmt.Errorf("err")
			})
			defer mock.Reset()
			_, ret := VerifyPathAndPermission("./")
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("filepath.EvalSymlinks failed", func() {
			mock := gomonkey.ApplyFunc(filepath.EvalSymlinks, func(path string) (string, error) {
				return "", fmt.Errorf("err")
			})
			defer mock.Reset()
			_, ret := VerifyPathAndPermission("./")
			convey.So(ret, convey.ShouldBeFalse)
		})
	})
}

// TestWatchFile for test watchFile
func TestWatchFile(t *testing.T) {
	convey.Convey("TestWatchFile", t, func() {
		convey.Convey("fsnotify.NewWatcher ok", func() {
			watcher, err := NewFileWatch()
			convey.So(err, convey.ShouldBeNil)
			convey.So(watcher, convey.ShouldNotBeNil)
		})
		convey.Convey("fsnotify.NewWatcher failed", func() {
			mock := gomonkey.ApplyFunc(fsnotify.NewWatcher, func() (*fsnotify.Watcher, error) {
				return nil, fmt.Errorf("error")
			})
			defer mock.Reset()
			watcher, err := NewFileWatch()
			convey.So(err, convey.ShouldNotBeNil)
			convey.So(watcher, convey.ShouldBeNil)
		})
		watcher, _ := NewFileWatch()
		convey.Convey("stat failed", func() {
			convey.So(watcher, convey.ShouldNotBeNil)
			mock := gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
				return nil, fmt.Errorf("err")
			})
			defer mock.Reset()
			err := watcher.WatchFile("")
			convey.So(err, convey.ShouldNotBeNil)
		})
		mockStat := gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
			return nil, nil
		})
		defer mockStat.Reset()
		convey.Convey("Add failed", func() {
			convey.So(watcher, convey.ShouldNotBeNil)
			mockWatchFile := gomonkey.ApplyMethod(reflect.TypeOf(new(fsnotify.Watcher)), "Add",
				func(_ *fsnotify.Watcher, name string) error { return fmt.Errorf("err") })
			defer mockWatchFile.Reset()
			err := watcher.WatchFile("")
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("Add ok", func() {
			convey.So(watcher, convey.ShouldNotBeNil)
			mockWatchFile := gomonkey.ApplyMethod(reflect.TypeOf(new(fsnotify.Watcher)), "Add",
				func(_ *fsnotify.Watcher, name string) error { return nil })
			defer mockWatchFile.Reset()
			err := watcher.WatchFile("")
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestGetDeviceListID for test GetDeviceListID
func TestGetDeviceListID(t *testing.T) {
	convey.Convey("TestGetDeviceListID", t, func() {
		convey.Convey("device name is invalid", func() {
			devices := []string{"Ascend910"}
			_, _, ret := GetDeviceListID(devices, "")
			convey.So(ret, convey.ShouldNotBeNil)
		})
		convey.Convey("physical device", func() {
			devices := []string{"Ascend910-0"}
			_, ascendVisibleDevices, ret := GetDeviceListID(devices, "")
			convey.So(ret, convey.ShouldBeNil)
			convey.So(len(ascendVisibleDevices), convey.ShouldEqual, 1)
		})
		convey.Convey("virtual device", func() {
			devices := []string{"Ascend910-2c-100-0"}
			_, ascendVisibleDevices, ret := GetDeviceListID(devices, VirtualDev)
			convey.So(ret, convey.ShouldBeNil)
			convey.So(len(ascendVisibleDevices), convey.ShouldEqual, 1)
		})
	})
}

// TestGetPodConfiguration for test GetPodConfiguration
func TestGetPodConfiguration(t *testing.T) {
	convey.Convey("TestGetPodConfiguration", t, func() {
		convey.Convey("Marshal failed", func() {
			mockMarshal := gomonkey.ApplyFunc(json.Marshal, func(v interface{}) ([]byte, error) {
				return nil, fmt.Errorf("err")
			})
			defer mockMarshal.Reset()
			devices := map[string]string{"100": DefaultDeviceIP}
			phyDevMapVirtualDev := map[string]string{"100": "0"}
			deviceType := "Ascend910-2c"
			ret := GetPodConfiguration(phyDevMapVirtualDev, devices, "pod-name", DefaultDeviceIP, deviceType)
			convey.So(ret, convey.ShouldBeEmpty)
		})
		convey.Convey("Marshal ok", func() {
			devices := map[string]string{"100": DefaultDeviceIP}
			phyDevMapVirtualDev := map[string]string{"100": "0"}
			deviceType := "Ascend910-2c"
			ret := GetPodConfiguration(phyDevMapVirtualDev, devices, "pod-name", DefaultDeviceIP, deviceType)
			convey.So(ret, convey.ShouldNotBeEmpty)
		})
	})
}

// TestNewSignWatcher for test NewSignWatcher
func TestNewSignWatcher(t *testing.T) {
	convey.Convey("TestNewSignWatcher", t, func() {
		signChan := NewSignWatcher(syscall.SIGHUP)
		convey.So(signChan, convey.ShouldNotBeNil)
	})
}
