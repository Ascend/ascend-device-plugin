/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2020-2021. All rights reserved.
 */

package huawei

import (
	"fmt"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"strings"
	"testing"
)

// TestHwAscend710Manager_GetNPUs for GetNPUs
func TestHwAscend710Manager_GetNPUs(t *testing.T) {
	resultDevMap := make(map[string]string)
	for i := 0; i < npuTestNum; i++ {
		resultDevMap[fmt.Sprintf("Ascend710-%d", i)] = ""
	}
	hdm := createFake710HwDevManager("ascend710", false, false, false)
	err := hdm.manager.GetNPUs(&hdm.allDevs, &hdm.allDevTypes, hdm.manager.GetMatchingDeviType())
	if err != nil {
		t.Fatalf("TestHwAscend710Manager_GetNPUs Run Failed")
	}
	if hdm.allDevTypes[0] != "Ascend710" {
		t.Fatalf("TestHwAscend710Manager_GetNPUs Run Failed")
	}
	for _, dev := range hdm.allDevs {
		_, ok := resultDevMap[dev.ID]
		if !ok {
			t.Fatalf("TestHwAscend710Manager_GetNPUs Run Failed")
		}
	}
	t.Logf("TestHwAscend710Manager_GetNPUs Run Pass")
}

// TestHwAscend710Manager_GetDevState for GetDevState
func TestHwAscend710Manager_GetDevState(t *testing.T) {
	hdm := createFake710HwDevManager("ascend710", false, false, false)
	err := hdm.manager.GetNPUs(&hdm.allDevs, &hdm.allDevTypes, hdm.manager.GetMatchingDeviType())
	if err != nil {
		t.Fatal(err)
	}
	for _, dev := range hdm.allDevs {
		state := hdm.manager.GetDevState(dev.ID, hdm.manager.GetDmgr())
		if strings.Contains(dev.ID, "3") && state != pluginapi.Unhealthy {
			t.Fatalf("TestHwAscend710Manager_GetDevState Run Failed %v", dev)

		} else if !strings.Contains(dev.ID, "3") && state == pluginapi.Unhealthy {
			t.Fatalf("TestHwAscend710Manager_GetDevState Run Failed %v", dev)
		}
	}
	t.Logf("TestHwAscend710Manager_GetDevState Run Pass")
}

// TestHwAscend710Manager_GetDevPath for GetDevPath
func TestHwAscend710Manager_GetDevPath(t *testing.T) {
	hdm := createFake710HwDevManager("ascend710", false, false, false)
	containerPath, hostPath := hdm.manager.GetDevPath("0", physicalDev)
	if hostPath != containerPath && hostPath != "/dev/davinci0" {
		t.Fatal("TestHwAscend710Manager_GetDevPath Run Failed")
	}
	t.Logf("TestHwAscend710Manager_GetDevPath Run Pass")
}

func createFake710HwDevManager(mode string, fdFlag, useAscendDocker, volcanoType bool) *HwDevManager {
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
	hdm.manager = NewHwAscend710Manager()
	hdm.manager.SetDmgr(newFakeDeviceManager())
	return hdm
}
