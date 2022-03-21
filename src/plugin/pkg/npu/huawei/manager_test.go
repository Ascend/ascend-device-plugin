/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2020-2021. All rights reserved.
 */

package huawei

import (
	"Ascend-device-plugin/src/plugin/pkg/npu/common"
	"Ascend-device-plugin/src/plugin/pkg/npu/dsmi"
	"fmt"
	"go.uber.org/atomic"
	"huawei.com/npu-exporter/hwlog"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"os"
	"syscall"
	"testing"
	"time"
)

const (
	sleepNumTwo   = 2
	serverSockFd  = "/var/lib/kubelet/device-plugins/davinci-mini.sock"
	serverSock310 = "/var/lib/kubelet/device-plugins/Ascend310.sock"
)

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
		dmgr:     dsmi.NewFakeDeviceManager(),
		stopFlag: atomic.NewBool(false),
	}
	return fakeHwDevManager
}

func deleteServerSocketByDevManager(serverSocket string, manager *HwDevManager) {
	time.Sleep(sleepNumTwo * time.Second)
	manager.stopFlag.Store(true)
	hwlog.RunLog.Infof("remove serverSocket: %s", serverSocket)
	if err := os.Remove(serverSocket); err != nil {
		fmt.Println(err)
	}
	hwlog.RunLog.Infof("remove serverSocket success")
}

// TestSignalWatch for testSingalWatch
func TestSignalWatch(t *testing.T) {
	f, err := os.Create(serverSockFd)
	defer f.Close()
	if err != nil {
		t.Fatal("TestSignalWatch Run FAiled, reason is failed to create sock file")
	}
	if err := f.Chmod(common.SocketChmod); err != nil {
		t.Fatal("TestHwDevManager_Serve Run FAiled, reason is failed to Chmod")
	}
	watcher := NewFileWatch()
	err = watcher.watchFile(pluginapi.DevicePluginPath)
	if err != nil {
		t.Errorf("failed to create file watcher. %v", err)
	}
	defer watcher.fileWatcher.Close()
	hwlog.RunLog.Infof("Starting OS signs watcher.")
	osSignChan := newSignWatcher(syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
	hdm := HwDevManager{}
	useVolcanoType = false
	hps := NewHwPluginServe(&hdm, "")
	var restart bool
	go deleteServerSocket(serverSockFd)
	restart = hdm.signalWatch(watcher.fileWatcher, osSignChan, restart, hps, "")
	t.Logf("TestSignalWatch Run Pass")
}

func deleteServerSocket(serverSocket string) {
	time.Sleep(sleepNumTwo * time.Second)
	if err := os.Remove(serverSocket); err != nil {
		fmt.Println(err)
	}
}
