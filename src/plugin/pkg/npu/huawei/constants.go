/*
* Copyright(C) 2020. Huawei Technologies Co.,Ltd. All rights reserved.
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
* http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package huawei

const (
	// All HUAWEI Ascend910 cards should be mounted with hiAIHDCDevice and hiAIManagerDevice
	// If the driver installed correctly, these two devices will be there.
	hiAIHDCDevice     = "/dev/hisi_hdc"
	hiAIManagerDevice = "/dev/davinci_manager"
	hiAISVMDevice     = "/dev/devmm_svm"

	hiAi200RCSVM0            = "/dev/svm0"
	hiAi200RCLog             = "/dev/log_drv"
	hiAi200RCEventSched      = "/dev/event_sched"
	hiAi200RCUpgrade         = "/dev/upgrade"
	hiAi200RCHiDvpp          = "/dev/hi_dvpp"
	hiAi200RCMemoryBandwidth = "/dev/memory_bandwidth"
	hiAi200RCTsAisle         = "/dev/ts_aisle"

	// resource Name
	resourceNamePrefix  = "huawei.com/"
	hiAIAscend310Prefix = "Ascend310"
	hiAIAscend910Prefix = "Ascend910"
	hiAIAscend710Prefix = "Ascend710"
	hiAIAscendfdPrefix  = "davinci-mini"
	hiAISlogdConfig     = "/etc/slog.conf"
	hiAIMaxDeviceNum    = 64
	idSplitNum          = 2
	// The unit is seconds
	sleepTime = 5

	// logger setting

	// LogName save log file
	LogName = "devicePlugin.log"
	// FileMaxSize each log file size
	FileMaxSize = 20
	// MaxBackups log file max backup
	MaxBackups = 8
	// MaxAge the log file last time
	MaxAge                  = 10
	podDeviceKey            = "ascend.kubectl.kubernetes.io/ascend-910-configuration" // config map name
	ascendVisibleDevicesEnv = "ASCEND_VISIBLE_DEVICES"                                // visible env
	ascendRuntimeOptionsEnv = "ASCEND_RUNTIME_OPTIONS"                                // virtual runtime option env
	// LogChmod log file mode
	LogChmod = 0640
	// BackupLogChmod backup log file mode
	BackupLogChmod = 0400
	socketChmod    = 0600

	huaweiAscend910  = "huawei.com/Ascend910"
	podPredicateTime = "predicate-time"
	runMode310       = "ascend310"
	runMode910       = "ascend910"
	runMode710       = "ascend710"
	retryTime        = 3
	interval         = 1
	timeout          = 10
	maxChipName      = 32

	virtualDevicesPattern = "Ascend910-(2|4|8|16)c"
	pwr2CSuffix           = "Ascend910-2c"
	pwr4CSuffix           = "Ascend910-4c"
	pwr8CSuffix           = "Ascend910-8c"
	pwr16CSuffix          = "Ascend910-16c"

	logicIDIndexInVirtualDevID910 = 3
	huaweiUnHealthAscend910       = "huawei.com/Ascend910-Unhealthy"
	huaweiRecoverAscend910        = "huawei.com/Ascend910-Recover"

	// FunctionNotFound for describe dsmi interface function is exist or not
	FunctionNotFound = "-99998"

	// MaxVirtualDevNum is the max virtual devices number
	MaxVirtualDevNum = 128

	resetZero = 0
)
