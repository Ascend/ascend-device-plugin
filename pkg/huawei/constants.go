/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2020-2022. All rights reserved.
 */
// Package huawei const
package huawei

const (
	// resource Name
	resourceNamePrefix   = "huawei.com/"
	hiAIAscend310Prefix  = "Ascend310"
	hiAIAscend910Prefix  = "Ascend910"
	hiAIAscend310PPrefix = "Ascend310P"
	hiAIAscendfdPrefix   = "davinci-mini"
	hiAIMaxDeviceNum     = 64
	hiAIMaxCardNum       = 64
	hiAIMaxDevNumInCard  = 4
	// The unit is seconds
	sleepTime = 5

	// MaxBackups log file max backup
	MaxBackups = 30
	// MaxAge the log file last time
	MaxAge = 7
	// config map name
	pod910DeviceKey  = "ascend.kubectl.kubernetes.io/ascend-910-configuration"
	pod310PDeviceKey = "ascend.kubectl.kubernetes.io/ascend-310P-configuration"
	// visible env
	ascendVisibleDevicesEnv = "ASCEND_VISIBLE_DEVICES"
	// virtual runtime option env
	ascendRuntimeOptionsEnv = "ASCEND_RUNTIME_OPTIONS"

	huaweiAscend910  = "huawei.com/Ascend910"
	huaweiAscend310P = "huawei.com/Ascend310P"
	huaweiAscend310  = "huawei.com/Ascend310"
	podPredicateTime = "predicate-time"
	pod2kl           = "huawei.com/kltDev"
	podRealAlloc     = "huawei.com/AscendReal"
	interval         = 1
	timeout          = 10

	logicIDIndexInVirtualDevID910  = 3
	huaweiNetworkUnHealthAscend910 = "huawei.com/Ascend910-NetworkUnhealthy"
	huaweiNetworkRecoverAscend910  = "huawei.com/Ascend910-NetworkRecover"
	huaweiRecoverAscend910         = "huawei.com/Ascend910-Recover"
	huaweiUnHealthAscend910        = "huawei.com/Ascend910-Unhealthy"
	huaweiUnHealthAscend310P       = "huawei.com/Ascend310P-Unhealthy"
	huaweiUnHealthAscend310        = "huawei.com/Ascend310-Unhealthy"
	// FunctionNotFound for describe devmanager interface function is exist or not
	FunctionNotFound = "-99998"

	// MaxVirtualDevNum is the max virtual devices number
	MaxVirtualDevNum = 128

	sleep2ListW = 3

	initMapCap = 5

	// for format change
	baseDec   = 10
	bitSize   = 64
	bitSize32 = 32

	chip310PCore1C     = "Ascend310P-1c"
	chip310PCore2C     = "Ascend310P-2c"
	chip310PCore4C     = "Ascend310P-4c"
	chip310PCore4C3Cpu = "Ascend310P-4c.3cpu"
	chip310PCore2C1Cpu = "Ascend310P-2c.1cpu"

	chip910Core2C  = "Ascend910-2c"
	chip910Core4C  = "Ascend910-4c"
	chip910Core8C  = "Ascend910-8c"
	chip910Core16C = "Ascend910-16c"
)
