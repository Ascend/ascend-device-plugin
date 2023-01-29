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
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/npu-exporter/v3/devmanager"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/kubeclient"
)

func createFake310pManager() *HwAscend310PManager {
	manager := NewHwAscend310PManager()
	manager.SetDmgr(&devmanager.DeviceManagerMock{})
	return manager
}

func TestHwAscend310PManagerGetNPUs(t *testing.T) {
	convey.Convey("310p get npu", t, func() {
		manager := createFake310pManager()
		allInfo, err := manager.GetNPUs()
		convey.So(err, convey.ShouldBeNil)
		convey.So(allInfo.AllDevTypes[0], convey.ShouldEqual, common.Ascend310P)
		convey.So(allInfo.AllDevs[0].DeviceName, convey.ShouldEqual,
			fmt.Sprintf("%s-%d", common.Ascend310P, allInfo.AllDevs[0].PhyID))
	})
}

func TestDoWithVolcanoListAndWatch310p(t *testing.T) {
	convey.Convey("310p DoWithVolcanoListAndWatch", t, func() {
		manager := createFake310pManager()
		fakeKubeInteractor := &kubeclient.ClientK8s{Clientset: nil, NodeName: "NODE_NAME"}
		manager.SetKubeClient(fakeKubeInteractor)
		allInfo, err := manager.GetNPUs()
		convey.So(err, convey.ShouldBeNil)
		groupDevice := ClassifyDevices(allInfo.AllDevs, allInfo.AllDevTypes)

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

		manager.DoWithVolcanoListAndWatch(groupDevice)
	})
}
