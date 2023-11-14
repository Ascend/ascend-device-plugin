/* Copyright(C) 2022. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Package server holds the implementation of registration to kubelet, k8s device plugin interface and grpc service.
package server

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc"
	"k8s.io/kubelet/pkg/apis/podresources/v1alpha1"
	"k8s.io/kubernetes/pkg/kubelet/apis/podresources"

	"Ascend-device-plugin/pkg/common"
)

const (
	sockMode = 0755
)

func init() {
	if _, err := os.Stat(socketPath); err == nil {
		return
	}
	if _, err := os.OpenFile(socketPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, sockMode); err != nil {
		fmt.Errorf("err: %#v", err)
		return
	}
	if err := os.Chmod(socketPath, os.ModeSocket); err != nil {
		fmt.Errorf("err: %#v", err)
		return
	}
}

// TestPodResourceStart1 for test the interface Start part 2
func TestPodResourceStart1(t *testing.T) {
	pr := NewPodResource()
	convey.Convey("test start", t, func() {
		convey.Convey("VerifyPath failed", func() {
			mockVerifyPath := gomonkey.ApplyFunc(common.VerifyPathAndPermission, func(verifyPath string,
				waitSecond int) (string, bool) {
				return "", false
			})
			defer mockVerifyPath.Reset()
			convey.So(pr.start(), convey.ShouldNotBeNil)
		})
		convey.Convey("VerifyPath ok", func() {
			mockVerifyPath := gomonkey.ApplyFunc(common.VerifyPathAndPermission, func(verifyPath string,
				waitSecond int) (string, bool) {
				return "", true
			})
			defer mockVerifyPath.Reset()
			err := pr.start()
			convey.So(err, convey.ShouldBeNil)
		})
	})
}

// TestPodResourceStart2 for test the interface Start part 2
func TestPodResourceStart2(t *testing.T) {
	pr := NewPodResource()
	convey.Convey("test start", t, func() {
		convey.Convey("GetClient failed", func() {
			mockGetClient := gomonkey.ApplyFunc(podresources.GetV1alpha1Client, func(socket string,
				connectionTimeout time.Duration, maxMsgSize int) (v1alpha1.PodResourcesListerClient,
				*grpc.ClientConn, error) {
				return nil, nil, fmt.Errorf("err")
			})
			defer mockGetClient.Reset()
			convey.So(pr.start(), convey.ShouldNotBeNil)
		})
		convey.Convey("start ok", func() {
			mockGetClient := gomonkey.ApplyFunc(podresources.GetV1alpha1Client, func(socket string,
				connectionTimeout time.Duration, maxMsgSize int) (v1alpha1.PodResourcesListerClient,
				*grpc.ClientConn, error) {
				return nil, nil, nil
			})
			defer mockGetClient.Reset()
			funcStub := gomonkey.ApplyFunc(common.VerifyPathAndPermission,
				func(verifyPathAndPermission string, waitSecond int) (string, bool) {
					return verifyPathAndPermission, true
				})
			defer funcStub.Reset()
			convey.So(pr.start(), convey.ShouldBeNil)
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
			pr.stop()
			convey.So(pr.conn, convey.ShouldBeNil)
		})
		convey.Convey("close ok", func() {
			pr := &PodResource{conn: &grpc.ClientConn{}}
			mockClose := gomonkey.ApplyMethod(reflect.TypeOf(new(grpc.ClientConn)), "Close",
				func(_ *grpc.ClientConn) error { return nil })
			defer mockClose.Reset()
			pr.stop()
			convey.So(pr.conn, convey.ShouldBeNil)
		})
	})
}

type FakeClient struct{}

// List is to get pod resource
func (c *FakeClient) List(ctx context.Context, in *v1alpha1.ListPodResourcesRequest,
	opts ...grpc.CallOption) (*v1alpha1.ListPodResourcesResponse, error) {
	out := new(v1alpha1.ListPodResourcesResponse)
	return out, nil
}
