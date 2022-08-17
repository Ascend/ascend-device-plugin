// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package device a series of device function
package device

import (
	"errors"

	"huawei.com/npu-exporter/devmanager"
	"huawei.com/npu-exporter/hwlog"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/kubeclient"
	"Ascend-device-plugin/pkg/server"
)

// HwDevManager manages huawei device devices.
type HwDevManager struct {
	groupDevice map[string][]*common.NpuDevice
	ServerMap   map[string]server.Server
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
