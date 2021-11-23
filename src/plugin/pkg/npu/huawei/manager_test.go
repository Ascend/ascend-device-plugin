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
	"Ascend-device-plugin/src/plugin/pkg/npu/hwlog"
	"go.uber.org/atomic"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"os"
	"syscall"
	"testing"
	"time"
)

const sleepNumTwo = 2

// TestHwDevManager_GetNPUs for getNpus
func TestHwDevManager_GetNPUs(t *testing.T) {
	fakeHwDevManager := createFakeDevManager("")
	err := fakeHwDevManager.GetNPUs()
	if err != nil {
		t.Fatal(err)
	}
	fakeHwDevManager = createFakeDevManager("ascend910")
	err = fakeHwDevManager.GetNPUs()
	if err != nil {
		t.Fatal(err)
	}
	fakeHwDevManager = createFakeDevManager("ascend710")
	err = fakeHwDevManager.GetNPUs()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("TestHwDevManager_GetNPUs Run Pass")
}

func createFakeDevManager(runMode string) *HwDevManager {
	fakeHwDevManager := &HwDevManager{
		runMode:  runMode,
		dmgr:     newFakeDeviceManager(),
		stopFlag: atomic.NewBool(false),
	}
	return fakeHwDevManager
}

// TestHwDevManager_Serve for serve
func TestHwDevManager_Serve(t *testing.T) {
	fakeHwDevManager := createFakeDevManager("")
	errDir := os.MkdirAll("/var/lib/kubelet/device-plugins/",os.ModePerm)
	if errDir != nil {
		t.Fatal("TestHwDevManager_Serve Run FAiled, reason is failed to create folder file")
	}
	f, err := os.Create(serverSock310)
	if err != nil {
		t.Fatal("TestHwDevManager_Serve Run FAiled, reason is failed to create sock file")
	}
	f.Chmod(socketChmod)
	f.Close()
	go deleteServerSocketByDevManager(serverSock310, fakeHwDevManager)
	fakeHwDevManager.Serve("Ascend310", "/var/lib/kubelet/device-plugins/",
		"Ascend310.sock", NewFakeHwPluginServe)
	t.Logf("TestHwDevManager_Serve Run Pass")
}

func deleteServerSocketByDevManager(serverSocket string, manager *HwDevManager) {
	time.Sleep(sleepNumTwo * time.Second)
	manager.stopFlag.Store(true)
	hwlog.Infof("remove serverSocket: %s", serverSocket)
	os.Remove(serverSocket)
}

// TestSignalWatch for testSingalWatch
func TestSignalWatch(t *testing.T) {
	f, err := os.Create(serverSockFd)
	if err != nil {
		t.Fatal("TestSignalWatch Run FAiled, reason is failed to create sock file")
	}
	f.Chmod(socketChmod)
	f.Close()
	watcher := NewFileWatch()
	err = watcher.watchFile(pluginapi.DevicePluginPath)
	if err != nil {
		t.Errorf("failed to create file watcher. %v", err)
	}
	defer watcher.fileWatcher.Close()
	hwlog.Infof("Starting OS signs watcher.")
	osSignChan := newSignWatcher(syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	hdm := HwDevManager{}
	useVolcanoType = true
	hps := NewHwPluginServe(&hdm, "", "")
	var restart bool
	go deleteServerSocket(serverSockFd)
	restart = hdm.signalWatch(watcher.fileWatcher, osSignChan, restart, hps, "")
	t.Logf("TestSignalWatch Run Pass")
}

func deleteServerSocket(serverSocket string) {
	time.Sleep(sleepNumTwo * time.Second)
	os.Remove(serverSocket)
}
