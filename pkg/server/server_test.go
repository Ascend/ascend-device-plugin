// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package server holds the implementation of registration to kubelet, k8s device plugin interface and grpc service.
package server

import (
	"errors"
	"net"
	"os"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"Ascend-device-plugin/pkg/common"
)

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

		socketWatcher, err := common.NewFileWatch()
		convey.So(err, convey.ShouldBeNil)
		err = ps.Start(socketWatcher)
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

		socketWatcher, err := common.NewFileWatch()
		convey.So(err, convey.ShouldBeNil)
		err = ps.Start(socketWatcher)
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

		socketWatcher, err := common.NewFileWatch()
		convey.So(err, convey.ShouldBeNil)
		err = ps.Start(socketWatcher)
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

		socketWatcher, err := common.NewFileWatch()
		convey.So(err, convey.ShouldBeNil)
		err = ps.Start(socketWatcher)
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

		socketWatcher, err := common.NewFileWatch()
		convey.So(err, convey.ShouldBeNil)
		err = ps.Start(socketWatcher)
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

		socketWatcher, err := common.NewFileWatch()
		convey.So(err, convey.ShouldBeNil)
		err = ps.Start(socketWatcher)
		convey.So(err, convey.ShouldNotBeNil)
	})
}
