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
	"golang.org/x/net/context"
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
	annonationTest2 = "{\"pod_name\":\"pod_name\",\"server_id\":\"0.0.0.0\",\"devices\":[{\"device_id\":\"1\"," +
		"\"device_ip\":\"0.0.0.1\"},{\"device_id\":\"0\",\"device_ip\":\"0.0.0.0\"}]}"
	sleepTestFour = 4
)

// TestPluginAPI_ListAndWatch for listAndWatch
func TestPluginAPI_ListAndWatch(t *testing.T) {
	hdm := createFakeDevManager("ascend910")
	hdm.SetParameters(false, false, true, true, sleepTime)
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
		ObjectMeta: metav1.ObjectMeta{Annotations: make(map[string]string), Labels: make(map[string]string)},
	}
	podList := &v1.PodList{}
	pod1 := v1.Pod{
		ObjectMeta: metav1.ObjectMeta{Annotations: make(map[string]string), Labels: make(map[string]string)},
	}
	podList.Items = append(podList.Items, pod1)
	node1.Annotations[huaweiAscend910] = "Ascend910-1,Ascend910-2"
	pod1.Annotations[huaweiAscend910] = "Ascend910-1"
	mockK8s := mock_kubernetes.NewMockInterface(ctrl)
	mockV1 := mock_v1.NewMockCoreV1Interface(ctrl)
	mockNode := mock_v1.NewMockNodeInterface(ctrl)
	mockPod := mock_v1.NewMockPodInterface(ctrl)
	mockNode.EXPECT().Get(context.Background(), gomock.Any(), metav1.GetOptions{}).AnyTimes().Return(node1, nil)
	mockNode.EXPECT().Patch(context.Background(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(node1, nil)
	mockPod.EXPECT().List(context.Background(), gomock.Any()).AnyTimes().Return(podList, nil)
	mockV1.EXPECT().Pods(gomock.Any()).AnyTimes().Return(mockPod)
	mockV1.EXPECT().Nodes().AnyTimes().Return(mockNode)
	mockK8s.EXPECT().CoreV1().AnyTimes().Return(mockV1)
	fakeKubeInteractor := &KubeInteractor{
		clientset: mockK8s,
		nodeName:  "NODE_NAME",
	}
	for _, devType := range devTypes {
		mockstream := mock_v1beta1.NewMockDevicePlugin_ListAndWatchServer(ctrl)
		mockstream.EXPECT().Send(&pluginapi.ListAndWatchResponse{}).Return(nil)
		pluginSocket := fmt.Sprintf("%s.sock", devType)
		pluginSockPath := path.Join(pluginapi.DevicePluginPath, pluginSocket)
		fakePluginAPI = createFakePluginAPI(hdm, devType, pluginSockPath, fakeKubeInteractor)
		go changeBreakFlag(fakePluginAPI)
		err := fakePluginAPI.ListAndWatch(&pluginapi.Empty{}, mockstream)
		if err != nil {
			t.Fatal(err)
		}
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
		unHealthDevice: sets.String{},
	},
		outbreak: atomic.NewBool(false),
	}
}

// TestAddAnnotation for test AddAnnotation
func TestAddAnnotation(t *testing.T) {
	hdm := createFakeDevManager("ascend910")
	hdm.SetParameters(false, false, true, true, sleepTime)
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
		if IsVirtualDev(devType) {
			continue
		}
		pluginSocket := fmt.Sprintf("%s.sock", devType)
		pluginSockPath := path.Join(pluginapi.DevicePluginPath, pluginSocket)
		fakePluginAPI = createFakePluginAPI(hdm, devType, pluginSockPath, fakeKubeInteractor)
	}
	devices := make(map[string]string, 2)
	devices["0"] = "0.0.0.0"
	devices["1"] = "0.0.0.1"
	annonationString := fakePluginAPI.addAnnotation(devices, "pod_name", "0.0.0.0")
	if annonationString == annonationTest1 || annonationString == annonationTest2 {
		t.Logf("TestAddAnnotation Run Pass")
	} else {
		t.Fatal("TestAddAnnotation Run Failed")
	}
}

// TestAllocate for test Allocate
func TestAllocate(t *testing.T) {
	hdm := createFakeDevManager("ascend910")
	hdm.SetParameters(false, true, false, true, sleepTime)
	if err := hdm.GetNPUs(); err != nil {
		t.Fatal(err)
	}
	devTypes := hdm.GetDevType()
	if len(devTypes) == 0 {
		t.Fatal("TestPluginAPI_Allocate Run Failed")
	}
	devicesIDs := []string{"Ascend910-8c-1-1"}
	var containerRequests []*pluginapi.ContainerAllocateRequest
	tmp := &pluginapi.ContainerAllocateRequest{
		DevicesIDs: devicesIDs,
	}
	containerRequests = append(containerRequests, tmp)
	requests := pluginapi.AllocateRequest{
		ContainerRequests: containerRequests,
	}
	ctrl := gomock.NewController(t)
	mockK8s := mock_kubernetes.NewMockInterface(ctrl)
	mockV1 := mock_v1.NewMockCoreV1Interface(ctrl)
	mockNode := mock_v1.NewMockNodeInterface(ctrl)
	mockPod := mock_v1.NewMockPodInterface(ctrl)
	mockV1.EXPECT().Pods(gomock.Any()).AnyTimes().Return(mockPod)
	mockV1.EXPECT().Nodes().AnyTimes().Return(mockNode)
	mockK8s.EXPECT().CoreV1().AnyTimes().Return(mockV1)
	fakeKubeInteractor := &KubeInteractor{
		clientset: mockK8s,
		nodeName:  "NODE_NAME",
	}
	pluginSockPath := fmt.Sprintf("%s.sock", "Ascend910")
	fakePluginAPI := createFakePluginAPI(hdm, "Ascend910", pluginSockPath, fakeKubeInteractor)
	var ctx context.Context
	fakePluginAPI.hps.devices["Ascend910-8c-1-1"] = &npuDevice{
		devType: "Ascend910-8c",
		pciID:   "",
		ID:      "Ascend910-8c-1-1",
		Health:  pluginapi.Healthy,
	}
	_, requestErrs := fakePluginAPI.Allocate(ctx, &requests)
	if requestErrs != nil {
		t.Fatal("TestPluginAPI_Allocate Run Failed")
	}

	t.Logf("TestPluginAPI_Allocate Run Pass")
}
