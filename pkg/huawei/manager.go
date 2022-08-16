/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2020-2022. All rights reserved.
 */
// Package huawei manager
package huawei

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"go.uber.org/atomic"
	"huawei.com/npu-exporter/devmanager"
	npuCommon "huawei.com/npu-exporter/devmanager/common"
	"huawei.com/npu-exporter/hwlog"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"Ascend-device-plugin/pkg/common"
)

// HwDevManager manages huawei device devices.
type HwDevManager struct {
	manager     devManager
	runMode     string
	allDevTypes []string
	allDevs     []common.NpuDevice
	defaultDevs []string
	stopFlag    *atomic.Bool
	dmgr        devmanager.DeviceInterface
}

type devManager interface {
	GetNPUs(*[]common.NpuDevice, *[]string, string) error
	GetDevPath(string, string) (string, string)
	GetDevState(string, devmanager.DeviceInterface) string
	SetDmgr(devmanager.DeviceInterface)
	GetDmgr() devmanager.DeviceInterface
	GetMatchingDeviType() string
	GetPhyDevMapVirtualDev() map[int32]string
	DoWithVolcanoListAndWatch(*HwPluginServe)
	GetDeviceNetworkState(int32, *common.NpuDevice) (string, error)
	GetAnnotationMap(sets.String, []string) map[string]string
}

// NewHwDevManager function is used to new a dev manager.
func NewHwDevManager(mode string) *HwDevManager {
	devM, err := devmanager.AutoInit("")
	if err != nil {
		hwlog.RunLog.Errorf("init hw dev manager failed, err: %v", err)
		return nil
	}
	switch devM.DevType {
	case npuCommon.Ascend310:
		mode = common.RunMode310
	case npuCommon.Ascend310P:
		mode = common.RunMode310P
	case npuCommon.Ascend910:
		mode = common.RunMode910
	default:
	}
	return &HwDevManager{
		runMode:  mode,
		dmgr:     devM,
		stopFlag: atomic.NewBool(false),
	}
}

// GetNPUs get npu types
func (hdm *HwDevManager) GetNPUs() error {
	switch hdm.runMode {
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
	hwlog.RunLog.Infof("device plugin start")
	hdm.manager.SetDmgr(hdm.dmgr)

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
func (hdm *HwDevManager) Serve(devType string, stop chan struct{}) {
	if stop == nil {
		hwlog.RunLog.Errorf("stop channel is nil")
		return
	}
	// start sockPath monitor
	hwlog.RunLog.Infof("starting the inspection of register devType %v", devType)
	pluginSockPath, watcher, err := hdm.createSignWatchServe(devType)
	if err != nil {
		return
	}
	defer func() {
		if watcher == nil {
			return
		}
		if err := watcher.fileWatcher.Close(); err != nil {
			hwlog.RunLog.Errorf("close file watcher, err: %s", err.Error())
		}
	}()
	restart := true
	var hps HwPluginServeInterface
	for !hdm.stopFlag.Load() {
		if hdm.stopFlag.Load() {
			break
		}
		if restart {
			var err error
			restart, err = hdm.reStartServe(&hps, devType, pluginSockPath)
			if err != nil {
				return
			}
			time.Sleep(sleepTime * time.Second)
		}
		// Monitor file signals and system signals
		osSignChan := newSignWatcher(syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
		restart = hdm.signalWatch(watcher.fileWatcher, osSignChan, restart, hps, pluginSockPath)
	}
	stopCount.Add(1)
	if int(stopCount.Load()) == len(hdm.GetDevType()) {
		hdm.dmgr.ShutDown()
		stop <- struct{}{}
	}
}

func (hdm *HwDevManager) createSignWatchServe(devType string) (string, *FileWatch, error) {
	realDevSockPath, isOk := VerifyPath(v1beta1.DevicePluginPath)
	if !isOk {
		hwlog.RunLog.Errorf("socket path verify failed!")
		return "", nil, fmt.Errorf("socket path verify failed")
	}
	pluginSockPath := path.Join(realDevSockPath, fmt.Sprintf("%s.sock", devType))
	hwlog.RunLog.Infof("starting socket file watcher.")
	watcher := NewFileWatch()
	if watcher == nil {
		hwlog.RunLog.Errorf("failed to create file watcher")
		return "", nil, fmt.Errorf("failed to create file watcher")
	}
	if err := watcher.watchFile(realDevSockPath); err != nil {
		hwlog.RunLog.Errorf("failed to create file watcher, err: %s", err.Error())
		return "", watcher, err
	}
	return pluginSockPath, watcher, nil
}

func (hdm *HwDevManager) reStartServe(hps *HwPluginServeInterface, devType, pluginSockPath string) (bool, error) {
	if (*hps) != nil {
		(*hps).Stop()
	}
	// restart serve
	*hps = NewHwPluginServe(hdm, devType)
	if *hps == nil {
		hwlog.RunLog.Errorf("failed to create kube interactor")
		return false, fmt.Errorf("failed to create kube interactor")
	}
	preStart(*hps)
	if err := (*hps).Start(pluginSockPath); err != nil {
		hwlog.RunLog.Errorf("Could not contact Kubelet, retrying. " +
			"Did you enable the device plugin feature gate?")
		return true, nil
	}
	return false, nil
}

func preStart(hps HwPluginServeInterface) {
	for {
		err := hps.GetDevByType()
		if err == nil {
			break
		}
		// Use non-default level to avoid log spam.
		if logFlag {
			hwlog.RunLog.Warnf("hwPluginServe preStart, message: %s", err.Error())
		}
		logFlag = false
		time.Sleep(sleepTime * time.Second)
	}
	logFlag = true
	hwlog.RunLog.Infof("starting device-plugin server")
}

func (hdm *HwDevManager) signalWatch(watcher *fsnotify.Watcher, sigs chan os.Signal, restart bool,
	hps HwPluginServeInterface, pluginSockPath string) bool {
	if sigs == nil {
		return false
	}
	// start sockPath monitor
	select {
	case event, signEnd := <-watcher.Events:
		if signEnd == false {
			hwlog.RunLog.Infof("no watcher event, channel closed")
			return restart
		}
		if event.Name == pluginSockPath && event.Op&fsnotify.Remove == fsnotify.Remove {
			hwlog.RunLog.Warnf("notify: sock file deleted, please check !")
		}
		if event.Name == v1beta1.KubeletSocket && event.Op&fsnotify.Create == fsnotify.Create {
			hwlog.RunLog.Infof("notify: kubelet.sock file created, restarting.")
			return true
		}

	case s, signEnd := <-sigs:
		if signEnd == false {
			hwlog.RunLog.Infof("no watcher sign event, channel closed")
			return restart
		}
		switch s {
		case syscall.SIGHUP:
			hwlog.RunLog.Infof("Received SIGHUP, restarting.")
			return true
		default:
			hwlog.RunLog.Infof("Received signal: %s, shutting down.", s.String())
			hdm.stopFlag.Store(true)
			hps.Stop()
		}
	}
	return restart
}
