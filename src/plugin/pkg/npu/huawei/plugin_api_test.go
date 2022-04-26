/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2020-2022. All rights reserved.
 */

package huawei

import (
	"testing"
	"time"

	. "github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	"github.com/smartystreets/goconvey/convey"
	"go.uber.org/atomic"
	"golang.org/x/net/context"
	"huawei.com/npu-exporter/hwlog"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"Ascend-device-plugin/src/plugin/pkg/npu/common"
	"Ascend-device-plugin/src/plugin/pkg/npu/dsmi"
	mock_v1beta1 "Ascend-device-plugin/src/plugin/pkg/npu/huawei/mock_kubelet_v1beta1"
	"Ascend-device-plugin/src/plugin/pkg/npu/huawei/mock_kubernetes"
	"Ascend-device-plugin/src/plugin/pkg/npu/huawei/mock_v1"
)

const (
	device1         = `"devices":[{"device_id":"0","device_ip":"127.0.0.0"},{"device_id":"1","device_ip":"127.0.0.1"}]`
	device2         = `"devices":[{"device_id":"1","device_ip":"127.0.0.1"},{"device_id":"0","device_ip":"127.0.0.0"}]`
	annonationTest1 = `{"pod_name":"pod_name","server_id":"127.0.0.0",` + device1 + `}`
	annonationTest2 = `{"pod_name":"pod_name","server_id":"127.0.0.0",` + device2 + `}`
	sleepTestFour   = 4
)

// TestPluginAPIListAndWatch for listAndWatch
func TestPluginAPIListAndWatch(t *testing.T) {
	hdm := createFakeDevManager("ascend910")
	o := Option{GetFdFlag: false, UseAscendDocker: false, UseVolcanoType: false, ListAndWatchPeriod: sleepTime,
		AutoStowingDevs: true, KubeConfig: ""}
	hdm.SetParameters(o)
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
		ObjectMeta: metav1.ObjectMeta{Annotations: make(map[string]string), Labels: make(map[string]string)}}
	podList := &v1.PodList{}
	pod1 := v1.Pod{ObjectMeta: metav1.ObjectMeta{Annotations: make(map[string]string), Labels: make(map[string]string)}}
	podList.Items = append(podList.Items, pod1)
	node1.Annotations[huaweiAscend910] = "Ascend910-1,Ascend910-2"
	pod1.Annotations[huaweiAscend910] = "Ascend910-1"
	mockK8s := mock_kubernetes.NewMockInterface(ctrl)
	mockV1 := mock_v1.NewMockCoreV1Interface(ctrl)
	mockNode := mock_v1.NewMockNodeInterface(ctrl)
	mockPod := mock_v1.NewMockPodInterface(ctrl)
	mockNode.EXPECT().Get(context.Background(), gomock.Any(), metav1.GetOptions{}).AnyTimes().Return(node1, nil)
	mockNode.EXPECT().Patch(context.Background(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(),
		gomock.Any()).AnyTimes().Return(node1, nil)
	mockPod.EXPECT().List(context.Background(), gomock.Any()).AnyTimes().Return(podList, nil)
	mockV1.EXPECT().Pods(gomock.Any()).AnyTimes().Return(mockPod)
	mockV1.EXPECT().Nodes().AnyTimes().Return(mockNode)
	mockK8s.EXPECT().CoreV1().AnyTimes().Return(mockV1)
	fakeKubeInteractor := &KubeInteractor{clientset: mockK8s, nodeName: "NODE_NAME"}
	hdm.allDevs = getTestDevs()
	for _, devType := range devTypes {
		mockstream := mock_v1beta1.NewMockDevicePlugin_ListAndWatchServer(ctrl)
		mockstream.EXPECT().Send(&v1beta1.ListAndWatchResponse{}).AnyTimes().Return(nil)
		fakePluginAPI = createFakePluginAPI(hdm, devType, fakeKubeInteractor)
		go changeBreakFlag(fakePluginAPI)
		err := fakePluginAPI.ListAndWatch(&v1beta1.Empty{}, mockstream)
		if err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("TestPluginAPI_ListAndWatch Run Pass")
}

func getTestDevs() []common.NpuDevice {
	return []common.NpuDevice{
		{
			DevType:       "Ascend710",
			ID:            "Ascend710-0",
			Health:        "Health",
			NetworkHealth: "Health",
		},
	}
}

func changeBreakFlag(api *pluginAPI) {
	time.Sleep(sleepTestFour * time.Second)
	api.outbreak.Store(true)
}

func createFakePluginAPI(hdm *HwDevManager, devType string, ki *KubeInteractor) *pluginAPI {
	return &pluginAPI{hps: &HwPluginServe{
		devType:        devType,
		hdm:            hdm,
		runMode:        hdm.runMode,
		devices:        make(map[string]*common.NpuDevice),
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
	o := Option{
		GetFdFlag:          false,
		UseAscendDocker:    false,
		UseVolcanoType:     true,
		ListAndWatchPeriod: sleepTime,
		AutoStowingDevs:    true,
		KubeConfig:         "",
	}
	hdm.SetParameters(o)
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
		if common.IsVirtualDev(devType) {
			continue
		}
		fakePluginAPI = createFakePluginAPI(hdm, devType, fakeKubeInteractor)
	}
	devices := make(map[string]string, 1)
	devices["0"] = "127.0.0.0"
	devices["1"] = "127.0.0.1"
	annonationString := fakePluginAPI.addAnnotation(devices, "pod_name", "127.0.0.0")
	if annonationString == annonationTest1 || annonationString == annonationTest2 {
		t.Logf("TestAddAnnotation Run Pass")
	} else {
		t.Fatal("TestAddAnnotation Run Failed")
	}
}

// TestAllocate for test Allocate
func TestAllocate(t *testing.T) {
	hdm := createFakeDevManager("ascend910")
	o := Option{GetFdFlag: false, UseAscendDocker: false, UseVolcanoType: false, ListAndWatchPeriod: sleepTime,
		AutoStowingDevs: true, KubeConfig: ""}
	hdm.SetParameters(o)
	if err := hdm.GetNPUs(); err != nil {
		t.Fatal(err)
	}
	devTypes := hdm.GetDevType()
	if len(devTypes) == 0 {
		t.Fatal("TestPluginAPI_Allocate Run Failed")
	}
	devicesIDs := []string{"Ascend910-8c-1-1"}
	var containerRequests []*v1beta1.ContainerAllocateRequest
	tmp := &v1beta1.ContainerAllocateRequest{
		DevicesIDs: devicesIDs,
	}
	containerRequests = append(containerRequests, tmp)
	requests := v1beta1.AllocateRequest{
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
	fakePluginAPI := createFakePluginAPI(hdm, "Ascend910", fakeKubeInteractor)
	var ctx context.Context
	fakePluginAPI.hps.devices["Ascend910-8c-1-1"] = &common.NpuDevice{
		DevType: "Ascend910-8c",
		PciID:   "",
		ID:      "Ascend910-8c-1-1",
		Health:  v1beta1.Healthy,
	}

	_, requestErrs := fakePluginAPI.Allocate(ctx, &requests)
	if requestErrs != nil {
		t.Fatal("TestPluginAPI_Allocate Run Failed")
	}

	t.Logf("TestPluginAPI_Allocate Run Pass")
}

// TestGetDevicePluginOptions for test GetDevicePluginOptions
func TestGetDevicePluginOptions(t *testing.T) {
	var ctx context.Context
	var pe v1beta1.Empty
	hdm := createFakeDevManager("ascend910")
	fakeKubeInteractor := &KubeInteractor{}
	fakePluginAPI := createFakePluginAPI(hdm, "Ascend910", fakeKubeInteractor)
	_, err := fakePluginAPI.GetDevicePluginOptions(ctx, &pe)
	if err != nil {
		t.Fatal("TestGetDevicePluginOptions Run Failed")
	}
	t.Logf("TestGetDevicePluginOptions Run Pass")
}

// TestGetNPUAnnotationOfPod for test getNPUAnnotationOfPod
func TestGetNPUAnnotationOfPod(t *testing.T) {
	pods := mockPodList()
	var res []v1.Pod
	ascendVisibleDevices := make(map[string]string, MaxVirtualDevNum)
	hdm := createFakeDevManager("ascend910")
	fakeKubeInteractor := &KubeInteractor{}
	fakePluginAPI := createFakePluginAPI(hdm, "Ascend910", fakeKubeInteractor)
	for _, pod := range pods {
		if err := fakePluginAPI.checkPodNameAndSpace(pod.Name, podNameMaxLength); err != nil {
			t.Fatal("TestGetNPUAnnotationOfPod Run Failed")
		}
		if err := fakePluginAPI.checkPodNameAndSpace(pod.Namespace, podNameSpaceMaxLength); err != nil {
			t.Fatal("TestGetNPUAnnotationOfPod Run Failed")
		}
		if fakePluginAPI.getNPUResourceNumOfPod(&pod) >= 0 && fakePluginAPI.isAscendAssignedPod(&pod) &&
			!fakePluginAPI.isShouldDeletePod(&pod) {
			res = append(res, pod)
		}
	}
	oldPod := getOldestPod(pods)
	if oldPod == nil {
		t.Fatal("TestGetNPUAnnotationOfPod Run Failed")
	}
	allocateDevice := sets.NewString()
	err := fakePluginAPI.getNPUAnnotationOfPod(oldPod, &allocateDevice, 1)
	if err != nil {
		t.Fatal("TestGetNPUAnnotationOfPod Run Failed")
	}
	err = fakePluginAPI.getAscendVisiDevsWithVolcano(allocateDevice, &ascendVisibleDevices)
	if err != nil {
		t.Fatal("TestGetNPUAnnotationOfPod Run Failed")
	}
	t.Logf("TestGetNPUAnnotationOfPod Run Pass")
}

func mockPodList() []v1.Pod {
	annotationTag := resourceNamePrefix + "Ascend910"
	annotations := make(map[string]string, 1)
	annotations[annotationTag] = "Ascend910-0"
	annotations["predicate-time"] = "1626785193048251590"
	return []v1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "mindx-dls-npu-1p-default-2p-0",
				Namespace:   "btg-test",
				Annotations: annotations,
			},
			Spec: v1.PodSpec{
				Containers: []v1.Container{
					{Resources: v1.ResourceRequirements{
						Limits: v1.ResourceList{
							"huawei.com/ascend910": resource.Quantity{},
						},
					}},
				},
			},
			Status: v1.PodStatus{
				Reason: "UnexpectedAdmissionError",
				ContainerStatuses: []v1.ContainerStatus{
					{State: v1.ContainerState{
						Waiting: &v1.ContainerStateWaiting{},
					}},
				},
			},
		},
	}
}

func TestCheckDeviceNetworkHealthStatus(t *testing.T) {
	convey.Convey("checkDeviceNetworkHealthStatus", t, func() {
		convey.Convey("network status unchanged", func() {
			patches := ApplyFunc(hwlog.RunLog.Error, func(args ...interface{}) {
				return
			})
			defer patches.Reset()
			hdm := createFake910HwDevManager("ascend910", false, false, false)
			hdm.dmgr = dsmi.NewFakeDeviceManager()
			fakeKubeInteractor := &KubeInteractor{}
			device := &common.NpuDevice{
				DevType:       "Ascend910",
				PciID:         "",
				ID:            "Ascend910-0",
				Health:        v1beta1.Healthy,
				NetworkHealth: v1beta1.Healthy,
			}
			fakePluginAPI := createFakePluginAPI(hdm, "Ascend910", fakeKubeInteractor)
			ret := fakePluginAPI.checkDeviceNetworkHealthStatus(device)
			convey.So(ret, convey.ShouldBeFalse)
		})
	})
}

func TestCheckDeviceNetworkStatusChange(t *testing.T) {
	convey.Convey("network health status", t, func() {
		convey.Convey("network status changed to unhealthy", func() {
			patches := ApplyFunc(hwlog.RunLog.Error, func(args ...interface{}) {
				return
			})
			defer patches.Reset()
			hdm := createFake910HwDevManager("ascend910", false, false, false)
			hdm.dmgr = dsmi.NewFakeDeviceManager()
			fakeKubeInteractor := &KubeInteractor{}
			device := &common.NpuDevice{
				DevType:       "Ascend910",
				PciID:         "",
				ID:            "Ascend910-3",
				Health:        v1beta1.Healthy,
				NetworkHealth: v1beta1.Healthy,
			}
			fakePluginAPI := createFakePluginAPI(hdm, "Ascend910", fakeKubeInteractor)
			ret := fakePluginAPI.checkDeviceNetworkHealthStatus(device)
			convey.So(ret, convey.ShouldBeTrue)
		})
	})
}

func TestDonotCheckNetworkStatus(t *testing.T) {
	convey.Convey("do not check device network status", t, func() {
		convey.Convey("virtual device don't check network healthy", func() {
			hdm := createFakeDevManager("ascend910")
			fakeKubeInteractor := &KubeInteractor{}
			device := &common.NpuDevice{
				DevType:       "Ascend910-8c",
				PciID:         "",
				ID:            "Ascend910-8c-1-1",
				Health:        v1beta1.Healthy,
				NetworkHealth: v1beta1.Healthy,
			}
			fakePluginAPI := createFakePluginAPI(hdm, "Ascend910-8c", fakeKubeInteractor)
			ret := fakePluginAPI.checkDeviceNetworkHealthStatus(device)
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("device id error", func() {
			patches := ApplyFunc(hwlog.RunLog.Errorf, func(format string, args ...interface{}) {
				return
			})
			patches2 := ApplyFunc(hwlog.RunLog.Error, func(args ...interface{}) {
				return
			})
			defer patches.Reset()
			defer patches2.Reset()
			hdm := createFake910HwDevManager("ascend910", false, false, false)
			fakeKubeInteractor := &KubeInteractor{}
			device := &common.NpuDevice{
				DevType:       "Ascend910",
				PciID:         "",
				ID:            "Ascend910-1000",
				Health:        v1beta1.Healthy,
				NetworkHealth: v1beta1.Healthy,
			}
			fakePluginAPI := createFakePluginAPI(hdm, "Ascend910", fakeKubeInteractor)
			ret := fakePluginAPI.checkDeviceNetworkHealthStatus(device)
			convey.So(ret, convey.ShouldBeFalse)
		})
	})
}

// TestGetAscendVisiDevsWithVolcano for getAscendVisiDevsWithVolcano
func TestGetAscendVisiDevsWithVolcano(t *testing.T) {
	hdm := createFake910HwDevManager("ascend910", false, false, false)
	fakeKubeInteractor := &KubeInteractor{}
	fakePluginAPI := createFakePluginAPI(hdm, "Ascend910", fakeKubeInteractor)
	allocateDevice := sets.NewString()
	ascendVisibleDevices := make(map[string]string, MaxVirtualDevNum)

	convey.Convey("isExecTimingUpdate", t, func() {
		convey.Convey("IsPatchSuccess is false", func() {
			fakePluginAPI.getAscendVisiDevsWithVolcano(allocateDevice, &ascendVisibleDevices)
			convey.So(GetAnnotationObj().IsUpdateComplete.Load(), convey.ShouldBeFalse)
		})
	})
}

// TestGetPreferredAllocation for GetPreferredAllocation
func TestGetPreferredAllocation(t *testing.T) {
	hdm := createFake910HwDevManager("ascend910", false, false, false)
	fakeKubeInteractor := &KubeInteractor{}
	fakePluginAPI := createFakePluginAPI(hdm, "Ascend910", fakeKubeInteractor)
	var ctx context.Context
	req := &v1beta1.PreferredAllocationRequest{}
	fakePluginAPI.GetPreferredAllocation(ctx, req)
}
