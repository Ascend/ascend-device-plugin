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
	"huawei.com/npu-exporter/v5/devmanager"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	"Ascend-device-plugin/pkg/common"
	"Ascend-device-plugin/pkg/kubeclient"
)

func createFake910Manager() *HwAscend910Manager {
	manager := NewHwAscend910Manager()
	manager.SetDmgr(&devmanager.DeviceManagerMock{})
	return manager
}

func TestHwAscend910ManagerGetNPUs(t *testing.T) {
	convey.Convey("910 test GetNPUs", t, func() {
		manager := createFake910Manager()
		allInfo, err := manager.GetNPUs()
		convey.So(err, convey.ShouldBeNil)
		convey.So(allInfo.AllDevTypes[0], convey.ShouldEqual, common.Ascend910)
		convey.So(allInfo.AllDevs[0].DeviceName, convey.ShouldEqual,
			fmt.Sprintf("%s-%d", common.Ascend910, allInfo.AllDevs[0].PhyID))
	})
}

func TestDoWithVolcanoListAndWatch910(t *testing.T) {
	convey.Convey("910 test DoWithVolcanoListAndWatch", t, func() {
		manager := createFake910Manager()
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
			"GetDeviceInfoCMCache", func(_ *kubeclient.ClientK8s) *common.NodeDeviceInfoCache {
				nodeDeviceData := common.NodeDeviceInfoCache{DeviceInfo: common.NodeDeviceInfo{
					DeviceList: map[string]string{common.Ascend910: "Ascend910-1"},
					UpdateTime: time.Now().Unix()}}
				nodeDeviceData.CheckCode = common.MakeDataHash(nodeDeviceData.DeviceInfo)
				return &nodeDeviceData
			})
		mockPatchNodeState := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)),
			"PatchNodeState", func(_ *kubeclient.ClientK8s, curNode,
				newNode *v1.Node) (*v1.Node, []byte, error) {
				return &v1.Node{}, nil, nil
			})
		mockCreateConfigMap := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)),
			"WriteDeviceInfoDataIntoCMCache", func(_ *kubeclient.ClientK8s,
				deviceInfo map[string]string) error {
				return nil
			})
		mockNodeBack := gomonkey.ApplyMethod(reflect.TypeOf(new(kubeclient.ClientK8s)), "GetNode",
			func(_ *kubeclient.ClientK8s) (*v1.Node, error) {
				curNode := &v1.Node{}
				curNode.Labels = make(map[string]string, 1)
				return curNode, nil
			})
		defer func() {
			mockGetPodsUsedNpu.Reset()
			mockGetConfigMap.Reset()
			mockPatchNodeState.Reset()
			mockCreateConfigMap.Reset()
			mockNodeBack.Reset()
		}()

		manager.DoWithVolcanoListAndWatch(groupDevice)

	})
}

func TestToStandardDeviceFmt(t *testing.T) {
	convey.Convey("910 test toStandardDeviceFmt", t, func() {
		hnm := NewHwAscend910Manager()
		devices := sets.String{}.Insert("test910")
		res := hnm.toStandardDeviceFmt(devices)
		convey.So(len(res), convey.ShouldEqual, 1)
	})
}

func TestGetPatchLabel(t *testing.T) {
	convey.Convey("910 getPatchLabel", t, func() {
		hnm := NewHwAscend910Manager()
		devices := sets.String{}.Insert("100-1")
		devices.Insert("100-2")
		res := hnm.getPatchLabel(devices)
		convey.So(res, convey.ShouldBeIn, []string{"1.2", "2.1"})
	})
}
