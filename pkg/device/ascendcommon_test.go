// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package device a series of device function
package device

import (
	"fmt"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"huawei.com/npu-exporter/devmanager"
	npuCommon "huawei.com/npu-exporter/devmanager/common"
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
