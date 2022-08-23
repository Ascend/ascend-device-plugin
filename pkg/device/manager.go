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
	"huawei.com/mindx/common/hwlog"
	"huawei.com/npu-exporter/devmanager"
	"k8s.io/api/core/v1"
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
	if err := hdm.initPluginServer(devM, client); err != nil {
		hwlog.RunLog.Errorf("init plugin server failed, err: %v", err)
		return nil
	}
	if common.ParamOption.UseVolcanoType {
		hdm.ServerMap[common.PodResourceSeverKey] = server.NewPodResource()
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
	if err := hdm.manager.GetNPUs(&hdm.AllDevs, &hdm.AllDevTypes); err != nil {
		return err
	}
	if len(hdm.AllDevTypes) == 0 {
		return fmt.Errorf("no devices type found")
	}
	return nil
}

func (hdm *HwDevManager) initPluginServer(dmgr devmanager.DeviceInterface, client *kubeclient.ClientK8s) error {
	hdm.ServerMap = make(map[string]server.InterfaceServer, len(hdm.AllDevTypes))
	hdm.groupDevice = ClassifyDevices(hdm.AllDevs, hdm.AllDevTypes)
	defaultDevices, err := common.GetDefaultDevices(common.ParamOption.GetFdFlag)
	if err != nil {
		hwlog.RunLog.Error("get default device error")
		return err
	}
	for _, devcieType := range hdm.AllDevTypes {
		hdm.ServerMap[devcieType] = server.NewPluginServer(dmgr, client, devcieType,
			hdm.groupDevice[devcieType], defaultDevices)
	}
	return nil
}

// ListenDevice ListenDevice coroutine
func (hdm *HwDevManager) ListenDevice(ctx context.Context) {
	hwlog.RunLog.Info("starting the listen device")
	go hdm.Serve(ctx)
	for {
		select {
		case _, ok := <-ctx.Done():
			if !ok {
				hwlog.RunLog.Info("catch stop signal channel is closed")
			}
			hwlog.RunLog.Info("listen device stop")
			break
		default:
			time.Sleep(time.Duration(common.ParamOption.ListAndWatchPeriod) * time.Second)
			hdm.notifyToK8s()
			hdm.useVolcanoNotify()
		}
	}
}

func (hdm *HwDevManager) notifyToK8s() {
	for _, devType := range hdm.AllDevTypes {
		classifyDev := hdm.groupDevice[devType]
		isDevStateChange := hdm.manager.IsDeviceStatusChange(classifyDev, devType)
		if !isDevStateChange {
			continue
		}
		// if any device state or network state change, sure notify k8s
		serverMap, ok := hdm.ServerMap[devType]
		if !ok {
			hwlog.RunLog.Warnf("server map (%s) not exist", devType)
			continue
		}
		if !serverMap.(*server.PluginServer).Notify(classifyDev) {
			hwlog.RunLog.Warnf("deviceType(%s) notify failed, server may not start, please check", devType)
		}

	}
}

func (hdm *HwDevManager) useVolcanoNotify() {
	if !common.ParamOption.UseVolcanoType {
		return
	}
	common.DpStartReset.Do(func() {
		if err := hdm.manager.GetKubeClient().AnnotationReset(); err != nil {
			hwlog.RunLog.Warn("device plugin first reset annotation and config map error")
		}
	})
	hdm.manager.DoWithVolcanoListAndWatch(hdm.groupDevice)
	if err := hdm.updatePodAnnotation(); err != nil {
		hwlog.RunLog.Error(err)
	}
}

// SignCatch stop system sign catch
func (hdm *HwDevManager) SignCatch(cancel context.CancelFunc) {
	osSignChan := common.NewSignWatcher(syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGKILL)
	if osSignChan == nil {
		hwlog.RunLog.Error("the stop signal is not initialized")
		return
	}
	select {
	case s, signEnd := <-osSignChan:
		if signEnd == false {
			hwlog.RunLog.Info("catch stop signal channel is closed")
			return
		}
		hwlog.RunLog.Infof("Received signal: %s, shutting down.", s.String())
		cancel()
		hdm.DeleteDeviceInfo()
		hdm.manager.GetDmgr().ShutDown()
	}
}

// DeleteDeviceInfo delete the device info configmap
func (hdm *HwDevManager) DeleteDeviceInfo() {
	client := hdm.manager.GetKubeClient()
	if !common.ParamOption.UseVolcanoType || client == nil {
		return
	}
	if err := client.DeleteConfigMap(); err != nil {
		hwlog.RunLog.Errorf("delete device info configmap failed, error is %#v", err)
		return
	}
	hwlog.RunLog.Infof("delete device info configmap")
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
		if !hdm.ServerMap[deviceType].GetRestartFlag() {
			continue
		}
		if err := hdm.ServerMap[deviceType].Start(socketWatcher); err != nil {
			hwlog.RunLog.Errorf("Could not contact Kubelet for %s, retrying. "+
				"Did you enable the device plugin feature gate?", deviceType)
			success = false
		} else {
			hdm.ServerMap[deviceType].SetRestartFlag(false)
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

func (hdm *HwDevManager) updatePodAnnotation() error {
	element, exist := hdm.ServerMap[common.PodResourceSeverKey]
	if !exist {
		return fmt.Errorf("not found pod resource client")
	}
	prClient, ok := element.(*server.PodResource)
	if !ok {
		return fmt.Errorf("serverMap convert pod resource client failed")
	}
	podResource, err := prClient.GetPodResource()
	if err != nil {
		return fmt.Errorf("get pod resource failed, %s", err.Error())
	}
	podList, err := hdm.manager.GetKubeClient().GetPodList()
	if err != nil {
		return err
	}
	serverID, err := hdm.manager.GetKubeClient().GetNodeServerID()
	if err != nil {
		return fmt.Errorf("get node server id failed: %s", err.Error())
	}
	for _, devType := range hdm.AllDevTypes {
		element, exist := hdm.ServerMap[devType]
		if !exist {
			return fmt.Errorf("not found %s plugin server", devType)
		}
		ps, ok := element.(*server.PluginServer)
		if !ok {
			return fmt.Errorf("serverMap convert %s failed", devType)
		}
		if err := hdm.updateSpecTypePodAnnotation(podList, devType, serverID, podResource, ps); err != nil {
			return err
		}
	}
	return nil
}

func (hdm *HwDevManager) updateSpecTypePodAnnotation(podList *v1.PodList, deviceType, serverID string,
	podDevice map[string]server.PodDevice, pluginServer *server.PluginServer) error {
	pods, err := common.FilterPods(podList, common.GetPodPhaseBlackList(), deviceType, nil)
	if err != nil {
		return err
	}
	for _, pod := range pods {
		hwlog.RunLog.Debugf("pods: %s, %s, %s", pod.Name, pod.Status.Phase, pod.UID)
		if _, exist := pod.Annotations[common.PodRealAlloc]; exist {
			continue
		}
		podKey := pod.Namespace + common.UnderLine + pod.Name
		podResource, exist := podDevice[podKey]
		if !exist {
			hwlog.RunLog.Debugf("get %s klt device list failed, not in pod resource", podKey)
			continue
		}
		if podResource.ResourceName != common.ResourceNamePrefix+deviceType {
			hwlog.RunLog.Debugf("podKey %s resource name %s not equal device type %s", podKey,
				podResource.ResourceName, deviceType)
			continue
		}
		volDeviceList, err := pluginServer.GetRealAllocateDevices(podResource.DeviceIds)
		if err != nil {
			hwlog.RunLog.Debugf("get device list %#v failed, %s", podResource.DeviceIds, err.Error())
			continue
		}
		if err := hdm.manager.AddPodAnnotation(&pod, podResource.DeviceIds, volDeviceList, deviceType,
			serverID); err != nil {
			hwlog.RunLog.Errorf("update pod %s annotation failed, %s", podKey, err.Error())
		} else {
			hwlog.RunLog.Infof("update pod %s annotation success", podKey)
		}
	}
	return nil
}
