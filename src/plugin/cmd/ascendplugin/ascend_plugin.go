/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2020-2021. All rights reserved.
 */

// Package main implements initialization of the startup parameters of the device plugin.
package main

import (
	hwmanager "Ascend-device-plugin/src/plugin/pkg/npu/huawei"
	"flag"
	"fmt"
	"huawei.com/npu-exporter/hwlog"
	"os"
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
)

var (
	mode            = flag.String("mode", "", "Device plugin running mode: ascend310, ascend710, ascend910")
	fdFlag          = flag.Bool("fdFlag", false, "Whether to use fd system to manage device (default false)")
	useAscendDocker = flag.Bool("useAscendDocker", true, "Whether to use ascend docker")
	volcanoType     = flag.Bool("volcanoType", false,
		"Specifies whether to use volcano for scheduling when the chip type is Ascend310 or Ascend910 (default false)")
	version         = flag.Bool("version", false, "Output version information")
	edgeLogFile     = flag.String("edgeLogFile", "/var/alog/AtlasEdge_log/devicePlugin.log",
		"Log file path in edge scene")
	listWatchPeriod = flag.Int("listWatchPeriod", defaultListWatchPeriod,
		"Listen and watch device state's period, unit second, range [3, 60]")
	autoStowing = flag.Bool("autoStowing", true, "Whether to automatically stow the fixed device")
	logLevel    = flag.Int("logLevel", 0,
		"Log level, -1-debug, 0-info(default), 1-warning, 2-error, 3-dpanic, 4-panic, 5-fatal (default 0)")
	logMaxAge     = flag.Int("maxAge", hwmanager.MaxAge,
		"Maximum number of days for backup run log files, must be greater than or equal to 7 days")
	logFile       = flag.String("logFile", defaultLogPath,
		"The log file path, if the file size exceeds 20MB, will be rotate")
	logMaxBackups = flag.Int("maxBackups", hwmanager.MaxBackups,
		"Maximum number of backup log files, range is (0, 30]")
)

var (
	// BuildName show app name
	BuildName string
	// BuildVersion show app version
	BuildVersion string
)

func initLogModule(stopCh <-chan struct{}) {
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
		fmt.Printf("init hwlog error %v", err.Error())
		os.Exit(1)
	}
}

func main() {

	flag.Parse()

	if *version {
		fmt.Printf("%s version: %s\n", BuildName, BuildVersion)
		os.Exit(0)
	}

	if *listWatchPeriod < minListWatchPeriod || *listWatchPeriod > maxListWatchPeriod {
		fmt.Printf("list and watch period %d out of range\n", *listWatchPeriod)
		os.Exit(1)
	}

	stopCh := make(chan struct{})
	defer close(stopCh)
	initLogModule(stopCh)
	hwlog.RunLog.Infof("ascend device plugin starting and the version is %s", BuildVersion)

	neverStop := make(chan struct{})
	switch *mode {
	case "ascend310", "ascend910", "ascend710", "":
		hwlog.RunLog.Infof("ascend device plugin running mode: %s", *mode)
	default:
		hwlog.RunLog.Infof("unSupport mode: %s, waiting indefinitely", *mode)
		<-neverStop
	}

	hdm := hwmanager.NewHwDevManager(*mode)
	hdm.SetParameters(*fdFlag, *useAscendDocker, *volcanoType, *autoStowing, *listWatchPeriod)
	if err := hdm.GetNPUs(); err != nil {
		hwlog.RunLog.Errorf("no devices found. waiting indefinitely, err: %s", err.Error())
		<-neverStop
	}

	devTypes := hdm.GetDevType()
	if len(devTypes) == 0 {
		hwlog.RunLog.Errorf("no devices type found. waiting indefinitely")
		<-neverStop
	}

	for _, devType := range devTypes {
		hwlog.RunLog.Infof("ascend device serve started, devType: %s", devType)
		go hdm.Serve(devType, hwmanager.NewHwPluginServe)
	}

	<-neverStop
}
