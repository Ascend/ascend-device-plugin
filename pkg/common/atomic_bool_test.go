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
	"testing"

	"github.com/smartystreets/goconvey/convey"
)

// TestAtomicBool for test AtomicBool
func TestAtomicBool(t *testing.T) {
	convey.Convey("test AtomicBool", t, func() {
		flag := NewAtomicBool(false)
		ret := flag.Load()
		convey.So(ret, convey.ShouldBeFalse)
		flag.Store(true)
		ret = flag.Load()
		convey.So(ret, convey.ShouldBeTrue)
	})
}
