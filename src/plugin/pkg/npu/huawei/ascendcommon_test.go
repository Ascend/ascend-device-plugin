/*
* Copyright(C) 2021-2022. Huawei Technologies Co.,Ltd. All rights reserved.
 */

package huawei

import (
	"os"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"k8s.io/apimachinery/pkg/util/sets"

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
