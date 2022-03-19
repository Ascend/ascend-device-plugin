/*
* Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.
 */

package huawei

import (
	"Ascend-device-plugin/src/plugin/pkg/npu/common"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"k8s.io/apimachinery/pkg/util/sets"
)

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
