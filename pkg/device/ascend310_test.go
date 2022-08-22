// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package device a series of device function
package device

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/npu-exporter/devmanager"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/kubeclient"
)

func init() {
	stopCh := make(chan struct{})
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, stopCh)
}

func createFake310Manager(fdFlag, useAscendDocker bool) *HwDevManager {
	hdm := &HwDevManager{}
	common.ParamOption = common.Option{
		GetFdFlag:       fdFlag,
		UseAscendDocker: useAscendDocker,
	}
	hdm.manager = NewHwAscend310Manager()
	hdm.manager.SetDmgr(&devmanager.DeviceManagerMock{})
	return hdm
}

func TestHwAscend310ManagerGetNPUs(t *testing.T) {
	convey.Convey("test GetNPUs", t, func() {
		convey.Convey("310 get npu", func() {
			hdm := createFake310Manager(false, false)
			err := hdm.manager.GetNPUs(&hdm.AllDevs, &hdm.AllDevTypes)
			convey.So(err, convey.ShouldBeNil)
			convey.So(hdm.AllDevTypes[0], convey.ShouldEqual, common.Ascend310)
			convey.So(hdm.AllDevs[0].DeviceName, convey.ShouldEqual,
				fmt.Sprintf("%s-%d", common.Ascend310, hdm.AllDevs[0].PhyID))
		})
		convey.Convey("310 get npu use fd", func() {
			hdm := createFake310Manager(true, false)
			err := hdm.manager.GetNPUs(&hdm.AllDevs, &hdm.AllDevTypes)
			convey.So(err, convey.ShouldBeNil)
			convey.So(hdm.AllDevTypes[0], convey.ShouldEqual, common.AscendfdPrefix)
			convey.So(hdm.AllDevs[0].DeviceName, convey.ShouldEqual,
				fmt.Sprintf("%s-%d", common.AscendfdPrefix, hdm.AllDevs[0].PhyID))
		})
	})
}

func TestDoWithVolcanoListAndWatch310(t *testing.T) {
	convey.Convey("310 test DoWithVolcanoListAndWatch", t, func() {
		hdm := createFake310Manager(false, true)
		fakeKubeInteractor := &kubeclient.ClientK8s{Clientset: nil, NodeName: "NODE_NAME"}
		hdm.manager.SetKubeClient(fakeKubeInteractor)
		err := hdm.manager.GetNPUs(&hdm.AllDevs, &hdm.AllDevTypes)
		convey.So(err, convey.ShouldBeNil)
		hdm.groupDevice = ClassifyDevices(hdm.AllDevs, hdm.AllDevTypes)

		mockGetPodsUsedNpu := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)),
			"GetPodsUsedNpu", func(_ *kubeclient.ClientK8s, devType string) sets.String {
				return nil
			})
		mockGetConfigMap := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)),
			"GetConfigMap", func(_ *kubeclient.ClientK8s) (*v1.ConfigMap, error) {
				nodeDeviceData := common.NodeDeviceInfoCache{
					DeviceInfo: common.NodeDeviceInfo{
						DeviceList: map[string]string{common.Ascend310: "Ascend310-1"},
						UpdateTime: time.Now().Unix(),
					},
				}
				nodeDeviceData.CheckCode = common.MakeDataHash(nodeDeviceData.DeviceInfo)
				data := common.MarshalData(nodeDeviceData)

				return &v1.ConfigMap{Data: map[string]string{
					common.DeviceInfoCMDataKey: string(data)},
				}, nil
			})
		mockCreateConfigMap := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)),
			"WriteDeviceInfoDataIntoCM", func(_ *kubeclient.ClientK8s,
				deviceInfo map[string]string) (*v1.ConfigMap, error) {
				return &v1.ConfigMap{}, nil
			})
		defer func() {
			mockGetPodsUsedNpu.Reset()
			mockGetConfigMap.Reset()
			mockCreateConfigMap.Reset()
		}()

		hdm.manager.DoWithVolcanoListAndWatch(hdm.groupDevice)
	})
}
