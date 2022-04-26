// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package vnpumanager for parse configMap llt
package vnpumanager

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"

	"huawei.com/npu-exporter/hwlog"

	"Ascend-device-plugin/src/plugin/pkg/npu/common"
)

func init() {
	stopCh := make(chan struct{})
	defer close(stopCh)
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, stopCh)
}

// TestGetVNpuCfg get vnpu info, for create virtual device
func TestGetVNpuCfg(t *testing.T) {
	convey.Convey("GetVNpuCfg", t, func() {
		convey.Convey("getVNpuCMFromK8s failed", func() {
			mock := gomonkey.ApplyFunc(getVNpuCMFromK8s, func(_ kubernetes.Interface, _, _ string) (*v1.ConfigMap,
				error) {
				return nil, fmt.Errorf("err")
			})
			defer mock.Reset()
			_, _, err := GetVNpuCfg(nil)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("vnpu config key not exist", func() {
			mock := gomonkey.ApplyFunc(getVNpuCMFromK8s, func(_ kubernetes.Interface, _, _ string) (*v1.ConfigMap,
				error) {
				return &v1.ConfigMap{}, nil
			})
			defer mock.Reset()
			_, _, err := GetVNpuCfg(nil)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("GetCfgContent failed", func() {
			mockCfg := gomonkey.ApplyFunc(getVNpuCMFromK8s, func(_ kubernetes.Interface, _, _ string) (*v1.ConfigMap,
				error) {
				var data = map[string]string{common.VNpuCfgKey: `{"CheckCode": 10086}`}
				return &v1.ConfigMap{Data: data}, nil
			})
			defer mockCfg.Reset()

			mockGetCfgContent := gomonkey.ApplyFunc(GetCfgContent, func(data string) (string, []CardVNPUs, error) {
				return "", nil, fmt.Errorf("err")
			})
			defer mockGetCfgContent.Reset()
			_, _, err := GetVNpuCfg(nil)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("GetCfgContent ok", func() {
			mockCfg := gomonkey.ApplyFunc(getVNpuCMFromK8s, func(_ kubernetes.Interface, _, _ string) (*v1.ConfigMap,
				error) {
				var data = map[string]string{common.VNpuCfgKey: `{"CheckCode": 10086}`}
				return &v1.ConfigMap{Data: data}, nil
			})
			defer mockCfg.Reset()

			mockGetCfgContent := gomonkey.ApplyFunc(GetCfgContent, func(data string) (string, []CardVNPUs, error) {
				return "", []CardVNPUs{{CardName: "Ascend710"}}, nil
			})
			defer mockGetCfgContent.Reset()
			_, _, err := GetVNpuCfg(nil)
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestGetCfgContent test GetCfgContent
func TestGetCfgContent(t *testing.T) {
	convey.Convey("GetVNpuCfg", t, func() {
		convey.Convey("Unmarshal failed", func() {
			mock := gomonkey.ApplyFunc(json.Unmarshal, func(data []byte, v interface{}) error {
				return fmt.Errorf("err")
			})
			defer mock.Reset()
			_, _, err := GetCfgContent("")
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("CheckCode is 0", func() {
			data := map[string]string{
				common.VNpuCfgKey: `{"CheckCode": 0, "Nodes": [{"NodeName": "centos-6543","Cards": [{"CardName": 
				"Ascend710-2","Req": ["Ascend710-4c"],"Alloc": []}]}]}`,
			}
			_, _, err := GetCfgContent(data[common.VNpuCfgKey])
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("getCurNodeCfg no ok", func() {
			data := map[string]string{
				common.VNpuCfgKey: `{"CheckCode": 10086, "Nodes": [{"NodeName": "centos-6543","Cards": [{"CardName": 
				"Ascend710-2","Req": ["Ascend710-4c"],"Alloc": []}]}]}`,
			}
			_, _, err := GetCfgContent(data[common.VNpuCfgKey])
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("getCurNodeCfg ok", func() {
			data := map[string]string{
				common.VNpuCfgKey: `{"CheckCode": 10086, "Nodes": [{"NodeName": "","Cards": [{"CardName": 
			"Ascend710-2","Req": ["Ascend710-4c"],"Alloc": []}]}]}`,
			}
			_, _, err := GetCfgContent(data[common.VNpuCfgKey])
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestGetCurNodeCfg test getCurNodeCfg
func TestGetCurNodeCfg(t *testing.T) {
	vNpuCtn := NodeVNPUs{NodeName: "centos-6543"}
	nodeName := ""
	if _, ret := getCurNodeCfg(vNpuCtn, nodeName); ret {
		t.Fatalf("TestGetCurNodeCfg Run Failed, expect false, but true")
	}

	vNpuCtn = NodeVNPUs{NodeName: "centos-6543"}
	nodeName = "centos-6543"
	if _, ret := getCurNodeCfg(vNpuCtn, nodeName); !ret {
		t.Fatalf("TestGetCurNodeCfg Run Failed, expect true, but false")
	}
}
