/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2020-2021. All rights reserved.
 */

// Package huawei for 310p ut.
package huawei

import (
	"fmt"
	"strings"
	"testing"

	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"Ascend-device-plugin/src/plugin/pkg/npu/dsmi"
)

// TestHwAscend310PManagerGetNPUs for GetNPUs
func TestHwAscend310PManagerGetNPUs(t *testing.T) {
	resultDevMap := make(map[string]string)
	for i := 0; i < npuTestNum; i++ {
		resultDevMap[fmt.Sprintf("Ascend310P-%d", i)] = ""
	}
	resultDevMap["Ascend310P-4c-1-1"] = ""
	resultDevMap["Ascend310P-8c-2-1"] = ""
	hdm := createFake310PHwDevManager("ascend310P", false, false, false)
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
		if state != v1beta1.Healthy {
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

func createFake310PHwDevManager(mode string, fdFlag, useAscendDocker, volcanoType bool) *HwDevManager {
	hdm := NewHwDevManager(mode)
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
	hdm.manager.SetDmgr(dsmi.NewFakeDeviceManager())
	return hdm
}
