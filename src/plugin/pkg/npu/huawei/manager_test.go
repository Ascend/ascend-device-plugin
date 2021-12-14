/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2020-2021. All rights reserved.
 */

package huawei

import (
	"fmt"
	"go.uber.org/atomic"
	"huawei.com/npu-exporter/hwlog"
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
	errDir := os.MkdirAll("/var/lib/kubelet/device-plugins/", os.ModePerm)
	if errDir != nil {
		t.Fatal("TestHwDevManager_Serve Run FAiled, reason is failed to create folder file")
	}
	f, err := os.Create(serverSock310)
	if err != nil {
		hwlog.RunLog.Info(err)
		t.Fatal("TestHwDevManager_Serve Run FAiled, reason is failed to create sock file")
	}

	if err := f.Chmod(socketChmod); err != nil {
		t.Fatal("TestHwDevManager_Serve Run FAiled, reason is failed to Chmod")
	}
	if err := f.Close(); err != nil {
		t.Fatal("TestHwDevManager_Serve Run FAiled, reason is failed to Close")
	}
	go deleteServerSocketByDevManager(serverSock310, fakeHwDevManager)
	fakeHwDevManager.Serve("Ascend310", "/var/lib/kubelet/device-plugins/",
		"Ascend310.sock", NewFakeHwPluginServe)
	t.Logf("TestHwDevManager_Serve Run Pass")
}

func deleteServerSocketByDevManager(serverSocket string, manager *HwDevManager) {
	time.Sleep(sleepNumTwo * time.Second)
	manager.stopFlag.Store(true)
	hwlog.RunLog.Infof("remove serverSocket: %s", serverSocket)
	if err := os.Remove(serverSocket); err != nil {
		fmt.Println(err)
	}

}

// TestSignalWatch for testSingalWatch
func TestSignalWatch(t *testing.T) {
	f, err := os.Create(serverSockFd)
	if err != nil {
		t.Fatal("TestSignalWatch Run FAiled, reason is failed to create sock file")
	}
	if err := f.Chmod(socketChmod); err != nil {
		t.Fatal("TestHwDevManager_Serve Run FAiled, reason is failed to Chmod")
	}
	if err := f.Close(); err != nil {
		t.Fatal("TestHwDevManager_Serve Run FAiled, reason is failed to Close")
	}
	watcher := NewFileWatch()
	err = watcher.watchFile(pluginapi.DevicePluginPath)
	if err != nil {
		t.Errorf("failed to create file watcher. %v", err)
	}
	defer watcher.fileWatcher.Close()
	hwlog.RunLog.Infof("Starting OS signs watcher.")
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
	if err := os.Remove(serverSocket); err != nil {
		fmt.Println(err)
	}
}
