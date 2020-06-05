/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2019-2024. All rights reserved.
 * Description: manager.go
 * Create: 19-11-20 下午8:52
 */

package huawei

import (
	"github.com/fsnotify/fsnotify"
	"go.uber.org/zap"
	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"
	"os"
	"path"
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

// GetFdFlag to describe FdFlag
var GetFdFlag bool

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
	if hdm.runMode == "pci" {
		// hdm.manager = NewHwPCIManager()
		return nil
	} else if hdm.runMode == "vnpu" {
		// hdm.manager = NewHwVNPUManager()
		return nil
	} else if hdm.runMode == "ascend310" {
		hdm.manager = NewHwAscend310Manager()
	} else {
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
func (hdm *HwDevManager) Serve(devType, socketPath, k8sSocket, pluginSocket string, fdFlag *bool) {
	GetFdFlag = *fdFlag
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

	for {
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
		logger.Error("hwPluginServe.PreStart() failed", zap.String("err", err.Error()))
		time.Sleep(sleepTime * time.Second)
	}
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
