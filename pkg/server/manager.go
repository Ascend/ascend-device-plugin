/* Copyright(C) 2022. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Package server holds the implementation of registration to kubelet, k8s pod resource interface.
package server

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"huawei.com/npu-exporter/v3/common-utils/hwlog"
	"huawei.com/npu-exporter/v3/devmanager"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/device"
	"Ascend-device-plugin/pkg/kubeclient"
)

// HwDevManager manages huawei device devices.
type HwDevManager struct {
	groupDevice       map[string][]*common.NpuDevice
	ServerMap         map[string]InterfaceServer
	allInfo           common.NpuAllInfo
	manager           device.DevManager
	RunMode           string
	isEmptyKubeCfgErr bool
	RealCardType      string
}

// NewHwDevManager function is used to new a dev manager.
func NewHwDevManager(devM devmanager.DeviceInterface) *HwDevManager {
	var hdm HwDevManager
	if err := hdm.setAscendManager(devM); err != nil {
		hwlog.RunLog.Errorf("init hw dev manager failed, err: %#v", err)
		return nil
	}
	if err := hdm.setAllDeviceAndType(); err != nil {
		hwlog.RunLog.Errorf("set all device and type failed, err: %#v", err)
		return nil
	}
	if err := hdm.initPluginServer(); err != nil {
		hwlog.RunLog.Errorf("init plugin server failed, err: %#v", err)
		return nil
	}
	return &hdm
}

func (hdm *HwDevManager) setAscendManager(dmgr devmanager.DeviceInterface) error {
	switch dmgr.GetDevType() {
	case common.Ascend310:
		if !common.ParamOption.PresetVDevice {
			return fmt.Errorf("only 310p and 910 support dynamic virtual instance")
		}
		hdm.RunMode = common.Ascend310
		hdm.RealCardType = common.Ascend310
		hdm.manager = device.NewHwAscend310Manager()
	case common.Ascend910, common.Ascend910B:
		hdm.RunMode = common.Ascend910
		hdm.RealCardType = dmgr.GetDevType()
		hdm.manager = device.NewHwAscend910Manager()
	case common.Ascend310P:
		hdm.RunMode = common.Ascend310P
		hdm.RealCardType = common.Ascend310P
		hdm.manager = device.NewHwAscend310PManager()
	default:
		hwlog.RunLog.Error("found an unsupported device type")
		return fmt.Errorf("an unsupported device type")
	}
	hdm.manager.SetDmgr(dmgr)
	productType, err := hdm.manager.GetDmgr().GetProductType()
	if err != nil {
		return err
	}
	common.ParamOption.ProductType = productType
	return hdm.UpdateServerType()
}

// UpdateServerType update server type, like Ascend910-32
func (hdm *HwDevManager) UpdateServerType() error {
	kubeClient, err := kubeclient.NewClientK8s()
	if err != nil {
		hdm.isEmptyKubeCfgErr = strings.Contains(err.Error(), clientcmd.ErrEmptyConfig.Error())
		if !hdm.isEmptyKubeCfgErr {
			hwlog.RunLog.Errorf("init kubeclient failed err: %#v", err)
			return err
		}
		if common.ParamOption.UseVolcanoType {
			hwlog.RunLog.Warnf("not exist kube config, maybe it's edge scene")
			return errors.New("using volcano need kubeConfig, but not found")
		}
		return nil
	}
	if kubeClient == nil {
		return errors.New("kube client is nil")
	}
	hdm.manager.SetKubeClient(kubeClient)
	hwlog.RunLog.Info("init kube client success")
	aiCoreCount, err := hdm.manager.GetChipAiCoreCount()
	if err != nil {
		hwlog.RunLog.Errorf("get chip aicore count failed, err: %#v", err)
		return err
	}
	common.ParamOption.AiCoreCount = aiCoreCount
	return hdm.updateNodeServerType(aiCoreCount)

}

func (hdm *HwDevManager) updateNodeServerType(aiCoreCount int32) error {
	oldNode, err := hdm.manager.GetKubeClient().GetNode()
	if err != nil {
		hwlog.RunLog.Errorf("failed to get node, err: %#v", err)
		return err
	}
	if oldNode == nil {
		hwlog.RunLog.Error("invalid node")
		return fmt.Errorf("invalid node")
	}
	if _, ok := oldNode.Labels[common.ServerTypeLabelKey]; ok {
		return nil
	}
	newNode := oldNode.DeepCopy()
	newNode.Labels[common.ServerTypeLabelKey] = hdm.RealCardType + common.MiddelLine + strconv.Itoa(int(aiCoreCount))
	for i := 0; i < common.RetryUpdateCount; i++ {
		if _, _, err = hdm.manager.GetKubeClient().PatchNodeState(oldNode, newNode); err == nil {
			hwlog.RunLog.Infof("update server type success")
			return nil
		}
		hwlog.RunLog.Warnf("failed to patch server type to node, retry count:%d", i+1)
		time.Sleep(time.Second)
	}
	return fmt.Errorf("update server type to node label failed")
}

func (hdm *HwDevManager) setAllDeviceAndType() error {
	var err error
	if hdm.allInfo, err = hdm.manager.GetNPUs(); err != nil {
		return err
	}
	if len(hdm.allInfo.AllDevTypes) == 0 {
		return fmt.Errorf("no devices type found")
	}
	return nil
}

func (hdm *HwDevManager) initPluginServer() error {
	hdm.ServerMap = make(map[string]InterfaceServer, len(hdm.allInfo.AllDevTypes))
	hdm.groupDevice = device.ClassifyDevices(hdm.allInfo.AllDevs, hdm.allInfo.AllDevTypes)
	defaultDevices, err := common.GetDefaultDevices(common.ParamOption.GetFdFlag)
	if err != nil {
		hwlog.RunLog.Error("get default device error")
		return err
	}
	if !common.ParamOption.PresetVDevice {
		hdm.ServerMap[common.AiCoreResourceName] = NewPluginServer(common.AiCoreResourceName,
			hdm.allInfo.AICoreDevs, defaultDevices, hdm.manager)
		return nil
	}
	for _, deviceType := range hdm.allInfo.AllDevTypes {
		hdm.ServerMap[deviceType] = NewPluginServer(deviceType, hdm.groupDevice[deviceType], defaultDevices,
			hdm.manager)
	}
	return nil
}

// GetNPUs will set device default health, actually, it should be based on the last status if exist
func (hdm *HwDevManager) updateDeviceHealth(curAllDevs []common.NpuDevice) {
	lastAllDevs := make(map[string]int, len(hdm.allInfo.AllDevs))
	for index, dev := range hdm.allInfo.AllDevs {
		lastAllDevs[dev.DeviceName] = index
	}
	for i, dev := range curAllDevs {
		if index, exist := lastAllDevs[dev.DeviceName]; exist && index < len(hdm.allInfo.AllDevs) {
			curAllDevs[i].Health = hdm.allInfo.AllDevs[index].Health
			curAllDevs[i].NetworkHealth = hdm.allInfo.AllDevs[index].NetworkHealth
		}
	}
}

func (hdm *HwDevManager) updateDevice() error {
	if common.ParamOption.PresetVDevice {
		return nil
	}
	element, exist := hdm.ServerMap[common.AiCoreResourceName]
	if !exist {
		return fmt.Errorf("not found %s plugin server", common.AiCoreResourceName)
	}
	pluginServer, ok := element.(*PluginServer)
	if !ok {
		return fmt.Errorf("serverMap convert %s failed", common.AiCoreResourceName)
	}
	err := pluginServer.DestroyNotUsedVNPU()
	if err != nil {
		return err
	}
	allInfo, err := hdm.manager.GetNPUs()
	if err != nil {
		return err
	}
	if err := hdm.manager.CheckDeviceTypeLabel(); err != nil {
		hwlog.RunLog.Warnf("device type label may not correct, %v", err)
	}
	hdm.updateDeviceHealth(allInfo.AllDevs)
	hdm.groupDevice = device.ClassifyDevices(allInfo.AllDevs, allInfo.AllDevTypes)
	hdm.allInfo = allInfo
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
			return
		default:
			time.Sleep(time.Duration(common.ParamOption.ListAndWatchPeriod) * time.Second)
			common.LockAllDeviceInfo()
			if err := hdm.updateDevice(); err != nil {
				hwlog.RunLog.Error(err)
				common.UnlockAllDeviceInfo()
				continue
			}
			hdm.notifyToK8s()
			hdm.useVolcanoNotify()
			common.UnlockAllDeviceInfo()
		}
	}
}

func (hdm *HwDevManager) pluginNotify(classifyDev []*common.NpuDevice, devType string) {
	serverMap, ok := hdm.ServerMap[devType]
	if !ok {
		hwlog.RunLog.Warnf("server map (%s) not exist", devType)
		return
	}
	pluginServer, ok := serverMap.(*PluginServer)
	if !ok {
		hwlog.RunLog.Warnf("pluginServer (%s) not ok", devType)
		return
	}
	if !pluginServer.Notify(classifyDev) {
		hwlog.RunLog.Warnf("deviceType(%s) notify failed, server may not start, please check", devType)
	}
}

func (hdm *HwDevManager) notifyToK8s() {
	isDevStateChange := hdm.manager.IsDeviceStatusChange(hdm.groupDevice, hdm.allInfo.AICoreDevs, hdm.RunMode)
	for devType, isChanged := range isDevStateChange {
		if !isChanged {
			continue
		}
		if !common.ParamOption.PresetVDevice {
			hdm.pluginNotify(hdm.allInfo.AICoreDevs, common.AiCoreResourceName)
			return
		}
		hdm.pluginNotify(hdm.groupDevice[devType], devType)
	}
}

func (hdm *HwDevManager) useVolcanoNotify() {
	if hdm.isEmptyKubeCfgErr {
		return
	}
	common.DpStartReset.Do(func() {
		if err := hdm.manager.GetKubeClient().AnnotationReset(); err != nil {
			hwlog.RunLog.Warn("device plugin first reset annotation and config map error")
		}
	})
	if err := hdm.updatePodAnnotation(); err != nil {
		hwlog.RunLog.Error(err)
	}
	if !common.ParamOption.UseVolcanoType {
		return
	}
	hdm.manager.DoWithVolcanoListAndWatch(hdm.groupDevice)
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
		hdm.stopAllSever()
		hdm.manager.GetDmgr().ShutDown()
	}
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
			hwlog.RunLog.Errorf("close file watcher, err: %#v", err)
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
			hwlog.RunLog.Info("notify: kubelet.sock file created, restarting.")
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
	hwlog.RunLog.Info("stop all server done")
}

func (hdm *HwDevManager) setRestartForAll() {
	for deviceType := range hdm.ServerMap {
		hdm.ServerMap[deviceType].SetRestartFlag(true)
	}
}

func (hdm *HwDevManager) startAllServer(socketWatcher *common.FileWatch) bool {
	success := true
	for deviceType, serverInterface := range hdm.ServerMap {
		if !serverInterface.GetRestartFlag() {
			continue
		}
		if err := serverInterface.Start(socketWatcher); err != nil {
			hwlog.RunLog.Errorf("Could not contact Kubelet for %s, retrying. "+
				"Did you enable the device plugin feature gate?", deviceType)
			success = false
		} else {
			serverInterface.SetRestartFlag(false)
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
	serverID, err := hdm.manager.GetKubeClient().GetNodeServerID()
	if err != nil {
		return fmt.Errorf("get node server id failed: %#v", err)
	}
	if !common.ParamOption.PresetVDevice {
		return hdm.updateSpecTypePodAnnotation(common.AiCoreResourceName, serverID)
	}
	for _, devType := range hdm.allInfo.AllDevTypes {
		// for 310P vnpu no need update
		if common.IsVirtualDev(devType) && !strings.HasPrefix(devType, common.Ascend910) {
			continue
		}
		if err := hdm.updateSpecTypePodAnnotation(devType, serverID); err != nil {
			hwlog.RunLog.Warnf("update pod annotation failed, %#v", err)
		}
	}
	return nil
}

func (hdm *HwDevManager) updateSpecTypePodAnnotation(deviceType, serverID string) error {
	element, exist := hdm.ServerMap[deviceType]
	if !exist {
		return fmt.Errorf("not found %s plugin server", deviceType)
	}
	pluginServer, ok := element.(*PluginServer)
	if !ok {
		return fmt.Errorf("serverMap convert %s failed", deviceType)
	}
	podDeviceInfo, err := pluginServer.GetKltAndRealAllocateDev()
	if err != nil {
		return err
	}
	for _, deviceInfo := range podDeviceInfo {
		hwlog.RunLog.Debugf("pods: %s, %s, %s", deviceInfo.Pod.Name, deviceInfo.Pod.Status.Phase, deviceInfo.Pod.UID)
		_, existDeviceKey := deviceInfo.Pod.Annotations[common.Pod910DeviceKey]
		_, existRealAlloc := deviceInfo.Pod.Annotations[common.ResourceNamePrefix+common.PodRealAlloc]
		if existDeviceKey || existRealAlloc {
			continue
		}
		if len(deviceInfo.KltDevice) == 0 || len(deviceInfo.RealDevice) == 0 {
			hwlog.RunLog.Warnf("%s %s klt device or real device is empty", deviceInfo.Pod.Namespace,
				deviceInfo.Pod.Name)
			continue
		}
		hwlog.RunLog.Debugf("%s, %d, %v", deviceInfo.Pod.Name, len(deviceInfo.KltDevice), deviceInfo.RealDevice)
		if err := hdm.manager.AddPodAnnotation(&deviceInfo.Pod, deviceInfo.KltDevice, deviceInfo.RealDevice,
			deviceType, serverID); err != nil {
			hwlog.RunLog.Errorf("update pod %s_%s annotation failed, %#v", deviceInfo.Pod.Namespace,
				deviceInfo.Pod.Name, err)
		} else {
			hwlog.RunLog.Infof("update pod %s_%s annotation success", deviceInfo.Pod.Namespace, deviceInfo.Pod.Name)
		}
	}
	return nil
}
