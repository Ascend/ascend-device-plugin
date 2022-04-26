/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2022. All rights reserved.
 */

// Package huawei using informer update cache for hps.devices
package huawei

import (
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"go.uber.org/atomic"
	"k8s.io/apimachinery/pkg/util/sets"

	"Ascend-device-plugin/src/plugin/pkg/npu/common"
	"Ascend-device-plugin/src/plugin/pkg/npu/dsmi"
)

// TestUpdateHpsCache for test updateHpsCache
func TestUpdateHpsCache(t *testing.T) {
	fakeHwDevManager := &HwDevManager{
		runMode:     "ascend910",
		dmgr:        dsmi.NewFakeDeviceManager(),
		stopFlag:    atomic.NewBool(false),
		allDevTypes: []string{"Ascend910"},
	}
	fakeHwDevManager.manager = NewFakeHwAscend910Manager()
	fakeHwDevManager.manager.SetDmgr(dsmi.NewFakeDeviceManager())
	mockData := gomonkey.ApplyFunc(registerNewServer, func(*HwDevManager, []string) { return })
	defer mockData.Reset()
	updateHpsCache(fakeHwDevManager)

	mockNPU := gomonkey.ApplyFunc(fakeHwDevManager.manager.GetNPUs, func(*[]common.NpuDevice, *[]string,
		string) error {
		return fmt.Errorf("err")
	})
	defer mockNPU.Reset()
	updateHpsCache(fakeHwDevManager)
}

// TestGetDiffDevCount for test getDiffDevCount
func TestGetDiffDevCount(t *testing.T) {
	fakeHwDevManager := &HwDevManager{
		runMode:     common.RunMode710,
		dmgr:        dsmi.NewFakeDeviceManager(),
		stopFlag:    atomic.NewBool(false),
		allDevTypes: []string{"Ascend710"},
	}
	newDevTypes := []common.NpuDevice{{DevType: "Ascend710"}}
	getDiffDevCount(fakeHwDevManager, newDevTypes)
}

// TestRegisterNewServer for test registerNewServer
func TestRegisterNewServer(t *testing.T) {
	fakeHwDevManager := &HwDevManager{
		runMode:     "ascend910",
		dmgr:        dsmi.NewFakeDeviceManager(),
		stopFlag:    atomic.NewBool(false),
		allDevTypes: []string{"Ascend910"},
	}

	newDevTypes := []string{"Ascend910"}
	registerNewServer(fakeHwDevManager, newDevTypes)
}

// TestGetSpecDevTypes for test getSpecDevTypes
func TestGetSpecDevTypes(t *testing.T) {
	devices := []common.NpuDevice{{}}
	devType := "Ascend710"
	if ret := getSpecDevTypes(devices, devType); len(ret) != 0 {
		t.Fatalf("TestGetSpecDevTypes Run Failed, expect nil, but %v", ret)
	}

	devices = []common.NpuDevice{{DevType: "Ascend710", ID: "0"}}
	if ret := getSpecDevTypes(devices, devType); len(ret) != 1 || ret[0] != devices[0].ID {
		t.Fatalf("TestGetSpecDevTypes Run Failed, expect 0, but %v", ret)
	}
}

// TestIsDevEqual for test isDevEqual
func TestIsDevEqual(t *testing.T) {
	oldDevs := []string{"Ascend710-0"}
	newDevs := []string{"Ascend710-1"}
	expect := true
	if ret := isDevEqual(oldDevs, newDevs); ret != expect {
		t.Fatalf("TestIsDevEqual Run Failed, expect %v, but %v", expect, ret)
	}

	oldDevs = []string{"Ascend710-0"}
	newDevs = []string{"Ascend710-0"}
	expect = false
	if ret := isDevEqual(oldDevs, newDevs); ret != expect {
		t.Fatalf("TestIsDevEqual Run Failed, expect %v, but %v", expect, ret)
	}
}

// TestConvertToSets for test convertToSets
func TestConvertToSets(t *testing.T) {
	if ret := convertToSets([]string{}); len(ret) != 0 {
		t.Fatalf("TestConvertToSets Run Failed, expect empty, but %v", ret)
	}

	devTypes := []string{"Ascend710"}
	expect := sets.String{}
	expect.Insert("Ascend710")
	ret := convertToSets(devTypes)
	if _, ok := ret[devTypes[0]]; !ok {
		t.Fatalf("TestConvertToSets Run Failed, expect exist Ascend710, but %v", ret)
	}
}
