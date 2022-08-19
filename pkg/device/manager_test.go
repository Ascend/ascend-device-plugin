// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package device a series of device function
package device

import (
	"fmt"
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
