/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2020-2022. All rights reserved.
 */

package huawei

import (
	"math"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"go.uber.org/atomic"
	"golang.org/x/net/context"
	"huawei.com/npu-exporter/devmanager"
	"huawei.com/npu-exporter/hwlog"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"Ascend-device-plugin/src/plugin/pkg/npu/common"
)

const (
	device1         = `"devices":[{"device_id":"0","device_ip":"127.0.0.0"},{"device_id":"1","device_ip":"127.0.0.1"}]`
	device2         = `"devices":[{"device_id":"1","device_ip":"127.0.0.1"},{"device_id":"0","device_ip":"127.0.0.0"}]`
	annonationTest1 = `{"pod_name":"pod_name","server_id":"127.0.0.0",` + device1 + `}`
	annonationTest2 = `{"pod_name":"pod_name","server_id":"127.0.0.0",` + device2 + `}`
	sleepTestFour   = 4
	unHealthyCode   = 2
)

// TestPluginAPIListAndWatchWithoutVolcano for listAndWatch with out volcano schedule
func TestPluginAPIListAndWatchWithoutVolcano(t *testing.T) {
	hdm := setParams(false, common.RunMode910)
	if err := hdm.GetNPUs(); err != nil {
		t.Fatal(err)
	}
	devTypes := hdm.GetDevType()
	if len(devTypes) == 0 {
		t.Fatal("TestPluginAPIListAndWatchWithoutVolcano Run Failed")
	}
	mockKube := gomonkey.ApplyFunc(sendDevToKubelet,
		func(_ *v1beta1.ListAndWatchResponse, _ v1beta1.DevicePlugin_ListAndWatchServer) error {
			return nil
		})
	fakeKubeInteractor := &KubeInteractor{clientset: nil, nodeName: "NODE_NAME"}
	for _, devType := range devTypes {
		fakePluginAPI := createFakePluginAPI(hdm, devType, fakeKubeInteractor)
		go changeBreakFlag(fakePluginAPI)
		err := fakePluginAPI.ListAndWatch(&v1beta1.Empty{}, nil)
		if err != nil {
			t.Fatal(err)
		}
	}
	mockKube.Reset()
	t.Logf("TestPluginAPIListAndWatchWithoutVolcano Run Pass")
}

// TestUpdatePodRealAllocate for update pod real using devices
func TestUpdatePodRealAllocate(t *testing.T) {
	hdm := setParams(true, common.RunMode910)
	if err := hdm.GetNPUs(); err != nil {
		t.Fatal(err)
	}
	fakeKubeInteractor := &KubeInteractor{clientset: nil, nodeName: "NODE_NAME"}
	podList := getTestPodList(huaweiAscend910, "Ascend910-0")
	mockPod := gomonkey.ApplyFunc(getPodList, func(_ *KubeInteractor) (*v1.PodList, error) {
		return podList, nil
	})
	fakePluginAPI := createFakePluginAPI(hdm, hiAIAscend910Prefix, fakeKubeInteractor)
	fakePluginAPI.updatePodRealAllocate(podPhaseBlackList)
	mockPod.Reset()
	if len(fakePluginAPI.hps.vol2KlDevMap) != 0 {
		t.Fatal("TestUpdatePodRealAllocate Run Failed")
	}
	t.Logf("TestUpdatePodRealAllocate Run Pass")
}

func getTestPodList(ascendType, ascendValue string) *v1.PodList {
	annotations := make(map[string]string, 1)
	annotations[ascendType] = ascendValue
	annotations[podPredicateTime] = strconv.FormatUint(math.MaxUint64, baseDec)
	containers := getContainers()
	podList := []v1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:        "mindx-dls-npu-1p-default-2p-0",
				Namespace:   "btg-test",
				Annotations: annotations,
			},
			Status: v1.PodStatus{
				Phase: v1.PodPending,
			},
			Spec: v1.PodSpec{
				Containers: containers,
			},
		},
	}
	return &v1.PodList{
		Items: podList,
	}
}

func getContainers() []v1.Container {
	limits := resource.NewQuantity(1, resource.DecimalExponent)
	container := v1.Container{
		Resources: v1.ResourceRequirements{
			Limits: v1.ResourceList{
				v1.ResourceName(huaweiAscend910): *limits,
			},
		},
	}
	return []v1.Container{
		container,
	}
}

func getTestNode(ascendType, ascendValue string) *v1.Node {
	annotations := make(map[string]string, 1)
	annotations[ascendType] = ascendValue
	labels := make(map[string]string, 1)
	labels[huaweiRecoverAscend910] = "0"
	return &v1.Node{
		Status: v1.NodeStatus{
			Allocatable: v1.ResourceList{
				v1.ResourceName(ascendType): resource.Quantity{},
			},
		},
		ObjectMeta: metav1.ObjectMeta{
			Annotations: annotations,
			Labels:      labels,
		},
	}
}

func setParams(volcanoType bool, runMode string) *HwDevManager {
	hdm := createFakeDevManager(runMode)
	hdm.SetParameters(
		Option{GetFdFlag: false, UseAscendDocker: false, UseVolcanoType: volcanoType,
			ListAndWatchPeriod: sleepTime, AutoStowingDevs: true, KubeConfig: ""})
	return hdm
}

func changeBreakFlag(api *pluginAPI) {
	time.Sleep(sleepTestFour * time.Second)
	api.hps.outbreak.Store(true)
	if api.hps.stopCh == nil {
		return
	}
	<-api.hps.stopCh
}

func createFakePluginAPI(hdm *HwDevManager, devType string, ki *KubeInteractor) *pluginAPI {
	return &pluginAPI{
		hps: &HwPluginServe{
			devType:        devType,
			hdm:            hdm,
			runMode:        hdm.runMode,
			devices:        make(map[string]*common.NpuDevice),
			kubeInteractor: ki,
			healthDevice:   sets.String{},
			unHealthDevice: sets.String{},
			stopCh:         make(chan struct{}),
			outbreak:       atomic.NewBool(false),
			vol2KlDevMap:   make(map[string]string, maxDevicesNum),
		},
	}
}

// TestAddAnnotation for test AddAnnotation
func TestAddAnnotation(t *testing.T) {
	hdm := setParams(true, common.RunMode910)
	if err := hdm.GetNPUs(); err != nil {
		t.Fatal(err)
	}
	devTypes := hdm.GetDevType()
	if len(devTypes) == 0 {
		t.Fatal("TestAddAnnotation Run Failed")
	}
	var fakePluginAPI *pluginAPI
	fakeKubeInteractor := &KubeInteractor{
		clientset: nil,
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
	if fakePluginAPI == nil {
		t.Fatal("TestAddAnnotation Run Failed: create fake plugin api failed")
	}
	annotationString := fakePluginAPI.addAnnotation(devices, "pod_name", "127.0.0.0")
	if annotationString == annonationTest1 || annotationString == annonationTest2 {
		t.Logf("TestAddAnnotation Run Pass")
	} else {
		t.Fatal("TestAddAnnotation Run Failed")
	}
}

// TestAllocate for test Allocate
func TestAllocate(t *testing.T) {
	hdm := setParams(false, common.RunMode910)
	if err := hdm.GetNPUs(); err != nil {
		t.Fatal(err)
	}
	devTypes := hdm.GetDevType()
	if len(devTypes) == 0 {
		t.Fatal("TestAllocate Run Failed")
	}
	fakeKubeInteractor := &KubeInteractor{clientset: nil, nodeName: "NODE_NAME"}
	fakePluginAPI := createFakePluginAPI(hdm, "Ascend910", fakeKubeInteractor)
	fakePluginAPI.hps.devices["Ascend910-8c-1-1"] = getTestDevs()

	if _, requestErrs := fakePluginAPI.Allocate(nil, getK8sRequest()); requestErrs != nil {
		t.Fatal("TestAllocate Run Failed")
	}
	t.Logf("TestAllocate Run Pass")
}

// TestAllocateWithVolcano for test Allocate with volcano
func TestAllocateWithVolcano(t *testing.T) {
	hdm := setParams(true, common.RunMode910)
	if err := hdm.GetNPUs(); err != nil {
		t.Fatal(err)
	}
	fakeKubeInteractor := &KubeInteractor{clientset: nil, nodeName: "NODE_NAME"}
	node := getTestNode(huaweiAscend910, "Ascend910-8c-1-1")
	podList := getTestPodList(huaweiAscend910, "Ascend910-8c-1-1")
	mockPod := gomonkey.ApplyFunc(getPodList, func(_ *KubeInteractor) (*v1.PodList, error) {
		return podList, nil
	})
	mockUpdate := gomonkey.ApplyFunc(tryUpdatePodAnnotation, func(_ *HwPluginServe, _ *v1.Pod,
		_ map[string]string) error {
		return nil
	})
	mockUsedNpu := gomonkey.ApplyFunc(getNodeNpuUsed, func(usedDevices *sets.String, hps *HwPluginServe) {
		return
	})
	mockNodeCtx := gomonkey.ApplyFunc(getNodeWithBackgroundCtx, func(_ *KubeInteractor) (*v1.Node, error) {
		return node, nil
	})
	mockState := gomonkey.ApplyFunc(patchNodeState, func(_ *KubeInteractor, _, _ *v1.Node) (*v1.Node, []byte, error) {
		return node, nil, nil
	})
	defer func() {
		mockPod.Reset()
		mockUpdate.Reset()
		mockUsedNpu.Reset()
		mockNodeCtx.Reset()
		mockState.Reset()
	}()
	fakePluginAPI := createFakePluginAPI(hdm, "Ascend910", fakeKubeInteractor)
	fakePluginAPI.hps.devices["Ascend910-8c-1-1"] = getTestDevs()
	if _, requestErrs := fakePluginAPI.Allocate(nil, getK8sRequest()); requestErrs != nil {
		t.Fatal("TestAllocateWithVolcano Run Failed")
	}
	t.Logf("TestAllocateWithVolcano Run Pass")
}

func getK8sRequest() *v1beta1.AllocateRequest {
	containerRequests := []*v1beta1.ContainerAllocateRequest{
		{
			DevicesIDs: []string{"Ascend910-8c-1-1"},
		},
	}
	return &v1beta1.AllocateRequest{
		ContainerRequests: containerRequests,
	}
}

func getTestDevs() *common.NpuDevice {
	return &common.NpuDevice{
		DevType: "Ascend910-8c",
		PciID:   "",
		ID:      "Ascend910-8c-1-1",
		Health:  v1beta1.Healthy,
	}
}

// TestGetNPUByStatus test get npu by status
func TestGetNPUByStatus(t *testing.T) {
	hdm := setParams(true, common.RunMode910)
	if err := hdm.GetNPUs(); err != nil {
		t.Fatal(err)
	}
	fakeKubeInteractor := &KubeInteractor{clientset: nil, nodeName: "NODE_NAME"}
	fakePluginAPI := createFakePluginAPI(hdm, "Ascend910", fakeKubeInteractor)
	podList := getTestPodList(huaweiAscend910, "Ascend910-8c-1-1")
	mockPod := gomonkey.ApplyFunc(getPodList, func(_ *KubeInteractor) (*v1.PodList, error) {
		return podList, nil
	})
	var useNpu []string
	getFailed := getNPUByStatus(fakePluginAPI.hps, &useNpu)
	mockPod.Reset()
	if getFailed {
		t.Fatal("TestGetNPUByStatus Run Failed")
	}
	t.Logf("TestGetNPUByStatus Run Pass")
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
	mock := gomonkey.ApplyFunc(tryUpdatePodAnnotation, func(hps *HwPluginServe, pod *v1.Pod,
		annotation map[string]string) error {
		return nil
	})
	defer mock.Reset()
	pods := mockPodList()
	var res []v1.Pod
	hdm := createFakeDevManager("ascend910")
	fakeKubeInteractor := &KubeInteractor{}
	fakePluginAPI := createFakePluginAPI(hdm, "Ascend910", fakeKubeInteractor)
	fakePluginAPI.hps.devices["Ascend910-0"] = &common.NpuDevice{}
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
	oldPod := getOldestPod(pods, fakePluginAPI.hps)
	if oldPod == nil {
		t.Fatal("TestGetNPUAnnotationOfPod Run Failed")
	}

	allocateDevice, err := fakePluginAPI.getVolAllocateDevice(oldPod)
	if err != nil {
		t.Fatalf("TestGetNPUAnnotationOfPod Run Failed, error is %v", err)
	}
	_, err = fakePluginAPI.getDeviceListIP(allocateDevice)
	if err != nil {
		t.Fatalf("TestGetNPUAnnotationOfPod Run Failed, error is %v", err)
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
			patches := gomonkey.ApplyFunc(hwlog.RunLog.Error, func(args ...interface{}) {
				return
			})
			defer patches.Reset()
			hdm := createFake910HwDevManager("ascend910", false, false, false)
			hdm.dmgr = &devmanager.DeviceManagerMock{}
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
			patches := gomonkey.ApplyFunc(hwlog.RunLog.Error, func(args ...interface{}) {
				return
			})
			defer patches.Reset()
			mock := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)), "GetDeviceNetWorkHealth",
				func(_ *devmanager.DeviceManagerMock, _ int32) (uint32, error) { return unHealthyCode, nil })
			defer mock.Reset()
			hdm := createFake910HwDevManager("ascend910", false, false, false)
			hdm.dmgr = &devmanager.DeviceManagerMock{}
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
			patches := gomonkey.ApplyFunc(hwlog.RunLog.Errorf, func(format string, args ...interface{}) {
				return
			})
			patches2 := gomonkey.ApplyFunc(hwlog.RunLog.Error, func(args ...interface{}) {
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
	var allocateDevice []string

	convey.Convey("isExecTimingUpdate", t, func() {
		convey.Convey("IsPatchSuccess is false", func() {
			fakePluginAPI.getDeviceListIP(allocateDevice)
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
