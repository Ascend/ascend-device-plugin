/* Copyright(C) 2022. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Package main implements initialization of the startup parameters of the device plugin.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"huawei.com/npu-exporter/v5/common-utils/hwlog"
	"huawei.com/npu-exporter/v5/devmanager"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/server"
)

const (
	// socket name
	defaultLogPath = "/var/log/mindx-dl/devicePlugin/devicePlugin.log"

	// defaultListWatchPeriod is the default listening device state's period
	defaultListWatchPeriod = 5

	// maxListWatchPeriod is the max listening device state's period
	maxListWatchPeriod = 60
	// minListWatchPeriod is the min listening device state's period
	minListWatchPeriod = 3
	maxRunModeLength   = 10
	maxLogLineLength   = 1024
)

var (
	mode = flag.String("mode", "", "Device plugin running mode: ascend310, ascend310P, "+
		"ascend910. This parameter will be deprecated in future versions")
	fdFlag          = flag.Bool("fdFlag", false, "Whether to use fd system to manage device (default false)")
	useAscendDocker = flag.Bool("useAscendDocker", true, "Whether to use ascend docker. "+
		"This parameter will be deprecated in future versions")
	volcanoType = flag.Bool("volcanoType", false,
		"Specifies whether to use volcano for scheduling when the chip type is Ascend310 or Ascend910 (default false)")
	version     = flag.Bool("version", false, "Output version information")
	edgeLogFile = flag.String("edgeLogFile", "/var/alog/AtlasEdge_log/devicePlugin.log",
		"Log file path in edge scene")
	listWatchPeriod = flag.Int("listWatchPeriod", defaultListWatchPeriod,
		"Listen and watch device state's period, unit second, range [3, 60]")
	autoStowing = flag.Bool("autoStowing", true, "Whether to automatically stow the fixed device")
	logLevel    = flag.Int("logLevel", 0,
		"Log level, -1-debug, 0-info, 1-warning, 2-error, 3-critical(default 0)")
	logMaxAge = flag.Int("maxAge", common.MaxAge,
		"Maximum number of days for backup run log files, range [7, 700] days")
	logFile = flag.String("logFile", defaultLogPath,
		"The log file path, if the file size exceeds 20MB, will be rotate")
	logMaxBackups = flag.Int("maxBackups", common.MaxBackups,
		"Maximum number of backup log files, range is (0, 30]")
	presetVirtualDevice = flag.Bool("presetVirtualDevice", true, "Open the static of "+
		"computing power splitting function, only support Ascend910 and Ascend310P")
	use310PMixedInsert = flag.Bool("use310PMixedInsert", false, "Whether to use mixed insert "+
		"ascend310P-V, ascend310P-VPro, ascend310P-IPro card mode")
	hotReset = flag.Int("hotReset", -1, "set hot reset mode: -1-close, 0-infer, 1-train")
)

var (
	// BuildName show app name
	BuildName string
	// BuildVersion show app version
	BuildVersion string
	// BuildScene show app staring scene
	BuildScene string
)

func initLogModule(ctx context.Context) error {
	var loggerPath string
	loggerPath = *logFile
	if *fdFlag {
		loggerPath = *edgeLogFile
	}
	if !common.CheckFileUserSameWithProcess(loggerPath) {
		return fmt.Errorf("check log file failed")
	}
	hwLogConfig := hwlog.LogConfig{
		LogFileName:   loggerPath,
		LogLevel:      *logLevel,
		MaxBackups:    *logMaxBackups,
		MaxAge:        *logMaxAge,
		MaxLineLength: maxLogLineLength,
	}
	if err := hwlog.InitRunLogger(&hwLogConfig, ctx); err != nil {
		fmt.Printf("hwlog init failed, error is %#v\n", err)
		return err
	}
	return nil
}

func checkParam() bool {
	if *listWatchPeriod < minListWatchPeriod || *listWatchPeriod > maxListWatchPeriod {
		hwlog.RunLog.Errorf("list and watch period %d out of range", *listWatchPeriod)
		return false
	}
	if !(*presetVirtualDevice) {
		hwlog.RunLog.Error("presetVirtualDevice not support set to false now")
		return false
	}
	if len(*mode) > maxRunModeLength {
		hwlog.RunLog.Error("run mode param length invalid")
		return false
	}
	if *use310PMixedInsert && *volcanoType {
		hwlog.RunLog.Error("use310PMixedInsert is ture, volcanoType should be false")
		return false
	}
	switch *hotReset {
	case common.HotResetClose, common.HotResetInfer, common.HotResetTrain:
	default:
		hwlog.RunLog.Error("hot reset mode param invalid")
		return false
	}
	if (*hotReset) == common.HotResetTrain {
		hwlog.RunLog.Warn("hotReset to 1 is a reserved value")
	}
	switch *mode {
	case common.RunMode310, common.RunMode910, common.RunMode310P, "":
	default:
		hwlog.RunLog.Errorf("unSupport mode: %s", *mode)
		return false
	}
	if BuildScene != common.EdgeScene && BuildScene != common.CenterScene {
		hwlog.RunLog.Error("unSupport build scene, only support edge and center")
		return false
	}
	return true
}

func main() {
	flag.Parse()
	if *version {
		fmt.Printf("%s version: %s\n", BuildName, BuildVersion)
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	if err := initLogModule(ctx); err != nil {
		return
	}
	if !checkParam() {
		return
	}
	hwlog.RunLog.Infof("ascend device plugin starting and the version is %s", BuildVersion)
	hwlog.RunLog.Infof("ascend device plugin starting scene is %s", BuildScene)
	setParameters()
	hdm, err := InitFunction()
	if err != nil {
		return
	}
	setUseAscendDocker()

	go hdm.ListenDevice(ctx)
	hdm.SignCatch(cancel)
}

// InitFunction init function
func InitFunction() (*server.HwDevManager, error) {
	devM, err := devmanager.AutoInit("")
	if err != nil {
		hwlog.RunLog.Errorf("init devmanager failed, err: %#v", err)
		return nil, err
	}
	hdm := server.NewHwDevManager(devM)
	if hdm == nil {
		hwlog.RunLog.Error("init device manager failed")
		return nil, fmt.Errorf("init device manager failed")
	}
	hwlog.RunLog.Info("init device manager success")
	return hdm, nil
}

func setParameters() {
	common.ParamOption = common.Option{
		GetFdFlag:          *fdFlag,
		UseAscendDocker:    *useAscendDocker,
		UseVolcanoType:     *volcanoType,
		AutoStowingDevs:    *autoStowing,
		ListAndWatchPeriod: *listWatchPeriod,
		PresetVDevice:      *presetVirtualDevice,
		Use310PMixedInsert: *use310PMixedInsert,
		HotReset:           *hotReset,
		BuildScene:         BuildScene,
	}
}

func setUseAscendDocker() {
	*useAscendDocker = true
	ascendDocker := os.Getenv("ASCEND_DOCKER_RUNTIME")
	if ascendDocker != "True" {
		*useAscendDocker = false
		hwlog.RunLog.Debugf("get ASCEND_DOCKER_RUNTIME from env is: %#v", ascendDocker)
	}
	if common.ParamOption.Use310PMixedInsert {
		*useAscendDocker = false
		hwlog.RunLog.Debugf("310P mixed insert mode do not use ascend docker")
	}
	if len(common.ParamOption.ProductTypes) == 1 && common.ParamOption.ProductTypes[0] == common.Atlas200ISoc {
		*useAscendDocker = false
		hwlog.RunLog.Debugf("your device-type is: %v", common.Atlas200ISoc)
	}

	common.ParamOption.UseAscendDocker = *useAscendDocker
	hwlog.RunLog.Infof("device-plugin set ascend docker as: %v", *useAscendDocker)
}
