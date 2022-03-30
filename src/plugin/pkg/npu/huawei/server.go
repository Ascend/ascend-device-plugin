/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2020-2022. All rights reserved.
 */

package huawei

import (
	"fmt"
	"time"

	"go.uber.org/atomic"
	"google.golang.org/grpc"
	"huawei.com/npu-exporter/hwlog"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"Ascend-device-plugin/src/plugin/pkg/npu/common"
)

// HwPluginServe show plugin data
type HwPluginServe struct {
	hdm            *HwDevManager
	devices        map[string]*common.NpuDevice
	grpcServer     *grpc.Server
	vol2KlDevMap   map[string]string
	kubeInteractor *KubeInteractor
	healthDevice   sets.String
	unHealthDevice sets.String
	devType        string
	runMode        string
}

// HwPluginServeInterface the interface of PluginServer
type HwPluginServeInterface interface {
	GetDevByType() error
	Start(pluginSocketPath string) error
	setSocket()
	Stop()
	Register() error
}

// NewHwPluginServe new a device plugin server
func NewHwPluginServe(hdm *HwDevManager, devType string) HwPluginServeInterface {
	var ki *KubeInteractor
	var err error
	if useVolcanoType {
		ki, err = NewKubeInteractor()
		if err != nil {
			hwlog.RunLog.Errorf("cannot create kube interactor, err: %v", err)
			return nil
		}
	}
	return &HwPluginServe{
		devType:        devType,
		hdm:            hdm,
		runMode:        hdm.runMode,
		devices:        make(map[string]*common.NpuDevice, hiAIMaxDeviceNum),
		kubeInteractor: ki,
		healthDevice:   sets.String{},
		unHealthDevice: sets.String{},
	}
}

// GetDevByType get dev by type
func (hps *HwPluginServe) GetDevByType() error {
	allDevs := hps.hdm.allDevs
	if len(allDevs) == 0 {
		return fmt.Errorf("no device found")
	}
	hps.devices = make(map[string]*common.NpuDevice, 1)
	for i := range allDevs {
		dev := &allDevs[i]
		if dev.DevType == hps.devType {
			hps.devices[dev.ID] = dev
		}
	}
	if len(hps.devices) == 0 {
		return fmt.Errorf("no %s device found", hps.devType)
	}
	return nil
}

// Start starts the gRPC server of the device plugin
func (hps *HwPluginServe) Start(pluginSocketPath string) error {
	netListen, err := createNetListen(pluginSocketPath)
	if err != nil {
		return err
	}
	hps.setSocket()

	// noinspection ALL
	go hps.grpcServer.Serve(netListen)

	// Wait for grpcServer
	for len(hps.grpcServer.GetServiceInfo()) <= 0 {
		time.Sleep(time.Second)
	}
	hwlog.RunLog.Infof("device plugin start serving.")

	// Registers To Kubelet.
	if err = hps.Register(); err == nil {
		hwlog.RunLog.Infof("register to kubelet success.")
		return nil
	}
	hps.grpcServer.Stop()
	time.Sleep(sleepTime * time.Second)
	hwlog.RunLog.Errorf("register to kubelet failed, err: %s", err.Error())
	return err
}

func (hps *HwPluginServe) setSocket() {
	hps.grpcServer = grpc.NewServer()
	// Registers service.
	plugin := &pluginAPI{hps: hps, outbreak: atomic.NewBool(false)}
	v1beta1.RegisterDevicePluginServer(hps.grpcServer, plugin)
}

// Stop the gRPC server
func (hps *HwPluginServe) Stop() {
	if hps.grpcServer == nil {
		return
	}
	hps.grpcServer.Stop()
	hps.grpcServer = nil

	return
}
