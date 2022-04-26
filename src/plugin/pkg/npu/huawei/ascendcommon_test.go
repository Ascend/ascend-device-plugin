/*
* Copyright(C) 2021-2022. Huawei Technologies Co.,Ltd. All rights reserved.
 */

package huawei

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"Ascend-device-plugin/src/plugin/pkg/npu/common"
	"Ascend-device-plugin/src/plugin/pkg/npu/dsmi"
)

const testLogicID = 3

// TestUnhealthyState for UnhealthyState
func TestUnhealthyState(t *testing.T) {
	if err := UnhealthyState(1, uint32(testLogicID), "healthState", dsmi.NewFakeDeviceManager()); err != nil {
		t.Errorf("TestUnhealthyState Run Failed")
	}
	t.Logf("TestUnhealthyState Run Pass")
}

// TestGetPhyIDByName for PhyIDByName
func TestGetPhyIDByName(t *testing.T) {
	if phyID, err := GetPhyIDByName("Ascend310-3"); err != nil || dsmi.UnHealthyTestLogicID != phyID {
		t.Errorf("TestGetLogicIDByName Run Failed")
	}

	if _, err := GetPhyIDByName("Ascend310-1000"); err == nil {
		t.Errorf("TestGetLogicIDByName Run Failed")
	}
	t.Logf("TestGetLogicIDByName Run Pass")
}

// TestGetDefaultDevices for GetDefaultDevices
func TestGetDefaultDevices(t *testing.T) {
	if _, err := os.Stat(common.HiAIHDCDevice); err != nil {
		if err = createFile(common.HiAIHDCDevice); err != nil {
			t.Fatal("TestGetDefaultDevices Run Failed")
		}
	}

	if _, err := os.Stat(common.HiAIManagerDevice); err != nil {
		if err = createFile(common.HiAIManagerDevice); err != nil {
			t.Fatal("TestGetDefaultDevices Run Failed")
		}
	}

	if _, err := os.Stat(common.HiAISVMDevice); err != nil {
		if err = createFile(common.HiAISVMDevice); err != nil {
			t.Fatal("TestGetDefaultDevices Run Failed")
		}
	}
	var defaultDeivces []string
	if err := GetDefaultDevices(&defaultDeivces); err != nil {
		t.Errorf("TestGetDefaultDevices Run Failed")
	}
	defaultMap := make(map[string]string)
	defaultMap[common.HiAIHDCDevice] = ""
	defaultMap[common.HiAIManagerDevice] = ""
	defaultMap[common.HiAISVMDevice] = ""
	defaultMap[common.HiAi200RCEventSched] = ""
	defaultMap[common.HiAi200RCHiDvpp] = ""
	defaultMap[common.HiAi200RCLog] = ""
	defaultMap[common.HiAi200RCMemoryBandwidth] = ""
	defaultMap[common.HiAi200RCSVM0] = ""
	defaultMap[common.HiAi200RCTsAisle] = ""
	defaultMap[common.HiAi200RCUpgrade] = ""

	for _, str := range defaultDeivces {
		if _, ok := defaultMap[str]; !ok {
			t.Errorf("TestGetDefaultDevices Run Failed")
		}
	}
	t.Logf("TestGetDefaultDevices Run Pass")
}

func createFile(filePath string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := f.Chmod(common.SocketChmod); err != nil {
		return err
	}
	return nil
}

func TestGetNewNetworkRecoverDev(t *testing.T) {
	convey.Convey("getNewNetworkRecoverDev test", t, func() {
		convey.Convey("autoStowing is true", func() {
			autoStowingDevs = true
			totalNetworkUnhealthDevices = sets.String{}
			emptySets := sets.String{}
			newNetworkRecoverDevSets, newNetworkUnhealthDevSets := getNewNetworkRecoverDev(emptySets, emptySets)
			convey.So(newNetworkRecoverDevSets, convey.ShouldEqual, emptySets)
			convey.So(newNetworkUnhealthDevSets, convey.ShouldEqual, totalNetworkUnhealthDevices)
		})
		convey.Convey("autoStowing is false", func() {
			autoStowingDevs = false
			totalNetworkUnhealthDevices = sets.String{}
			emptySets := sets.String{}
			newNetworkRecoverDevSets, newNetworkUnhealthDevSets := getNewNetworkRecoverDev(emptySets, emptySets)
			convey.So(newNetworkRecoverDevSets, convey.ShouldHaveSameTypeAs, emptySets)
			convey.So(newNetworkUnhealthDevSets, convey.ShouldHaveSameTypeAs, emptySets)
		})
	})
}

func TestGetDeviceID(t *testing.T) {
	convey.Convey("getDeviceID test", t, func() {
		convey.Convey("getDeviceID get error", func() {
			deviceName := ""
			ascendRuntimeOptions := common.VirtualDev
			_, _, err := common.GetDeviceID(deviceName, ascendRuntimeOptions)
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("ascendRuntimeOptions is physicalDev", func() {
			deviceName := "Ascend910-1"
			ascendRuntimeOptions := physicalDev
			_, virID, err := common.GetDeviceID(deviceName, ascendRuntimeOptions)
			convey.So(err, convey.ShouldBeNil)
			convey.So(virID, convey.ShouldBeEmpty)
		})
		convey.Convey("ascendRuntimeOptions is virtualDev", func() {
			deviceName := "Ascend910-2c-112-1"
			ascendRuntimeOptions := common.VirtualDev
			_, virID, err := common.GetDeviceID(deviceName, ascendRuntimeOptions)
			convey.So(err, convey.ShouldBeNil)
			convey.So(virID, convey.ShouldNotBeEmpty)
		})
	})
}

// TestReloadHealthDevice for reloadHealthDevice
func TestReloadHealthDevice(t *testing.T) {
	devices := map[string]*common.NpuDevice{"Ascend710": &common.NpuDevice{ID: "0", Health: "Healthy"},
		"Ascend710-1c": &common.NpuDevice{ID: "1", Health: "Unhealthy"}}
	hps := HwPluginServe{devices: devices}
	adc := ascendCommonFunction{}
	adc.reloadHealthDevice(&hps)
	if len(hps.healthDevice) != 1 {
		t.Fatalf("TestReloadHealthDevice Run Failed")
	}
	if len(hps.unHealthDevice) != 1 {
		t.Fatalf("TestReloadHealthDevice Run Failed")
	}
}

// TestUpdateAiCore for updateAiCore
func TestUpdateAiCore(t *testing.T) {
	convey.Convey("TestUpdateAiCore", t, func() {
		convey.Convey("GetDeviceList failed", func() {
			adc := ascendCommonFunction{dmgr: dsmi.NewFakeDeviceManager()}
			mock := gomonkey.ApplyMethod(reflect.TypeOf(new(dsmi.FakeDeviceManager)), "GetDeviceList",
				func(_ *dsmi.FakeDeviceManager, _ *[hiAIMaxDeviceNum]uint32) (int32, error) {
					return 0,
						fmt.Errorf("err")
				})
			defer mock.Reset()
			convey.So(adc.updateAiCore(), convey.ShouldBeEmpty)
		})
		convey.Convey("GetDeviceHealth failed", func() {
			adc := ascendCommonFunction{dmgr: dsmi.NewFakeDeviceManager()}
			mock := gomonkey.ApplyMethod(reflect.TypeOf(new(dsmi.FakeDeviceManager)), "GetDeviceHealth",
				func(_ *dsmi.FakeDeviceManager, _ int32) (uint32, error) { return 0, fmt.Errorf("err") })
			defer mock.Reset()
			convey.So(adc.updateAiCore(), convey.ShouldBeEmpty)
		})
		convey.Convey("GetPhyID failed", func() {
			adc := ascendCommonFunction{dmgr: dsmi.NewFakeDeviceManager()}
			mock := gomonkey.ApplyMethod(reflect.TypeOf(new(dsmi.FakeDeviceManager)), "GetPhyID",
				func(_ *dsmi.FakeDeviceManager, _ uint32) (uint32, error) { return 0, fmt.Errorf("err") })
			defer mock.Reset()
			convey.So(adc.updateAiCore(), convey.ShouldBeEmpty)
		})
		convey.Convey("GetVDevicesInfo failed", func() {
			adc := ascendCommonFunction{dmgr: dsmi.NewFakeDeviceManager()}
			mock := gomonkey.ApplyMethod(reflect.TypeOf(new(dsmi.FakeDeviceManager)), "GetVDevicesInfo",
				func(_ *dsmi.FakeDeviceManager, _ uint32) (dsmi.CgoDsmiVDevInfo, error) {
					return dsmi.CgoDsmiVDevInfo{}, fmt.Errorf("err")
				})
			defer mock.Reset()
			convey.So(adc.updateAiCore(), convey.ShouldNotBeEmpty)
		})
		convey.Convey("return not empty", func() {
			adc := ascendCommonFunction{dmgr: dsmi.NewFakeDeviceManager()}
			convey.So(adc.updateAiCore(), convey.ShouldNotBeEmpty)
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
			_, ret := VerifyPath("")
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("os.Stat failed", func() {
			mock := gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
				return nil, fmt.Errorf("err")
			})
			defer mock.Reset()
			_, ret := VerifyPath("./")
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("filepath.EvalSymlinks failed", func() {
			mock := gomonkey.ApplyFunc(filepath.EvalSymlinks, func(path string) (string, error) {
				return "", fmt.Errorf("err")
			})
			defer mock.Reset()
			_, ret := VerifyPath("./")
			convey.So(ret, convey.ShouldBeFalse)
		})
	})
}

// TestGetDevState for GetDevState
func TestGetDevState(t *testing.T) {
	convey.Convey("TestGetDevState", t, func() {
		convey.Convey("GetPhyIDByName failed", func() {
			mock := gomonkey.ApplyFunc(GetPhyIDByName, func(_ string) (uint32, error) {
				return 0, fmt.Errorf("err")
			})
			defer mock.Reset()
			adc := ascendCommonFunction{}
			convey.So(adc.GetDevState("", dsmi.NewFakeDeviceManager()), convey.ShouldEqual, v1beta1.Unhealthy)
		})
		convey.Convey("GetLogicID failed", func() {
			mock := gomonkey.ApplyMethod(reflect.TypeOf(new(dsmi.FakeDeviceManager)), "GetLogicID",
				func(_ *dsmi.FakeDeviceManager, _ uint32) (uint32, error) { return 0, fmt.Errorf("err") })
			defer mock.Reset()
			adc := ascendCommonFunction{}
			convey.So(adc.GetDevState("", dsmi.NewFakeDeviceManager()), convey.ShouldEqual, v1beta1.Unhealthy)
		})
		convey.Convey("GetDeviceHealth failed", func() {
			mock := gomonkey.ApplyMethod(reflect.TypeOf(new(dsmi.FakeDeviceManager)), "GetDeviceHealth",
				func(_ *dsmi.FakeDeviceManager, _ int32) (uint32, error) { return 0, fmt.Errorf("err") })
			defer mock.Reset()
			adc := ascendCommonFunction{}
			convey.So(adc.GetDevState("", dsmi.NewFakeDeviceManager()), convey.ShouldEqual, v1beta1.Unhealthy)
		})
		convey.Convey("GetDeviceHealth return unhealth, UnhealthyState failed", func() {
			mock := gomonkey.ApplyFunc(UnhealthyState, func(_, _ uint32, _ string, _ dsmi.DeviceMgrInterface) error {
				return fmt.Errorf("err")
			})
			defer mock.Reset()
			adc := ascendCommonFunction{}
			convey.So(adc.GetDevState("", dsmi.NewFakeDeviceManager()), convey.ShouldEqual, v1beta1.Unhealthy)
		})
	})
}
