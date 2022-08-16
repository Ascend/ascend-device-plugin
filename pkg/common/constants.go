// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package common a series of constants
package common

const (
	// HiAIHDCDevice hisi_hdc
	HiAIHDCDevice = "/dev/hisi_hdc"
	// HiAIManagerDevice davinci_manager
	HiAIManagerDevice = "/dev/davinci_manager"
	// HiAISVMDevice devmm_svm
	HiAISVMDevice = "/dev/devmm_svm"
	// HiAi200RCSVM0 svm0
	HiAi200RCSVM0 = "/dev/svm0"
	// HiAi200RCLog log_drv
	HiAi200RCLog = "/dev/log_drv"
	// HiAi200RCEventSched event_sched
	HiAi200RCEventSched = "/dev/event_sched"
	// HiAi200RCUpgrade upgrade
	HiAi200RCUpgrade = "/dev/upgrade"
	// HiAi200RCHiDvpp hi_dvpp
	HiAi200RCHiDvpp = "/dev/hi_dvpp"
	// HiAi200RCMemoryBandwidth memory_bandwidth
	HiAi200RCMemoryBandwidth = "/dev/memory_bandwidth"
	// HiAi200RCTsAisle ts_aisle
	HiAi200RCTsAisle = "/dev/ts_aisle"
)

const (
	// SocketChmod socket file mode
	SocketChmod = 0600
	// RunMode310 for 310 chip
	RunMode310 = "ascend310"
	// RunMode910 for 910 chip
	RunMode910 = "ascend910"
	// RunMode310P for 310P chip
	RunMode310P = "ascend310P"
)
