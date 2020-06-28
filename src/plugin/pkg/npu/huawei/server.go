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

import (
	"fmt"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"net"
	"os"
	"time"
)

// HwPluginServe show plugin data
type HwPluginServe struct {
	hdm         *HwDevManager
	devices     map[string]*npuDevice
	grpcServer  *grpc.Server
	devType     string
	runMode     string
	defaultDevs []string
	socket      string
}

// NewHwPluginServe new a device plugin server
func NewHwPluginServe(hdm *HwDevManager, devType string, socket string) *HwPluginServe {
	return &HwPluginServe{
		devType: devType,
		hdm:     hdm,
		runMode: hdm.runMode,
		devices: make(map[string]*npuDevice),
		socket:  socket,
	}
}

// GetDevByType get dev by type
func (hps *HwPluginServe) GetDevByType() error {
	allDevs := hps.hdm.allDevs
	if len(allDevs) == 0 {
		return fmt.Errorf("no device found")
	}

	for i := range allDevs {
		dev := &allDevs[i]
		if dev.devType == hps.devType {
			hps.devices[dev.ID] = dev
		}
	}
	if len(hps.devices) == 0 {
		return fmt.Errorf("no %s device found", hps.devType)
	}

	defaultDevs := hps.hdm.defaultDevs
	if len(defaultDevs) != 0 {
		for _, dev := range defaultDevs {
			hps.defaultDevs = append(hps.defaultDevs, dev)
		}
	}

	return nil
}

// Start starts the gRPC server of the device plugin
func (hps *HwPluginServe) Start(pluginPath, k8sSocket, pluginSocket, pluginSocketPath string) error {
	if _, err := os.Stat(pluginSocketPath); err == nil {
		logger.Info("Found exist sock file,now remove it.", zap.String("sockName", pluginSocketPath))
		os.Remove(pluginSocketPath)
	}
	lis, err := net.Listen("unix", pluginSocketPath)
	if err != nil {
		logger.Error("device plugin start failed.", zap.String("err", err.Error()))
	}
	err = os.Chmod(pluginSocketPath, logChmod)
	if err != nil {
		logger.Error("chmod error", zap.Error(err))
	}
	hps.socket = pluginSocketPath
	hps.grpcServer = grpc.NewServer()

	// Registers service.
	plugin := &pluginAPI{hps: hps}
	pluginapi.RegisterDevicePluginServer(plugin.hps.grpcServer, plugin)

	// noinspection ALL
	go hps.grpcServer.Serve(lis)

	// Wait for grpcServer
	for len(hps.grpcServer.GetServiceInfo()) <= 0 {
		time.Sleep(1 * time.Second)
	}
	logger.Info("device plugin start serving.")

	// Registers To Kubelet.
	resourceName := fmt.Sprintf("%s%s", resourceNamePrefix, hps.devType)
	k8sSocketPath := pluginapi.KubeletSocket
	countRegister := 0

reRegister:
	err = Register(k8sSocketPath, pluginSocket, resourceName)

	if err != nil {
		hps.grpcServer.Stop()
		logger.Error("register to kubelet failed.", zap.String("err", err.Error()))
		if countRegister < registerTimeout {
			countRegister++
			logger.Error("k8s device plugin register failed.", zap.Int("failed time", countRegister))
			time.Sleep(sleepTime * time.Second)
			goto reRegister
		}
		return nil
	}
	logger.Info("register to kubelet success.")
	return nil
}

// Stop the gRPC server
func (hps *HwPluginServe) Stop() error {
	if hps.grpcServer == nil {
		return nil
	}
	hps.grpcServer.Stop()
	hps.grpcServer = nil

	return hps.cleanSock()
}

// if device plugin stopped,the socket file should be removed
func (hps *HwPluginServe) cleanSock() error {

	if err := os.Remove(hps.socket); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}
