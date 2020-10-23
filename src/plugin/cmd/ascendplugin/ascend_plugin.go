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
	"go.uber.org/zap"
	"os"
)

const (
	// socket name
	dlogPath   = "/var/dlog"
	socketPath = "/var/lib/kubelet/device-plugins"
	logPath    = "/var/log/devicePlugin"
)

var (
	mode            = flag.String("mode", "", "device plugin running mode")
	fdFlag          = flag.Bool("fdFlag", false, "set the connect system is fd system")
	useAscendDocker = flag.Bool("useAscendDocker", true, "use ascend docker or not")
	volcanoType     = flag.Bool("volcanoType", false, "use volcano to schedue")
	version         = flag.Bool("version", false, "show k8s device plugin version ")
	logDir          = flag.String("logDir", "/var/alog/AtlasEdge_log", "log path")
)

var (
	// BuildName is k8s-device-plugin
	BuildName string
	// BuildVersion show app version
	BuildVersion string
)

func main() {

	flag.Parse()
	var loggerPath string

	loggerPath = fmt.Sprintf("%s/%s", logPath, hwmanager.LogName)
	if *fdFlag {
		loggerPath = fmt.Sprintf("%s/%s", *logDir, hwmanager.LogName)
	}
	err := hwmanager.NewLogger(loggerPath)
	if err != nil {
		fmt.Errorf("new logger is error %v", err.Error())
		os.Exit(1)
	}
	logger := hwmanager.GetLogger()
	if logger == nil {
		fmt.Errorf("logger is nil")
		os.Exit(1)
	}
	if *version {
		fmt.Printf("%s version: %s\n", BuildName, BuildVersion)
		os.Exit(0)
	}

	neverStop := make(chan struct{})
	switch *mode {
	case "ascend310", "pci", "vnpu", "ascend910", "":
		logger.Info("ascend device plugin running mode", zap.String("mode", *mode))
	default:
		logger.Info("unSupport mode, waiting indefinitely", zap.String("mode", *mode))
		<-neverStop
	}

	hdm := hwmanager.NewHwDevManager(*mode, dlogPath, loggerPath)
	hdm.SetParameters(*fdFlag, *useAscendDocker, *volcanoType)
	if err := hdm.GetNPUs(); err != nil {
		logger.Error("no devices found. waiting indefinitely", zap.String("err", err.Error()))
		<-neverStop
	}

	devTypes := hdm.GetDevType()
	if len(devTypes) == 0 {
		logger.Error("no devices type found. waiting indefinitely")
		<-neverStop
	}

	for _, devType := range devTypes {
		logger.Info("ascend device serve started", zap.String("devType", devType))
		pluginSocket := fmt.Sprintf("%s.sock", devType)
		go hdm.Serve(devType, socketPath, pluginSocket, hwmanager.NewHwPluginServe)
	}

	<-neverStop
}
