// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package server holds the implementation of registration to kubelet, k8s device plugin interface and grpc service.
package server

import (
	"sync"

	"google.golang.org/grpc"
	"huawei.com/npu-exporter/devmanager"
	"k8s.io/kubernetes/pkg/kubelet/apis/podresources/v1alpha1"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/kubeclient"
)

// InterfaceServer interface for object that keeps running for providing service
type InterfaceServer interface {
	Start(*common.FileWatch) error
	Stop()
	GetRestartFlag() bool
	SetRestartFlag(bool)
}

// PluginServer implements the interface of DevicePluginServer; manages the registration and lifecycle of grpc server
type PluginServer struct {
	kubeClient           *kubeclient.ClientK8s
	grpcServer           *grpc.Server
	isRunning            *common.AtomicBool
	cachedDevices        []common.NpuDevice
	deviceType           string
	ascendRuntimeOptions string
	defaultDevs          []string
	allocMapLock         sync.RWMutex
	cachedLock           sync.RWMutex
	reciChan             chan interface{}
	stop                 chan interface{}
	devManager           devmanager.DeviceInterface
	vol2KlDevMap         map[string]string
	restart              bool
}

// PodDevice define device info in pod
type PodDevice struct {
	ResourceName string
	DeviceIds    []string
}

// PodResource implements the get pod resource info
type PodResource struct {
	conn    *grpc.ClientConn
	client  v1alpha1.PodResourcesListerClient
	restart bool
}
