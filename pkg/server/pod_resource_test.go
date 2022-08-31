// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package server holds the implementation of registration to kubelet, k8s device plugin interface and grpc service.
package server

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc"
	"k8s.io/kubernetes/pkg/kubelet/apis/podresources"
	"k8s.io/kubernetes/pkg/kubelet/apis/podresources/v1alpha1"

	"Ascend-device-plugin/pkg/common"
)

// TestPodResourceStart1 for test the interface Start part 2
func TestPodResourceStart1(t *testing.T) {
	convey.Convey("invalid interface receiver", t, func() {
		var pr *PodResource
		convey.So(pr.Start(nil), convey.ShouldNotBeNil)
	})
	pr := NewPodResource()
	socketWatcher, err := common.NewFileWatch()
	if err != nil {
		t.Fatal(err)
	}
	convey.Convey("test start", t, func() {
		convey.Convey("VerifyPath failed", func() {
			mockVerifyPath := gomonkey.ApplyFunc(common.VerifyPathAndPermission, func(verifyPath string) (string,
				bool) {
				return "", false
			})
			defer mockVerifyPath.Reset()
			convey.So(pr.Start(socketWatcher), convey.ShouldNotBeNil)
		})
		mockVerifyPath := gomonkey.ApplyFunc(common.VerifyPathAndPermission, func(verifyPath string) (string,
			bool) {
			return "", true
		})
		defer mockVerifyPath.Reset()
		convey.Convey("WatchFile failed", func() {
			mockWatchFile := gomonkey.ApplyMethod(reflect.TypeOf(new(common.FileWatch)), "WatchFile",
				func(_ *common.FileWatch, fileName string) error { return fmt.Errorf("err") })
			defer mockWatchFile.Reset()
			err := pr.Start(socketWatcher)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}

// TestPodResourceStart2 for test the interface Start part 2
func TestPodResourceStart2(t *testing.T) {
	pr := NewPodResource()
	socketWatcher, err := common.NewFileWatch()
	if err != nil {
		t.Fatal(err)
	}
	convey.Convey("test start", t, func() {
		mockWatchFile := gomonkey.ApplyMethod(reflect.TypeOf(new(common.FileWatch)), "WatchFile",
			func(_ *common.FileWatch, fileName string) error { return nil })
		defer mockWatchFile.Reset()
		convey.Convey("GetClient failed", func() {
			mockGetClient := gomonkey.ApplyFunc(podresources.GetClient, func(socket string,
				connectionTimeout time.Duration, maxMsgSize int) (v1alpha1.PodResourcesListerClient,
				*grpc.ClientConn, error) {
				return nil, nil, fmt.Errorf("err")
			})
			defer mockGetClient.Reset()
			convey.So(pr.Start(socketWatcher), convey.ShouldNotBeNil)
		})
		convey.Convey("start ok", func() {
			mockGetClient := gomonkey.ApplyFunc(podresources.GetClient, func(socket string,
				connectionTimeout time.Duration, maxMsgSize int) (v1alpha1.PodResourcesListerClient,
				*grpc.ClientConn, error) {
				return nil, nil, nil
			})
			defer mockGetClient.Reset()
			funcStub := gomonkey.ApplyFunc(common.VerifyPathAndPermission,
				func(verifyPathAndPermission string) (string, bool) { return verifyPathAndPermission, true })
			defer funcStub.Reset()
			convey.So(pr.Start(socketWatcher), convey.ShouldBeNil)
		})
	})
}

// TestPodResourceStart for test the interface Stop
func TestPodResourceStop(t *testing.T) {
	convey.Convey("test start", t, func() {
		convey.Convey("close failed", func() {
			pr := &PodResource{conn: &grpc.ClientConn{}}
			mockClose := gomonkey.ApplyMethod(reflect.TypeOf(new(grpc.ClientConn)), "Close",
				func(_ *grpc.ClientConn) error { return fmt.Errorf("err") })
			defer mockClose.Reset()
			pr.Stop()
			convey.So(pr.conn, convey.ShouldBeNil)
		})
		convey.Convey("close ok", func() {
			pr := &PodResource{conn: &grpc.ClientConn{}}
			mockClose := gomonkey.ApplyMethod(reflect.TypeOf(new(grpc.ClientConn)), "Close",
				func(_ *grpc.ClientConn) error { return nil })
			defer mockClose.Reset()
			pr.Stop()
			convey.So(pr.conn, convey.ShouldBeNil)
		})
	})
}

// TestPodResourceGetRestartFlag for test the interface GetRestartFlag
func TestPodResourceGetRestartFlag(t *testing.T) {
	convey.Convey("test GetRestartFlag", t, func() {
		pr := &PodResource{conn: &grpc.ClientConn{}, restart: false}
		convey.So(pr.GetRestartFlag(), convey.ShouldNotBeNil)
	})
}

// TestPodResourceSetRestartFlag for test the interface SetRestartFlag
func TestPodResourceSetRestartFlag(t *testing.T) {
	convey.Convey("test GetRestartFlag", t, func() {
		pr := &PodResource{conn: &grpc.ClientConn{}, restart: false}
		pr.SetRestartFlag(true)
		convey.So(pr.restart, convey.ShouldBeTrue)
	})
}

// TestPodResourceGetPodResource1 for test the interface GetPodResource part 1
func TestPodResourceGetPodResource1(t *testing.T) {
	pr := &PodResource{client: &FakeClient{}, restart: false}
	convey.Convey("conn is nil", t, func() {
		_, err := pr.GetPodResource()
		convey.So(err, convey.ShouldNotBeNil)
	})
	pr.conn = &grpc.ClientConn{}
	podResourceResponse := v1alpha1.ListPodResourcesResponse{}
	convey.Convey("podResourceList failed", t, func() {
		mockList := gomonkey.ApplyMethod(reflect.TypeOf(new(FakeClient)), "List",
			func(_ *FakeClient, ctx context.Context, in *v1alpha1.ListPodResourcesRequest,
				opts ...grpc.CallOption) (*v1alpha1.ListPodResourcesResponse, error) {
				return &podResourceResponse, fmt.Errorf("error")
			})
		defer mockList.Reset()
		_, err := pr.GetPodResource()
		convey.So(err, convey.ShouldNotBeNil)
	})
	mockList := gomonkey.ApplyMethod(reflect.TypeOf(new(FakeClient)), "List",
		func(_ *FakeClient, ctx context.Context, in *v1alpha1.ListPodResourcesRequest,
			opts ...grpc.CallOption) (*v1alpha1.ListPodResourcesResponse,
			error) {
			return &podResourceResponse, nil
		})
	defer mockList.Reset()
	convey.Convey("the number of pods exceeds the upper limit", t, func() {
		podResourceResponse.PodResources = make([]*v1alpha1.PodResources, common.MaxPodLimit+1)
		_, err := pr.GetPodResource()
		convey.So(err, convey.ShouldNotBeNil)
	})
	convey.Convey("pod name syntax illegal", t, func() {
		podResourceResponse.PodResources = []*v1alpha1.PodResources{{Name: "invalid_name",
			Containers: make([]*v1alpha1.ContainerResources, common.MaxContainerLimit+1)}}
		_, err := pr.GetPodResource()
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("pod name syntax illegal", t, func() {
		podResourceResponse.PodResources = []*v1alpha1.PodResources{{Name: "pod-name", Namespace: "invalid_namespace",
			Containers: make([]*v1alpha1.ContainerResources, common.MaxContainerLimit+1)}}
		_, err := pr.GetPodResource()
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("the number of containers exceeds the upper limit", t, func() {
		podResourceResponse.PodResources = []*v1alpha1.PodResources{{Name: "pod-name", Namespace: "pod-namespace",
			Containers: make([]*v1alpha1.ContainerResources, common.MaxContainerLimit+1)}}
		_, err := pr.GetPodResource()
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestPodResourceGetPodResource2 for test the interface GetPodResource part 2
func TestPodResourceGetPodResource2(t *testing.T) {
	pr := &PodResource{conn: &grpc.ClientConn{}, client: &FakeClient{}, restart: false}
	podResourceResponse := v1alpha1.ListPodResourcesResponse{}
	mockList := gomonkey.ApplyMethod(reflect.TypeOf(new(FakeClient)), "List",
		func(_ *FakeClient, ctx context.Context, in *v1alpha1.ListPodResourcesRequest,
			opts ...grpc.CallOption) (*v1alpha1.ListPodResourcesResponse, error) {
			return &podResourceResponse, nil
		})
	defer mockList.Reset()
	convey.Convey("the number of containers device type exceeds the upper limit", t, func() {
		podResourceResponse.PodResources = []*v1alpha1.PodResources{{Name: "pod-name", Namespace: "pod-namespace",
			Containers: []*v1alpha1.ContainerResources{{Devices: make([]*v1alpha1.ContainerDevices,
				common.MaxDevicesNum+1)}}}}
		_, err := pr.GetPodResource()
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("containerDevice is nil", t, func() {
		podResourceResponse.PodResources = []*v1alpha1.PodResources{{Name: "pod-name", Namespace: "pod-namespace",
			Containers: []*v1alpha1.ContainerResources{{Devices: []*v1alpha1.ContainerDevices{nil}}}}}
		_, err := pr.GetPodResource()
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("not huawei resource", t, func() {
		podResourceResponse.PodResources = []*v1alpha1.PodResources{{Name: "pod-name", Namespace: "pod-namespace",
			Containers: []*v1alpha1.ContainerResources{{Devices: []*v1alpha1.ContainerDevices{{ResourceName: ""}}}}}}
		_, err := pr.GetPodResource()
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("the number of container device exceeds the upper limit", t, func() {
		podResourceResponse.PodResources = []*v1alpha1.PodResources{{Name: "pod-name", Namespace: "pod-namespace",
			Containers: []*v1alpha1.ContainerResources{{Devices: []*v1alpha1.ContainerDevices{{
				ResourceName: common.ResourceNamePrefix + common.Ascend910,
				DeviceIds:    make([]string, common.MaxDevicesNum+1)}}}}}}
		_, err := pr.GetPodResource()
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("length of device name is invalid", t, func() {
		podResourceResponse.PodResources = []*v1alpha1.PodResources{{Name: "pod-name", Namespace: "pod-namespace",
			Containers: []*v1alpha1.ContainerResources{{Devices: []*v1alpha1.ContainerDevices{{
				ResourceName: common.ResourceNamePrefix + common.Ascend910,
				DeviceIds:    []string{string(make([]byte, common.MaxDeviceNameLen+1))}}}}}}}
		_, err := pr.GetPodResource()
		convey.So(err, convey.ShouldBeNil)
	})
}

// TestPodResourceGetPodResource3 for test the interface GetPodResource part 3
func TestPodResourceGetPodResource3(t *testing.T) {
	pr := &PodResource{conn: &grpc.ClientConn{}, client: &FakeClient{}, restart: false}
	podResourceResponse := v1alpha1.ListPodResourcesResponse{}
	mockList := gomonkey.ApplyMethod(reflect.TypeOf(new(FakeClient)), "List",
		func(_ *FakeClient, ctx context.Context, in *v1alpha1.ListPodResourcesRequest,
			opts ...grpc.CallOption) (*v1alpha1.ListPodResourcesResponse, error) {
			return &podResourceResponse, nil
		})
	defer mockList.Reset()
	convey.Convey("get valid pod resource", t, func() {
		podResourceResponse.PodResources = []*v1alpha1.PodResources{{Name: "pod-name", Namespace: "pod-namespace",
			Containers: []*v1alpha1.ContainerResources{{Devices: []*v1alpha1.ContainerDevices{{
				ResourceName: common.ResourceNamePrefix + common.Ascend910,
				DeviceIds:    []string{common.Ascend910 + "-0"}}}}}}}
		_, err := pr.GetPodResource()
		convey.So(err, convey.ShouldBeNil)
	})
	convey.Convey("multi resource", t, func() {
		podResourceResponse.PodResources = []*v1alpha1.PodResources{{Name: "pod-name", Namespace: "pod-namespace",
			Containers: []*v1alpha1.ContainerResources{{Devices: []*v1alpha1.ContainerDevices{{
				ResourceName: common.ResourceNamePrefix + common.Ascend910,
				DeviceIds:    []string{common.Ascend910 + "-0"}}}},
				{Devices: []*v1alpha1.ContainerDevices{{
					ResourceName: common.ResourceNamePrefix + common.Ascend310,
					DeviceIds:    []string{common.Ascend310 + "-0"}}}}}}}
		_, err := pr.GetPodResource()
		convey.So(err, convey.ShouldBeNil)
	})
}

type FakeClient struct{}

// List is to get pod resource
func (c *FakeClient) List(ctx context.Context, in *v1alpha1.ListPodResourcesRequest,
	opts ...grpc.CallOption) (*v1alpha1.ListPodResourcesResponse, error) {
	out := new(v1alpha1.ListPodResourcesResponse)
	return out, nil
}
