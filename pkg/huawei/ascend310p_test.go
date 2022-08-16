/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2020-2021. All rights reserved.
 */

// Package huawei for 310p ut.
package huawei

import (
	"fmt"
	"strings"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"huawei.com/npu-exporter/devmanager"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"Ascend-device-plugin/pkg/common"
)

// TestHwAscend310PManagerGetNPUs for GetNPUs
func TestHwAscend310PManagerGetNPUs(t *testing.T) {
	resultDevMap := make(map[string]string)
	for i := 0; i < npuTestNum; i++ {
		resultDevMap[fmt.Sprintf("Ascend310P-%d", i)] = ""
	}
	resultDevMap["Ascend310P-4c-1-1"] = ""
	resultDevMap["Ascend310P-8c-2-1"] = ""
	hdm := createFake310PHwDevManager("", true, false, false)
	err := hdm.manager.GetNPUs(&hdm.allDevs, &hdm.allDevTypes, hdm.manager.GetMatchingDeviType())
	if err != nil {
		t.Fatalf("TestHwAscend310PManager_GetNPUs Run Failed")
	}
	if !strings.Contains(hdm.allDevTypes[0], "Ascend310P") {
		t.Fatalf("TestHwAscend310PManager_GetNPUs Run Failed")
	}
	for _, dev := range hdm.allDevs {
		_, ok := resultDevMap[dev.ID]
		if !ok {
			t.Fatalf("TestHwAscend310PManager_GetNPUs Run Failed")
		}
	}
	t.Logf("TestHwAscend310PManager_GetNPUs Run Pass")
}

// TestHwAscend310PManagerGetDevState for GetDevState
func TestHwAscend310PManagerGetDevState(t *testing.T) {
	hdm := createFake310PHwDevManager("ascend310P", false, false, false)
	err := hdm.manager.GetNPUs(&hdm.allDevs, &hdm.allDevTypes, hdm.manager.GetMatchingDeviType())
	if err != nil {
		t.Fatal(err)
	}
	for _, dev := range hdm.allDevs {
		state := hdm.manager.GetDevState(dev.ID, hdm.manager.GetDmgr())
		if strings.Contains(dev.ID, "-3") && state != v1beta1.Unhealthy {
			t.Fatalf("TestHwAscend310PManager_GetDevState Run Failed %v", dev)

		} else if !strings.Contains(dev.ID, "-3") && state == v1beta1.Unhealthy {
			t.Fatalf("TestHwAscend310PManager_GetDevState Run Failed %v", dev)
		}
	}
	t.Logf("TestHwAscend310PManager_GetDevState Run Pass")
}

// TestHwAscend310PManagerGetDevPath for GetDevPath
func TestHwAscend310PManagerGetDevPath(t *testing.T) {
	hdm := createFake310PHwDevManager("ascend310P", false, false, false)
	containerPath, hostPath := hdm.manager.GetDevPath("0", physicalDev)
	if hostPath != containerPath && hostPath != "/dev/davinci0" {
		t.Fatal("TestHwAscend310PManager_GetDevPath Run Failed")
	}
	t.Logf("TestHwAscend310PManager_GetDevPath Run Pass")
}

// Test310PListAndWatch test 310p listen and watch
func Test310PListAndWatch(t *testing.T) {
	hdm := setParams(false, common.RunMode310P)
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
	devices := map[string]*common.NpuDevice{"Ascend310": &common.NpuDevice{ID: "0", Health: "Healthy"}}
	hps := &HwPluginServe{devices: devices, hdm: hdm, devType: hiAIAscend310PPrefix}
	hdm.manager.DoWithVolcanoListAndWatch(hps)
	mockNode.Reset()
	mockNodeCtx.Reset()
	mockPatchNode.Reset()
	if len(totalDevices) != 1 || totalDevices.List()[0] != "0" {
		t.Fatal("Test310PListAndWatch Run Failed")
	}
	t.Logf("Test310PListAndWatch Run Pass")
}

func createFake310PHwDevManager(mode string, fdFlag, useAscendDocker, volcanoType bool) *HwDevManager {
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
	hdm.manager = NewHwAscend310PManager()
	hdm.manager.SetDmgr(&devmanager.DeviceManagerMock{})
	return hdm
}
