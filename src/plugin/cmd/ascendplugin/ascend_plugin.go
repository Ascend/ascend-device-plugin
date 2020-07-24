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
	kubeletSocket = "kubelet.sock"
)

var (
	dlogPath   = flag.String("dlog-path", "/var/dlog", "Path on the host that contains log")
	socketPath = flag.String("plugin-directory", "/var/lib/kubelet/device-plugins",
		"Path to create plugin socket")
	mode           = flag.String("mode", "", "device plugin running mode")
	timeInterval   = flag.String("timeInterval", "1", "check frequency of AI core health")
	checkNum       = flag.String("checkNum", "5", "check num of AI core health")
	restoreNum     = flag.String("restoreNum", "3", "restore num of AI core health")
	highThreshold  = flag.String("highThreshold", "90", "AI core high-level threshold of frequency")
	lowThreshold   = flag.String("lowThreshold", "80", "AI core low-level threshold of frequency")
	netDetect      = flag.Bool("netDetect", false, "detect device network health ")
	version        = flag.Bool("version", false, "show k8s device plugin version ")
	fdFlag         = flag.Bool("fdFlag", false, "set the connect system is fd system")
	useAscendDocer = flag.Bool("useAscendDocker", true, "use ascend docker or not")
	volcanoType    = flag.Bool("volcanoType", false, "use volcano to schedue")
)

var (
	// BuildName is k8s-device-plugin
	BuildName string
	// BuildVersion show app version
	BuildVersion string
	// BuildTime show build time
	BuildTime string
)

func main() {
	var log *zap.Logger
	log = hwmanager.ConfigLog(hwmanager.LogPath)

	flag.Parse()

	if *version {
		fmt.Printf("%s version: %s   buildTime:%s\n", BuildName, BuildVersion, BuildTime)
		os.Exit(0)
	}

	neverStop := make(chan struct{})
	switch *mode {
	case "ascend310", "pci", "vnpu", "ascend910", "":
		log.Info("ascend device plugin running mode", zap.String("mode", *mode))
	default:
		log.Info("unSupport mode, waiting indefinitely", zap.String("mode", *mode))
		<-neverStop
	}

	hdm := hwmanager.NewHwDevManager(*mode, *dlogPath)
	hdm.SetParameters(fdFlag, useAscendDocer, volcanoType)
	if err := hdm.GetNPUs(*timeInterval, *checkNum, *restoreNum, *highThreshold, *lowThreshold,
		*netDetect); err != nil {
		log.Error("no devices found. waiting indefinitely", zap.String("err", err.Error()))
		<-neverStop
	}

	devTypes := hdm.GetDevType()
	if len(devTypes) == 0 {
		log.Error("no devices type found. waiting indefinitely")
		<-neverStop
	}

	for _, devType := range devTypes {
		log.Info("ascend device serve started", zap.String("devType", devType))
		pluginSocket := fmt.Sprintf("%s.sock", devType)
		go hdm.Serve(devType, *socketPath, kubeletSocket, pluginSocket)
	}

	<-neverStop
}
