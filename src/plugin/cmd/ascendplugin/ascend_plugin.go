/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2020-2022. All rights reserved.
 */

// Package main implements initialization of the startup parameters of the device plugin.
package main

import (
	"flag"
	"fmt"

	"huawei.com/npu-exporter/hwlog"

	"Ascend-device-plugin/src/plugin/pkg/npu/common"
	"Ascend-device-plugin/src/plugin/pkg/npu/huawei"
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
	logMaxAge = flag.Int("maxAge", huawei.MaxAge,
		"Maximum number of days for backup run log files, must be greater than or equal to 7 days")
	logFile = flag.String("logFile", defaultLogPath,
		"The log file path, if the file size exceeds 20MB, will be rotate")
	logMaxBackups = flag.Int("maxBackups", huawei.MaxBackups,
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

func initLogModule(stopCh <-chan struct{}) error {
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
	if err := hwlog.InitRunLogger(&hwLogConfig, stopCh); err != nil {
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
	stopCh := make(chan struct{})
	defer close(stopCh)
	if err := initLogModule(stopCh); err != nil {
		return
	}
	hwlog.RunLog.Infof("ascend device plugin starting and the version is %s", BuildVersion)
	if len(*mode) > maxRunModeLength {
		hwlog.RunLog.Errorf("run mode param length invalid")
		return
	}
	neverStop := make(chan struct{})

	switch *mode {
	case common.RunMode310, common.RunMode910, common.RunMode310P, "":
		hwlog.RunLog.Infof("ascend device plugin running mode: %s", *mode)
	default:
		hwlog.RunLog.Infof("unSupport mode: %s, waiting indefinitely", *mode)
		<-neverStop
	}
	hdm := huawei.NewHwDevManager(*mode)
	if hdm == nil {
		hwlog.RunLog.Errorf("init device manager failed")
		return
	}
	if err := hdm.SetRunMode(); err != nil {
		hwlog.RunLog.Errorf("err to set Run mode, err: %v ", err)
		<-neverStop
	}
	hdm.SetParameters(getParams())
	if err := hdm.GetNPUs(); err != nil {
		hwlog.RunLog.Errorf("no devices found. waiting indefinitely, err: %s", err.Error())
		<-neverStop
	}
	if len(hdm.GetDevType()) == 0 {
		hwlog.RunLog.Errorf("no devices type found. waiting indefinitely")
		<-neverStop
	}
	if *volcanoType {
		if err := common.GetNodeNameFromEnv(); err != nil {
			hwlog.RunLog.Errorf("get node name failed. waiting indefinitely, err: %v", err)
			<-neverStop
		}
	}
	startDiffTypeServe(hdm, neverStop)
	<-neverStop
}

func getParams() huawei.Option {
	return huawei.Option{
		GetFdFlag:          *fdFlag,
		UseAscendDocker:    *useAscendDocker,
		UseVolcanoType:     *volcanoType,
		AutoStowingDevs:    *autoStowing,
		ListAndWatchPeriod: *listWatchPeriod,
		KubeConfig:         *kubeconfig,
		PresetVDevice:      *presetVirtualDevice,
	}
}

func startDiffTypeServe(hdm *huawei.HwDevManager, stop chan struct{}) {
	for _, devType := range hdm.GetDevType() {
		hwlog.RunLog.Infof("ascend device serve started, devType: %s", devType)
		go hdm.Serve(devType, stop)
	}
}
