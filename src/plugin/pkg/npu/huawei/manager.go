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
	"go.uber.org/zap"
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
	serves      map[string]*HwPluginServe
	manager     devManager
	dlogPath    string
	runMode     string
	allDevTypes []string
	allDevs     []npuDevice
	defaultDevs []string
}

var (
	// GetFdFlag to describe FdFlag
	GetFdFlag bool
	// UseAscendDocker to chose docker type
	UseAscendDocker bool
	useVolcanoType  bool
)

type devManager interface {
	GetNPUs(*[]npuDevice, *[]string) error
	GetDefaultDevs(*[]string) error
	GetDevState(string) string
	GetDevPath(string, *string, *string) error
	GetLogPath([]string, string, *string) error
}

// NewHwDevManager function is used to new a dev manager.
func NewHwDevManager(mode, dlogPath string) *HwDevManager {
	return &HwDevManager{
		dlogPath: dlogPath,
		runMode:  mode,
		serves:   make(map[string]*HwPluginServe),
	}
}

// GetNPUs get npu types
func (hdm *HwDevManager) GetNPUs(timeInterval, checkNum, restoreNum, highThreshold, lowThreshold string, netDetect bool) error {
	// start dsmi in contaioner
	err := enableContainerService()
	if err != nil {
		logger.Error("enable container Service failed. error", zap.String("error", err.Error()))
	}

	err = hdm.setRunMode()
	if err != nil {
		logger.Error("err to set Run mode ", zap.Error(err))
		return err
	}
	switch hdm.runMode {
	case runMode310:
		hdm.manager = NewHwAscend310Manager()
	case runMode910:
		hdm.manager = NewHwAscend910Manager(timeInterval, checkNum, restoreNum, highThreshold, lowThreshold, netDetect)
		logger.Info("device plugin start",
			zap.String("time", timeInterval),
			zap.String("check", checkNum),
			zap.String("restore", restoreNum),
			zap.String("high_threshold", highThreshold),
			zap.String("low_threshold", lowThreshold),
			zap.Bool("netDetect", netDetect))
	}

	if err := hdm.manager.GetDefaultDevs(&hdm.defaultDevs); err != nil {
		return err
	}

	if err := hdm.manager.GetNPUs(&hdm.allDevs, &hdm.allDevTypes); err != nil {
		return err
	}

	return nil
}

// GetDevType get dev type
func (hdm *HwDevManager) GetDevType() []string {
	return hdm.allDevTypes
}

// Serve start grpc server
func (hdm *HwDevManager) Serve(devType, socketPath, k8sSocket, pluginSocket string) {
	// start sockPath monitor
	logger.Info("the log path is :", zap.String("logPath", LogPath))
	pluginSockPath := path.Join(socketPath, pluginSocket)

	logger.Info("Starting socket file watcher.")
	watcher := NewFileWatch()
	err := watcher.watchFile(pluginapi.DevicePluginPath)
	if err != nil {
		logger.Error("failed to create file watcher.", zap.String("err", err.Error()))
	}
	defer watcher.fileWatcher.Close()

	logger.Info("Starting OS signs watcher.")
	osSignChan := newSignWatcher(syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	restart := true
	var hps *HwPluginServe
	nerverStop := true
	for nerverStop {
		if restart {
			if hps != nil {
				hps.Stop()
			}
			// start
			hps = NewHwPluginServe(hdm, devType, pluginSockPath)
			hdm.serves[devType] = hps
			preStart(hps, pluginSockPath)
			// end
			if err := hps.Start(socketPath, k8sSocket, pluginSocket, pluginSockPath); err != nil {
				logger.Error("Could not contact Kubelet, retrying. Did you enable the device plugin feature gate?")
			}
			restart = false

		}
		// Monitor file signals and system signals
		restart = hdm.signalWatch(watcher.fileWatcher, osSignChan, restart, hps)
	}

}

func preStart(hps *HwPluginServe, pluginSockPath string) {
	for {
		err := hps.GetDevByType()
		if err == nil {
			break
		}
		// Use non-default level to avoid log spam.
		if logFlag {
			logger.Error("hwPluginServe.PreStart() failed", zap.String("err", err.Error()))
		}
		logFlag = false
		time.Sleep(sleepTime * time.Second)
	}
	logFlag = true
	logger.Info("starting device-plugin server at:", zap.String("pluginSockPath", pluginSockPath))
}

func (hdm *HwDevManager) signalWatch(watcher *fsnotify.Watcher, sigs chan os.Signal, restart bool, hps *HwPluginServe) bool {

	// start sockPath monitor
	select {
	case event, signEnd := <-watcher.Events:
		if signEnd == false {
			logger.Info("no watcher event, channel closed")
			return restart
		}
		if event.Name == serverSock && event.Op&fsnotify.Remove == fsnotify.Remove {
			logger.Warn("notify: file deleted, please check !", zap.String("fileName", serverSock))
		}
		if event.Name == pluginapi.KubeletSocket && event.Op&fsnotify.Create == fsnotify.Create {
			logger.Info("notify: file created, restarting.", zap.String("fileName", pluginapi.KubeletSocket))
			restart = true
			return restart
		}

	case s, signEnd := <-sigs:
		if signEnd == false {
			logger.Info("no watcher sign event, channel closed")
			return restart
		}
		switch s {
		case syscall.SIGHUP:
			logger.Info("Received SIGHUP, restarting.")
			restart = true
			return restart
		default:
			logger.Info("Received signal, shutting down.", zap.String("signal", s.String()))
			hps.Stop()
			os.Exit(0)
		}
	}
	return restart

}

// SetParameters to set Parameters
func (hdm *HwDevManager) SetParameters(fdFlag, useAscendDocker, volcanoType *bool) {
	GetFdFlag = *fdFlag
	UseAscendDocker = *useAscendDocker
	useVolcanoType = *volcanoType
}

func (hdm *HwDevManager) setRunMode() error {
	if hdm.runMode != "" {
		return nil
	}
	devNum, err := getDeviceCount()
	if err != nil && devNum == 0 {
		return err
	}

	chipinfo, err := getChipInfo(0)
	if err != nil {
		return err
	}

	if strings.Contains(chipinfo.ChipName, "310") {
		hdm.runMode = runMode310
		return nil
	}

	hdm.runMode = runMode910
	return nil
}
