// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package device a series of device function
package device

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"huawei.com/npu-exporter/devmanager"
	"huawei.com/npu-exporter/hwlog"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/kubeclient"
	"Ascend-device-plugin/pkg/server"
)

// HwDevManager manages huawei device devices.
type HwDevManager struct {
	groupDevice map[string][]*common.NpuDevice
	ServerMap   map[string]server.InterfaceServer
	AllDevTypes []string
	AllDevs     []common.NpuDevice
	manager     devManager
	RunMode     string
}

// NewHwDevManager function is used to new a dev manager.
func NewHwDevManager(devM devmanager.DeviceInterface, client *kubeclient.ClientK8s) *HwDevManager {
	var hdm HwDevManager
	if err := hdm.setRunMode(devM.GetDevType()); err != nil {
		hwlog.RunLog.Errorf("set runmode failed, err: %v", err)
		return nil
	}
	if err := hdm.setAscendManager(devM, client); err != nil {
		hwlog.RunLog.Errorf("init hw dev manager failed, err: %v", err)
		return nil
	}
	if err := hdm.setAllDeviceAndType(); err != nil {
		hwlog.RunLog.Errorf("set all device and type failed, err: %v", err)
		return nil
	}

	return &hdm
}

func (hdm *HwDevManager) setRunMode(devType string) error {
	switch devType {
	case common.Ascend310:
		hdm.RunMode = common.RunMode310
	case common.Ascend310P:
		hdm.RunMode = common.RunMode310P
	case common.Ascend910:
		hdm.RunMode = common.RunMode910
	default:
		return errors.New("an unsupported device type")
	}
	return nil
}

func (hdm *HwDevManager) setAscendManager(dmgr devmanager.DeviceInterface, client *kubeclient.ClientK8s) error {
	switch hdm.RunMode {
	case common.RunMode310:
		hdm.manager = NewHwAscend310Manager()
	case common.RunMode910:
		hdm.manager = NewHwAscend910Manager()
	case common.RunMode310P:
		hdm.manager = NewHwAscend310PManager()
	default:
		hwlog.RunLog.Errorf("found an unsupported device type")
		return errors.New("an unsupported device type")
	}
	hdm.manager.SetDmgr(dmgr)
	if common.ParamOption.UseVolcanoType && client != nil {
		hdm.manager.SetKubeClient(client)
	}
	return nil
}

func (hdm *HwDevManager) setAllDeviceAndType() error {
	return hdm.manager.GetNPUs(&hdm.AllDevs, &hdm.AllDevTypes)
}

// Serve Serve function
func (hdm *HwDevManager) Serve(ctx context.Context) {
	// initiate a global socket path watcher
	hwlog.RunLog.Info("Serve start")
	watcher, err := common.NewFileWatch()
	if err != nil {
		hwlog.RunLog.Error("createSocketWatcher error")
		return
	}
	defer func() {
		if watcher == nil {
			hwlog.RunLog.Error("watcher is nil")
			return
		}
		if err := watcher.FileWatcher.Close(); err != nil {
			hwlog.RunLog.Errorf("close file watcher, err: %s", err.Error())
		}
	}()

	// create restart signal
	restartSignal := common.NewSignWatcher(syscall.SIGHUP)

	for {
		allSuccess := hdm.startAllServer(watcher)
		if hdm.handleEvents(ctx, restartSignal, watcher) {
			break
		}
		if !allSuccess {
			time.Sleep(common.SleepTime * time.Second)
		}
	}

	hdm.stopAllSever()
}

func (hdm *HwDevManager) handleEvents(ctx context.Context, restartSignal chan os.Signal,
	watcher *common.FileWatch) bool {

	if restartSignal == nil {
		hwlog.RunLog.Error("the restart signal is not initialized")
		return true
	}

	select {
	case <-ctx.Done():
		hwlog.RunLog.Info("stop signal received, stop device plugin")
		return true
	case sig, ok := <-restartSignal:
		if ok {
			hwlog.RunLog.Infof("restart signal %s received, restart device plugin", sig)
			hdm.setRestartForAll()
		}
	case event := <-watcher.FileWatcher.Events:
		if event.Op&fsnotify.Remove == fsnotify.Remove {
			_, deleteFile := filepath.Split(event.Name)
			hdm.handleDeleteEvent(deleteFile)
		}
		if event.Name == v1beta1.KubeletSocket && event.Op&fsnotify.Create == fsnotify.Create {
			hwlog.RunLog.Infof("notify: kubelet.sock file created, restarting.")
			hdm.setRestartForAll()
		}
	}
	return false
}

func (hdm *HwDevManager) stopAllSever() {
	for deviceType := range hdm.ServerMap {
		hwlog.RunLog.Infof("stop server type %s", deviceType)
		hdm.ServerMap[deviceType].Stop()
	}
	hwlog.RunLog.Infof("stop all server done")
}

func (hdm *HwDevManager) setRestartForAll() {
	for deviceType := range hdm.ServerMap {
		hdm.ServerMap[deviceType].SetRestartFlag(true)
	}
}

func (hdm *HwDevManager) startAllServer(socketWatcher *common.FileWatch) bool {
	success := true
	for deviceType := range hdm.ServerMap {
		if hdm.ServerMap[deviceType].GetRestartFlag() {
			if err := hdm.ServerMap[deviceType].Start(socketWatcher); err != nil {
				hwlog.RunLog.Errorf("Could not contact Kubelet for %s, retrying. "+
					"Did you enable the device plugin feature gate?", deviceType)
				success = false
			} else {
				hdm.ServerMap[deviceType].SetRestartFlag(false)
			}
		}
	}
	return success
}

func (hdm *HwDevManager) handleDeleteEvent(deleteFile string) {
	for deviceType := range hdm.ServerMap {
		candidateSocketFilename := fmt.Sprintf("%s.sock", deviceType)
		if candidateSocketFilename == deleteFile {
			hwlog.RunLog.Warnf("notify: sock file %s deleted, please check !", deleteFile)
		}
	}
}
