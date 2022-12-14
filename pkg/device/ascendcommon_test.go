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
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/npu-exporter/devmanager"
	npuCommon "huawei.com/npu-exporter/devmanager/common"
	"k8s.io/api/core/v1"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/kubeclient"
)

const (
	phyIDNum   = 1
	logicIDNum = 2
	vDevIDNum  = 3
	aiCoreNum  = 4
)

func TestIsDeviceStatusChange(t *testing.T) {
	tool := AscendTools{name: common.Ascend910, client: &kubeclient.ClientK8s{},
		dmgr: &devmanager.DeviceManagerMock{}}
	convey.Convey("test IsDeviceStatusChange true", t, func() {
		devices := []*common.NpuDevice{{
			Health: v1beta1.Healthy,
		}}
		res := tool.IsDeviceStatusChange(devices, common.Ascend910)
		convey.So(res, convey.ShouldBeTrue)
	})
}

func TestAssembleVirtualDevices(t *testing.T) {
	convey.Convey("test assembleVirtualDevices", t, func() {
		tool := AscendTools{name: common.Ascend910, client: &kubeclient.ClientK8s{},
			dmgr: &devmanager.DeviceManagerMock{}}

		var device []common.NpuDevice
		var deivceType []string
		davinCiDev := common.DavinCiDev{
			PhyID:        phyIDNum,
			LogicID:      logicIDNum,
			TemplateName: map[string]string{"vir16": common.Ascend910c16},
		}

		QueryInfo := npuCommon.CgoVDevQueryInfo{
			Computing: npuCommon.CgoComputingResource{Aic: aiCoreNum},
			Name:      "vir16",
		}
		vDevInfos := npuCommon.VirtualDevInfo{
			VDevInfo: []npuCommon.CgoVDevQueryStru{{QueryInfo: QueryInfo, VDevID: vDevIDNum}},
		}
		tool.assembleVirtualDevices(davinCiDev, vDevInfos, &device, &deivceType)
		testRes := common.NpuDevice{
			DevType:       common.Ascend910c16,
			DeviceName:    fmt.Sprintf("%s-%d-%d", common.Ascend910c16, vDevIDNum, phyIDNum),
			Health:        v1beta1.Healthy,
			NetworkHealth: v1beta1.Healthy,
			LogicID:       logicIDNum,
			PhyID:         phyIDNum,
		}
		convey.So(device, convey.ShouldContain, testRes)
	})
}

// TestAddPodAnnotation1 for test the interface AddPodAnnotation, part 1
func TestAddPodAnnotation1(t *testing.T) {
	tool := AscendTools{name: common.Ascend910, client: &kubeclient.ClientK8s{},
		dmgr: &devmanager.DeviceManagerMock{}}
	convey.Convey("test AddPodAnnotation 1", t, func() {
		convey.Convey("GetDeviceListID failed", func() {
			err := tool.AddPodAnnotation(nil, nil, []string{common.Ascend910}, common.Ascend910c2, "")
			convey.So(err, convey.ShouldNotBeNil)
		})
		mockTryUpdatePodAnnotation := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)),
			"TryUpdatePodAnnotation", func(_ *kubeclient.ClientK8s, pod *v1.Pod,
				annotation map[string]string) error {
				return nil
			})
		defer mockTryUpdatePodAnnotation.Reset()
		convey.Convey("physical device 310P", func() {
			tool.name = common.Ascend310P
			err := tool.AddPodAnnotation(&v1.Pod{}, nil, []string{common.Ascend310P + "-0"}, common.Ascend310P, "")
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("virtual device", func() {
			err := tool.AddPodAnnotation(&v1.Pod{}, nil, []string{common.Ascend310Pc2 + "-100-0"}, common.Ascend310Pc2,
				"")
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("ParseInt failed", func() {
			tool.name = common.Ascend910
			err := tool.AddPodAnnotation(&v1.Pod{}, nil, []string{common.Ascend910 + "-a"}, common.Ascend910, "")
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("GetLogicIDFromPhysicID failed", func() {
			mockGetLogicIDFromPhysicID := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)),
				"GetLogicIDFromPhysicID", func(_ *devmanager.DeviceManagerMock, physicID int32) (int32, error) {
					return 0, fmt.Errorf("error")
				})
			defer mockGetLogicIDFromPhysicID.Reset()
			err := tool.AddPodAnnotation(&v1.Pod{}, nil, []string{common.Ascend910 + "-0"}, common.Ascend910, "")
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestAddPodAnnotation2 for test the interface AddPodAnnotation, part 2
func TestAddPodAnnotation2(t *testing.T) {
	tool := AscendTools{name: common.Ascend910, client: &kubeclient.ClientK8s{},
		dmgr: &devmanager.DeviceManagerMock{}}
	convey.Convey("test AddPodAnnotation 2", t, func() {
		mockTryUpdatePodAnnotation := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)),
			"TryUpdatePodAnnotation", func(_ *kubeclient.ClientK8s, pod *v1.Pod,
				annotation map[string]string) error {
				return nil
			})
		defer mockTryUpdatePodAnnotation.Reset()
		mockGetLogicIDFromPhysicID := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)),
			"GetLogicIDFromPhysicID", func(_ *devmanager.DeviceManagerMock, physicID int32) (int32, error) {
				return 0, nil
			})
		defer mockGetLogicIDFromPhysicID.Reset()
		convey.Convey("GetDeviceIPAddress failed", func() {
			mockGetDeviceIPAddress := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)),
				"GetDeviceIPAddress", func(_ *devmanager.DeviceManagerMock, logicID int32) (string, error) {
					return "", fmt.Errorf("error")
				})
			defer mockGetDeviceIPAddress.Reset()
			err := tool.AddPodAnnotation(&v1.Pod{}, nil, []string{common.Ascend910 + "-0"}, common.Ascend910, "")
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("GetDeviceIPAddress ok", func() {
			mockGetDeviceIPAddress := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)),
				"GetDeviceIPAddress", func(_ *devmanager.DeviceManagerMock, logicID int32) (string, error) {
					return "", nil
				})
			defer mockGetDeviceIPAddress.Reset()
			err := tool.AddPodAnnotation(&v1.Pod{}, nil, []string{common.Ascend910 + "-0"}, common.Ascend910, "")
			convey.So(err, convey.ShouldBeNil)
		})
	})
}
