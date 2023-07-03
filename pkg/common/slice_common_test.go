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
