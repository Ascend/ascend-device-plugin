/*
* Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.
 */

package huawei

import (
	. "github.com/smartystreets/goconvey/convey"
	"k8s.io/apimachinery/pkg/util/sets"
	"testing"
)

func TestGetNewNetworkRecoverDev(t *testing.T) {
	Convey("getNewNetworkRecoverDev test", t, func() {
		Convey("autoStowing is true", func() {
			autoStowingDevs = true
			totalNetworkUnhealthDevices = sets.String{}
			emptySets := sets.String{}
			newNetworkRecoverDevSets, newNetworkUnhealthDevSets := getNewNetworkRecoverDev(emptySets, emptySets)
			So(newNetworkRecoverDevSets, ShouldEqual, emptySets)
			So(newNetworkUnhealthDevSets, ShouldEqual, totalNetworkUnhealthDevices)
		})
		Convey("autoStowing is false", func() {
			autoStowingDevs = false
			totalNetworkUnhealthDevices = sets.String{}
			emptySets := sets.String{}
			newNetworkRecoverDevSets, newNetworkUnhealthDevSets := getNewNetworkRecoverDev(emptySets, emptySets)
			So(newNetworkRecoverDevSets, ShouldHaveSameTypeAs, emptySets)
			So(newNetworkUnhealthDevSets, ShouldHaveSameTypeAs, emptySets)
		})
	})
}

func TestGetDeviceID(t *testing.T) {
	Convey("getDeviceID test", t, func() {
		Convey("getDeviceID get error", func() {
			deviceName := ""
			ascendRuntimeOptions := virtualDev
			_, _, err := getDeviceID(deviceName, ascendRuntimeOptions)
			So(err, ShouldBeError)
		})
		Convey("ascendRuntimeOptions is physicalDev", func() {
			deviceName := "Ascend910-1"
			ascendRuntimeOptions := physicalDev
			_, virID, err := getDeviceID(deviceName, ascendRuntimeOptions)
			So(err, ShouldBeNil)
			So(virID, ShouldBeEmpty)
		})
		Convey("ascendRuntimeOptions is virtualDev", func() {
			deviceName := "Ascend910-1"
			ascendRuntimeOptions := virtualDev
			_, virID, err := getDeviceID(deviceName, ascendRuntimeOptions)
			So(err, ShouldBeNil)
			So(virID, ShouldNotBeEmpty)
		})
	})
}
