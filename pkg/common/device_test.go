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
			ret := ConvertDevListToSets(devices, "")
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
