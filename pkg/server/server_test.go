// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package server holds the implementation of registration to kubelet, k8s device plugin interface and grpc service.
package server

import (
	"context"
	"errors"
	"net"
	"os"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc"
	"huawei.com/npu-exporter/hwlog"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"Ascend-device-plugin/pkg/common"
)

func init() {
	stopCh := make(chan struct{})
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, stopCh)
}

// TestPluginServerGetRestartFlag Test PluginServer GetRestartFlag()
func TestPluginServerGetRestartFlag(t *testing.T) {
	convey.Convey("test GetRestartFlag", t, func() {
		ps := &PluginServer{restart: false}
		convey.So(ps.GetRestartFlag(), convey.ShouldBeFalse)
	})
}

// TestPluginServerSetRestartFlag Test PluginServer SetRestartFlag()
func TestPluginServerSetRestartFlag(t *testing.T) {
	convey.Convey("test SetRestartFlag", t, func() {
		ps := &PluginServer{restart: false}
		ps.SetRestartFlag(true)
		convey.So(ps.GetRestartFlag(), convey.ShouldBeTrue)
	})
}

// TestPluginServerStop Test PluginServer Stop()
func TestPluginServerStop(t *testing.T) {
	convey.Convey("test Stop", t, func() {
		ps := &PluginServer{
			isRunning:  common.NewAtomicBool(false),
			grpcServer: grpc.NewServer()}

		ps.Stop()
		convey.So(ps.isRunning.Load(), convey.ShouldBeFalse)
	})
}

// TestPluginServerStartPart1 Test PluginServer Start()
func TestPluginServerStartPart1(t *testing.T) {
	convey.Convey("when serve func createNetListener verify path failed", t, func() {
		funcStub := gomonkey.ApplyFunc(common.VerifyPathAndPermission, func(VerifyPathAndPermission string) (string, bool) {
			return "", false
		})
		defer funcStub.Reset()

		ps := &PluginServer{
			isRunning:  common.NewAtomicBool(false),
			grpcServer: grpc.NewServer(),
			deviceType: common.Ascend910}

		socketWatcher := common.NewFileWatch()
		err := ps.Start(socketWatcher)
		convey.So(err.Error(), convey.ShouldEqual, "socket path verify failed")
	})

	convey.Convey("when serve func createNetListener watch file failed", t, func() {
		funcStub := gomonkey.ApplyFunc(common.VerifyPathAndPermission, func(VerifyPathAndPermission string) (string, bool) {
			return VerifyPathAndPermission, true
		})
		defer funcStub.Reset()

		watchStub := gomonkey.ApplyMethod(reflect.TypeOf(new(common.FileWatch)),
			"WatchFile", func(_ *common.FileWatch, _ string) error {
				return errors.New("watch file failed")
			})
		defer watchStub.Reset()

		ps := &PluginServer{
			isRunning:  common.NewAtomicBool(false),
			grpcServer: grpc.NewServer(),
			deviceType: common.Ascend910}

		socketWatcher := common.NewFileWatch()
		err := ps.Start(socketWatcher)
		convey.So(err.Error(), convey.ShouldEqual, "watch file failed")
	})
}

// TestPluginServerStartPart2 Test PluginServer Start()
func TestPluginServerStartPart2(t *testing.T) {
	convey.Convey("when serve func createNetListener delete socket file failed", t, func() {
		funcStub := gomonkey.ApplyFunc(common.VerifyPathAndPermission, func(VerifyPathAndPermission string) (string, bool) {
			return VerifyPathAndPermission, true
		})
		defer funcStub.Reset()

		watchStub := gomonkey.ApplyMethod(reflect.TypeOf(new(common.FileWatch)),
			"WatchFile", func(_ *common.FileWatch, _ string) error {
				return nil
			})
		defer watchStub.Reset()

		statStub := gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
			return nil, nil
		})
		defer statStub.Reset()

		removeStub := gomonkey.ApplyFunc(os.Remove, func(name string) error {
			return errors.New("remove file failed")
		})
		defer removeStub.Reset()

		ps := &PluginServer{
			isRunning:  common.NewAtomicBool(false),
			grpcServer: grpc.NewServer(),
			deviceType: common.Ascend910}

		socketWatcher := common.NewFileWatch()
		err := ps.Start(socketWatcher)
		convey.So(err.Error(), convey.ShouldEqual, "remove file failed")
	})
}

// TestPluginServerStartPart3 Test PluginServer Start()
func TestPluginServerStartPart3(t *testing.T) {
	convey.Convey("when serve func createNetListener create listener failed", t, func() {
		funcStub := gomonkey.ApplyFunc(common.VerifyPathAndPermission, func(VerifyPathAndPermission string) (string, bool) {
			return VerifyPathAndPermission, true
		})
		defer funcStub.Reset()

		watchStub := gomonkey.ApplyMethod(reflect.TypeOf(new(common.FileWatch)),
			"WatchFile", func(_ *common.FileWatch, _ string) error {
				return nil
			})
		defer watchStub.Reset()

		statStub := gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
			return nil, errors.New("not exist")
		})
		defer statStub.Reset()

		listenStub := gomonkey.ApplyFunc(net.Listen, func(network, address string) (net.Listener, error) {
			return nil, errors.New("create listener failed")
		})
		defer listenStub.Reset()

		ps := &PluginServer{
			isRunning:  common.NewAtomicBool(false),
			grpcServer: grpc.NewServer(),
			deviceType: common.Ascend910}

		socketWatcher := common.NewFileWatch()
		err := ps.Start(socketWatcher)
		convey.So(err.Error(), convey.ShouldEqual, "create listener failed")
	})
}

// TestPluginServerStartPart4 Test PluginServer Start()
func TestPluginServerStartPart4(t *testing.T) {
	convey.Convey("when serve func createNetListener change file mode failed", t, func() {
		funcStub := gomonkey.ApplyFunc(common.VerifyPathAndPermission, func(VerifyPathAndPermission string) (string, bool) {
			return VerifyPathAndPermission, true
		})
		defer funcStub.Reset()

		watchStub := gomonkey.ApplyMethod(reflect.TypeOf(new(common.FileWatch)),
			"WatchFile", func(_ *common.FileWatch, _ string) error {
				return nil
			})
		defer watchStub.Reset()

		statStub := gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
			return nil, errors.New("not exist")
		})
		defer statStub.Reset()

		listenStub := gomonkey.ApplyFunc(net.Listen, func(network, address string) (net.Listener, error) {
			return nil, nil
		})
		defer listenStub.Reset()

		modStub := gomonkey.ApplyFunc(os.Chmod, func(name string, mode os.FileMode) error {
			return errors.New("change file mode failed")
		})
		defer modStub.Reset()

		ps := &PluginServer{
			isRunning:  common.NewAtomicBool(false),
			grpcServer: grpc.NewServer(),
			deviceType: common.Ascend910}

		socketWatcher := common.NewFileWatch()
		err := ps.Start(socketWatcher)
		convey.So(err.Error(), convey.ShouldEqual, "change file mode failed")
	})
}

// TestPluginServerStartPart5 Test PluginServer Start()
func TestPluginServerStartPart5(t *testing.T) {
	convey.Convey("when register check socket failed", t, func() {
		funcStub := gomonkey.ApplyFunc(common.VerifyPathAndPermission, func(VerifyPathAndPermission string) (string, bool) {
			if VerifyPathAndPermission == v1beta1.DevicePluginPath {
				return VerifyPathAndPermission, true
			}
			return "", false
		})
		defer funcStub.Reset()

		watchStub := gomonkey.ApplyMethod(reflect.TypeOf(new(common.FileWatch)),
			"WatchFile", func(_ *common.FileWatch, _ string) error {
				return nil
			})
		defer watchStub.Reset()

		statStub := gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
			return nil, errors.New("not exist")
		})
		defer statStub.Reset()

		listenStub := gomonkey.ApplyFunc(net.Listen, func(network, address string) (net.Listener, error) {
			return nil, nil
		})
		defer listenStub.Reset()

		modStub := gomonkey.ApplyFunc(os.Chmod, func(name string, mode os.FileMode) error {
			return nil
		})
		defer modStub.Reset()

		var server *grpc.Server
		grpcStub := gomonkey.ApplyMethod(reflect.TypeOf(server),
			"Serve", func(_ *grpc.Server, _ net.Listener) error {
				return nil
			})
		defer grpcStub.Reset()
		grpcStub2 := gomonkey.ApplyMethod(reflect.TypeOf(server),
			"GetServiceInfo", func(_ *grpc.Server) map[string]grpc.ServiceInfo {
				return map[string]grpc.ServiceInfo{"1": grpc.ServiceInfo{}}
			})
		defer grpcStub2.Reset()

		ps := &PluginServer{
			isRunning:  common.NewAtomicBool(false),
			grpcServer: grpc.NewServer(),
			deviceType: common.Ascend910}

		socketWatcher := common.NewFileWatch()
		err := ps.Start(socketWatcher)
		convey.So(err.Error(), convey.ShouldEqual, "check kubelet socket file path failed")
	})
}

// TestPluginServerStartPart6 Test PluginServer Start()
func TestPluginServerStartPart6(t *testing.T) {
	convey.Convey("when register grpc Dial failed", t, func() {
		funcStub := gomonkey.ApplyFunc(common.VerifyPathAndPermission, func(VerifyPathAndPermission string) (string, bool) {
			return VerifyPathAndPermission, true
		})
		defer funcStub.Reset()

		watchStub := gomonkey.ApplyMethod(reflect.TypeOf(new(common.FileWatch)),
			"WatchFile", func(_ *common.FileWatch, _ string) error {
				return nil
			})
		defer watchStub.Reset()

		statStub := gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
			return nil, errors.New("not exist")
		})
		defer statStub.Reset()

		listenStub := gomonkey.ApplyFunc(net.Listen, func(network, address string) (net.Listener, error) {
			return nil, nil
		})
		defer listenStub.Reset()

		modStub := gomonkey.ApplyFunc(os.Chmod, func(name string, mode os.FileMode) error {
			return nil
		})
		defer modStub.Reset()

		var server *grpc.Server
		grpcStub := gomonkey.ApplyMethod(reflect.TypeOf(server),
			"Serve", func(_ *grpc.Server, _ net.Listener) error {
				return nil
			})
		defer grpcStub.Reset()
		grpcStub2 := gomonkey.ApplyMethod(reflect.TypeOf(server),
			"GetServiceInfo", func(_ *grpc.Server) map[string]grpc.ServiceInfo {
				return map[string]grpc.ServiceInfo{"1": grpc.ServiceInfo{}}
			})
		defer grpcStub2.Reset()

		dialStub := gomonkey.ApplyFunc(grpc.Dial, func(_ string,
			_ ...grpc.DialOption) (*grpc.ClientConn, error) {
			return nil, errors.New("dial failed")
		})
		defer dialStub.Reset()

		ps := &PluginServer{
			isRunning:  common.NewAtomicBool(false),
			grpcServer: grpc.NewServer(),
			deviceType: common.Ascend910}

		socketWatcher := common.NewFileWatch()
		err := ps.Start(socketWatcher)
		convey.So(err.Error(), convey.ShouldContainSubstring, "connect to kubelet fail:")
	})
}

// TestPluginServerStartPart7 Test PluginServer Start()
func TestPluginServerStartPart7(t *testing.T) {
	convey.Convey("when register client register failed", t, func() {
		funcStub := gomonkey.ApplyFunc(common.VerifyPathAndPermission, func(VerifyPathAndPermission string) (string, bool) {
			return VerifyPathAndPermission, true
		})
		defer funcStub.Reset()

		watchStub := gomonkey.ApplyMethod(reflect.TypeOf(new(common.FileWatch)),
			"WatchFile", func(_ *common.FileWatch, _ string) error { return nil })
		defer watchStub.Reset()

		statStub := gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
			return nil, errors.New("not exist")
		})
		defer statStub.Reset()

		listenStub := gomonkey.ApplyFunc(net.Listen, func(network, address string) (net.Listener, error) {
			return nil, nil
		})
		defer listenStub.Reset()

		modStub := gomonkey.ApplyFunc(os.Chmod, func(name string, mode os.FileMode) error { return nil })
		defer modStub.Reset()

		var server *grpc.Server
		grpcStub := gomonkey.ApplyMethod(reflect.TypeOf(server), "Serve",
			func(_ *grpc.Server, _ net.Listener) error { return nil })
		defer grpcStub.Reset()

		grpcStub2 := gomonkey.ApplyMethod(reflect.TypeOf(server),
			"GetServiceInfo", func(_ *grpc.Server) map[string]grpc.ServiceInfo {
				return map[string]grpc.ServiceInfo{"1": grpc.ServiceInfo{}}
			})
		defer grpcStub2.Reset()

		dialStub := gomonkey.ApplyFunc(grpc.Dial, func(_ string,
			_ ...grpc.DialOption) (*grpc.ClientConn, error) {
			return &grpc.ClientConn{}, nil
		})
		defer dialStub.Reset()

		connCloseStub := gomonkey.ApplyMethod(reflect.TypeOf(new(grpc.ClientConn)),
			"Close", func(_ *grpc.ClientConn) error { return nil })
		defer connCloseStub.Reset()

		newClientStub := gomonkey.ApplyFunc(v1beta1.NewRegistrationClient,
			func(_ *grpc.ClientConn) v1beta1.RegistrationClient { return &fakeConn{} })
		defer newClientStub.Reset()

		ps := &PluginServer{
			isRunning: common.NewAtomicBool(false), grpcServer: grpc.NewServer(), deviceType: common.Ascend910}

		socketWatcher := common.NewFileWatch()
		err := ps.Start(socketWatcher)
		convey.So(err.Error(), convey.ShouldContainSubstring, "register to kubelet fail")
	})
}

// TestPluginServerStartPart8 Test PluginServer Start()
func TestPluginServerStartPart8(t *testing.T) {
	convey.Convey("when register conn close failed", t, func() {
		funcStub := gomonkey.ApplyFunc(common.VerifyPathAndPermission, func(VerifyPathAndPermission string) (string, bool) {
			return VerifyPathAndPermission, true
		})
		defer funcStub.Reset()

		watchStub := gomonkey.ApplyMethod(reflect.TypeOf(new(common.FileWatch)),
			"WatchFile", func(_ *common.FileWatch, _ string) error { return nil })
		defer watchStub.Reset()

		statStub := gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
			return nil, errors.New("not exist")
		})
		defer statStub.Reset()

		listenStub := gomonkey.ApplyFunc(net.Listen, func(network, address string) (net.Listener, error) {
			return nil, nil
		})
		defer listenStub.Reset()

		modStub := gomonkey.ApplyFunc(os.Chmod, func(name string, mode os.FileMode) error { return nil })
		defer modStub.Reset()

		var server *grpc.Server
		grpcStub := gomonkey.ApplyMethod(reflect.TypeOf(server),
			"Serve", func(_ *grpc.Server, _ net.Listener) error { return nil })
		defer grpcStub.Reset()
		grpcStub2 := gomonkey.ApplyMethod(reflect.TypeOf(server),
			"GetServiceInfo", func(_ *grpc.Server) map[string]grpc.ServiceInfo {
				return map[string]grpc.ServiceInfo{"1": grpc.ServiceInfo{}}
			})
		defer grpcStub2.Reset()

		dialStub := gomonkey.ApplyFunc(grpc.Dial, func(_ string,
			_ ...grpc.DialOption) (*grpc.ClientConn, error) {
			return &grpc.ClientConn{}, nil
		})
		defer dialStub.Reset()

		connCloseStub := gomonkey.ApplyMethod(reflect.TypeOf(new(grpc.ClientConn)),
			"Close", func(_ *grpc.ClientConn) error { return errors.New("close failed") })
		defer connCloseStub.Reset()

		newClientStub := gomonkey.ApplyFunc(v1beta1.NewRegistrationClient,
			func(_ *grpc.ClientConn) v1beta1.RegistrationClient { return &fakeConn2{} })
		defer newClientStub.Reset()

		ps := &PluginServer{
			isRunning: common.NewAtomicBool(false), grpcServer: grpc.NewServer(), deviceType: common.Ascend910}

		socketWatcher := common.NewFileWatch()
		err := ps.Start(socketWatcher)
		convey.So(err, convey.ShouldBeNil)
	})
}

type fakeConn struct {
}

// Register fake implement
func (f *fakeConn) Register(ctx context.Context,
	in *v1beta1.RegisterRequest, opts ...grpc.CallOption) (*v1beta1.Empty, error) {
	return nil, errors.New("register failed")
}

type fakeConn2 struct {
}

// Register fake implement
func (f *fakeConn2) Register(ctx context.Context,
	in *v1beta1.RegisterRequest, opts ...grpc.CallOption) (*v1beta1.Empty, error) {
	return nil, nil
}
