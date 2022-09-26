// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package common a series of common function
package common

const (
	// Component component name
	Component = "device-plugin"
	// MaxBackups log file max backup
	MaxBackups = 30
	// MaxAge the log file last time
	MaxAge = 7

	// KubeEnvMaxLength k8s env name max length
	KubeEnvMaxLength = 230
	// PodNameMaxLength pod name max length
	PodNameMaxLength = 253
	// PodNameSpaceMaxLength pod name space max length
	PodNameSpaceMaxLength = 63
	// MaxPodLimit max pod num
	MaxPodLimit = 110
	// MaxContainerLimit max container num
	MaxContainerLimit = 300000
	// RetryUpdateCount is max number of retry resource update
	RetryUpdateCount = 3
	// MaxDeviceNameLen max length of device name, like "Ascend310P-4c.3cpu-100-0"
	MaxDeviceNameLen = 50
	// MaxGRPCRecvMsgSize 4MB
	MaxGRPCRecvMsgSize = 4 * 1024 * 1024
	// MaxGRPCConcurrentStreams limit on the number of concurrent streams to each ServerTransport.
	MaxGRPCConcurrentStreams = 64
	// MaxVirtualDeviceNum max num of virtual device
	MaxVirtualDeviceNum = 1024

	// DeviceInfoCMNameSpace namespace of device info configmap
	DeviceInfoCMNameSpace = "kube-system"
	// DeviceInfoCMNamePrefix device info configmap name prefix
	DeviceInfoCMNamePrefix = "mindx-dl-deviceinfo-"
	// DeviceInfoCMDataKey device info configmap data key
	DeviceInfoCMDataKey = "DeviceInfoCfg"

	runtimeEnvNum           = 2
	ascendVisibleDevicesEnv = "ASCEND_VISIBLE_DEVICES" // visible env
	ascendRuntimeOptionsEnv = "ASCEND_RUNTIME_OPTIONS" // virtual runtime option env
	// PodPredicateTime pod predicate time
	PodPredicateTime = "predicate-time"
	// Pod2kl pod annotation key, means kubelet allocate device
	Pod2kl = ResourceNamePrefix + "kltDev"
	// PodRealAlloc pod annotation key, means pod real mount device
	PodRealAlloc = ResourceNamePrefix + "AscendReal"
	// Pod910DeviceKey pod annotation key, for generate 910 hccl rank table
	Pod910DeviceKey = "ascend.kubectl.kubernetes.io/ascend-910-configuration"
	// Pod310PDeviceKey pod annotation key, for generate 310P hccl rank table
	Pod310PDeviceKey = "ascend.kubectl.kubernetes.io/ascend-310P-configuration"

	// PodResourceSeverKey for pod resource key
	PodResourceSeverKey = "podResource"
	// VirtualDev Virtual device tag
	VirtualDev = "VIRTUAL"
	// PhyDeviceLen like Ascend910-0 split length is 2
	PhyDeviceLen = 2
	// VirDeviceLen like Ascend910-2c-100-1 split length is 4
	VirDeviceLen = 4
	// MaxDevicesNum max device num
	MaxDevicesNum = 64
	// MaxCardNum max card num
	MaxCardNum = 64
	// MaxDevNumInCard max device num in card
	MaxDevNumInCard = 4
	// MaxRequestVirtualDeviceNum max request device num
	MaxRequestVirtualDeviceNum = 1
	// LabelDeviceLen like Ascend910-0 split length is 2
	LabelDeviceLen = 2
	// DefaultDeviceIP device ip address
	DefaultDeviceIP = "127.0.0.1"
	// NormalState health state
	NormalState = uint32(0)
	// GeneralAlarm health state
	GeneralAlarm = uint32(1)

	// SocketChmod socket file mode
	SocketChmod = 0600
	// RunMode310 for 310 chip
	RunMode310 = "ascend310"
	// RunMode910 for 910 chip
	RunMode910 = "ascend910"
	// RunMode310P for 310P chip
	RunMode310P = "ascend310P"

	// Interval interval time
	Interval = 1
	// Timeout time
	Timeout = 10
	// BaseDec base
	BaseDec = 10
	// BitSize base size
	BitSize = 64
	// BitSize32 base size 32
	BitSize32 = 32
	// SleepTime The unit is seconds
	SleepTime = 5
)

const (
	// ResourceNamePrefix prefix
	ResourceNamePrefix = "huawei.com/"
	// Ascend310P 310p
	Ascend310P = "Ascend310P"
	// Ascend310Pc1 Ascend310P 1 core
	Ascend310Pc1 = "Ascend310P-1c"
	// Ascend310Pc2 Ascend310P 2 core
	Ascend310Pc2 = "Ascend310P-2c"
	// Ascend310Pc4 Ascend310P 4 core
	Ascend310Pc4 = "Ascend310P-4c"
	// Ascend310Pc4Cpu3 Ascend310P 4core 3cpu
	Ascend310Pc4Cpu3 = "Ascend310P-4c.3cpu"
	// Ascend310Pc2Cpu1 Ascend310P 2core 1cpu
	Ascend310Pc2Cpu1 = "Ascend310P-2c.1cpu"
	// Ascend310Pc4Cpu4Dvpp Ascend310P 4core 4cpu dvpp
	Ascend310Pc4Cpu4Dvpp = "Ascend310P-4c.4cpu.dvpp"
	// Ascend310Pc4Cpu3Ndvpp Ascend310P 4core 3cpu ndvpp
	Ascend310Pc4Cpu3Ndvpp = "Ascend310P-4c.3cpu.ndvpp"
	// HuaweiAscend310P with prefix
	HuaweiAscend310P = ResourceNamePrefix + Ascend310P

	// Ascend910 910
	Ascend910 = "Ascend910"
	// Ascend910c2  Ascend910 2core
	Ascend910c2 = "Ascend910-2c"
	// Ascend910c4 Ascend910 4core
	Ascend910c4 = "Ascend910-4c"
	// Ascend910c8 Ascend910 8core
	Ascend910c8 = "Ascend910-8c"
	// Ascend910c16 Ascend910 16core
	Ascend910c16 = "Ascend910-16c"
	// HuaweiAscend910 with prefix
	HuaweiAscend910 = ResourceNamePrefix + Ascend910

	// Ascend310 310
	Ascend310 = "Ascend310"
	// HuaweiAscend310 with prefix
	HuaweiAscend310 = ResourceNamePrefix + Ascend310
	// AscendfdPrefix use in fd
	AscendfdPrefix = "davinci-mini"

	// HuaweiNetworkUnHealthAscend910 910 network unhealthy
	HuaweiNetworkUnHealthAscend910 = ResourceNamePrefix + "Ascend910-NetworkUnhealthy"
	// HuaweiUnHealthAscend910 unhealth
	HuaweiUnHealthAscend910 = ResourceNamePrefix + "Ascend910-Unhealthy"
	// HuaweiUnHealthAscend310P 310p unhealthy
	HuaweiUnHealthAscend310P = ResourceNamePrefix + "Ascend310P-Unhealthy"
	// HuaweiUnHealthAscend310 310 unhealthy
	HuaweiUnHealthAscend310 = ResourceNamePrefix + "Ascend310-Unhealthy"
	// HuaweiNetworkRecoverAscend910 910 network recover
	HuaweiNetworkRecoverAscend910 = ResourceNamePrefix + "Ascend910-NetworkRecover"
	// HuaweiRecoverAscend910 910 recover
	HuaweiRecoverAscend910 = ResourceNamePrefix + "Ascend910-Recover"
)

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
	// DefaultKubeConfig is default k8s config file
	DefaultKubeConfig = "/etc/mindx-dl/device-plugin/.config/config6"
	// DefaultKubeConfigBkp is default k8s config bake up file
	DefaultKubeConfigBkp = "/etc/mindx-dl/device-plugin/.config6"
)

const (
	// Atlas200ISoc 200 soc env
	Atlas200ISoc = "Atlas 200I SoC A1"
	// Atlas200SocXSMEM is xsmem_dev
	Atlas200SocXSMEM = "/dev/xsmem_dev"
	// Atlas200SocSYS is sys
	Atlas200SocSYS = "/dev/sys"
	// Atlas200SocVDEC is vdec
	Atlas200SocVDEC = "/dev/vdec"
	// Atlas200SocVPC is vpc
	Atlas200SocVPC = "/dev/vpc"
)

const (
	// RootUID is root user id
	RootUID = 0
	// RootGID is root group id
	RootGID = 0

	// DotSepDev if the separator between devices on labels
	DotSepDev = "."

	// CommaSepDev if the separator between devices on annotation
	CommaSepDev = ","
	// MiddelLine if the separator between devices for split id
	MiddelLine = "-"
	// UnderLine the separator between ids
	UnderLine = "_"
)
