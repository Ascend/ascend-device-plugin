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
	"github.com/fsnotify/fsnotify"
	"go.uber.org/atomic"
	"huawei.com/npu-exporter/hwlog"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"os"
	"path"
	"strings"
	"syscall"
	"time"
)

type npuDevice struct {
	devType string
	pciID   string
	ID      string
	Health  string
}

// HwDevManager manages huawei device devices.
type HwDevManager struct {
	manager     devManager
	dlogPath    string
	runMode     string
	allDevTypes []string
	allDevs     []npuDevice
	defaultDevs []string
	stopFlag    *atomic.Bool
	dmgr        DeviceMgrInterface
}

var (
	// GetFdFlag to describe FdFlag
	GetFdFlag bool
	// UseAscendDocker to chose docker type
	UseAscendDocker bool
	useVolcanoType  bool

	// listAndWatchPeriod set listening device state period
	listAndWatchPeriod int

	// autoStowingDevs auto stowing fixes devices or not
	autoStowingDevs bool
)

type devManager interface {
	GetNPUs(*[]npuDevice, *[]string, string) error
	GetLogPath([]string, string, string, *string) error
	GetDevPath(string, string) (string, string)
	GetDevState(string, DeviceMgrInterface) string
	SetDmgr(DeviceMgrInterface)
	GetDmgr() DeviceMgrInterface
	GetMatchingDeviType() string
	GetPhyDevMapVirtualDev() map[uint32]string
}

// NewHwDevManager function is used to new a dev manager.
func NewHwDevManager(mode, dlogPath, logPath string) *HwDevManager {
	return &HwDevManager{
		dlogPath: dlogPath,
		runMode:  mode,
		dmgr:     NewDeviceManager(),
		stopFlag: atomic.NewBool(false),
	}
}

// GetNPUs get npu types
func (hdm *HwDevManager) GetNPUs() error {

	err := hdm.setRunMode()
	if err != nil {
		hwlog.Errorf("err to set Run mode, err: %v ", err)
		return err
	}

	switch hdm.runMode {
	case runMode310:
		hdm.manager = NewHwAscend310Manager()
	case runMode910:
		hdm.manager = NewHwAscend910Manager()
	case runMode710:
		hdm.manager = NewHwAscend710Manager()
	}
	hwlog.Infof("device plugin start")
	hdm.manager.SetDmgr(hdm.dmgr)

	if err := getDefaultDevices(&hdm.defaultDevs); err != nil {
		return err
	}

	if err := hdm.manager.GetNPUs(&hdm.allDevs, &hdm.allDevTypes, hdm.manager.GetMatchingDeviType()); err != nil {
		return err
	}

	return nil
}

// GetDevType get dev type
func (hdm *HwDevManager) GetDevType() []string {
	return hdm.allDevTypes
}

// Serve start grpc server
func (hdm *HwDevManager) Serve(devType, socketPath, pluginSocket string, pluginServerFunc func(*HwDevManager, string, string) HwPluginServeInterface) {
	// start sockPath monitor
	if !VerifyPath(socketPath) {
		hwlog.Errorf("socket path verify failed")
		return
	}
	pluginSockPath := path.Join(socketPath, pluginSocket)

	hwlog.Infof("Starting socket file watcher.")
	watcher := NewFileWatch()
	err := watcher.watchFile(pluginapi.DevicePluginPath)
	if err != nil {
		hwlog.Errorf("failed to create file watcher, err: %s", err.Error())
	}
	defer watcher.fileWatcher.Close()

	hwlog.Infof("Starting OS signs watcher.")
	osSignChan := newSignWatcher(syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)

	restart := true
	var hps HwPluginServeInterface
	for !hdm.stopFlag.Load() {
		if hdm.stopFlag.Load() {
			break
		}
		if restart {
			if hps != nil {
				hps.Stop()
			}
			// start
			hps = pluginServerFunc(hdm, devType, pluginSockPath)
			preStart(hps, pluginSockPath)
			// end
			if err := hps.Start(pluginSocket, pluginSockPath); err != nil {
				hwlog.Errorf("Could not contact Kubelet, retrying. " +
					"Did you enable the device plugin feature gate?")
				restart = true
			} else {
				restart = false
			}
		}
		// Monitor file signals and system signals
		restart = hdm.signalWatch(watcher.fileWatcher, osSignChan, restart, hps, pluginSockPath)
	}

}

func preStart(hps HwPluginServeInterface, pluginSockPath string) {
	for {
		err := hps.GetDevByType()
		if err == nil {
			break
		}
		// Use non-default level to avoid log spam.
		if logFlag {
			hwlog.Errorf("hwPluginServe preStart failed, err: %s", err.Error())
		}
		logFlag = false
		time.Sleep(sleepTime * time.Second)
	}
	logFlag = true
	hwlog.Infof("starting device-plugin server")
}

func (hdm *HwDevManager) signalWatch(watcher *fsnotify.Watcher, sigs chan os.Signal, restart bool, hps HwPluginServeInterface, pluginSockPath string) bool {
	// start sockPath monitor
	select {
	case event, signEnd := <-watcher.Events:
		if signEnd == false {
			hwlog.Infof("no watcher event, channel closed")
			return restart
		}
		if event.Name == pluginSockPath && event.Op&fsnotify.Remove == fsnotify.Remove {
			hwlog.Warnf("notify: sock file deleted, please check !")
		}
		if event.Name == pluginapi.KubeletSocket && event.Op&fsnotify.Create == fsnotify.Create {
			hwlog.Infof("notify: kubelet.sock file created, restarting.")
			return true
		}

	case s, signEnd := <-sigs:
		if signEnd == false {
			hwlog.Infof("no watcher sign event, channel closed")
			return restart
		}
		switch s {
		case syscall.SIGHUP:
			hwlog.Infof("Received SIGHUP, restarting.")
			return true
		default:
			hwlog.Infof("Received signal: %s, shutting down.", s.String())
			hps.Stop()
			hdm.dmgr.ShutDown()
			os.Exit(0)
		}
	}
	return restart
}

// SetParameters to set Parameters
func (hdm *HwDevManager) SetParameters(fdFlag, useAscendDocker, volcanoType, autoStowing bool, listWatchPeriod int) {
	GetFdFlag = fdFlag
	UseAscendDocker = useAscendDocker
	useVolcanoType = volcanoType
	listAndWatchPeriod = listWatchPeriod
	autoStowingDevs = autoStowing
}

func (hdm *HwDevManager) setRunMode() error {
	if hdm.runMode != "" {
		return nil
	}
	devNum, err := hdm.dmgr.GetDeviceCount()
	if err != nil || devNum == 0 {
		return err
	}
	chipinfo, err := hdm.dmgr.GetChipInfo(0)
	if err != nil {
		return err
	}

	if strings.Contains(chipinfo.ChipName, "310") {
		hdm.runMode = runMode310
		return nil
	}

	if strings.Contains(chipinfo.ChipName, "710") {
		hdm.runMode = runMode710
		return nil
	}

	hdm.runMode = runMode910
	return nil
}
