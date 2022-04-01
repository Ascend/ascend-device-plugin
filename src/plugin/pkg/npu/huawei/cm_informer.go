// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package huawei using informer update cache for hps.devices
package huawei

import (
	"os"
	"path"
	"time"

	"huawei.com/npu-exporter/hwlog"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"Ascend-device-plugin/src/plugin/pkg/npu/common"
	"Ascend-device-plugin/src/plugin/pkg/npu/vnpumanager"
)

const (
	// waitingTimeTask wait timing task update 10s
	waitingTimeTask = 10
)

// ConfigMapAgent Agent for configMap Workers
type ConfigMapAgent struct {
	cmInformer        cache.SharedInformer
	cmInformerFactory informers.SharedInformerFactory
}

// NewConfigMapAgent new ConfigMapAgent
func NewConfigMapAgent(kubeClientSet kubernetes.Interface, hdm *HwDevManager) {
	stopCh := make(chan struct{})
	cmInformerFactory := informers.NewSharedInformerFactoryWithOptions(kubeClientSet, time.Second*sleep2ListW,
		informers.WithTweakListOptions(func(options *v1.ListOptions) {}))

	cmAgent := &ConfigMapAgent{
		cmInformerFactory: cmInformerFactory,
		cmInformer:        cmInformerFactory.Core().V1().ConfigMaps().Informer(),
	}
	defer runtime.HandleCrash()
	cmAgent.cmInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			newCM, ok := newObj.(v1.Object)
			if !ok {
				return
			}
			oldCM, ok := oldObj.(v1.Object)
			if !ok {
				return
			}
			newCardNPUs, isChange := isCMChange(newCM, oldCM)
			if isChange {
				parseCMData(newCardNPUs, hdm, kubeClientSet)
			}
		},
	})
	hwlog.RunLog.Infof("start configMap informer factory")
	go cmInformerFactory.Start(stopCh)
	hwlog.RunLog.Info("start running configMap informer")
	cmAgent.cmInformer.Run(stopCh)
}

func isCMChange(newCM, oldCM v1.Object) ([]vnpumanager.CardVNPUs, bool) {
	newCardNPUs := vnpumanager.ConvertCMToStruct(newCM)
	oldCardNPUs := vnpumanager.ConvertCMToStruct(oldCM)
	if newCardNPUs == nil || oldCardNPUs == nil {
		return nil, false
	}
	if len(newCardNPUs) != len(oldCardNPUs) {
		return nil, true
	}
	return newCardNPUs, vnpumanager.IsConfigMapChange(newCardNPUs, oldCardNPUs)
}

func parseCMData(newCardNPUs []vnpumanager.CardVNPUs, hdm *HwDevManager, kubeClient kubernetes.Interface) {
	m.Lock()
	defer m.Unlock()
	hwlog.RunLog.Infof("start sync informer info by update or add func")
	for ListenAnnotation.IsTimingComplete.Load() {
		time.Sleep(time.Second * waitingTimeTask)
	}
	var dcmiDevices []common.NpuDevice
	var dcmiDeviceTypes []string
	hwlog.RunLog.Infof("starting get old NPU info")
	if err := hdm.manager.GetNPUs(&dcmiDevices, &dcmiDeviceTypes, hdm.runMode); err != nil {
		hwlog.RunLog.Errorf("get NPU failed, err: %v\n", err)
		return
	}
	vnpumanager.DestroyVirtualDev(hdm.dmgr, dcmiDevices, newCardNPUs)
	vnpumanager.CreateVirtualDev(hdm.dmgr, newCardNPUs, hdm.runMode, kubeClient)
	updateHpsCache(hdm)
}

func updateHpsCache(hdm *HwDevManager) {
	hwlog.RunLog.Infof("start update multi-virtual device cache after create virtual device")
	var newDevices []common.NpuDevice
	var newDevTypes []string
	if err := hdm.manager.GetNPUs(&newDevices, &newDevTypes, hdm.runMode); err != nil {
		hwlog.RunLog.Errorf("get new NPU devices failed, err: %v\n", err)
		return
	}
	getDiffDevCount(hdm, newDevices)
	registerNewServer(hdm, newDevTypes)
}

func getDiffDevCount(hdm *HwDevManager, newDevices []common.NpuDevice) {
	pwrSuffix := []string{hiAIAscend910Prefix, pwr2CSuffix, pwr4CSuffix, pwr8CSuffix, pwr16CSuffix}
	if hdm.runMode == common.RunMode710 {
		pwrSuffix = []string{hiAIAscend710Prefix, chip710Core1C, chip710Core2C, chip710Core4C}
	}
	oldDevices := hdm.allDevs
	for _, devType := range pwrSuffix {
		listenDevCountIsChange[devType] = isDevCountChange(oldDevices, newDevices, devType)
	}
	hdm.allDevs = newDevices
}

func isDevCountChange(oldDevices, newDevices []common.NpuDevice, devType string) bool {
	return isDevEqual(getSpecDevTypes(oldDevices, devType), getSpecDevTypes(newDevices, devType))
}

func registerNewServer(hdm *HwDevManager, newDevTypes []string) {
	hwlog.RunLog.Infof("starting reRegister new type virtual device server")
	interDevTypes := getInterDevType(hdm.GetDevType(), newDevTypes)
	for devType := range getDiffDevType(hdm.GetDevType(), newDevTypes) {
		sockPath := path.Join(v1beta1.DevicePluginPath, devType)
		if _, err := os.Stat(sockPath); err == nil {
			continue
		}
		go hdm.Serve(devType)
		hdm.allDevTypes = append(hdm.allDevTypes, devType)
	}
	for devType := range interDevTypes {
		ServeUpdateMap[devType] = make(chan int, len(interDevTypes))
		ServeUpdateMap[devType] <- 1
	}
	hwlog.RunLog.Infof("reRegister new type virtual device server complete")
}

func getDiffDevType(devTypes, newDevTypes []string) sets.String {
	return convertToSets(newDevTypes).Difference(convertToSets(devTypes))
}

func getInterDevType(devTypes, newDevTypes []string) sets.String {
	return convertToSets(newDevTypes).Intersection(convertToSets(devTypes))
}

func getSpecDevTypes(devices []common.NpuDevice, devType string) []string {
	var devTypes []string
	for _, device := range devices {
		if device.DevType == devType {
			devTypes = append(devTypes, device.ID)
		}
	}
	return devTypes
}

func isDevEqual(oldDevs, newDevs []string) bool {
	return !convertToSets(oldDevs).Equal(convertToSets(newDevs))
}

func convertToSets(devTypes []string) sets.String {
	devSet := sets.String{}
	for _, devType := range devTypes {
		devSet.Insert(devType)
	}
	return devSet
}
