/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2022. All rights reserved.
 */

// Package huawei update cache for hps.devices
package huawei

import (
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/client-go/kubernetes"
)

// TestIsExecTimingUpdate for test isExecTimingUpdate
func TestIsExecTimingUpdate(t *testing.T) {
	convey.Convey("isExecTimingUpdate", t, func() {
		convey.Convey("IsPatchSuccess is false", func() {
			GetAnnotationObj().IsUpdateComplete.Store(false)
			GetAnnotationObj().IsPatchSuccess.Store(false)
			isExecTimingUpdate(nil)
			convey.So(GetAnnotationObj().IsUpdateComplete.Load(), convey.ShouldBeFalse)
		})
		convey.Convey("getAnnotationFromNode failed", func() {
			GetAnnotationObj().IsUpdateComplete.Store(false)
			GetAnnotationObj().IsPatchSuccess.Store(true)
			mockData := gomonkey.ApplyFunc(getAnnotationFromNode, func(_ kubernetes.Interface) (map[string]string,
				error) {
				return nil, fmt.Errorf("err")
			})
			defer mockData.Reset()
			isExecTimingUpdate(nil)
			convey.So(GetAnnotationObj().IsUpdateComplete.Load(), convey.ShouldBeFalse)
		})
		convey.Convey("patch success", func() {
			GetAnnotationObj().IsUpdateComplete.Store(false)
			GetAnnotationObj().IsPatchSuccess.Store(true)
			GetAnnotationObj().WaitUpdateAnnotation = map[string]string{resourceNamePrefix + "Ascend710": ""}
			mockData := gomonkey.ApplyFunc(getAnnotationFromNode, func(_ kubernetes.Interface) (map[string]string,
				error) {
				return map[string]string{resourceNamePrefix + "Ascend710": ""}, nil
			})
			defer mockData.Reset()
			isExecTimingUpdate(nil)
			convey.So(GetAnnotationObj().IsUpdateComplete.Load(), convey.ShouldBeTrue)
		})
		convey.Convey("sort list not equal", func() {
			GetAnnotationObj().IsUpdateComplete.Store(false)
			GetAnnotationObj().IsPatchSuccess.Store(true)
			GetAnnotationObj().WaitUpdateAnnotation = map[string]string{resourceNamePrefix + "Ascend710": "Ascend710" +
				"-0,Ascend710-1"}
			mockData := gomonkey.ApplyFunc(getAnnotationFromNode, func(_ kubernetes.Interface) (map[string]string,
				error) {
				return map[string]string{resourceNamePrefix + "Ascend710": "Ascend710-0"}, nil
			})
			defer mockData.Reset()
			isExecTimingUpdate(nil)
			convey.So(GetAnnotationObj().IsUpdateComplete.Load(), convey.ShouldBeFalse)
		})
	})
	t.Logf("TestIsExecTimingUpdate Run Pass")
}

// TestIsSpecDev for test isSpecDev
func TestIsSpecDev(t *testing.T) {
	type testParameter struct {
		annotationTag string
		ret           bool
	}
	var testCases = []testParameter{
		{annotationTag: resourceNamePrefix + hiAIAscend910Prefix, ret: true},
		{annotationTag: "", ret: false},
	}
	for _, testCase := range testCases {
		if ret := isSpecDev(testCase.annotationTag); ret != testCase.ret {
			t.Fatalf("TestIsSpecDev Run Failed, expect %v, but %v", testCase.ret, ret)
		}
	}
	t.Logf("TestIsSpecDev Run Pass")
}
