/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2020-2022. All rights reserved.
 */

package huawei

import (
	"fmt"
	"os"
	"reflect"
	"syscall"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"go.uber.org/atomic"
	"huawei.com/npu-exporter/hwlog"
	"huawei.com/npu-exporter/utils"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"Ascend-device-plugin/src/plugin/pkg/npu/common"
	"Ascend-device-plugin/src/plugin/pkg/npu/dsmi"
)

const (
	sleepNumTwo   = 2
	serverSockFd  = "/var/lib/kubelet/device-plugins/davinci-mini.sock"
	serverSock310 = "/var/lib/kubelet/device-plugins/Ascend310.sock"
)

// TestHwDevManagerGetNPUs for getNpus
func TestHwDevManagerGetNPUs(t *testing.T) {
	fakeHwDevManager := createFakeDevManager("")
	fakeHwDevManager.runMode = common.RunMode310
	err := fakeHwDevManager.GetNPUs()
	if err != nil {
		t.Fatal(err)
	}
	fakeHwDevManager = createFakeDevManager("ascend910")
	fakeHwDevManager.runMode = common.RunMode910
	err = fakeHwDevManager.GetNPUs()
	if err != nil {
		t.Fatal(err)
	}
	fakeHwDevManager = createFakeDevManager("ascend710")
	fakeHwDevManager.runMode = common.RunMode710
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
	if err := utils.MakeSureDir(serverSockFd); err != nil {
		t.Fatal("TestSignalWatch Run FAiled, reason is failed to create sock file dir")
	}
	f, err := os.Create(serverSockFd)
	defer f.Close()
	if err != nil {
		t.Fatal("TestSignalWatch Run FAiled, reason is failed to create sock file")
	}
	if err := f.Chmod(common.SocketChmod); err != nil {
		t.Fatal("TestHwDevManager_Serve Run FAiled, reason is failed to Chmod")
	}
	watcher := NewFileWatch()
	err = watcher.watchFile(v1beta1.DevicePluginPath)
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

// TestSetRunMode for SetRunMode
func TestSetRunMode(t *testing.T) {
	convey.Convey("TestSetRunMode", t, func() {
		convey.Convey("runMode is not empty", func() {
			hdm := &HwDevManager{runMode: common.RunMode310, dmgr: dsmi.NewFakeDeviceManager()}
			convey.So(hdm.SetRunMode(), convey.ShouldBeNil)
		})
		convey.Convey("GetDeviceCount failed", func() {
			hdm := &HwDevManager{dmgr: dsmi.NewFakeDeviceManager()}
			mock := gomonkey.ApplyMethod(reflect.TypeOf(new(dsmi.FakeDeviceManager)), "GetDeviceCount",
				func(_ *dsmi.FakeDeviceManager) (int32, error) { return 0, fmt.Errorf("err") })
			defer mock.Reset()
			convey.So(hdm.SetRunMode(), convey.ShouldNotBeNil)
		})
		convey.Convey("GetChipInfo failed", func() {
			hdm := &HwDevManager{dmgr: dsmi.NewFakeDeviceManager()}
			mock := gomonkey.ApplyMethod(reflect.TypeOf(new(dsmi.FakeDeviceManager)), "GetChipInfo",
				func(_ *dsmi.FakeDeviceManager, _ int32) (string, error) { return "", fmt.Errorf("err") })
			defer mock.Reset()
			convey.So(hdm.SetRunMode(), convey.ShouldNotBeNil)
		})
		convey.Convey("310", func() {
			hdm := &HwDevManager{dmgr: dsmi.NewFakeDeviceManager()}
			mock := gomonkey.ApplyMethod(reflect.TypeOf(new(dsmi.FakeDeviceManager)), "GetChipInfo",
				func(_ *dsmi.FakeDeviceManager, _ int32) (string, error) { return "310", nil })
			defer mock.Reset()
			convey.So(hdm.SetRunMode(), convey.ShouldBeNil)
		})
		convey.Convey("710", func() {
			hdm := &HwDevManager{dmgr: dsmi.NewFakeDeviceManager()}
			mock := gomonkey.ApplyMethod(reflect.TypeOf(new(dsmi.FakeDeviceManager)), "GetChipInfo",
				func(_ *dsmi.FakeDeviceManager, _ int32) (string, error) { return "710", nil })
			defer mock.Reset()
			convey.So(hdm.SetRunMode(), convey.ShouldBeNil)
		})
		convey.Convey("910", func() {
			hdm := &HwDevManager{dmgr: dsmi.NewFakeDeviceManager()}
			mock := gomonkey.ApplyMethod(reflect.TypeOf(new(dsmi.FakeDeviceManager)), "GetChipInfo",
				func(_ *dsmi.FakeDeviceManager, _ int32) (string, error) { return "910", nil })
			defer mock.Reset()
			convey.So(hdm.SetRunMode(), convey.ShouldBeNil)
		})
		convey.Convey("invalid mode", func() {
			hdm := &HwDevManager{dmgr: dsmi.NewFakeDeviceManager()}
			mock := gomonkey.ApplyMethod(reflect.TypeOf(new(dsmi.FakeDeviceManager)), "GetChipInfo",
				func(_ *dsmi.FakeDeviceManager, _ int32) (string, error) { return "110", nil })
			defer mock.Reset()
			convey.So(hdm.SetRunMode(), convey.ShouldNotBeNil)
		})
	})
}
