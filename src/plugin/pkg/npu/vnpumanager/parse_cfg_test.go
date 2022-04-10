// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package vnpumanager for parse configMap llt
package vnpumanager

import (
	"os"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
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
	t.Logf("Start UT TestGetVNpuCfg")
	var data = map[string]string{
		common.VNpuCfgKey: `{"CheckCode": 10086,
		"Nodes": [{"NodeName": "centos-6543","Cards": [{"CardName": 
		"Ascend710-2","Req": ["Ascend710-4c"],"Alloc": []}]}]}`,
	}
	var cm = v1.ConfigMap{
		Data: data,
	}
	mockCM := gomonkey.ApplyFunc(getVNpuCMFromK8s, func(_ kubernetes.Interface, _, _ string) (*v1.ConfigMap, error) {
		return &cm, nil
	})
	if err := os.Setenv("NODE_NAME", "centos-6543"); err != nil {
		t.Logf("UT TestGetVNpuCfg Failed, err: %v\n", err)
	}
	if _, _, err := GetVNpuCfg(nil); err != nil {
		t.Logf("UT TestGetVNpuCfg Failed, err: %v\n", err)
	}
	mockCM.Reset()
	t.Logf("UT TestGetVNpuCfg Success")
}
