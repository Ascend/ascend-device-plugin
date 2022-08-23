/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2020-2022. All rights reserved.
 */

package huawei

import (
	"os"
	"testing"

	"go.uber.org/atomic"
	"google.golang.org/grpc"
	"huawei.com/mindx/common/hwlog"
	"huawei.com/npu-exporter/devmanager"
	"k8s.io/apimachinery/pkg/util/sets"

	"Ascend-device-plugin/src/plugin/pkg/npu/common"
)

type fakeHwPluginServe struct {
	hdm            *HwDevManager
	devices        map[string]*common.NpuDevice
	grpcServer     *grpc.Server
	devType        string
	runMode        string
	defaultDevs    []string
	kubeInteractor *KubeInteractor
	healthDevice   sets.String
	unHealthDevice sets.String
}

// NewFakeHwPluginServe to create fakePlugin
func NewFakeHwPluginServe(hdm *HwDevManager, devType string) HwPluginServeInterface {
	return &fakeHwPluginServe{
		devType:        devType,
		hdm:            hdm,
		runMode:        hdm.runMode,
		devices:        make(map[string]*common.NpuDevice),
		healthDevice:   sets.String{},
		unHealthDevice: sets.String{},
	}
}

// TestStart for test Start
func TestStart(t *testing.T) {
	fakeHwDevManager := &HwDevManager{
		runMode:  "ascend910",
		dmgr:     &devmanager.DeviceManagerMock{},
		stopFlag: atomic.NewBool(false),
	}
	pluginSocket := "Ascend910.sock"
	pluginSocketPath := "/var/lib/kubelet/device-plugins/" + pluginSocket
	hps := NewHwPluginServe(fakeHwDevManager, "Ascend910")
	err := hps.Start(pluginSocketPath)
	kubeSocketPath := "/var/lib/kubelet/device-plugins/kubelet.sock"
	_, kubeErr := os.Stat(kubeSocketPath)
	if err != nil && kubeErr != nil && os.IsExist(kubeErr) {
		t.Fatal("TestStart Run Failed")
	}
	t.Logf("TestStart Run Pass")
}

// TestStop for test Stop
func TestStop(t *testing.T) {
	fakeHwDevManager := &HwDevManager{
		runMode:  "ascend910",
		stopFlag: atomic.NewBool(false),
	}
	hps := NewFakeHwPluginServe(fakeHwDevManager, "Ascend910")
	hps.Stop()

	hps.setSocket()
	hps.Stop()
	t.Logf("TestStop Run Pass")
}

// TestGetDevByType for test GetDevByType
func TestGetDevByType(t *testing.T) {
	fakeHwDevManager := &HwDevManager{
		runMode:  "Ascend310",
		dmgr:     &devmanager.DeviceManagerMock{},
		stopFlag: atomic.NewBool(false),
		allDevs:  []common.NpuDevice{},
	}
	hps := NewHwPluginServe(fakeHwDevManager, "Ascend310")
	if err := hps.GetDevByType(); err == nil {
		t.Fatalf("TestGetDevByType Run Failed, expect err, but nil")
	}

	fakeHwDevManager.allDevs = []common.NpuDevice{{ID: "0"}}
	hps = NewHwPluginServe(fakeHwDevManager, "Ascend310")
	if err := hps.GetDevByType(); err == nil {
		t.Fatalf("TestGetDevByType Run Failed, expect err, but nil")
	}

	fakeHwDevManager.allDevs = []common.NpuDevice{{ID: "0", DevType: "Ascend310"}}
	hps = NewHwPluginServe(fakeHwDevManager, "Ascend310")
	if err := hps.GetDevByType(); err != nil {
		t.Fatalf("TestGetDevByType Run Failed, err is %v", err)
	}

	t.Logf("TestGetDevByType Run Pass")
}

// GetDevByType by fake
func (hps *fakeHwPluginServe) GetDevByType() error {
	return nil
}

// Start starts the gRPC server of the device plugin
func (hps *fakeHwPluginServe) Start(pluginSocketPath string) error {
	hwlog.RunLog.Infof("device plugin start serving.")
	// Registers To Kubelet.
	err := hps.Register()
	if err == nil {
		hwlog.RunLog.Infof("register to kubelet success.")
		return nil
	}
	hwlog.RunLog.Errorf("register to kubelet failed, err: %s", err.Error())
	return err
}

func (hps *fakeHwPluginServe) setSocket() {
	// Registers service.
}

// Stop the gRPC server
func (hps *fakeHwPluginServe) Stop() {
	return
}

// Register function is use to register k8s devicePlugin to kubelet.
func (hps *fakeHwPluginServe) Register() error {
	return nil
}
