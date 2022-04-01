// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package huawei update cache for hps.devices
package huawei

import (
	"fmt"
	"time"

	"huawei.com/npu-exporter/hwlog"
	"k8s.io/client-go/kubernetes"

	"Ascend-device-plugin/src/plugin/pkg/npu/common"
	"Ascend-device-plugin/src/plugin/pkg/npu/vnpumanager"
)

// UpdateVNpuDevice to create and destroy virtual device
func UpdateVNpuDevice(hdm *HwDevManager, stopCh <-chan struct{}) {
	go func(hdm *HwDevManager) {
		for {
			if stopCh == nil {
				return
			}
			client, err := common.NewKubeClient(kubeConfig)
			if err != nil {
				fmt.Errorf("failed to create kube client: %v", err)
				return
			}
			if err := TimingUpdate(hdm, client); err != nil {
				hwlog.RunLog.Errorf("current timing update failed, waiting for next time, err: %v", err)
			}
			time.Sleep(time.Minute)
		}
	}(hdm)
	go func() {
		for {
			if stopCh == nil {
				return
			}
			InformerCmUpdate(hdm)
		}
	}()
}

// TimingUpdate each minute exec update function
func TimingUpdate(hdm *HwDevManager, client *kubernetes.Clientset) error {
	hwlog.RunLog.Infof("starting configMap timing update task")
	m.Lock()
	defer m.Unlock()
	var dcmiDevices []common.NpuDevice
	var dcmiDeviceTypes []string
	if err := hdm.manager.GetNPUs(&dcmiDevices, &dcmiDeviceTypes, hdm.manager.GetMatchingDeviType()); err != nil {
		return err
	}
	cardVNPUs, err := vnpumanager.GetVNpuCfg(client)
	if err != nil {
		return err
	}
	vnpumanager.DestroyVirtualDev(hdm.dmgr, dcmiDevices, cardVNPUs)
	vnpumanager.CreateVirtualDev(hdm.dmgr, cardVNPUs, hdm.runMode, client)
	updateHpsCache(hdm)
	hwlog.RunLog.Infof("configMap timing update task complete")
	return nil
}

// InformerCmUpdate update vnpu by configMap informer
func InformerCmUpdate(hdm *HwDevManager) {
	client, err := common.NewKubeClient(kubeConfig)
	if err != nil {
		fmt.Errorf("failed to create kube client: %v", err)
		return
	}
	NewConfigMapAgent(client, hdm)
}
