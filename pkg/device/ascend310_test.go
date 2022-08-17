/*
* Copyright(C) 2020-2022. Huawei Technologies Co.,Ltd. All rights reserved.
 */
// Package device ascend 310
package device

import (
	"fmt"
	"strings"
	"testing"

	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"huawei.com/npu-exporter/devmanager"
)

const (
	npuTestNum = 8
)

// TestHwAscend310ManagerGetNPUs for GetNPUs
func TestHwAscend310ManagerGetNPUs(t *testing.T) {
	resultDevMap := make(map[string]string)
	for i := 0; i < npuTestNum; i++ {
		resultDevMap[fmt.Sprintf("davinci-mini-%d", i)] = ""
	}
	hdm := createFakeHwDevManager("", true, false, false)
	err := hdm.manager.GetNPUs(&hdm.allDevs, &hdm.allDevTypes, hdm.manager.GetMatchingDeviType())
	if err != nil {
		t.Fatalf("TestHwAscend310Manager_GetNPUs Run Failed")
	}
	t.Logf("TestHwAscend310Manager_GetNPUs %+v ", hdm.allDevTypes)
	if hdm.allDevTypes[0] != "davinci-mini" {
		t.Fatalf("TestHwAscend310Manager_GetNPUs Run Failed")
	}
	for _, dev := range hdm.allDevs {
		_, ok := resultDevMap[dev.ID]
		if !ok {
			t.Fatalf("TestHwAscend310Manager_GetNPUs Run Failed")
		}
	}
	t.Logf("TestHwAscend310Manager_GetNPUs Run Pass")
}

// TestHwAscend310ManagerGetDevState for GetDevState
func TestHwAscend310ManagerGetDevState(t *testing.T) {
	hdm := createFakeHwDevManager("", true, false, false)
	err := hdm.manager.GetNPUs(&hdm.allDevs, &hdm.allDevTypes, hdm.manager.GetMatchingDeviType())
	if err != nil {
		t.Fatal(err)
	}
	for _, dev := range hdm.allDevs {
		state := hdm.manager.GetDevState(dev.ID, hdm.manager.GetDmgr())
		if strings.Contains(dev.ID, "3") && state != v1beta1.Unhealthy {
			t.Fatalf("TestHwAscend310Manager_GetDevState Run Failed %v", dev)

		} else if !strings.Contains(dev.ID, "3") && state == v1beta1.Unhealthy {
			t.Fatalf("TestHwAscend310Manager_GetDevState Run Failed %v", dev)
		}
	}
	t.Logf("TestHwAscend310Manager_GetDevState Run Pass")
}

// TestHwAscend310ManagerGetDevPath for GetDevPath
func TestHwAscend310ManagerGetDevPath(t *testing.T) {
	hdm := createFakeHwDevManager("", true, false, false)
	containerPath, hostPath := hdm.manager.GetDevPath("0", physicalDev)
	if hostPath != containerPath && hostPath != "/dev/davinci0" {
		t.Fatal("TestHwAscend310Manager_GetDevPath Run Failed")
	}
	t.Logf("TestHwAscend310Manager_GetDevPath Run Pass")
}

func createFakeHwDevManager(mode string, fdFlag, useAscendDocker, volcanoType bool) *HwDevManager {
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
	hdm.manager = NewHwAscend310Manager()
	hdm.manager.SetDmgr(&devmanager.DeviceManagerMock{})

	return hdm
}
