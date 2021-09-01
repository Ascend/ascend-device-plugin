/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2020-2021. All rights reserved.
 */

package huawei

import (
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
	"k8s.io/apimachinery/pkg/util/sets"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"sort"
	"strings"
	"testing"
)

// NewFakeHwAscend910Manager for newFakeHwAscend910
func NewFakeHwAscend910Manager() *HwAscend910Manager {
	return &HwAscend910Manager{}
}

// TestHwAscend910Manager_GetNPUs for getNpus
func TestHwAscend910Manager_GetNPUs(t *testing.T) {
	resultDevMap := make(map[string]empty.Empty)
	for i := 0; i < npuTestNum; i++ {
		resultDevMap[fmt.Sprintf("Ascend910-%d", i)] = empty.Empty{}
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
		if IsVirtualDev(dev.ID) {
			continue
		}
		if !ok {
			t.Fatalf("TestHwAscend910Manager_GetNPUs Run Failed")
		}
	}
	t.Logf("TestHwAscend910Manager_GetNPUs Run Pass")
}

// TestHwAscend910Manager_GetDevState for get DevState
func TestHwAscend910Manager_GetDevState(t *testing.T) {
	hdm := createFake910HwDevManager("", true, false, false)
	err := hdm.manager.GetNPUs(&hdm.allDevs, &hdm.allDevTypes, hdm.manager.GetMatchingDeviType())
	if err != nil {
		t.Fatal(err)
	}
	for _, dev := range hdm.allDevs {
		state := hdm.manager.GetDevState(dev.ID, hdm.manager.GetDmgr())
		if IsVirtualDev(dev.ID) {
			continue
		}
		if strings.Contains(dev.ID, "3") && state != pluginapi.Unhealthy {
			t.Fatalf("TestHwAscend910Manager_GetDevState Run Failed %v", dev)
		} else if !strings.Contains(dev.ID, "3") && state == pluginapi.Unhealthy {
			t.Fatalf("TestHwAscend910Manager_GetDevState Run Failed %v", dev)
		}
	}
	t.Logf("TestHwAscend910Manager_GetDevState Run Pass")
}

// TestHwAscend910Manager_GetDevPath for getDevPath
func TestHwAscend910Manager_GetDevPath(t *testing.T) {
	hdm := createFake910HwDevManager("", true, false, false)
	containerPath, hostPath := hdm.manager.GetDevPath("0", physicalDev)
	if hostPath != containerPath && hostPath != "/dev/davinci0" {
		t.Fatal("TestHwAscend910Manager_GetDevPath Run Failed")
	}
	t.Logf("TestHwAscend910Manager_GetDevPath Run Pass")
}

func createFake910HwDevManager(mode string, fdFlag, useAscendDocker, volcanoType bool) *HwDevManager {
	hdm := NewHwDevManager(mode)
	hdm.SetParameters(fdFlag, useAscendDocker, volcanoType, true, sleepTime)
	hdm.manager = NewFakeHwAscend910Manager()
	hdm.manager.SetDmgr(newFakeDeviceManager())
	return hdm
}

func TestGroupDevByPower(t *testing.T) {
	groupDevByPower(sets.String{}, hiAIAscend310Prefix)
	t.Logf("TestGroupDevByPower Run Pass")
}
