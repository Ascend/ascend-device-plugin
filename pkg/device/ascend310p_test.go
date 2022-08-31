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
	"huawei.com/npu-exporter/devmanager"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/kubeclient"
)

func createFake310pManager() *HwDevManager {
	hdm := &HwDevManager{}
	hdm.manager = NewHwAscend310PManager()
	hdm.manager.SetDmgr(&devmanager.DeviceManagerMock{})
	return hdm
}

func TestHwAscend310PManagerGetNPUs(t *testing.T) {
	convey.Convey("310p get npu", t, func() {
		hdm := createFake310pManager()
		err := hdm.manager.GetNPUs(&hdm.AllDevs, &hdm.AllDevTypes)
		convey.So(err, convey.ShouldBeNil)
		convey.So(hdm.AllDevTypes[0], convey.ShouldEqual, common.Ascend310P)
		convey.So(hdm.AllDevs[0].DeviceName, convey.ShouldEqual,
			fmt.Sprintf("%s-%d", common.Ascend310P, hdm.AllDevs[0].PhyID))
	})
}

func TestDoWithVolcanoListAndWatch310p(t *testing.T) {
	convey.Convey("310p DoWithVolcanoListAndWatch", t, func() {
		hdm := createFake310pManager()
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
						DeviceList: map[string]string{common.Ascend310P: "Ascend310p-1"},
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