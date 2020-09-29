/*
* Copyright(C) 2020. Huawei Technologies Co.,Ltd. All rights reserved.
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You may obtain a copy of the License at
*
* http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
 */

package huawei

import (
	mock_v1beta1 "Ascend-device-plugin/src/plugin/pkg/npu/huawei/mock_kubelet_v1beta1"
	"Ascend-device-plugin/src/plugin/pkg/npu/huawei/mock_kubernetes"
	"Ascend-device-plugin/src/plugin/pkg/npu/huawei/mock_v1"
	"fmt"
	"github.com/golang/mock/gomock"
	"go.uber.org/atomic"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"path"
	"testing"
	"time"
)

const (
	annonationTest1 = "{\"pod_name\":\"pod_name\",\"server_id\":\"0.0.0.0\",\"devices\":[{\"device_id\":\"0\"," +
		"\"device_ip\":\"0.0.0.0\"},{\"device_id\":\"1\",\"device_ip\":\"0.0.0.1\"}]}"
	runTime2      = 2
	runTime3      = 3
	runTime5      = 5
	sleepTestFour = 4
)

// TestPluginAPI_ListAndWatch for listAndWatch
func TestPluginAPI_ListAndWatch(t *testing.T) {
	hdm := createFakeDevManager("ascend910")
	hdm.SetParameters(false, false, true)
	if err := hdm.GetNPUs(); err != nil {
		t.Fatal(err)
	}
	devTypes := hdm.GetDevType()
	if len(devTypes) == 0 {
		t.Fatal("TestPluginAPI_ListAndWatch Run Failed")
	}
	var fakePluginAPI *pluginAPI
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	node1 := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{Annotations: make(map[string]string)},
	}
	podList := &v1.PodList{}
	pod1 := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{Annotations: make(map[string]string)},
	}
	podList.Items = append(podList.Items, pod1)
	node1.Annotations[huaweiAscend910] = "Ascend910-1,Ascend910-2"
	pod1.Annotations[huaweiAscend910] = "Ascend910-1"
	mockK8s := mock_kubernetes.NewMockInterface(ctrl)
	mockV1 := mock_v1.NewMockCoreV1Interface(ctrl)
	mockNode := mock_v1.NewMockNodeInterface(ctrl)
	mockPod := mock_v1.NewMockPodInterface(ctrl)
	mockNode.EXPECT().Get(gomock.Any(), metav1.GetOptions{}).Return(node1, nil).Times(runTime2)
	mockNode.EXPECT().Patch(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(node1, nil)
	mockPod.EXPECT().List(gomock.Any()).Return(podList, nil).Times(runTime2)
	mockV1.EXPECT().Pods(gomock.Any()).Return(mockPod).Times(runTime2)
	mockV1.EXPECT().Nodes().Return(mockNode).Times(runTime3)
	mockK8s.EXPECT().CoreV1().Return(mockV1).Times(runTime5)
	mockstream := mock_v1beta1.NewMockDevicePlugin_ListAndWatchServer(ctrl)
	mockstream.EXPECT().Send(&pluginapi.ListAndWatchResponse{}).Return(nil)
	fakeKubeInteractor := &KubeInteractor{
		clientset: mockK8s,
		nodeName:  "NODE_NAME",
	}

	for _, devType := range devTypes {
		pluginSocket := fmt.Sprintf("%s.sock", devType)
		pluginSockPath := path.Join(pluginapi.DevicePluginPath, pluginSocket)
		fakePluginAPI = createFakePluginAPI(hdm, devType, pluginSockPath, fakeKubeInteractor)
	}
	go changeBreakFlag(fakePluginAPI)
	err := fakePluginAPI.ListAndWatch(&pluginapi.Empty{}, mockstream)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("TestPluginAPI_ListAndWatch Run Pass")
}

func changeBreakFlag(api *pluginAPI) {
	time.Sleep(sleepTestFour * time.Second)
	api.outbreak.Store(true)
}

func createFakePluginAPI(hdm *HwDevManager, devType string, socket string, ki *KubeInteractor) *pluginAPI {
	return &pluginAPI{hps: &HwPluginServe{
		devType:        devType,
		hdm:            hdm,
		runMode:        hdm.runMode,
		devices:        make(map[string]*npuDevice),
		socket:         socket,
		kubeInteractor: ki,
		healthDevice:   sets.String{},
	},
		outbreak: atomic.NewBool(false),
	}
}

// TestAddAnnotation for test AddAnnotation
func TestAddAnnotation(t *testing.T) {
	hdm := createFakeDevManager("ascend910")
	hdm.SetParameters(false, false, true)
	if err := hdm.GetNPUs(); err != nil {
		t.Fatal(err)
	}
	devTypes := hdm.GetDevType()
	if len(devTypes) == 0 {
		t.Fatal("TestPluginAPI_ListAndWatch Run Failed")
	}
	var fakePluginAPI *pluginAPI
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockK8s := mock_kubernetes.NewMockInterface(ctrl)
	fakeKubeInteractor := &KubeInteractor{
		clientset: mockK8s,
		nodeName:  "NODE_NAME",
	}
	for _, devType := range devTypes {
		pluginSocket := fmt.Sprintf("%s.sock", devType)
		pluginSockPath := path.Join(pluginapi.DevicePluginPath, pluginSocket)
		fakePluginAPI = createFakePluginAPI(hdm, devType, pluginSockPath, fakeKubeInteractor)
	}
	annonationString := fakePluginAPI.addAnnotation("0,1", "pod_name", "0.0.0.0")
	if annonationString == annonationTest1 {
		t.Logf("TestAddAnnotation Run Pass")
	} else {
		t.Fatal("TestAddAnnotation Run Failed")
	}
}
