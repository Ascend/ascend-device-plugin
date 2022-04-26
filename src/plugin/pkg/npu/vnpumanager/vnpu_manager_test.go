// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package vnpumanager using for create and destroy device llt
package vnpumanager

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"

	"Ascend-device-plugin/src/plugin/pkg/npu/common"
	"Ascend-device-plugin/src/plugin/pkg/npu/dsmi"
)

// TestCreateVirtualDev test create virtual devices
func TestCreateVirtualDev(t *testing.T) {
	t.Logf("Start UT TestCreateVirtualDev")
	var cardVNPUs = [][]CardVNPUs{
		{{CardName: "", Req: []string{}, Alloc: []string{}}},
		{{CardName: "Ascend710-2A", Req: []string{}, Alloc: []string{}}},
		{{CardName: "Ascend710-2", Req: []string{"Ascend710-1-4c-2"}, Alloc: []string{}}},
		{{CardName: "Ascend710-2", Req: []string{"Ascend710-200c"}, Alloc: []string{}}},
		{{CardName: "Ascend710-2", Req: []string{"Ascend710-4c"}, Alloc: []string{}}},
		{{CardName: "Ascend710-2", Req: []string{}, Alloc: []string{}}},
	}
	for _, cardVNPU := range cardVNPUs {
		CreateVirtualDev(dsmi.NewFakeDeviceManager(), cardVNPU, common.RunMode710)
	}
	t.Logf("UT TestCreateVirtualDev Success")
}

// TestDestroyVirtualDev test destroy virtual devices
func TestDestroyVirtualDev(t *testing.T) {
	t.Logf("Start UT TestDestroyVirtualDev")

	convey.Convey("TestDestroyVirtualDev", t, func() {
		convey.Convey("destroy", func() {
			dcmiDevices := []common.NpuDevice{{ID: "huawei.com/Ascend710-2c-100-0"}}
			cardVNPUs := []CardVNPUs{{CardName: "Ascend710-2", Req: []string{}, Alloc: []string{"Ascend710-2c-100-0"}}}
			DestroyVirtualDev(dsmi.NewFakeDeviceManager(), dcmiDevices, cardVNPUs, common.NodeName)
		})
		convey.Convey("invalid deviceID", func() {
			dcmiDevices := []common.NpuDevice{{ID: "huawei.com/Ascend710-2c-100-0a"}}
			cardVNPUs := []CardVNPUs{{CardName: "Ascend710-2", Req: []string{}, Alloc: []string{"Ascend710-2c-100-0"}}}
			DestroyVirtualDev(dsmi.NewFakeDeviceManager(), dcmiDevices, cardVNPUs, common.NodeName)
		})
		convey.Convey("destroyRetry failed", func() {
			dcmiDevices := []common.NpuDevice{{ID: "huawei.com/Ascend710-2c-100-0"}}
			cardVNPUs := []CardVNPUs{{CardName: "Ascend710-2", Req: []string{}, Alloc: []string{"Ascend710-2c-100-0"}}}
			mock := gomonkey.ApplyFunc(destroyRetry, func(_ dsmi.DeviceMgrInterface, _ int, virID string) error {
				return fmt.Errorf("err")
			})
			defer mock.Reset()
			DestroyVirtualDev(dsmi.NewFakeDeviceManager(), dcmiDevices, cardVNPUs, common.NodeName)
		})
	})
}

// TestCreateRetry test createRetry
func TestCreateRetry(t *testing.T) {
	convey.Convey("TestCreateRetry", t, func() {
		convey.Convey("GetLogicID failed", func() {
			mock := gomonkey.ApplyMethod(reflect.TypeOf(new(dsmi.FakeDeviceManager)), "GetLogicID",
				func(_ *dsmi.FakeDeviceManager, phyID uint32) (uint32, error) {
					return 0, fmt.Errorf("err")
				})
			defer mock.Reset()
			ret := createRetry(dsmi.NewFakeDeviceManager(), "0", common.RunMode710, CardVNPUs{
				CardName: "Ascend710-2", Req: []string{"Ascend710-2c-100-0"}, Alloc: []string{}})
			convey.So(ret, convey.ShouldNotBeNil)
		})
		convey.Convey("CreateVirtualDevice failed", func() {
			mock := gomonkey.ApplyMethod(reflect.TypeOf(new(dsmi.FakeDeviceManager)), "CreateVirtualDevice",
				func(_ *dsmi.FakeDeviceManager, logicID uint32, runMode string, vNPUs []string) error {
					return fmt.Errorf("err")
				})
			defer mock.Reset()
			ret := createRetry(dsmi.NewFakeDeviceManager(), "0", common.RunMode710, CardVNPUs{
				CardName: "Ascend710-2", Req: []string{"Ascend710-2c-100-0"}, Alloc: []string{}})
			convey.So(ret, convey.ShouldNotBeNil)
		})
	})
}

// TestDestroyRetry test destroyRetry
func TestDestroyRetry(t *testing.T) {
	convey.Convey("TestDestroyRetry", t, func() {
		convey.Convey("GetLogicID failed", func() {
			mock := gomonkey.ApplyMethod(reflect.TypeOf(new(dsmi.FakeDeviceManager)), "GetLogicID",
				func(_ *dsmi.FakeDeviceManager, phyID uint32) (uint32, error) {
					return 0, fmt.Errorf("err")
				})
			defer mock.Reset()
			ret := destroyRetry(dsmi.NewFakeDeviceManager(), 0, "0")
			convey.So(ret, convey.ShouldNotBeNil)
		})
		convey.Convey("invalid virId", func() {
			ret := destroyRetry(dsmi.NewFakeDeviceManager(), 0, "0a")
			convey.So(ret, convey.ShouldNotBeNil)
		})
		convey.Convey("DestroyVirtualDevice failed", func() {
			mock := gomonkey.ApplyMethod(reflect.TypeOf(new(dsmi.FakeDeviceManager)), "DestroyVirtualDevice",
				func(_ *dsmi.FakeDeviceManager, logicID uint32, vDevID uint32) error {
					return fmt.Errorf("err")
				})
			defer mock.Reset()
			ret := destroyRetry(dsmi.NewFakeDeviceManager(), 0, "0")
			convey.So(ret, convey.ShouldNotBeNil)
		})
	})
}

// TestGetNeedCreateDev test getNeedCreateDev
func TestGetNeedCreateDev(t *testing.T) {
	convey.Convey("TestGetNeedCreateDev", t, func() {
		convey.Convey("return empty", func() {
			cardVNPU := CardVNPUs{CardName: "Ascend710-0", Req: []string{}, Alloc: []string{}}
			logicID := uint32(0)
			ret := getNeedCreateDev(cardVNPU, dsmi.NewFakeDeviceManager(), logicID)
			convey.So(len(ret), convey.ShouldEqual, 0)
		})
		convey.Convey("GetVDevicesInfo failed", func() {
			cardVNPU := CardVNPUs{CardName: "Ascend710-0", Req: []string{"Ascend710-2c-100-0"}, Alloc: []string{}}
			logicID := uint32(0)
			fakeDsmi := dsmi.NewFakeDeviceManager()
			mock := gomonkey.ApplyMethod(reflect.TypeOf(new(dsmi.FakeDeviceManager)), "GetVDevicesInfo",
				func(_ *dsmi.FakeDeviceManager, logicID uint32) (dsmi.CgoDsmiVDevInfo, error) {
					return dsmi.CgoDsmiVDevInfo{}, fmt.Errorf("err")
				})
			defer mock.Reset()
			ret := getNeedCreateDev(cardVNPU, fakeDsmi, logicID)
			convey.So(ret, convey.ShouldBeNil)
		})
		convey.Convey("alloc type not meet", func() {
			cardVNPU := CardVNPUs{CardName: "Ascend710-0", Req: []string{"Ascend710-4c-100-0"}, Alloc: []string{}}
			logicID := uint32(0)
			fakeDsmi := dsmi.NewFakeDeviceManager()
			mock := gomonkey.ApplyMethod(reflect.TypeOf(new(dsmi.FakeDeviceManager)), "GetVDevicesInfo",
				func(_ *dsmi.FakeDeviceManager, logicID uint32) (dsmi.CgoDsmiVDevInfo, error) {
					return dsmi.CgoDsmiVDevInfo{CgoDsmiSubVDevInfos: []dsmi.CgoDsmiSubVDevInfo{
						{Spec: dsmi.CgoDsmiVdevSpecInfo{CoreNum: "2"}}}}, nil
				})
			defer mock.Reset()
			ret := getNeedCreateDev(cardVNPU, fakeDsmi, logicID)
			convey.So(len(ret), convey.ShouldEqual, 1)
		})
		convey.Convey("alloc type meet", func() {
			cardVNPU := CardVNPUs{CardName: "Ascend710-0", Req: []string{"Ascend710-2c-100-0"}, Alloc: []string{}}
			logicID := uint32(0)
			fakeDsmi := dsmi.NewFakeDeviceManager()
			mock := gomonkey.ApplyMethod(reflect.TypeOf(new(dsmi.FakeDeviceManager)), "GetVDevicesInfo",
				func(_ *dsmi.FakeDeviceManager, logicID uint32) (dsmi.CgoDsmiVDevInfo, error) {
					return dsmi.CgoDsmiVDevInfo{CgoDsmiSubVDevInfos: []dsmi.CgoDsmiSubVDevInfo{
						{Spec: dsmi.CgoDsmiVdevSpecInfo{CoreNum: "2"}}}}, nil
				})
			defer mock.Reset()
			ret := getNeedCreateDev(cardVNPU, fakeDsmi, logicID)
			convey.So(ret, convey.ShouldBeNil)
		})
	})
}

// TestGetData test getData
func TestGetData(t *testing.T) {
	type testParameter struct {
		count   int
		devCore string
	}
	var testCases = []testParameter{
		{count: 0},
		{count: 1, devCore: "1c"},
	}
	for _, testCase := range testCases {
		getData(testCase.count, testCase.devCore)
	}
}

// TestGetCoreAndCount test getCoreAndCount
func TestGetCoreAndCount(t *testing.T) {
	var testCases = [][]string{
		[]string{},
		[]string{"1c"},
	}
	for _, testCase := range testCases {
		getCoreAndCount(testCase)
	}
}

// TestGetNeedDestroyDev test getNeedDestroyDev
func TestGetNeedDestroyDev(t *testing.T) {
	type testParameter struct {
		dcmiDevices []common.NpuDevice
		cardVNPUs   []CardVNPUs
		nodeName    string
	}

	var testCases = []testParameter{
		{dcmiDevices: []common.NpuDevice{}},
		{dcmiDevices: []common.NpuDevice{{ID: "Ascend710"}}},
		{dcmiDevices: []common.NpuDevice{{ID: "Ascend710-2c-100-0"}}, nodeName: ""},
		{dcmiDevices: []common.NpuDevice{{ID: "Ascend710-2c-100-0"}},
			cardVNPUs: []CardVNPUs{
				{
					CardName: "Ascend710-0",
					Req:      []string{"Ascend710-2c-100-0"},
					Alloc:    []string{"Ascend710-2c-100-0"},
				},
			},
			nodeName: "nodeName"},
	}
	for _, testCase := range testCases {
		getNeedDestroyDev(testCase.dcmiDevices, testCase.cardVNPUs, testCase.nodeName)
	}
}

// TestIsInVNpuCfg test isInVNpuCfg
func TestIsInVNpuCfg(t *testing.T) {
	type testParameter struct {
		devName   string
		deviceID  string
		cardVNPUs []CardVNPUs
		ret       bool
	}

	var testCases = []testParameter{
		{cardVNPUs: []CardVNPUs{}, ret: false},
		{cardVNPUs: []CardVNPUs{{CardName: "Ascend710-2-2"}}, ret: false},
		{deviceID: "2", cardVNPUs: []CardVNPUs{{CardName: "Ascend710-0"}}, ret: false},
		{deviceID: "0", cardVNPUs: []CardVNPUs{{CardName: "Ascend710-0"}}, ret: false},
		{deviceID: "0", cardVNPUs: []CardVNPUs{{CardName: "Ascend710-0", Req: []string{"Ascend710-2c-100-0"}}},
			ret: true},
		{deviceID: "0", cardVNPUs: []CardVNPUs{{CardName: "Ascend710-0", Req: []string{"Ascend710-2c-100-0"},
			Alloc: []string{"Ascend710-2c-100-0"}}}, ret: false},
		{devName: "Ascend710-2c-100-0", deviceID: "0", cardVNPUs: []CardVNPUs{{CardName: "Ascend710-0",
			Req: []string{"Ascend710-2c-100-0"}, Alloc: []string{"Ascend710-2c-100-0"}}}, ret: true},
	}
	for _, testCase := range testCases {
		if ret := isInVNpuCfg(testCase.devName, testCase.deviceID, testCase.cardVNPUs); ret != testCase.ret {
			t.Fatalf("TestIsInVNpuCfg Run Failed, expect %v, but %v", testCase.ret, ret)
		}
	}
}

// TestIsReqAndAllocStable test isReqAndAllocStable
func TestIsReqAndAllocStable(t *testing.T) {
	type testParameter struct {
		cardVNPU CardVNPUs
		ret      bool
	}
	var testCases = []testParameter{
		{cardVNPU: CardVNPUs{CardName: "Ascend710-2", Req: []string{}, Alloc: []string{"Ascend710-0"}}, ret: false},
		{cardVNPU: CardVNPUs{CardName: "Ascend710-2", Req: []string{}, Alloc: []string{}}, ret: true},
		{cardVNPU: CardVNPUs{CardName: "Ascend710-2", Req: []string{"Ascend710-2c-100-0"}, Alloc: []string{}},
			ret: false},
	}
	for _, testCase := range testCases {
		if ret := isReqAndAllocStable(testCase.cardVNPU); ret != testCase.ret {
			t.Fatalf("TestIsReqAndAllocStable Run Failed, expect %v, but %v", testCase.ret, ret)
		}
	}
}
