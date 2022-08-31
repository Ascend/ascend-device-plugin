/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2020-2022. All rights reserved.
 */

// Package main implements initialization of the startup parameters of the device plugin.
package main

import (
	"context"
	"flag"
	"fmt"

	"huawei.com/mindx/common/hwlog"
	"huawei.com/npu-exporter/devmanager"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/device"
	"Ascend-device-plugin/pkg/kubeclient"
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
)

var (
	mode            = flag.String("mode", "", "Device plugin running mode: ascend310, ascend310P, ascend910")
	fdFlag          = flag.Bool("fdFlag", false, "Whether to use fd system to manage device (default false)")
	useAscendDocker = flag.Bool("useAscendDocker", true, "Whether to use ascend docker")
	volcanoType     = flag.Bool("volcanoType", false,
		"Specifies whether to use volcano for scheduling when the chip type is Ascend310 or Ascend910 (default false)")
	version     = flag.Bool("version", false, "Output version information")
	edgeLogFile = flag.String("edgeLogFile", "/var/alog/AtlasEdge_log/devicePlugin.log",
		"Log file path in edge scene")
	listWatchPeriod = flag.Int("listWatchPeriod", defaultListWatchPeriod,
		"Listen and watch device state's period, unit second, range [3, 60]")
	autoStowing = flag.Bool("autoStowing", true, "Whether to automatically stow the fixed device")
	logLevel    = flag.Int("logLevel", 0,
		"Log level, -1-debug, 0-info(default), 1-warning, 2-error, 3-dpanic, 4-panic, 5-fatal (default 0)")
	logMaxAge = flag.Int("maxAge", common.MaxAge,
		"Maximum number of days for backup run log files, must be greater than or equal to 7 days")
	logFile = flag.String("logFile", defaultLogPath,
		"The log file path, if the file size exceeds 20MB, will be rotate")
	logMaxBackups = flag.Int("maxBackups", common.MaxBackups,
		"Maximum number of backup log files, range is (0, 30]")
	kubeconfig = flag.String("kubeConfig", "", "Path to a kubeconfig. "+
		"Only required if out-of-cluster.")
	presetVirtualDevice = flag.Bool("presetVirtualDevice", true, "Open the static of "+
		"computing power splitting function, only support Ascend910 and Ascend310P")
)

var (
	// BuildName show app name
	BuildName string
	// BuildVersion show app version
	BuildVersion string
)

func initLogModule(ctx context.Context) error {
	var loggerPath string
	loggerPath = *logFile
	if *fdFlag {
		loggerPath = *edgeLogFile
	}
	hwLogConfig := hwlog.LogConfig{
		LogFileName: loggerPath,
		LogLevel:    *logLevel,
		MaxBackups:  *logMaxBackups,
		MaxAge:      *logMaxAge,
	}
	if err := hwlog.InitRunLogger(&hwLogConfig, ctx); err != nil {
		fmt.Printf("hwlog init failed, error is %v", err)
		return err
	}
	return nil
}

func checkParam() bool {
	if *listWatchPeriod < minListWatchPeriod || *listWatchPeriod > maxListWatchPeriod {
		fmt.Printf("list and watch period %d out of range\n", *listWatchPeriod)
		return false
	}
	if !(*presetVirtualDevice) {
		fmt.Println("presetVirtualDevice can be only set to true")
		return false
	}
	if len(*mode) > maxRunModeLength {
		fmt.Println("run mode param length invalid")
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
	if !checkParam() {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	if err := initLogModule(ctx); err != nil {
		return
	}
	hwlog.RunLog.Infof("ascend device plugin starting and the version is %s", BuildVersion)

	setParameters()
	hdm, err := InitFunction()
	if err != nil {
		return
	}

	go hdm.ListenDevice(ctx)
	hdm.SignCatch(cancel)
}

// InitFunction init function
func InitFunction() (*device.HwDevManager, error) {
	var err error
	devM, err := devmanager.AutoInit("")
	if err != nil {
		hwlog.RunLog.Errorf("init devmanager failed, err: %v", err)
		return nil, err
	}
	var kubeClient *kubeclient.ClientK8s
	if common.ParamOption.UseVolcanoType {
		kubeClient, err = kubeclient.NewClientK8s(common.ParamOption.KubeConfig)
		if err != nil {
			hwlog.RunLog.Errorf("init kubeclient failed err: %#v", err)
			return nil, err
		}
		hwlog.RunLog.Infof("init kubeclient success")
	}
	hdm := device.NewHwDevManager(devM, kubeClient)
	if hdm == nil {
		hwlog.RunLog.Errorf("init device manager failed")
		return nil, err
	}
	hwlog.RunLog.Infof("init device manager success")
	return hdm, nil
}

func setParameters() {
	common.ParamOption = common.Option{
		GetFdFlag:          *fdFlag,
		UseAscendDocker:    *useAscendDocker,
		UseVolcanoType:     *volcanoType,
		AutoStowingDevs:    *autoStowing,
		ListAndWatchPeriod: *listWatchPeriod,
		KubeConfig:         *kubeconfig,
		PresetVDevice:      *presetVirtualDevice,
	}
}
