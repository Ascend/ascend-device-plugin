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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"syscall"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/fsnotify/fsnotify"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/npu-exporter/v5/common-utils/hwlog"
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

// TestLockAllDeviceInfo for test LockAllDeviceInfo
func TestLockAllDeviceInfo(t *testing.T) {
	convey.Convey("test LockAllDeviceInfo", t, func() {
		convey.Convey("LockAllDeviceInfo success", func() {
			LockAllDeviceInfo()
			UnlockAllDeviceInfo()
		})
	})
}

// TestSetAscendRuntimeEnv for test SetAscendRuntimeEnv
func TestSetAscendRuntimeEnv(t *testing.T) {
	convey.Convey("test SetAscendRuntimeEnv", t, func() {
		id := 100
		devices := []int{id}
		resp := v1beta1.ContainerAllocateResponse{}
		SetAscendRuntimeEnv(devices, "", &resp)
		convey.So(resp.Envs[AscendVisibleDevicesEnv], convey.ShouldEqual, strconv.Itoa(id))
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

// TestSet200SocDefaultDevices for test set200SocDefaultDevices
func TestSet200SocDefaultDevices(t *testing.T) {
	convey.Convey("test set200SocDefaultDevices", t, func() {
		convey.Convey("os.Stat err", func() {
			mockStat := gomonkey.ApplyFuncReturn(os.Stat, nil, errors.New("failed"))
			defer mockStat.Reset()
			_, err := set200SocDefaultDevices()
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("device is exist", func() {
			mockStat := gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
				if name == HiAi200RCEventSched {
					return nil, fmt.Errorf("err")
				}
				return nil, nil
			})
			defer mockStat.Reset()
			ret, err := set200SocDefaultDevices()
			convey.So(ret, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestSet310BDefaultDevices for test set310BDefaultDevices
func TestSet310BDefaultDevices(t *testing.T) {
	convey.Convey("test set310BDefaultDevices", t, func() {
		convey.Convey("os.Stat err", func() {
			mockStat := gomonkey.ApplyFuncReturn(os.Stat, nil, errors.New("failed"))
			defer mockStat.Reset()
			convey.So(len(set310BDefaultDevices()), convey.ShouldEqual, 0)
		})
		convey.Convey("device is exist", func() {
			mockStat := gomonkey.ApplyFuncReturn(os.Stat, nil, nil)
			defer mockStat.Reset()
			convey.So(len(set310BDefaultDevices()), convey.ShouldNotEqual, 0)
		})
	})
}

func TestFilterPods1(t *testing.T) {
	convey.Convey("test FilterPods part1", t, func() {
		convey.Convey("The number of container exceeds the upper limit", func() {
			pods := []v1.Pod{{Spec: v1.PodSpec{Containers: make([]v1.Container, MaxContainerLimit+1)}}}
			res := FilterPods(pods, Ascend910, nil)
			convey.So(res, convey.ShouldBeEmpty)
		})
		convey.Convey("annotationTag not exist", func() {
			pods := []v1.Pod{{Spec: v1.PodSpec{Containers: []v1.Container{{Resources: v1.
				ResourceRequirements{Limits: v1.ResourceList{}}}}}}}
			res := FilterPods(pods, Ascend910, nil)
			convey.So(res, convey.ShouldBeEmpty)
		})
		convey.Convey("annotationTag exist, device is virtual", func() {
			limits := resource.NewQuantity(1, resource.DecimalExponent)
			pods := []v1.Pod{{Spec: v1.PodSpec{Containers: []v1.Container{{Resources: v1.
				ResourceRequirements{Limits: v1.ResourceList{ResourceNamePrefix + Ascend910c2: *limits}}}}}}}
			res := FilterPods(pods, Ascend910c2, nil)
			convey.So(len(res), convey.ShouldEqual, 1)
		})
		convey.Convey("limitsDevNum exceeds the upper limit", func() {
			limits := resource.NewQuantity(MaxDevicesNum*MaxAICoreNum+1, resource.DecimalExponent)
			pods := []v1.Pod{{Spec: v1.PodSpec{Containers: []v1.Container{{Resources: v1.
				ResourceRequirements{Limits: v1.ResourceList{ResourceNamePrefix + Ascend910c2: *limits}}}}}}}
			res := FilterPods(pods, Ascend910c2, nil)
			convey.So(res, convey.ShouldBeEmpty)
		})
		convey.Convey("no assigned flag", func() {
			limits := resource.NewQuantity(1, resource.DecimalExponent)
			pods := []v1.Pod{
				{Spec: v1.PodSpec{Containers: []v1.Container{{Resources: v1.ResourceRequirements{Limits: v1.
					ResourceList{ResourceNamePrefix + Ascend910: *limits}}}}}}}
			res := FilterPods(pods, Ascend910, nil)
			convey.So(res, convey.ShouldBeEmpty)
		})
		convey.Convey("had assigned flag", func() {
			limits := resource.NewQuantity(1, resource.DecimalExponent)
			pods := []v1.Pod{
				{Spec: v1.PodSpec{Containers: []v1.Container{{Resources: v1.ResourceRequirements{Limits: v1.
					ResourceList{HuaweiAscend910: *limits}}}}},
					ObjectMeta: metav1.ObjectMeta{Name: "test3", Namespace: "test3",
						Annotations: map[string]string{PodPredicateTime: "1", HuaweiAscend910: "Ascend910-1"}},
				},
			}
			res := FilterPods(pods, Ascend910, nil)
			convey.So(len(res), convey.ShouldEqual, 1)
		})
	})
}

func TestFilterPods2(t *testing.T) {
	convey.Convey("test FilterPods part2", t, func() {
		limits := resource.NewQuantity(1, resource.DecimalExponent)
		pods := []v1.Pod{
			{Spec: v1.PodSpec{Containers: []v1.Container{{Resources: v1.ResourceRequirements{Limits: v1.
				ResourceList{HuaweiAscend910: *limits}}}}},
				ObjectMeta: metav1.ObjectMeta{Name: "test3", Namespace: "test3",
					Annotations:       map[string]string{PodPredicateTime: "1", HuaweiAscend910: "Ascend910-1"},
					DeletionTimestamp: &metav1.Time{}},
				Status: v1.PodStatus{ContainerStatuses: make([]v1.ContainerStatus, 1),
					Reason: "UnexpectedAdmissionError"},
			},
		}
		convey.Convey("DeletionTimestamp is not nil", func() {
			res := FilterPods(pods, Ascend910, nil)
			convey.So(res, convey.ShouldBeEmpty)
		})
		pods[0].DeletionTimestamp = nil
		convey.Convey("The number of container status exceeds the upper limit", func() {
			pods[0].Status.ContainerStatuses = make([]v1.ContainerStatus, 1)
			res := FilterPods(pods, Ascend910, nil)
			convey.So(res, convey.ShouldBeEmpty)
		})
		convey.Convey("Waiting.Message is not nil", func() {
			pods[0].Status.ContainerStatuses = []v1.ContainerStatus{{State: v1.ContainerState{Waiting: &v1.
				ContainerStateWaiting{Message: "PreStartContainer check failed"}}}}
			res := FilterPods(pods, Ascend910, nil)
			convey.So(res, convey.ShouldBeEmpty)
		})
		convey.Convey("pod.Status.Reason is UnexpectedAdmissionError", func() {
			pods[0].Status = v1.PodStatus{ContainerStatuses: []v1.ContainerStatus{},
				Reason: "UnexpectedAdmissionError"}
			res := FilterPods(pods, Ascend910, nil)
			convey.So(res, convey.ShouldBeEmpty)
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
			_, ret := VerifyPathAndPermission("", 0)
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("os.Stat failed", func() {
			mock := gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
				return nil, fmt.Errorf("err")
			})
			defer mock.Reset()
			_, ret := VerifyPathAndPermission("./", 0)
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("filepath.EvalSymlinks failed", func() {
			mock := gomonkey.ApplyFunc(filepath.EvalSymlinks, func(path string) (string, error) {
				return "", fmt.Errorf("err")
			})
			defer mock.Reset()
			_, ret := VerifyPathAndPermission("./", 0)
			convey.So(ret, convey.ShouldBeFalse)
		})
	})
}

// TestCheckPodNameAndSpace for test CheckPodNameAndSpace
func TestCheckPodNameAndSpace(t *testing.T) {
	convey.Convey("test CheckPodNameAndSpace", t, func() {
		convey.Convey("beyond max length", func() {
			podPara, maxLength := "abc", 1
			convey.So(CheckPodNameAndSpace(podPara, maxLength), convey.ShouldNotBeNil)
		})
		convey.Convey("device is exist", func() {
			podPara, maxLength := "abc", PodNameMaxLength
			convey.So(CheckPodNameAndSpace(podPara, maxLength), convey.ShouldBeNil)
			podPara = "abc_d"
			convey.So(CheckPodNameAndSpace(podPara, maxLength), convey.ShouldNotBeNil)
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
		convey.Convey("device num excceed max num", func() {
			devices := make([]string, MaxDevicesNum+1)
			_, _, ret := GetDeviceListID(devices, "")
			convey.So(ret, convey.ShouldNotBeNil)
		})
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
			devices := map[int]string{100: DefaultDeviceIP}
			phyDevMapVirtualDev := map[int]int{100: 0}
			deviceType := "Ascend910-2c"
			ret := GetPodConfiguration(phyDevMapVirtualDev, devices, "pod-name", DefaultDeviceIP, deviceType)
			convey.So(ret, convey.ShouldBeEmpty)
		})
		convey.Convey("Marshal ok", func() {
			devices := map[int]string{100: DefaultDeviceIP}
			phyDevMapVirtualDev := map[int]int{100: 0}
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

// TestCheckFileUserSameWithProcess for test CheckFileUserSameWithProcess
func TestCheckFileUserSameWithProcess(t *testing.T) {
	convey.Convey("CheckFileUserSameWithProcess", t, func() {
		loggerPath := "/home/test"
		convey.Convey("user is root", func() {
			mockGetuid := gomonkey.ApplyFuncReturn(syscall.Getuid, RootUID)
			defer mockGetuid.Reset()
			convey.So(CheckFileUserSameWithProcess(loggerPath), convey.ShouldBeTrue)
		})
		convey.Convey("user is not root", func() {
			mockGetuid := gomonkey.ApplyFuncReturn(syscall.Getuid, 1)
			defer mockGetuid.Reset()
			convey.So(CheckFileUserSameWithProcess(loggerPath), convey.ShouldBeFalse)
		})
	})
}

// TestIsContainAtlas300IDuo for test IsContainAtlas300IDuo
func TestIsContainAtlas300IDuo(t *testing.T) {
	convey.Convey("IsContainAtlas300IDuo", t, func() {
		convey.Convey("IsContainAtlas300IDuo success", func() {
			ParamOption.ProductTypes = nil
			convey.So(IsContainAtlas300IDuo(), convey.ShouldBeFalse)
			ParamOption.ProductTypes = []string{Atlas300IDuo}
			convey.So(IsContainAtlas300IDuo(), convey.ShouldBeTrue)
		})
	})
}
