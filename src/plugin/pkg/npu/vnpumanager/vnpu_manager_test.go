// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package vnpumanager, using for create and destroy device llt
package vnpumanager

import (
	"github.com/agiledragon/gomonkey/v2"
	"k8s.io/client-go/kubernetes"
	"testing"

	"Ascend-device-plugin/src/plugin/pkg/npu/common"
	"Ascend-device-plugin/src/plugin/pkg/npu/dsmi"
)

// TestCreateVirtualDev test create virtual devices
func TestCreateVirtualDev(t *testing.T) {
	t.Logf("Start UT TestCreateVirtualDev")
	var cardVNPUs = []CardVNPUs{
		{
			CardName: "Ascend710-2",
			Req:      []string{"Ascend710-4c"},
			Alloc:    []string{},
		},
	}
	k8sMock := gomonkey.ApplyFunc(getAnnotationFromNode, func(_ kubernetes.Interface, _, _ string) ([]string, error){
		return []string{"Ascend910-2c-100-0","Ascend910-16c-130-0"}, nil
	})
	CreateVirtualDev(dsmi.NewFakeDeviceManager(), cardVNPUs, common.RunMode710, nil)
	k8sMock.Reset()
	t.Logf("UT TestCreateVirtualDev Success")
}

// TestDestroyVirtualDev test destroy virtual devices
func TestDestroyVirtualDev(t *testing.T) {
	t.Logf("Start UT TestDestroyVirtualDev")
	var dcmiDevices = []common.NpuDevice{
		{
			ID: "huawei.com/Ascend710-2c-100-0",
		},
	}
	var cardVNPUs = []CardVNPUs{
		{
			CardName: "Ascend710-2",
			Req:      []string{},
			Alloc:    []string{"Ascend710-2c-100-0"},
		},
	}
	k8sMock := gomonkey.ApplyFunc(getAnnotationFromNode, func(_ kubernetes.Interface, _, _ string) ([]string, error){
		return []string{"Ascend910-2c-100-0","Ascend910-16c-130-0"}, nil
	})
	DestroyVirtualDev(dsmi.NewFakeDeviceManager(), dcmiDevices, cardVNPUs, "ascend710", nil)
	k8sMock.Reset()
	t.Logf("UT TestDestroyVirtualDev Success")
}
