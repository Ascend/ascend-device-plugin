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
	"huawei.com/npu-exporter/v5/devmanager"
	npuCommon "huawei.com/npu-exporter/v5/devmanager/common"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/kubeclient"
)

const (
	phyIDNum    = 1
	logicIDNum  = 2
	vDevIDNum   = 3
	aiCoreNum   = 4
	aiCoreCount = 8
	vDevChipID  = 100
)

// TestIsDeviceStatusChange testIsDeviceStatusChange
func TestIsDeviceStatusChange(t *testing.T) {
	tool := AscendTools{name: common.Ascend910, client: &kubeclient.ClientK8s{},
		dmgr: &devmanager.DeviceManagerMock{}}
	convey.Convey("test IsDeviceStatusChange true", t, func() {
		devices := map[string][]*common.NpuDevice{common.Ascend910: {{Health: v1beta1.Healthy}}}
		aiCoreDevice := []*common.NpuDevice{{Health: v1beta1.Healthy}}
		res := tool.UpdateHealthyAndGetChange(devices, aiCoreDevice, common.Ascend910)
		convey.So(res, convey.ShouldNotBeNil)
	})
	tool = AscendTools{name: common.Ascend310P, client: &kubeclient.ClientK8s{},
		dmgr: &devmanager.DeviceManagerMockErr{}}
	convey.Convey("test IsDeviceStatusChange which chip is unhealthy ", t, func() {
		devices := map[string][]*common.NpuDevice{common.Ascend310P: {{Health: v1beta1.Unhealthy}}}
		aiCoreDevice := []*common.NpuDevice{{Health: v1beta1.Unhealthy}}
		res := tool.UpdateHealthyAndGetChange(devices, aiCoreDevice, common.Ascend310P)
		convey.So(res, convey.ShouldNotBeNil)
	})
}

// TestAssembleVirtualDevices testAssembleVirtualDevices
func TestAssembleVirtualDevices(t *testing.T) {
	convey.Convey("test assembleVirtualDevices", t, func() {
		tool := AscendTools{name: common.Ascend910, client: &kubeclient.ClientK8s{},
			dmgr: &devmanager.DeviceManagerMock{}}

		var device []common.NpuDevice
		var deivceType []string
		davinCiDev := common.DavinCiDev{
			PhyID:   phyIDNum,
			LogicID: logicIDNum,
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

// TestCreateVirtualDevice testCreateVirtualDevice
func TestCreateVirtualDevice(t *testing.T) {
	tool := AscendTools{name: common.Ascend310P, client: &kubeclient.ClientK8s{},
		dmgr: &devmanager.DeviceManagerMock{}}
	convey.Convey("test CreateVirtualDevice", t, func() {
		convey.Convey("CreateVirtualDevice success", func() {
			mockGetLogicIDFromPhysicID := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)),
				"GetLogicIDFromPhysicID", func(_ *devmanager.DeviceManagerMock, physicID int32) (int32, error) {
					return 0, nil
				})
			mockCreate := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)),
				"CreateVirtualDevice", func(_ *devmanager.DeviceManagerMock, logicID int32,
					vDevInfo npuCommon.CgoCreateVDevRes) (npuCommon.CgoCreateVDevOut, error) {
					return npuCommon.CgoCreateVDevOut{}, nil
				})
			defer mockCreate.Reset()
			defer mockGetLogicIDFromPhysicID.Reset()
			_, err := tool.CreateVirtualDevice(0, "vir01")
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestDestroyVirtualDevice testDestroyVirtualDevice
func TestDestroyVirtualDevice(t *testing.T) {
	tool := AscendTools{name: common.Ascend310P, client: &kubeclient.ClientK8s{},
		dmgr: &devmanager.DeviceManagerMock{}}
	convey.Convey("test DestroyVirtualDevice", t, func() {
		convey.Convey("DestroyVirtualDevice success", func() {
			mockGetLogicIDFromPhysicID := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)),
				"GetLogicIDFromPhysicID", func(_ *devmanager.DeviceManagerMock, physicID int32) (int32, error) {
					return 0, nil
				})
			mockDestroy := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)),
				"DestroyVirtualDevice", func(_ *devmanager.DeviceManagerMock, _ int32, _ uint32) error {
					return nil
				})
			defer mockDestroy.Reset()
			defer mockGetLogicIDFromPhysicID.Reset()
			err := tool.DestroyVirtualDevice("Ascend310P-1c-100-0")
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestGetChipAiCoreCount testGetChipAiCoreCount
func TestGetChipAiCoreCount(t *testing.T) {
	tool := AscendTools{name: common.Ascend310P, client: &kubeclient.ClientK8s{},
		dmgr: &devmanager.DeviceManagerMock{}}
	res := getVirtualDevInfo(aiCoreNum)
	mockLogicIDs := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)),
		"GetDeviceList", func(_ *devmanager.DeviceManagerMock) (int32, []int32, error) {
			return 1, []int32{0}, nil
		})
	mockVirtual := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)),
		"GetVirtualDeviceInfo", func(_ *devmanager.DeviceManagerMock, _ int32) (
			npuCommon.VirtualDevInfo, error) {
			return res, nil
		})
	defer mockVirtual.Reset()
	defer mockLogicIDs.Reset()
	convey.Convey("test GetChipAiCoreCount 1", t, func() {
		convey.Convey("GetChipAiCoreCount failed", func() {
			_, err := tool.GetChipAiCoreCount()
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
	res = getVirtualDevInfo(aiCoreCount)
	convey.Convey("test GetChipAiCoreCount 2", t, func() {
		convey.Convey("GetChipAiCoreCount success", func() {
			_, err := tool.GetChipAiCoreCount()
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func getVirtualDevInfo(aic float32) npuCommon.VirtualDevInfo {
	return npuCommon.VirtualDevInfo{
		TotalResource: npuCommon.CgoSocTotalResource{
			Computing: npuCommon.CgoComputingResource{
				Aic: aic,
			},
		},
		VDevInfo: []npuCommon.CgoVDevQueryStru{
			{
				VDevID: vDevChipID,
			},
		},
	}
}

// TestAppendVGroupInfo testAppendVGroupInfo
func TestAppendVGroupInfo(t *testing.T) {
	tool := AscendTools{name: common.Ascend310P, client: &kubeclient.ClientK8s{},
		dmgr: &devmanager.DeviceManagerMock{}}
	res := getVirtualDevInfo(aiCoreCount)
	convey.Convey("test AppendVGroupInfo", t, func() {
		convey.Convey("AppendVGroupInfo success", func() {
			mockGetLogicIDFromPhysicID := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)),
				"GetLogicIDFromPhysicID", func(_ *devmanager.DeviceManagerMock, physicID int32) (int32, error) {
					return 0, nil
				})
			mockVirtual := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)),
				"GetVirtualDeviceInfo", func(_ *devmanager.DeviceManagerMock, _ int32) (
					npuCommon.VirtualDevInfo, error) {
					return res, nil
				})
			defer mockVirtual.Reset()
			defer mockGetLogicIDFromPhysicID.Reset()
			allocateDevice := []string{
				"Ascend310P-1c-100-0",
			}
			tool.AppendVGroupInfo(allocateDevice)
			convey.So(len(allocateDevice), convey.ShouldEqual, 1)
		})
	})
}

// TestCheckDeviceTypeLabel testCheckDeviceTypeLabel
func TestCheckDeviceTypeLabel(t *testing.T) {
	tool := AscendTools{name: common.Ascend310P, client: &kubeclient.ClientK8s{},
		dmgr: &devmanager.DeviceManagerMock{}}
	node := getMockNode()
	convey.Convey("test CheckDeviceTypeLabel", t, func() {
		convey.Convey("CheckDeviceTypeLabel get node failed", func() {
			mockNode := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)), "GetNode",
				func(_ *kubeclient.ClientK8s) (*v1.Node, error) {
					return nil, fmt.Errorf("failed to get node")
				})
			defer mockNode.Reset()
			err := tool.CheckDeviceTypeLabel()
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("CheckDeviceTypeLabel success", func() {
			mockNode := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)), "GetNode",
				func(_ *kubeclient.ClientK8s) (*v1.Node, error) {
					return node, nil
				})
			defer mockNode.Reset()
			delete(node.Labels, common.ServerTypeLabelKey)
			err := tool.CheckDeviceTypeLabel()
			convey.So(err, convey.ShouldNotBeNil)
			common.ParamOption.AiCoreCount = aiCoreCount
			node.Labels[common.ServerTypeLabelKey] = "Ascend310P-8"
			err = tool.CheckDeviceTypeLabel()
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

func getMockNode() *v1.Node {
	labels := make(map[string]string, 1)
	labels[common.ServerTypeLabelKey] = "Ascend310P-8"
	return &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Labels: labels,
		},
	}
}

// TestAssemble310PMixedPhyDevices test assemble310PMixedPhyDevices
func TestAssemble310PMixedPhyDevices(t *testing.T) {
	convey.Convey("test assembleVirtualDevices", t, func() {
		tool := AscendTools{name: common.Ascend310P, client: &kubeclient.ClientK8s{},
			dmgr: &devmanager.DeviceManagerMock{}}
		var device []common.NpuDevice
		var deivceType []string
		davinCiDev := common.DavinCiDev{
			PhyID:   phyIDNum,
			LogicID: logicIDNum,
		}
		mockProductType := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)),
			"GetProductType",
			func(_ *devmanager.DeviceManagerMock, cardID int32, deviceID int32) (string, error) {
				return "Atlas 300V Pro", nil
			})
		defer mockProductType.Reset()
		productTypeMap := common.Get310PProductType()
		tool.assemble310PMixedPhyDevices(davinCiDev, &device, &deivceType)
		testRes := common.NpuDevice{
			DevType:       productTypeMap["Atlas 300V Pro"],
			DeviceName:    fmt.Sprintf("%s-%d", productTypeMap["Atlas 300V Pro"], phyIDNum),
			Health:        v1beta1.Healthy,
			NetworkHealth: v1beta1.Healthy,
			LogicID:       logicIDNum,
			PhyID:         phyIDNum,
		}
		convey.So(device, convey.ShouldContain, testRes)
	})
}
