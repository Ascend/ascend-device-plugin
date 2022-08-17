/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2020-2022. All rights reserved.
 */
// Package device manager
package device

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
	"huawei.com/npu-exporter/devmanager"
	npuCommon "huawei.com/npu-exporter/devmanager/common"
	"huawei.com/npu-exporter/hwlog"
	"huawei.com/npu-exporter/utils"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"Ascend-device-plugin/pkg/common"
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
	fakeHwDevManager = createFakeDevManager("ascend310P")
	fakeHwDevManager.runMode = common.RunMode310P
	err = fakeHwDevManager.GetNPUs()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("TestHwDevManager_GetNPUs Run Pass")
}

func createFakeDevManager(runMode string) *HwDevManager {
	fakeHwDevManager := &HwDevManager{
		runMode:  runMode,
		dmgr:     &devmanager.DeviceManagerMock{},
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

// TestSetRunModeFailed for SetRunMode
func TestSetRunModeFailed(t *testing.T) {
	convey.Convey("TestSetRunMode", t, func() {
		convey.Convey("GetDeviceCount failed", func() {
			hdm := &HwDevManager{dmgr: &devmanager.DeviceManagerMock{}}
			mock := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)), "GetDeviceCount",
				func(_ *devmanager.DeviceManagerMock) (int32, error) { return 0, fmt.Errorf("err") })
			defer mock.Reset()
			convey.So(hdm.SetRunMode(), convey.ShouldNotBeNil)
		})
		convey.Convey("GetChipInfo failed", func() {
			hdm := &HwDevManager{dmgr: &devmanager.DeviceManagerMock{}}
			mock := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)), "GetChipInfo",
				func(_ *devmanager.DeviceManagerMock, _ int32) (*npuCommon.ChipInfo, error) {
					return &npuCommon.ChipInfo{
						Type:    "ascend",
						Name:    "910",
						Version: "v1",
					}, fmt.Errorf("err")
				})
			defer mock.Reset()
			convey.So(hdm.SetRunMode(), convey.ShouldNotBeNil)
		})
		convey.Convey("invalid mode", func() {
			hdm := &HwDevManager{dmgr: &devmanager.DeviceManagerMock{}}
			mock := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)), "GetChipInfo",
				func(_ *devmanager.DeviceManagerMock, _ int32) (*npuCommon.ChipInfo, error) {
					return &npuCommon.ChipInfo{}, nil
				})
			defer mock.Reset()
			convey.So(hdm.SetRunMode(), convey.ShouldNotBeNil)
		})
	})
}

// TestSetRunModeFailed for SetRunMode
func TestSetRunModeSuccess(t *testing.T) {
	convey.Convey("TestSetRunMode", t, func() {
		convey.Convey("runMode is not empty", func() {
			hdm := &HwDevManager{runMode: common.RunMode310, dmgr: &devmanager.DeviceManagerMock{}}
			convey.So(hdm.SetRunMode(), convey.ShouldBeNil)
		})
		convey.Convey("310", func() {
			hdm := &HwDevManager{dmgr: &devmanager.DeviceManagerMock{}}
			mock := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)), "GetChipInfo",
				func(_ *devmanager.DeviceManagerMock, _ int32) (*npuCommon.ChipInfo, error) {
					return &npuCommon.ChipInfo{
						Type:    "ascend",
						Name:    "910",
						Version: "v1",
					}, nil
				})
			defer mock.Reset()
			convey.So(hdm.SetRunMode(), convey.ShouldBeNil)
		})
		convey.Convey("310P", func() {
			hdm := &HwDevManager{dmgr: &devmanager.DeviceManagerMock{}}
			mock := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)), "GetChipInfo",
				func(_ *devmanager.DeviceManagerMock, _ int32) (*npuCommon.ChipInfo, error) {
					return &npuCommon.ChipInfo{
						Type:    "ascend",
						Name:    "910",
						Version: "v1",
					}, nil
				})
			defer mock.Reset()
			convey.So(hdm.SetRunMode(), convey.ShouldBeNil)
		})
		convey.Convey("910", func() {
			hdm := &HwDevManager{dmgr: &devmanager.DeviceManagerMock{}}
			mock := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)), "GetChipInfo",
				func(_ *devmanager.DeviceManagerMock, _ int32) (*npuCommon.ChipInfo, error) {
					return &npuCommon.ChipInfo{
						Type:    "ascend",
						Name:    "910",
						Version: "v1",
					}, nil
				})
			defer mock.Reset()
			convey.So(hdm.SetRunMode(), convey.ShouldBeNil)
		})
	})
}

// TestNewHwDevManager for NewHwDevManager
func TestNewHwDevManager(t *testing.T) {
	convey.Convey("TestNewHwDevManager", t, func() {
		convey.Convey("AutoInit failed", func() {
			mock := gomonkey.ApplyFunc(devmanager.AutoInit, func(_ string) (*devmanager.DeviceManager, error) {
				return nil, fmt.Errorf("error")
			})
			defer mock.Reset()
			devM := NewHwDevManager("")
			convey.So(devM, convey.ShouldBeNil)
		})
		convey.Convey("DevType is 910", func() {
			mock0 := gomonkey.ApplyFunc(devmanager.AutoInit, func(_ string) (*devmanager.DeviceManager, error) {
				return &devmanager.DeviceManager{DevType: npuCommon.Ascend910}, nil
			})
			defer mock0.Reset()
			devM0 := NewHwDevManager("")
			convey.So(devM0, convey.ShouldNotBeNil)
			convey.So(devM0.runMode, convey.ShouldEqual, common.RunMode910)
		})
	})
}
