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
	hiAIDavinciPrefix = "/dev/davinci"
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
	hiAIAscendfdPrefix  = "davinci-mini"
	hiAISlogdConfig     = "/etc/slog.conf"
	hiAIMaxDeviceNum    = 64
	idSplitNum          = 2
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

	huaweiAscend910  = "huawei.com/Ascend910"
	podPredicateTime = "predicate-time"
	runMode310       = "ascend310"
	runMode910       = "ascend910"
	retryTime        = 3
	interval         = 1
	timeout          = 10
	maxChipName      = 32
)
