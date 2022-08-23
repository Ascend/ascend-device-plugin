// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package huawei get data from kubelet check point file
package huawei

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/utils"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"k8s.io/kubernetes/pkg/kubelet/cm/devicemanager/checkpoint"
)

// TestGetEnvVisibleDevices test get env visible devices
func TestGetEnvVisibleDevices(t *testing.T) {
	t.Logf("Start UT TestGetEnvVisibleDevices")
	convey.Convey("getEnvVisibleDevices", t, func() {
		convey.Convey("invalid input", func() {
			validDeviceIDs := getEnvVisibleDevices(nil)
			convey.So(validDeviceIDs, convey.ShouldBeNil)
		})
		convey.Convey("response.Unmarshal failed", func() {
			response := v1beta1.ContainerAllocateResponse{}
			allocResp, err := response.Marshal()
			convey.So(err, convey.ShouldBeNil)
			mockUnmarshal := gomonkey.ApplyMethod(reflect.TypeOf(new(v1beta1.ContainerAllocateResponse)), "Unmarshal",
				func(_ *v1beta1.ContainerAllocateResponse, _ []byte) error { return fmt.Errorf("err") })
			defer mockUnmarshal.Reset()
			validDeviceIDs := getEnvVisibleDevices(allocResp)
			convey.So(validDeviceIDs, convey.ShouldBeNil)
		})
		convey.Convey("not exist env", func() {
			response := v1beta1.ContainerAllocateResponse{}
			allocResp, err := response.Marshal()
			convey.So(err, convey.ShouldBeNil)
			validDeviceIDs := getEnvVisibleDevices([]byte(allocResp))
			convey.So(validDeviceIDs, convey.ShouldBeNil)
		})
		convey.Convey("invalid device id", func() {
			response := v1beta1.ContainerAllocateResponse{Envs: map[string]string{ascendVisibleDevicesEnv: "xxx"}}
			allocResp, err := response.Marshal()
			convey.So(err, convey.ShouldBeNil)
			validDeviceIDs := getEnvVisibleDevices(allocResp)
			convey.So(validDeviceIDs, convey.ShouldBeNil)
		})
		convey.Convey("valid device id", func() {
			response := v1beta1.ContainerAllocateResponse{Envs: map[string]string{ascendVisibleDevicesEnv: "0"}}
			allocResp, err := response.Marshal()
			convey.So(err, convey.ShouldBeNil)
			validDeviceIDs := getEnvVisibleDevices(allocResp)
			convey.So(validDeviceIDs, convey.ShouldNotBeNil)
		})
	})
	t.Logf("UT TestGetEnvVisibleDevices Success")
}

// TestGetKubeletCheckPoint test get kubelet check point
func TestGetKubeletCheckPoint(t *testing.T) {
	t.Logf("Start UT TestGetKubeletCheckPoint")
	convey.Convey("GetKubeletCheckPoint", t, func() {
		convey.Convey("utils.ReadLimitBytes failed", func() {
			mock := gomonkey.ApplyFunc(utils.ReadLimitBytes, func(path string, limitLength int) ([]byte, error) {
				return nil, fmt.Errorf("err")
			})
			defer mock.Reset()
			_, err := GetKubeletCheckPoint(kubeletCheckPointFile)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("unmarshal failed", func() {
			mockRead := gomonkey.ApplyFunc(utils.ReadLimitBytes, func(path string, limitLength int) ([]byte, error) {
				return nil, nil
			})
			defer mockRead.Reset()
			mockUnmarshal := gomonkey.ApplyFunc(json.Unmarshal, func(data []byte, v interface{}) error {
				return fmt.Errorf("err")
			})
			defer mockUnmarshal.Reset()
			_, err := GetKubeletCheckPoint(kubeletCheckPointFile)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("VerifyChecksum failed", func() {
			mockRead := gomonkey.ApplyFunc(utils.ReadLimitBytes, func(path string, limitLength int) ([]byte, error) {
				return nil, nil
			})
			defer mockRead.Reset()
			mockUnmarshal := gomonkey.ApplyFunc(json.Unmarshal, func(data []byte, v interface{}) error {
				return nil
			})
			defer mockUnmarshal.Reset()
			mockVerify := gomonkey.ApplyMethod(reflect.TypeOf(new(checkpoint.Data)), "VerifyChecksum",
				func(_ *checkpoint.Data) error { return fmt.Errorf("err") })
			defer mockVerify.Reset()
			_, err := GetKubeletCheckPoint(kubeletCheckPointFile)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
	t.Logf("UT TestGetKubeletCheckPoint Success")
}

// TestCheckDevType test check device type
func TestCheckDevType(t *testing.T) {
	t.Logf("Start UT TestCheckDevType")
	convey.Convey("checkDevType", t, func() {
		convey.Convey("check 910 failed", func() {
			ret := checkDevType("xxx", hiAIAscend910Prefix)
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("check 310P failed", func() {
			ret := checkDevType("xxx", hiAIAscend310PPrefix)
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("check 310 failed", func() {
			ret := checkDevType("xxx", hiAIAscend310Prefix)
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("found 910", func() {
			ret := checkDevType(hiAIAscend910Prefix, hiAIAscend910Prefix)
			convey.So(ret, convey.ShouldBeTrue)
		})
	})
	t.Logf("UT TestCheckDevType Success")
}

// TestGetAnnotation test get annotation
func TestGetAnnotation(t *testing.T) {
	t.Logf("Start UT TestGetAnnotation")
	convey.Convey("GetAnnotation", t, func() {
		convey.Convey("empty data Request", func() {
			data := CheckpointData{ResourceName: resourceNamePrefix + hiAIAscend910Prefix}
			_, _, err := GetAnnotation(data, hiAIAscend910Prefix)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("invalid resource name", func() {
			data := CheckpointData{ResourceName: "xxx", Request: []string{"Ascend910-0-x"}}
			_, _, err := GetAnnotation(data, hiAIAscend910Prefix)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("invalid device type", func() {
			data := CheckpointData{ResourceName: resourceNamePrefix + "xxx", Request: []string{"Ascend910-0-x"}}
			_, _, err := GetAnnotation(data, hiAIAscend910Prefix)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("invalid device id", func() {
			data := CheckpointData{ResourceName: resourceNamePrefix + hiAIAscend910Prefix,
				Request: []string{"Ascend910-0-x"}}
			_, _, err := GetAnnotation(data, hiAIAscend910Prefix)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("virtual device", func() {
			data := CheckpointData{ResourceName: resourceNamePrefix + hiAIAscend910Prefix,
				Request: []string{"Ascend910-8c-197-6"}, Response: []string{"198"}}
			request, responseDeviceName, err := GetAnnotation(data, hiAIAscend910Prefix)
			convey.So(request, convey.ShouldNotBeNil)
			convey.So(responseDeviceName, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldBeNil)
		})
		convey.Convey("physical device", func() {
			data := CheckpointData{ResourceName: resourceNamePrefix + hiAIAscend910Prefix,
				Request: []string{"Ascend910-0"}, Response: []string{"1"}}
			request, responseDeviceName, err := GetAnnotation(data, hiAIAscend910Prefix)
			convey.So(request, convey.ShouldNotBeNil)
			convey.So(responseDeviceName, convey.ShouldNotBeNil)
			convey.So(err, convey.ShouldBeNil)
		})
	})
	t.Logf("UT TestGetKubeletCheckPoint Success")
}
