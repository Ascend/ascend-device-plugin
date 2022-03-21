/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2020-2022. All rights reserved.
 */

package huawei

const (
	// resource Name
	resourceNamePrefix  = "huawei.com/"
	hiAIAscend310Prefix = "Ascend310"
	hiAIAscend910Prefix = "Ascend910"
	hiAIAscend710Prefix = "Ascend710"
	hiAIAscendfdPrefix  = "davinci-mini"
	hiAISlogdConfig     = "/etc/slog.conf"
	hiAIMaxDeviceNum    = 64
	// The unit is seconds
	sleepTime = 5

	// MaxBackups log file max backup
	MaxBackups = 30
	// MaxAge the log file last time
	MaxAge                  = 7
	podDeviceKey            = "ascend.kubectl.kubernetes.io/ascend-910-configuration" // config map name
	ascendVisibleDevicesEnv = "ASCEND_VISIBLE_DEVICES"                                // visible env
	ascendRuntimeOptionsEnv = "ASCEND_RUNTIME_OPTIONS"                                // virtual runtime option env

	huaweiAscend910  = "huawei.com/Ascend910"
	podPredicateTime = "predicate-time"
	pod2kl           = "huawei/kltDev"
	interval         = 1
	timeout          = 10

	virtualDevicesPattern = "Ascend910-(2|4|8|16)c"
	pwr2CSuffix           = "Ascend910-2c"
	pwr4CSuffix           = "Ascend910-4c"
	pwr8CSuffix           = "Ascend910-8c"
	pwr16CSuffix          = "Ascend910-16c"

	logicIDIndexInVirtualDevID910  = 3
	huaweiNetworkUnHealthAscend910 = "huawei.com/Ascend910-NetworkUnhealthy"
	huaweiNetworkRecoverAscend910  = "huawei.com/Ascend910-NetworkRecover"
	huaweiRecoverAscend910         = "huawei.com/Ascend910-Recover"
	huaweiUnHealthAscend910        = "huawei.com/Ascend910-Unhealthy"
	huaweiUnHealthAscend710        = "huawei.com/Ascend710-Unhealthy"
	huaweiUnHealthAscend310        = "huawei.com/Ascend310-Unhealthy"
	// FunctionNotFound for describe dsmi interface function is exist or not
	FunctionNotFound = "-99998"

	// MaxVirtualDevNum is the max virtual devices number
	MaxVirtualDevNum = 128

	sleep2ListW = 3

	initMapCap = 5

	// for format change
	baseDec   = 10
	bitSize   = 64
	bitSize32 = 32
)
