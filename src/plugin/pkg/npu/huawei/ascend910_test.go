/*
* Copyright(C) 2020. Huawei Technologies Co.,Ltd. All rights reserved.
*
 * Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
* http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*/

package huawei

import (
	"fmt"
	"github.com/golang/protobuf/ptypes/empty"
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
		if IsOneOfVirtualDeviceType(dev.ID) {
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
		if IsOneOfVirtualDeviceType(dev.ID) {
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
	var hostPath string
	var containerPath string
	hdm.manager.GetDevPath("0", "", &hostPath, &containerPath)
	if hostPath != containerPath && hostPath != "/dev/davinci0" {
		t.Fatal("TestHwAscend910Manager_GetDevPath Run Failed")
	}
	t.Logf("TestHwAscend910Manager_GetDevPath Run Pass")
}

// TestHwAscend910Manager_GetLogPath for getLogPath
func TestHwAscend910Manager_GetLogPath(t *testing.T) {
	hdm := createFake910HwDevManager("", true, false, false)

	var logPath string
	devID := make([]string, 0)
	devID = append(devID, "Ascend910-0")
	fmt.Printf("deviceId%v, %d", devID, len(devID))
	err := hdm.manager.GetLogPath(devID, "/var/dlog", "", &logPath)
	if err != nil {
		t.Fatal(err)
	}
	splitstring := strings.Split(logPath, "_")
	if len(splitstring) != splitTestStringNum || !strings.Contains(logPath, "0") {
		t.Fail()
	}
	t.Logf("TestHwAscend910Manager_GetLogPath Run Pass ")
}

func createFake910HwDevManager(mode string, fdFlag, useAscendDocker, volcanoType bool) *HwDevManager {
	hdm := NewHwDevManager(mode, "/var/dlog", "/var/log/devicePlugin/")
	hdm.SetParameters(fdFlag, useAscendDocker, volcanoType)
	hdm.manager = NewFakeHwAscend910Manager()
	hdm.manager.SetDmgr(newFakeDeviceManager())
	return hdm
}
