// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package server holds the implementation of registration to kubelet, k8s device plugin interface and grpc service.
package server

import (
	"context"
	"fmt"
	"net"
	"os"
	"path"
	"time"

	"google.golang.org/grpc"
	"huawei.com/mindx/common/hwlog"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"Ascend-device-plugin/pkg/common"
)

// Start starts the gRPC server, registers the device plugin with the Kubelet
func (ps *PluginServer) Start(socketWatcher *common.FileWatch) error {
	// clean
	ps.Stop()

	var err error

	// Start gRPC server
	if err = ps.serve(socketWatcher); err != nil {
		return err
	}

	// Registers To Kubelet.
	if err = ps.register(); err == nil {
		hwlog.RunLog.Infof("register %s to kubelet success.", ps.deviceType)
		return nil
	}
	ps.Stop()
	hwlog.RunLog.Errorf("register to kubelet failed, err: %#v", err)
	return err
}

// Stop the gRPC server
func (ps *PluginServer) Stop() {
	ps.isRunning.Store(false)

	if ps.grpcServer == nil {
		return
	}
	ps.stopListAndWatch()
	ps.grpcServer.Stop()

	return
}

// GetRestartFlag get restart flag
func (ps *PluginServer) GetRestartFlag() bool {
	return ps.restart
}

// SetRestartFlag set restart flag
func (ps *PluginServer) SetRestartFlag(flag bool) {
	ps.restart = flag
}

// serve starts the gRPC server of the device plugin.
func (ps *PluginServer) serve(socketWatcher *common.FileWatch) error {
	netListener, err := createNetListener(socketWatcher, ps.deviceType)
	if err != nil {
		return err
	}
	ps.grpcServer = grpc.NewServer()
	v1beta1.RegisterDevicePluginServer(ps.grpcServer, ps)
	go func() {
		if err := ps.grpcServer.Serve(netListener); err != nil {
			hwlog.RunLog.Errorf("GRPC server for '%s' crashed with error: %#v", ps.deviceType, err)
		}
	}()

	// Wait for grpcServer
	for len(ps.grpcServer.GetServiceInfo()) <= 0 {
		time.Sleep(time.Second)
	}
	hwlog.RunLog.Infof("device plugin (%s) start serving.", ps.deviceType)

	return nil
}

// register function is use to register k8s devicePlugin to kubelet.
func (ps *PluginServer) register() error {
	realKubeletSockPath, ok := common.VerifyPathAndPermission(v1beta1.KubeletSocket)
	if !ok {
		return fmt.Errorf("check kubelet socket file path failed")
	}

	conn, err := grpc.Dial(realKubeletSockPath, grpc.WithInsecure(),
		grpc.WithContextDialer(
			func(ctx context.Context, addr string) (net.Conn, error) {
				if deadline, ok := ctx.Deadline(); ok {
					return net.DialTimeout("unix", addr, time.Until(deadline))
				}
				return net.DialTimeout("unix", addr, 0)
			}))

	if err != nil {
		hwlog.RunLog.Errorf("connect to kubelet failed, err: %#v", err)
		return fmt.Errorf("connect to kubelet fail: %#v", err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			hwlog.RunLog.Errorf("close kubelet connect failed, err: %#v", err)
		}
	}()

	client := v1beta1.NewRegistrationClient(conn)
	reqt := &v1beta1.RegisterRequest{
		Version:      v1beta1.Version,
		Endpoint:     fmt.Sprintf("%s.sock", ps.deviceType),
		ResourceName: common.ResourceNamePrefix + ps.deviceType,
	}
	if _, err = client.Register(context.Background(), reqt); err != nil {
		return fmt.Errorf("register to kubelet fail: %#v", err)
	}
	return nil
}

// need privilege
func createNetListener(socketWatcher *common.FileWatch, deviceType string) (net.Listener, error) {
	realSocketPath, ok := common.VerifyPathAndPermission(v1beta1.DevicePluginPath)
	if !ok {
		hwlog.RunLog.Error("socket path verify failed!")
		return nil, fmt.Errorf("socket path verify failed")
	}

	if err := socketWatcher.WatchFile(realSocketPath); err != nil {
		hwlog.RunLog.Errorf("failed to create file watcher, err: %#v", err)
		return nil, err
	}

	pluginSocketPath := path.Join(realSocketPath, fmt.Sprintf("%s.sock", deviceType))
	if _, err := os.Stat(pluginSocketPath); err == nil {
		hwlog.RunLog.Infof("Found exist sock file, sockName is: %s, now remove it.", path.Base(pluginSocketPath))
		if err = os.Remove(pluginSocketPath); err != nil {
			hwlog.RunLog.Error("failed to remove sock file")
			return nil, err
		}
	}
	netListen, err := net.Listen("unix", pluginSocketPath)
	if err != nil {
		hwlog.RunLog.Errorf("device plugin start failed, err: %#v", err)
		return nil, err
	}

	if err = os.Chmod(pluginSocketPath, common.SocketChmod); err != nil {
		hwlog.RunLog.Errorf("change file: %s mode error", path.Base(pluginSocketPath))
	}
	return netListen, err
}
