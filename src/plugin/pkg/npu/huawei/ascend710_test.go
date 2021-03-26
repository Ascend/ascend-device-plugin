/*
* Copyright(C) 2021. Huawei Technologies Co.,Ltd. All rights reserved.
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
	"strings"
	"testing"
)

// NewFakeHwAscend710Manager used to create ascend 710 manager
func NewFakeHwAscend710Manager() *HwAscend710Manager {
	return &HwAscend710Manager{}
}

// TestHwAscend710Manager_GetNPUs for GetNPUs
func TestHwAscend710Manager_GetNPUs(t *testing.T) {
	resultDevMap := make(map[string]empty.Empty)
	for i := 0; i < npuTestNum; i++ {
		resultDevMap[fmt.Sprintf("Ascend710-%d", i)] = empty.Empty{}
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
	var hostPath string
	var containerPath string
	hdm.manager.GetDevPath("0", "", &hostPath, &containerPath)
	if hostPath != containerPath && hostPath != "/dev/davinci0" {
		t.Fatal("TestHwAscend710Manager_GetDevPath Run Failed")
	}
	t.Logf("TestHwAscend710Manager_GetDevPath Run Pass")
}

// TestHwAscend710Manager_GetLogPath for getLogPath
func TestHwAscend710Manager_GetLogPath(t *testing.T) {
	hdm := createFake710HwDevManager("ascend710", false, false, false)

	var logPath string
	devID := make([]string, 0)
	devID = append(devID, "Ascend710-0")
	t.Logf("deviceId%v, %d", devID, len(devID))
	err := hdm.manager.GetLogPath(devID, "/var/dlog", "", &logPath)
	if err != nil {
		t.Fatal(err)
	}
	splitstring := strings.Split(logPath, "_")
	if len(splitstring) != splitTestStringNum || !strings.Contains(logPath, "0") {
		t.Fail()
	}
	t.Logf("TestHwAscend710Manager_GetLogPath Run Pass ")
}

func createFake710HwDevManager(mode string, fdFlag, useAscendDocker, volcanoType bool) *HwDevManager {
	hdm := NewHwDevManager(mode, "/var/dlog", "/var/log/devicePlugin/")
	hdm.SetParameters(fdFlag, useAscendDocker, volcanoType)
	hdm.manager = NewFakeHwAscend710Manager()
	hdm.manager.SetDmgr(newFakeDeviceManager())
	return hdm
}
