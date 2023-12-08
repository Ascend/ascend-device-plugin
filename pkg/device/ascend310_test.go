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

// Package device a series of device function
package device

import (
	"context"
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/npu-exporter/v5/common-utils/hwlog"
	"huawei.com/npu-exporter/v5/devmanager"
	"k8s.io/apimachinery/pkg/util/sets"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/kubeclient"
)

func init() {
	hwLogConfig := hwlog.LogConfig{
		OnlyToStdout: true,
	}
	hwlog.InitRunLogger(&hwLogConfig, context.Background())
}

func createFake310Manager(fdFlag, useAscendDocker bool) *HwAscend310Manager {
	common.ParamOption = common.Option{
		GetFdFlag:       fdFlag,
		UseAscendDocker: useAscendDocker,
		PresetVDevice:   true,
	}
	manager := NewHwAscend310Manager()
	manager.SetDmgr(&devmanager.DeviceManagerMock{})
	return manager
}

func TestHwAscend310ManagerGetNPUs(t *testing.T) {
	convey.Convey("test GetNPUs", t, func() {
		convey.Convey("310 get npu", func() {
			manager := createFake310Manager(false, false)
			allInfo, err := manager.GetNPUs()
			convey.So(err, convey.ShouldBeNil)
			convey.So(allInfo.AllDevTypes[0], convey.ShouldEqual, common.Ascend310)
			convey.So(allInfo.AllDevs[0].DeviceName, convey.ShouldEqual,
				fmt.Sprintf("%s-%d", common.Ascend310, allInfo.AllDevs[0].PhyID))
		})
		convey.Convey("310 get npu use fd", func() {
			manager := createFake310Manager(true, false)
			allInfo, err := manager.GetNPUs()
			convey.So(err, convey.ShouldBeNil)
			convey.So(allInfo.AllDevTypes[0], convey.ShouldEqual, common.AscendfdPrefix)
			convey.So(allInfo.AllDevs[0].DeviceName, convey.ShouldEqual,
				fmt.Sprintf("%s-%d", common.AscendfdPrefix, allInfo.AllDevs[0].PhyID))
		})
	})
}

func TestDoWithVolcanoListAndWatch310(t *testing.T) {
	convey.Convey("310 test DoWithVolcanoListAndWatch", t, func() {
		manager := createFake310Manager(false, true)
		fakeKubeInteractor := &kubeclient.ClientK8s{Clientset: nil, NodeName: "NODE_NAME"}
		manager.SetKubeClient(fakeKubeInteractor)
		allInfo, err := manager.GetNPUs()
		convey.So(err, convey.ShouldBeNil)
		groupDevice := ClassifyDevices(allInfo.AllDevs, allInfo.AllDevTypes)

		mockGetPodsUsedNpu := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)),
			"GetPodsUsedNpu", func(_ *kubeclient.ClientK8s) sets.String {
				return nil
			})
		mockGetConfigMap := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)),
			"GetDeviceInfoCMCache", func(_ *kubeclient.ClientK8s) *common.NodeDeviceInfoCache {
				nodeDeviceData := common.NodeDeviceInfoCache{
					DeviceInfo: common.NodeDeviceInfo{
						DeviceList: map[string]string{common.Ascend310: "Ascend310-1"},
						UpdateTime: time.Now().Unix(),
					},
				}
				nodeDeviceData.CheckCode = common.MakeDataHash(nodeDeviceData.DeviceInfo)

				return &nodeDeviceData
			})
		mockCreateConfigMap := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)),
			"WriteDeviceInfoDataIntoCM", func(_ *kubeclient.ClientK8s,
				deviceInfo map[string]string, manuallySeparateNPU string) (*common.NodeDeviceInfoCache, error) {
				return &common.NodeDeviceInfoCache{}, nil
			})
		defer func() {
			mockGetPodsUsedNpu.Reset()
			mockGetConfigMap.Reset()
			mockCreateConfigMap.Reset()
		}()
		manager.client.SetNodeDeviceInfoCache(createFakeDeviceInfo())
		manager.DoWithVolcanoListAndWatch(groupDevice)
	})
}
