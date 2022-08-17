/*
* Copyright(C) 2021-2022. Huawei Technologies Co.,Ltd. All rights reserved.
 */
// Package device ascend commmon
package device

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/smartystreets/goconvey/convey"
	"huawei.com/npu-exporter/devmanager"
	npuCommon "huawei.com/npu-exporter/devmanager/common"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"Ascend-device-plugin/pkg/common"
)

const (
	testLogicID  = 3
	testVirDevID = 100
)

// TestUnhealthyState for UnhealthyState
func TestUnhealthyState(t *testing.T) {
	if err := UnhealthyState(1, testLogicID, "healthState", &devmanager.DeviceManagerMock{}); err != nil {
		t.Errorf("TestUnhealthyState Run Failed")
	}
	t.Logf("TestUnhealthyState Run Pass")
}

// TestGetPhyIDByName for PhyIDByName
func TestGetPhyIDByName(t *testing.T) {
	if phyID, err := GetPhyIDByName("Ascend310-3"); err != nil || testLogicID != phyID {
		t.Errorf("TestGetLogicIDByName Run Failed")
	}

	if _, err := GetPhyIDByName("Ascend310-1000"); err == nil {
		t.Errorf("TestGetLogicIDByName Run Failed")
	}
	t.Logf("TestGetLogicIDByName Run Pass")
}

// TestGetDefaultDevices for GetDefaultDevices
func TestGetDefaultDevices(t *testing.T) {
	if _, err := os.Stat(common.HiAIHDCDevice); err != nil {
		if err = createFile(common.HiAIHDCDevice); err != nil {
			t.Fatal("TestGetDefaultDevices Run Failed")
		}
	}

	if _, err := os.Stat(common.HiAIManagerDevice); err != nil {
		if err = createFile(common.HiAIManagerDevice); err != nil {
			t.Fatal("TestGetDefaultDevices Run Failed")
		}
	}

	if _, err := os.Stat(common.HiAISVMDevice); err != nil {
		if err = createFile(common.HiAISVMDevice); err != nil {
			t.Fatal("TestGetDefaultDevices Run Failed")
		}
	}
	var defaultDeivces []string
	if err := GetDefaultDevices(&defaultDeivces); err != nil {
		t.Errorf("TestGetDefaultDevices Run Failed")
	}
	defaultMap := make(map[string]string)
	defaultMap[common.HiAIHDCDevice] = ""
	defaultMap[common.HiAIManagerDevice] = ""
	defaultMap[common.HiAISVMDevice] = ""
	defaultMap[common.HiAi200RCEventSched] = ""
	defaultMap[common.HiAi200RCHiDvpp] = ""
	defaultMap[common.HiAi200RCLog] = ""
	defaultMap[common.HiAi200RCMemoryBandwidth] = ""
	defaultMap[common.HiAi200RCSVM0] = ""
	defaultMap[common.HiAi200RCTsAisle] = ""
	defaultMap[common.HiAi200RCUpgrade] = ""

	for _, str := range defaultDeivces {
		if _, ok := defaultMap[str]; !ok {
			t.Errorf("TestGetDefaultDevices Run Failed")
		}
	}
	t.Logf("TestGetDefaultDevices Run Pass")
}

func createFile(filePath string) error {
	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := f.Chmod(common.SocketChmod); err != nil {
		return err
	}
	return nil
}

func TestGetNewNetworkRecoverDev(t *testing.T) {
	convey.Convey("getNewNetworkRecoverDev test", t, func() {
		convey.Convey("autoStowing is true", func() {
			autoStowingDevs = true
			totalNetworkUnhealthDevices = sets.String{}
			emptySets := sets.String{}
			newNetworkRecoverDevSets, newNetworkUnhealthDevSets := getNewNetworkRecoverDev(emptySets, emptySets)
			convey.So(newNetworkRecoverDevSets, convey.ShouldEqual, emptySets)
			convey.So(newNetworkUnhealthDevSets, convey.ShouldEqual, totalNetworkUnhealthDevices)
		})
		convey.Convey("autoStowing is false", func() {
			autoStowingDevs = false
			totalNetworkUnhealthDevices = sets.String{}
			emptySets := sets.String{}
			newNetworkRecoverDevSets, newNetworkUnhealthDevSets := getNewNetworkRecoverDev(emptySets, emptySets)
			convey.So(newNetworkRecoverDevSets, convey.ShouldHaveSameTypeAs, emptySets)
			convey.So(newNetworkUnhealthDevSets, convey.ShouldHaveSameTypeAs, emptySets)
		})
	})
}

func TestGetDeviceID(t *testing.T) {
	convey.Convey("getDeviceID test", t, func() {
		convey.Convey("getDeviceID get error", func() {
			deviceName := ""
			ascendRuntimeOptions := common.VirtualDev
			_, _, err := common.GetDeviceID(deviceName, ascendRuntimeOptions)
			convey.So(err, convey.ShouldBeError)
		})
		convey.Convey("ascendRuntimeOptions is physicalDev", func() {
			deviceName := "Ascend910-1"
			ascendRuntimeOptions := physicalDev
			_, virID, err := common.GetDeviceID(deviceName, ascendRuntimeOptions)
			convey.So(err, convey.ShouldBeNil)
			convey.So(virID, convey.ShouldBeEmpty)
		})
		convey.Convey("ascendRuntimeOptions is virtualDev", func() {
			deviceName := "Ascend910-2c-112-1"
			ascendRuntimeOptions := common.VirtualDev
			_, virID, err := common.GetDeviceID(deviceName, ascendRuntimeOptions)
			convey.So(err, convey.ShouldBeNil)
			convey.So(virID, convey.ShouldNotBeEmpty)
		})
	})
}

// TestReloadHealthDevice for reloadHealthDevice
func TestReloadHealthDevice(t *testing.T) {
	devices := map[string]*common.NpuDevice{"Ascend310P": &common.NpuDevice{ID: "0", Health: "Healthy"},
		"Ascend310P-1c": &common.NpuDevice{ID: "1", Health: "Unhealthy"}}
	hps := HwPluginServe{devices: devices}
	adc := ascendCommonFunction{}
	adc.reloadHealthDevice(&hps)
	if len(hps.healthDevice) != 1 {
		t.Fatalf("TestReloadHealthDevice Run Failed")
	}
	if len(hps.unHealthDevice) != 1 {
		t.Fatalf("TestReloadHealthDevice Run Failed")
	}
}

// TestVerifyPath for VerifyPath
func TestVerifyPath(t *testing.T) {
	convey.Convey("TestVerifyPath", t, func() {
		convey.Convey("filepath.Abs failed", func() {
			mock := gomonkey.ApplyFunc(filepath.Abs, func(path string) (string, error) {
				return "", fmt.Errorf("err")
			})
			defer mock.Reset()
			_, ret := VerifyPath("")
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("os.Stat failed", func() {
			mock := gomonkey.ApplyFunc(os.Stat, func(name string) (os.FileInfo, error) {
				return nil, fmt.Errorf("err")
			})
			defer mock.Reset()
			_, ret := VerifyPath("./")
			convey.So(ret, convey.ShouldBeFalse)
		})
		convey.Convey("filepath.EvalSymlinks failed", func() {
			mock := gomonkey.ApplyFunc(filepath.EvalSymlinks, func(path string) (string, error) {
				return "", fmt.Errorf("err")
			})
			defer mock.Reset()
			_, ret := VerifyPath("./")
			convey.So(ret, convey.ShouldBeFalse)
		})
	})
}

// TestGetDevState for GetDevState
func TestGetDevState(t *testing.T) {
	convey.Convey("TestGetDevState", t, func() {
		convey.Convey("GetPhyIDByName failed", func() {
			mock := gomonkey.ApplyFunc(GetPhyIDByName, func(_ string) (uint32, error) {
				return 0, fmt.Errorf("err")
			})
			defer mock.Reset()
			adc := ascendCommonFunction{}

			convey.So(adc.GetDevState("", &devmanager.DeviceManagerMock{}), convey.ShouldEqual, v1beta1.Unhealthy)
		})
		convey.Convey("GetLogicID failed", func() {
			mock := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)), "GetLogicIDFromPhysicID",
				func(_ *devmanager.DeviceManagerMock, _ int32) (int32, error) { return 0, fmt.Errorf("err") })
			defer mock.Reset()
			adc := ascendCommonFunction{}
			convey.So(adc.GetDevState("", &devmanager.DeviceManagerMock{}), convey.ShouldEqual, v1beta1.Unhealthy)
		})
		convey.Convey("GetDeviceHealth failed", func() {
			mock := gomonkey.ApplyMethod(reflect.TypeOf(new(devmanager.DeviceManagerMock)), "GetDeviceHealth",
				func(_ *devmanager.DeviceManagerMock, _ int32) (uint32, error) { return 0, fmt.Errorf("err") })
			defer mock.Reset()
			adc := ascendCommonFunction{}
			convey.So(adc.GetDevState("", &devmanager.DeviceManagerMock{}), convey.ShouldEqual, v1beta1.Unhealthy)
		})
		convey.Convey("GetDeviceHealth return unhealth, UnhealthyState failed", func() {
			mock := gomonkey.ApplyFunc(UnhealthyState, func(_ uint32, _ int32, _ string,
				_ devmanager.DeviceInterface) error {
				return fmt.Errorf("err")
			})
			defer mock.Reset()
			adc := ascendCommonFunction{}
			convey.So(adc.GetDevState("", &devmanager.DeviceManagerMock{}), convey.ShouldEqual, v1beta1.Unhealthy)
		})
	})
}

// TestDoWithVolcanoListAndWatch test 310 listen and watch
func TestDoWithVolcanoListAndWatch(t *testing.T) {
	hdm := setParams(false, common.RunMode310)
	if err := hdm.GetNPUs(); err != nil {
		t.Fatal(err)
	}
	mockNode := gomonkey.ApplyFunc(getNodeNpuUsed, func(usedDevices *sets.String, hps *HwPluginServe) {
		return
	})
	mockNodeCtx := gomonkey.ApplyFunc(getNodeWithTodoCtx, func(_ *KubeInteractor) (*v1.Node, error) {
		return nil, nil
	})
	mockPatchNode := gomonkey.ApplyFunc(patchNodeWithTodoCtx, func(_ *KubeInteractor, _ []byte) (*v1.Node, error) {
		return nil, nil
	})
	devices := map[string]*common.NpuDevice{"Ascend310": &common.NpuDevice{ID: "0", Health: "Healthy"}}
	hps := &HwPluginServe{devices: devices}
	hdm.manager.DoWithVolcanoListAndWatch(hps)
	mockNode.Reset()
	mockNodeCtx.Reset()
	mockPatchNode.Reset()
	if len(totalDevices) != 1 || totalDevices.List()[0] != "0" {
		t.Fatal("TestDoWithVolcanoListAndWatch Run Failed")
	}
	t.Logf("TestDoWithVolcanoListAndWatch Run Pass")
}

// TestAssembleSpecVirtualDevice test assembleSpecVirtualDevice
func TestAssembleSpecVirtualDevice(t *testing.T) {
	phyID := int32(0)
	runMode := hiAIAscend910Prefix
	convey.Convey("TestAssembleSpecVirtualDevice", t, func() {
		convey.Convey("aicore is 0", func() {
			adc := ascendCommonFunction{}
			vDevInfo := npuCommon.CgoVDevQueryStru{VDevID: uint32(0)}
			_, _, err := adc.assembleSpecVirtualDevice(runMode, phyID, vDevInfo)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("template name is invalid", func() {
			adc := ascendCommonFunction{}
			vDevID := uint32(testVirDevID)
			aiCore := float32(1)
			invalidTemplateName := "vir04x"
			vDevInfo := npuCommon.CgoVDevQueryStru{
				VDevID: vDevID,
				QueryInfo: npuCommon.CgoVDevQueryInfo{
					Name:      invalidTemplateName,
					Computing: npuCommon.CgoComputingResource{Aic: aiCore},
				},
			}
			_, _, err := adc.assembleSpecVirtualDevice(runMode, phyID, vDevInfo)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("template name is valid", func() {
			adc := ascendCommonFunction{}
			vDevID := uint32(testVirDevID)
			aiCore := float32(1)
			templateName := "vir04"
			vDevInfo := npuCommon.CgoVDevQueryStru{
				VDevID: vDevID,
				QueryInfo: npuCommon.CgoVDevQueryInfo{
					Name:      templateName,
					Computing: npuCommon.CgoComputingResource{Aic: aiCore},
				},
			}
			getDevTypeGet, exist := getDevTypeByTemplateName(runMode, templateName)
			convey.So(exist, convey.ShouldBeTrue)
			vDevType, devID, err := adc.assembleSpecVirtualDevice(runMode, phyID, vDevInfo)
			convey.So(err, convey.ShouldBeNil)
			convey.So(vDevType, convey.ShouldEqual, getDevTypeGet)
			getDevID := fmt.Sprintf("%s-%d-%d", vDevType, vDevInfo.VDevID, phyID)
			convey.So(devID, convey.ShouldEqual, getDevID)
		})
	})
}

// TestAssembleVirtualDevices test assembleVirtualDevices
func TestAssembleVirtualDevices(t *testing.T) {
	phyID := int32(0)
	runMode := hiAIAscend910Prefix
	convey.Convey("TestAssembleVirtualDevices", t, func() {
		convey.Convey("call assembleSpecVirtualDevice failed", func() {
			adc := ascendCommonFunction{}
			vDevInfos := npuCommon.VirtualDevInfo{VDevInfo: []npuCommon.CgoVDevQueryStru{{VDevID: uint32(0)}}}
			devices, deviTypes, vDevID := adc.assembleVirtualDevices(phyID, vDevInfos, runMode)
			convey.So(devices, convey.ShouldBeNil)
			convey.So(deviTypes, convey.ShouldBeNil)
			convey.So(vDevID, convey.ShouldBeNil)
		})
		convey.Convey("call assembleSpecVirtualDevice success", func() {
			adc := ascendCommonFunction{}
			vDevID := uint32(testVirDevID)
			templateName := "vir04"
			aiCore := float32(1)
			vDevInfos := npuCommon.VirtualDevInfo{
				VDevInfo: []npuCommon.CgoVDevQueryStru{
					{
						VDevID: vDevID,
						QueryInfo: npuCommon.CgoVDevQueryInfo{
							Name:      templateName,
							Computing: npuCommon.CgoComputingResource{Aic: aiCore},
						},
					},
				},
			}
			devices, deviTypes, ids := adc.assembleVirtualDevices(phyID, vDevInfos, runMode)
			convey.So(devices, convey.ShouldNotBeNil)
			convey.So(deviTypes, convey.ShouldNotBeNil)
			convey.So(ids, convey.ShouldNotBeNil)
		})
	})
}

// TestGetUnHealthDev test getUnHealthDev
func TestGetUnHealthDev(t *testing.T) {
	convey.Convey("TestGetUnHealthDev", t, func() {
		convey.Convey("autoStowingDevs true", func() {
			autoStowingDevsSave := autoStowingDevs
			autoStowingDevs = true
			device910 := sets.String{}
			device910.Insert("Ascend910-0")
			listenUHDev := sets.String{}
			annotationUHDev := sets.String{}
			labelsRecoverDev := sets.String{}
			_, newAscend910 := getUnHealthDev(listenUHDev, annotationUHDev, labelsRecoverDev, device910)
			convey.So(newAscend910, convey.ShouldNotBeNil)
			autoStowingDevs = autoStowingDevsSave
		})
		convey.Convey("autoStowingDevs false", func() {
			autoStowingDevsSave := autoStowingDevs
			autoStowingDevs = false
			device910 := sets.String{}
			device910.Insert("Ascend910-0")
			listenUHDev := sets.String{}
			annotationUHDev := sets.String{}
			labelsRecoverDev := sets.String{}
			_, newAscend910 := getUnHealthDev(listenUHDev, annotationUHDev, labelsRecoverDev, device910)
			convey.So(newAscend910, convey.ShouldNotBeNil)
			autoStowingDevs = autoStowingDevsSave
		})
	})
}

// TestSetUnHealthyDev test setUnHealthyDev
func TestSetUnHealthyDev(t *testing.T) {
	convey.Convey("TestSetUnHealthyDev", t, func() {
		convey.Convey("IsVirtualDev false", func() {
			totalUHDevicesSave := totalUHDevices
			totalUHDevices = sets.String{}
			device := common.NpuDevice{ID: "Ascend910-0"}
			adc := ascendCommonFunction{}
			adc.setUnHealthyDev("Ascend910", &device)
			convey.So(totalUHDevices.Len(), convey.ShouldEqual, 1)
			totalUHDevices = totalUHDevicesSave
		})
		convey.Convey("GetDeviceID failed ", func() {
			totalUHDevicesSave := totalUHDevices
			totalUHDevices = sets.String{}
			device := common.NpuDevice{ID: "Ascend910-2c-100-0-0"}
			mock := gomonkey.ApplyFunc(common.GetDeviceID, func(deviceName string,
				ascendRuntimeOptions string) (string, string, error) {
				return "", "", fmt.Errorf("error")
			})
			defer mock.Reset()
			adc := ascendCommonFunction{}
			adc.setUnHealthyDev("Ascend910", &device)
			convey.So(totalUHDevices.Len(), convey.ShouldEqual, 0)
			totalUHDevices = totalUHDevicesSave
		})
		convey.Convey("GetDeviceID has ", func() {
			totalUHDevicesSave := totalUHDevices
			totalUHDevices = sets.String{}
			totalUHDevices.Insert("Ascend910-0")
			device := common.NpuDevice{ID: "Ascend910-2c-100-0"}
			adc := ascendCommonFunction{}
			adc.setUnHealthyDev("Ascend910", &device)
			convey.So(totalUHDevices.Len(), convey.ShouldEqual, 1)
			totalUHDevices = totalUHDevicesSave
		})
		convey.Convey("GetDeviceID not has ", func() {
			totalUHDevicesSave := totalUHDevices
			totalUHDevices = sets.String{}
			device := common.NpuDevice{ID: "Ascend910-2c-100-0"}
			adc := ascendCommonFunction{}
			adc.setUnHealthyDev("Ascend910", &device)
			convey.So(totalUHDevices.Len(), convey.ShouldEqual, 1)
			totalUHDevices = totalUHDevicesSave
		})
	})
}
