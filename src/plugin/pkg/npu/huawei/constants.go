/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2019-2024. All rights reserved.
 * Description: constants.go
 * Create: 19-11-20 下午9:05
 */

package huawei

const (
	// All HUAWEI Ascend910 cards should be mounted with hiAIHDCDevice and hiAIManagerDevice
	// If the driver installed correctly, these two devices will be there.
	hiAIHDCDevice     = "/dev/hisi_hdc"
	hiAIManagerDevice = "/dev/davinci_manager"
	hiAIDavinciPrefix = "/dev/davinci"
	hiAISVMDevice     = "/dev/devmm_svm"

	// resource Name
	resourceNamePrefix  = "huawei.com/"
	hiAIAscend310Prefix = "Ascend310"
	hiAIAscend910Prefix = "Ascend910"
	hiAIAscendfdPrefix  = "davinci-mini"
	hiAISlogdConfig     = "/etc/slog.conf"
	hiAIMaxDeviceNum    = 64
	idSplitNum          = 2
	dieIDNum            = 5
	// The unit is seconds
	sleepTime = 5
	// if register failed three times then exit
	registerTimeout = 3

	// device socket path
	serverSock = "/var/lib/kubelet/device-plugins/Ascend910.sock"

	// logger setting

	// LogPath save log file
	LogPath                 = "/var/log/devicePlugin/devicePlugin.log"
	fileMaxSize             = 1000                                                   // each log file size
	maxBackups              = 20                                                     // max backup
	maxAge                  = 28                                                     // the log file last time
	podDeviceKey            = "atlas.kubectl.kubernetes.io/ascend-910-configuration" // config map name
	ascendVisibleDevicesEnv = "ASCEND_VISIBLE_DEVICES"                               // visible env
	logChmod                = 0640
)
