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
	"go.uber.org/atomic"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/util/sets"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"os"
	"time"
)

// HwPluginServe show plugin data
type HwPluginServe struct {
	hdm            *HwDevManager
	devices        map[string]*npuDevice
	grpcServer     *grpc.Server
	devType        string
	runMode        string
	defaultDevs    []string
	socket         string
	kubeInteractor *KubeInteractor
	healthDevice   sets.String
}

// HwPluginServeInterface the interface of PluginServer
type HwPluginServeInterface interface {
	GetDevByType() error
	Start(pluginSocket, pluginSocketPath string) error
	setSocket(pluginSocketPath string)
	Stop() error
	cleanSock() error
	Register(k8sSocketPath, pluginSocket, resourceName string) error
}

// NewHwPluginServe new a device plugin server
func NewHwPluginServe(hdm *HwDevManager, devType string, socket string) HwPluginServeInterface {
	var ki *KubeInteractor
	var err error
	if useVolcanoType {
		ki, err = NewKubeInteractor()
		if err != nil {
			logger.Error("cannot create kube interactor.", zap.Error(err))
		}
	}
	return &HwPluginServe{
		devType:        devType,
		hdm:            hdm,
		runMode:        hdm.runMode,
		devices:        make(map[string]*npuDevice, hiAIMaxDeviceNum),
		socket:         socket,
		kubeInteractor: ki,
		healthDevice:   sets.String{},
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
func (hps *HwPluginServe) Start(pluginSocket, pluginSocketPath string) error {
	netListen, err := createNetListen(pluginSocketPath)
	if err != nil {
		return err
	}
	hps.setSocket(pluginSocketPath)

	// noinspection ALL
	go hps.grpcServer.Serve(netListen)

	// Wait for grpcServer
	for len(hps.grpcServer.GetServiceInfo()) <= 0 {
		time.Sleep(1 * time.Second)
	}
	logger.Info("device plugin start serving.")

	// Registers To Kubelet.
	resourceName := fmt.Sprintf("%s%s", resourceNamePrefix, hps.devType)
	k8sSocketPath := pluginapi.KubeletSocket
	err = hps.Register(k8sSocketPath, pluginSocket, resourceName)
	if err == nil {
		logger.Info("register to kubelet success.")
		return nil
	}
	hps.grpcServer.Stop()
	time.Sleep(sleepTime * time.Second)
	logger.Error("register to kubelet failed.", zap.String("err", err.Error()))
	return err
}

func (hps *HwPluginServe) setSocket(pluginSocketPath string) {
	hps.socket = pluginSocketPath
	hps.grpcServer = grpc.NewServer()
	// Registers service.
	plugin := &pluginAPI{hps: hps, outbreak: atomic.NewBool(false)}
	pluginapi.RegisterDevicePluginServer(plugin.hps.grpcServer, plugin)
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
