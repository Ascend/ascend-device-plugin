// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package server holds the implementation of registration to kubelet, k8s device plugin interface and grpc service.
package server

import (
	"sync"

	"google.golang.org/grpc"
	"k8s.io/api/core/v1"
	"k8s.io/kubernetes/pkg/kubelet/apis/podresources/v1alpha1"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/device"
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
	manager              device.DevManager
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
	klt2RealDevMap       map[string]string
	restart              bool
}

// PodDevice define device info in pod
type PodDevice struct {
	ResourceName string
	DeviceIds    []string
}

// PodResource implements the get pod resource info
type PodResource struct {
	conn   *grpc.ClientConn
	client v1alpha1.PodResourcesListerClient
}

// PodDeviceInfo define device info of pod, include kubelet allocate and real allocate device
type PodDeviceInfo struct {
	Pod        v1.Pod
	KltDevice  []string
	RealDevice []string
}
