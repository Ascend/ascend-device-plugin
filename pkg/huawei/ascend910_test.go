/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2020-2022. All rights reserved.
 */
// Package huawei ascend 910
package huawei

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/agiledragon/gomonkey/v2"

	"huawei.com/npu-exporter/devmanager"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"Ascend-device-plugin/pkg/common"
)

// NewFakeHwAscend910Manager for newFakeHwAscend910
func NewFakeHwAscend910Manager() *HwAscend910Manager {
	return &HwAscend910Manager{}
}

// TestHwAscend910ManagerGetNPUs for getNpus
func TestHwAscend910ManagerGetNPUs(t *testing.T) {
	resultDevMap := make(map[string]string)
	for i := 0; i < npuTestNum; i++ {
		resultDevMap[fmt.Sprintf("Ascend910-%d", i)] = ""
	}
	hdm := createFake910HwDevManager("ascend910", false, false, false)
	err := hdm.manager.GetNPUs(&hdm.allDevs, &hdm.allDevTypes, hdm.manager.GetMatchingDeviType())
	if err != nil {
		t.Fatalf("TestHwAscend910Manager_GetNPUs Run Failed")
	}
	sort.Strings(hdm.allDevTypes)
	if hdm.allDevTypes[0] != "Ascend910" {
		t.Fatalf("TestHwAscend910Manager_GetNPUs Run Failed")
	}
	for _, dev := range hdm.allDevs {
		_, ok := resultDevMap[dev.ID]
		if common.IsVirtualDev(dev.ID) {
			continue
		}
		if !ok {
			t.Fatalf("TestHwAscend910Manager_GetNPUs Run Failed")
		}
	}
	t.Logf("TestHwAscend910Manager_GetNPUs Run Pass")
}

// TestHwAscend910ManagerGetDevState for get DevState
func TestHwAscend910ManagerGetDevState(t *testing.T) {
	hdm := createFake910HwDevManager("", true, false, false)
	err := hdm.manager.GetNPUs(&hdm.allDevs, &hdm.allDevTypes, hdm.manager.GetMatchingDeviType())
	if err != nil {
		t.Fatal(err)
	}
	for _, dev := range hdm.allDevs {
		state := hdm.manager.GetDevState(dev.ID, hdm.manager.GetDmgr())
		if common.IsVirtualDev(dev.ID) {
			continue
		}
		if strings.Contains(dev.ID, "3") && state != v1beta1.Unhealthy {
			t.Fatalf("TestHwAscend910Manager_GetDevState Run Failed %v", dev)
		} else if !strings.Contains(dev.ID, "3") && state == v1beta1.Unhealthy {
			t.Fatalf("TestHwAscend910Manager_GetDevState Run Failed %v", dev)
		}
	}
	t.Logf("TestHwAscend910Manager_GetDevState Run Pass")
}

// TestHwAscend910ManagerGetDevPath for getDevPath
func TestHwAscend910ManagerGetDevPath(t *testing.T) {
	hdm := createFake910HwDevManager("", true, false, false)
	containerPath, hostPath := hdm.manager.GetDevPath("0", physicalDev)
	if hostPath != containerPath && hostPath != "/dev/davinci0" {
		t.Fatal("TestHwAscend910Manager_GetDevPath Run Failed")
	}
	t.Logf("TestHwAscend910Manager_GetDevPath Run Pass")
}

func createFake910HwDevManager(mode string, fdFlag, useAscendDocker, volcanoType bool) *HwDevManager {
	hdm := &HwDevManager{}
	o := Option{
		GetFdFlag:          fdFlag,
		UseAscendDocker:    useAscendDocker,
		UseVolcanoType:     volcanoType,
		ListAndWatchPeriod: sleepTime,
		AutoStowingDevs:    true,
		KubeConfig:         "",
	}
	hdm.SetParameters(o)
	hdm.manager = NewFakeHwAscend910Manager()
	hdm.manager.SetDmgr(&devmanager.DeviceManagerMock{})
	return hdm
}

func TestGroupDevByPower(t *testing.T) {
	hdm := createFake910HwDevManager("", true, false, false)
	hdm.manager.GetAnnotationMap(sets.String{}, []string{hiAIAscend910Prefix})
	t.Logf("TestGroupDevByPower Run Pass")
}

// Test910PListAndWatch test 910 listen and watch
func Test910PListAndWatch(t *testing.T) {
	hdm := setParams(false, common.RunMode910)
	if err := hdm.GetNPUs(); err != nil {
		t.Fatal(err)
	}
	mockNode := gomonkey.ApplyFunc(getNodeNpuUsed, func(usedDevices *sets.String, hps *HwPluginServe) {
		return
	})
	mockNodeCtx := gomonkey.ApplyFunc(getNodeWithTodoCtx, func(_ *KubeInteractor) (*v1.Node, error) {
		return nil, nil
	})
	mockPatchNode := gomonkey.ApplyFunc(patchNodeWithTodoCtx, func(_ *KubeInteractor, _ []byte) (*v1.Node, error) {
		return nil, nil
	})
	devices := map[string]*common.NpuDevice{"Ascend910": &common.NpuDevice{ID: "0", Health: "Healthy"}}
	hps := &HwPluginServe{devices: devices, hdm: hdm, devType: hiAIAscend910Prefix}
	totalNetworkUnhealthDevices = sets.String{}
	hdm.manager.DoWithVolcanoListAndWatch(hps)
	mockNode.Reset()
	mockNodeCtx.Reset()
	mockPatchNode.Reset()
	if len(totalDevices) != 1 || totalDevices.List()[0] != "0" {
		t.Fatal("Test910PListAndWatch Run Failed")
	}
	t.Logf("Test910PListAndWatch Run Pass")
}
