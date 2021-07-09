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
	socketPath = "/var/lib/kubelet/device-plugins"
	defaultLogPath    = "/var/log/mindx-dl/devicePlugin/devicePlugin.log"

	// defaultListWatchPeriod is the default listening device state's period
	defaultListWatchPeriod = 5

	// maxListWatchPeriod is the max listening device state's period
	maxListWatchPeriod = 60
	// minListWatchPeriod is the min listening device state's period
	minListWatchPeriod = 3
)

var (
	mode            = flag.String("mode", "", "device plugin running mode")
	fdFlag          = flag.Bool("fdFlag", false, "set the connect system is fd system")
	useAscendDocker = flag.Bool("useAscendDocker", true, "use ascend docker or not")
	volcanoType     = flag.Bool("volcanoType", false, "use volcano to schedue")
	version         = flag.Bool("version", false, "show k8s device plugin version ")
	edgeLogFile     = flag.String("edgeLogFile", "/var/alog/AtlasEdge_log/devicePlugin.log",
		"edge log file path")
	listWatchPeriod = flag.Int("listWatchPeriod", defaultListWatchPeriod, "listen and "+
		"watch device state's period, unit is second, scope is [3, 60]")
	autoStowing     = flag.Bool("autoStowing", true, "auto stowing fixes devices or not")
	logLevel        = flag.Int("logLevel", 0,
		"log level, -1-debug, 0-info(default), 1-warning, 2-error, 3-dpanic, 4-panic, 5-fatal")
	logMaxAge       = flag.Int("maxAge", hwmanager.MaxAge, "maximum number of days for backup log files")
	logIsCompress   = flag.Bool("isCompress", false,
		"whether backup files need to be compressed (default false)")
	logFile         = flag.String("logFile", defaultLogPath, "log file path")
	logMaxBackups   = flag.Int("maxBackups", hwmanager.MaxBackups, "maximum number of backup log files")
)

var (
	// BuildName is k8s-device-plugin
	BuildName string
	// BuildVersion show app version
	BuildVersion string
)

func initLogModule( stopCh <-chan struct{}) {
	var loggerPath string
	loggerPath = *logFile
	if *fdFlag {
		loggerPath = *edgeLogFile
	}
	hwLogConfig := hwlog.LogConfig{
		LogFileName:   loggerPath,
		LogLevel:      *logLevel,
		MaxBackups:    *logMaxBackups,
		MaxAge:        *logMaxAge,
		IsCompress:    *logIsCompress,
	}
	if err := hwlog.Init(&hwLogConfig, stopCh); err != nil {
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

	neverStop := make(chan struct{})
	switch *mode {
	case "ascend310", "pci", "vnpu", "ascend910", "ascend710", "":
		hwlog.Infof("ascend device plugin running mode: %s", *mode)
	default:
		hwlog.Infof("unSupport mode: %s, waiting indefinitely", *mode)
		<-neverStop
	}

	hdm := hwmanager.NewHwDevManager(*mode)
	hdm.SetParameters(*fdFlag, *useAscendDocker, *volcanoType, *autoStowing, *listWatchPeriod)
	if err := hdm.GetNPUs(); err != nil {
		hwlog.Errorf("no devices found. waiting indefinitely, err: %s", err.Error())
		<-neverStop
	}

	devTypes := hdm.GetDevType()
	if len(devTypes) == 0 {
		hwlog.Errorf("no devices type found. waiting indefinitely")
		<-neverStop
	}

	for _, devType := range devTypes {
		hwlog.Infof("ascend device serve started, devType: %s", devType)
		pluginSocket := fmt.Sprintf("%s.sock", devType)
		go hdm.Serve(devType, socketPath, pluginSocket, hwmanager.NewHwPluginServe)
	}

	<-neverStop
}
