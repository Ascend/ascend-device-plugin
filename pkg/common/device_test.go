// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package common a series of common function
package common

import (
	"strconv"
	"strings"
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

// TestConvertDevListToSets for test ConvertDevListToSets
func TestConvertDevListToSets(t *testing.T) {
	convey.Convey("test ConvertDevListToSets", t, func() {
		convey.Convey("devices is empty", func() {
			ret := ConvertDevListToSets("", "")
			convey.So(ret.Len(), convey.ShouldEqual, 0)
		})
		convey.Convey("length of deviceInfo more then MaxDevicesNum", func() {
			devices := ""
			for i := 0; i <= MaxDevicesNum; i++ {
				devices += strconv.Itoa(i) + "."
			}
			ret := ConvertDevListToSets(devices, "")
			convey.So(ret.Len(), convey.ShouldEqual, 0)
		})
		convey.Convey("sepType is DotSepDev, ParseInt failed", func() {
			devices := "a.b.c"
			ret := ConvertDevListToSets(devices, DotSepDev)
			convey.So(ret.Len(), convey.ShouldEqual, 0)
		})
		convey.Convey("sepType is DotSepDev, ParseInt ok", func() {
			devices := "0.1.2"
			ret := ConvertDevListToSets(devices, DotSepDev)
			convey.So(ret.Len(), convey.ShouldEqual, len(strings.Split(devices, ".")))
		})
		convey.Convey("match Ascend910", func() {
			devices := "Ascend910-0.Ascend910-1.Ascend910-2"
			ret := ConvertDevListToSets(devices, DotSepDev)
			convey.So(ret.Len(), convey.ShouldEqual, 0)
		})
		convey.Convey("not match Ascend910", func() {
			devices := "0.1.2"
			ret := ConvertDevListToSets(devices, "")
			convey.So(ret.Len(), convey.ShouldEqual, 0)
		})
	})
}

// TestIsVirtualDev for test IsVirtualDev
func TestIsVirtualDev(t *testing.T) {
	convey.Convey("test IsVirtualDev", t, func() {
		convey.Convey("virtual device", func() {
			ret := IsVirtualDev("Ascend910")
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("physical device", func() {
			ret := IsVirtualDev("Ascend910-2c-100-0")
			convey.So(ret, convey.ShouldBeTrue)
		})
	})
}

// TestGetVNPUSegmentInfo for testGetVNPUSegmentInfo
func TestGetVNPUSegmentInfo(t *testing.T) {
	deviceInfos := []string{"0", "vir02"}
	convey.Convey("test GetVNPUSegmentInfo", t, func() {
		convey.Convey("GetVNPUSegmentInfo success", func() {
			_, _, err := GetVNPUSegmentInfo(deviceInfos)
			convey.So(err, convey.ShouldBeNil)
		})
		deviceInfos = []string{"65", "vir02"}
		convey.Convey("GetVNPUSegmentInfo failed with upper limit id", func() {
			_, _, err := GetVNPUSegmentInfo(deviceInfos)
			convey.So(err, convey.ShouldNotBeNil)
		})
		deviceInfos = []string{"x", "vir02"}
		convey.Convey("GetVNPUSegmentInfo failed with invalid id", func() {
			_, _, err := GetVNPUSegmentInfo(deviceInfos)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestFakeAiCoreDevice for testFakeAiCoreDevice
func TestFakeAiCoreDevice(t *testing.T) {
	dev := DavinCiDev{
		LogicID: 0,
		PhyID:   0,
	}
	aiCoreDevices := make([]*NpuDevice, 0)
	ParamOption.AiCoreCount = MinAICoreNum
	convey.Convey("test FakeAiCoreDevice", t, func() {
		convey.Convey("FakeAiCoreDevice success", func() {
			FakeAiCoreDevice(dev, &aiCoreDevices)
			convey.So(len(aiCoreDevices), convey.ShouldEqual, MinAICoreNum)
		})
	})
}
