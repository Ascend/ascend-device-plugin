// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

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
