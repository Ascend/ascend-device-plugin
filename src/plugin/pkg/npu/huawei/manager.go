/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2020-2021. All rights reserved.
 */

package huawei

import (
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"go.uber.org/atomic"
	"huawei.com/npu-exporter/hwlog"
	"k8s.io/apimachinery/pkg/util/sets"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"os"
	"path"
	"strings"
	"syscall"
	"time"
)

type npuDevice struct {
	devType       string
	pciID         string
	ID            string
	Health        string
	networkHealth string
}

// HwDevManager manages huawei device devices.
type HwDevManager struct {
	manager     devManager
	runMode     string
	allDevTypes []string
	allDevs     []npuDevice
	defaultDevs []string
	stopFlag    *atomic.Bool
	dmgr        DeviceMgrInterface
}

// Option option
type Option struct {
	// GetFdFlag to describe FdFlag
	GetFdFlag bool
	// UseAscendDocker to chose docker type
	UseAscendDocker bool
	UseVolcanoType  bool

	// ListAndWatchPeriod set listening device state period
	ListAndWatchPeriod int

	// AutoStowingDevs auto stowing fixes devices or not
	AutoStowingDevs bool

	KubeConfig string
}

var (
	// GetFdFlag to describe FdFlag
	GetFdFlag bool
	// UseAscendDocker to chose docker type
	UseAscendDocker bool
	useVolcanoType  bool

	// ListAndWatchPeriod set listening device state period
	listAndWatchPeriod int

	// AutoStowingDevs auto stowing fixes devices or not
	autoStowingDevs bool

	kubeConfig string
	// switch error log
	logFlag = true
)

type devManager interface {
	GetNPUs(*[]npuDevice, *[]string, string) error
	GetDevPath(string, string) (string, string)
	GetDevState(string, DeviceMgrInterface) string
	SetDmgr(DeviceMgrInterface)
	GetDmgr() DeviceMgrInterface
	GetMatchingDeviType() string
	GetPhyDevMapVirtualDev() map[uint32]string
	DoWithVolcanoListAndWatch(*HwPluginServe, bool)
	GetDeviceNetworkState(int32, *npuDevice) (string, error)
	GetAnnotationMap(sets.String, string) map[string]string
}

// NewHwDevManager function is used to new a dev manager.
func NewHwDevManager(mode string) *HwDevManager {
	return &HwDevManager{
		runMode:  mode,
		dmgr:     NewDeviceManager(),
		stopFlag: atomic.NewBool(false),
	}
}

// GetNPUs get npu types
func (hdm *HwDevManager) GetNPUs() error {

	if err := hdm.setRunMode(); err != nil {
		hwlog.RunLog.Errorf("err to set Run mode, err: %v ", err)
		return err
	}

	switch hdm.runMode {
	case runMode310:
		hdm.manager = NewHwAscend310Manager()
	case runMode910:
		hdm.manager = NewHwAscend910Manager()
	case runMode710:
		hdm.manager = NewHwAscend710Manager()
	default:
		hwlog.RunLog.Errorf("found an unsupported device type")
		return errors.New("an unsupported device type")
	}
	hwlog.RunLog.Infof("device plugin start")
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
func (hdm *HwDevManager) Serve(devType string, pluginServerFunc func(*HwDevManager,
	string) HwPluginServeInterface) {
	// start sockPath monitor
	hwlog.RunLog.Infof("Starting check device socket file path.")
	realDevSockPath, isOk := VerifyPath(pluginapi.DevicePluginPath)
	if !isOk {
		hwlog.RunLog.Errorf("socket path verify failed!")
		return
	}
	pluginSockPath := path.Join(realDevSockPath, fmt.Sprintf("%s.sock", devType))
	hwlog.RunLog.Infof("Starting socket file watcher.")
	watcher := NewFileWatch()
	if err := watcher.watchFile(realDevSockPath); err != nil {
		hwlog.RunLog.Errorf("failed to create file watcher, err: %s", err.Error())
	}
	defer func() {
		if err := watcher.fileWatcher.Close(); err != nil {
			hwlog.RunLog.Errorf("close file watcher, err: %s", err.Error())
		}
	}()

	hwlog.RunLog.Infof("Starting OS signs watcher.")
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
			hps = pluginServerFunc(hdm, devType)
			preStart(hps)
			// end
			if err := hps.Start(pluginSockPath); err != nil {
				hwlog.RunLog.Errorf("Could not contact Kubelet, retrying. " +
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

func preStart(hps HwPluginServeInterface) {
	for {
		err := hps.GetDevByType()
		if err == nil {
			break
		}
		// Use non-default level to avoid log spam.
		if logFlag {
			hwlog.RunLog.Errorf("hwPluginServe preStart failed, err: %s", err.Error())
		}
		logFlag = false
		time.Sleep(sleepTime * time.Second)
	}
	logFlag = true
	hwlog.RunLog.Infof("starting device-plugin server")
}

func (hdm *HwDevManager) signalWatch(watcher *fsnotify.Watcher, sigs chan os.Signal, restart bool, hps HwPluginServeInterface, pluginSockPath string) bool {
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
		if event.Name == pluginapi.KubeletSocket && event.Op&fsnotify.Create == fsnotify.Create {
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
			hps.Stop()
			hdm.dmgr.ShutDown()
			os.Exit(0)
		}
	}
	return restart
}

// SetParameters to set Parameters
func (hdm *HwDevManager) SetParameters(option Option) {
	GetFdFlag = option.GetFdFlag
	UseAscendDocker = option.UseAscendDocker
	useVolcanoType = option.UseVolcanoType
	listAndWatchPeriod = option.ListAndWatchPeriod
	autoStowingDevs = option.AutoStowingDevs
	kubeConfig = option.KubeConfig
}

func (hdm *HwDevManager) setRunMode() error {
	if hdm.runMode != "" {
		return nil
	}
	devNum, err := hdm.dmgr.GetDeviceCount()
	if err != nil {
		return err
	}
	chipName := ""
	for i := int32(0); i < devNum; i++ {
		chipName, err = hdm.dmgr.GetChipInfo(i)
		if err == nil {
			break
		}
		if i == devNum-1 {
			return err
		}
	}
	if strings.Contains(chipName, "310") {
		hdm.runMode = runMode310
		return nil
	}

	if strings.Contains(chipName, "710") {
		hdm.runMode = runMode710
		return nil
	}

	if strings.Contains(chipName, "910") {
		hdm.runMode = runMode910
		return nil
	}

	return fmt.Errorf("failed to get ascend device run mode")
}
