// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package server holds the implementation of registration to kubelet, k8s device plugin interface and grpc service.
package server

import (
	"context"
	"errors"

	"huawei.com/npu-exporter/hwlog"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

func (ps *PluginServer) stopListAndWatch() {
	if ps.isRunning.Load() {
		ps.stop <- struct{}{}
	}
}

// ListAndWatch is to send device info to kubelet
func (ps *PluginServer) ListAndWatch(empty *v1beta1.Empty, stream v1beta1.DevicePlugin_ListAndWatchServer) error {
	return nil
}

// Allocate is called by kubelet to mount device to k8s pod.
func (ps *PluginServer) Allocate(ctx context.Context, requests *v1beta1.AllocateRequest) (*v1beta1.AllocateResponse,
	error) {
	return &v1beta1.AllocateResponse{}, nil
}

// GetPreferredAllocation implement the kubelet device plugin interface
func (ps *PluginServer) GetPreferredAllocation(context.Context, *v1beta1.PreferredAllocationRequest) (
	*v1beta1.PreferredAllocationResponse, error) {
	return nil, errors.New("not support")
}

// GetDevicePluginOptions is Standard interface to kubelet.
func (ps *PluginServer) GetDevicePluginOptions(ctx context.Context, e *v1beta1.Empty) (*v1beta1.DevicePluginOptions,
	error) {
	return &v1beta1.DevicePluginOptions{}, nil
}

// PreStartContainer is Standard interface to kubelet with empty implement.
func (ps *PluginServer) PreStartContainer(ctx context.Context,
	r *v1beta1.PreStartContainerRequest) (*v1beta1.PreStartContainerResponse, error) {
	hwlog.RunLog.Infof("PreStart just call in UT.")
	return &v1beta1.PreStartContainerResponse{}, nil
}
