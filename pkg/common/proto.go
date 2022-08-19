// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package common a series of common function
package common

import (
	"sync"

	"github.com/fsnotify/fsnotify"
	"k8s.io/apimachinery/pkg/util/sets"
)

var (
	// ParamOption for option
	ParamOption Option
	// DpStartReset for reset configmap
	DpStartReset sync.Once
)

// NodeDeviceInfoCache record node NPU device information. Will be solidified into cm.
type NodeDeviceInfoCache struct {
	DeviceInfo NodeDeviceInfo
	CheckCode  string
}

// NodeDeviceInfo record node NPU device information. Will be solidified into cm.
type NodeDeviceInfo struct {
	DeviceList map[string]string
	UpdateTime int64
}

// NpuDevice npu device description
type NpuDevice struct {
	DevType       string
	DeviceName    string
	Health        string
	NetworkHealth string
	IP            string
	LogicID       int32
	PhyID         int32
}

// DavinCiDev davinci device
type DavinCiDev struct {
	TemplateName map[string]string
	LogicID      int32
	PhyID        int32
}

// Device id for Instcance
type Device struct { // Device
	DeviceID string `json:"device_id"` // device id
	DeviceIP string `json:"device_ip"` // device ip
}

// Instance is for annotation
type Instance struct { // Instance
	PodName  string   `json:"pod_name"`  // pod Name
	ServerID string   `json:"server_id"` // serverdId
	Devices  []Device `json:"devices"`   // dev
}

// Option option
type Option struct {
	GetFdFlag          bool // to describe FdFlag
	UseAscendDocker    bool // UseAscendDocker to chose docker type
	UseVolcanoType     bool
	ListAndWatchPeriod int  // set listening device state period
	AutoStowingDevs    bool // auto stowing fixes devices or not
	KubeConfig         string
	PresetVDevice      bool
}

// GetAllDeviceInfoTypeList Get All Device Info Type List
func GetAllDeviceInfoTypeList() map[string]struct{} {
	return map[string]struct{}{HuaweiUnHealthAscend910: {}, HuaweiNetworkUnHealthAscend910: {},
		ResourceNamePrefix + Ascend910: {}, ResourceNamePrefix + Ascend910c2: {},
		ResourceNamePrefix + Ascend910c4: {}, ResourceNamePrefix + Ascend910c8: {},
		ResourceNamePrefix + Ascend910c16: {}, ResourceNamePrefix + Ascend310: {},
		ResourceNamePrefix + Ascend310P: {}, ResourceNamePrefix + Ascend310Pc1: {},
		ResourceNamePrefix + Ascend310Pc2: {}, ResourceNamePrefix + Ascend310Pc4: {},
		ResourceNamePrefix + Ascend310Pc2Cpu1: {}, ResourceNamePrefix + Ascend310Pc4Cpu3: {},
		HuaweiUnHealthAscend310P: {}, HuaweiUnHealthAscend310: {}}
}

// FileWatch is used to watch sock file
type FileWatch struct {
	FileWatcher *fsnotify.Watcher
}

// DevStatusSet contain different states devices
type DevStatusSet struct {
	UnHealthyDevice    sets.String
	NetUnHealthyDevice sets.String
	FreeHealthyDevice  map[string]sets.String
}
