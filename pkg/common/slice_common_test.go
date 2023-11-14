/* Copyright(C) 2023. Huawei Technologies Co.,Ltd. All rights reserved.
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
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

var (
	testInt64Source = []int64{1}
	stringSource    = []string{"1"}
)

// TestContains for test Contains
func TestContains(t *testing.T) {
	convey.Convey("test Contains", t, func() {
		convey.Convey("Contains success", func() {
			tool := int32Tool{}
			sources := []int32{1}
			existVal, unExistVal := int32(1), int32(2)
			convey.So(tool.Contains(sources, existVal), convey.ShouldBeTrue)
			convey.So(tool.Contains(sources, unExistVal), convey.ShouldBeFalse)
		})
	})
}

// TestSameElement for test SameElement
func TestSameElement(t *testing.T) {
	convey.Convey("test SameElement", t, func() {
		convey.Convey("SameElement success", func() {
			tool := int64Tool{}
			existVal, unExistVal := []int64{1}, []int64{2}
			convey.So(tool.SameElement(testInt64Source, existVal), convey.ShouldBeTrue)
			convey.So(tool.SameElement(testInt64Source, unExistVal), convey.ShouldBeFalse)
		})
	})
}

// TestRemove for test Remove
func TestRemove(t *testing.T) {
	convey.Convey("test Remove", t, func() {
		convey.Convey("Remove success", func() {
			tool := int64Tool{}
			var emptySources []int64
			existVal, unExistVal := int64(1), int64(2)
			convey.So(len(tool.Remove(emptySources, existVal)), convey.ShouldEqual, 0)
			convey.So(len(tool.Remove(testInt64Source, existVal)), convey.ShouldEqual, 0)
			convey.So(len(tool.Remove(testInt64Source, unExistVal)), convey.ShouldEqual, len(testInt64Source))
		})
	})
}

// TestIndex for test Index
func TestIndex(t *testing.T) {
	convey.Convey("test Index", t, func() {
		convey.Convey("Index success", func() {
			tool := int64Tool{}
			existVal, unExistVal := int64(1), int64(2)
			convey.So(tool.Index(testInt64Source, existVal), convey.ShouldEqual, 0)
			convey.So(tool.Index(testInt64Source, unExistVal), convey.ShouldEqual, -1)
		})
	})
}

// TestToHexString for test ToHexString
func TestToHexString(t *testing.T) {
	convey.Convey("test ToHexString", t, func() {
		convey.Convey("ToHexString success", func() {
			tool := int64Tool{}
			convey.So(tool.ToHexString(testInt64Source), convey.ShouldEqual, "1")
		})
		convey.Convey("ToHexString multiple numbers success", func() {
			tool := int64Tool{}
			int64Slice := []int64{1, 3, 5}
			convey.So(tool.ToHexString(int64Slice), convey.ShouldEqual, "1,3,5")
		})
	})
}

// TestStringSameElement for test string SameElement
func TestStringSameElement(t *testing.T) {
	convey.Convey("test string SameElement", t, func() {
		convey.Convey("SameElement success", func() {
			tool := stringTool{}
			existVal, unExistVal := []string{"1"}, []string{"2"}
			convey.So(tool.SameElement(stringSource, existVal), convey.ShouldBeTrue)
			convey.So(tool.SameElement(stringSource, unExistVal), convey.ShouldBeFalse)
		})
	})
}

// TestHexStringToInt for test string HexStringToInt
func TestHexStringToInt(t *testing.T) {
	convey.Convey("test string HexStringToInt", t, func() {
		convey.Convey("HexStringToInt success", func() {
			tool := stringTool{}
			hexString := []string{"a"}
			errHexString := []string{"xx"}
			convey.So(tool.HexStringToInt(hexString)[0], convey.ShouldEqual, 10)
			convey.So(len(tool.HexStringToInt(errHexString)), convey.ShouldEqual, 0)
		})
	})
}
