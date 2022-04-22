/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2020-2022. All rights reserved.
 */

package huawei

import (
	"os"
	"testing"

	"go.uber.org/atomic"
	"google.golang.org/grpc"
	"huawei.com/npu-exporter/hwlog"
	"k8s.io/apimachinery/pkg/util/sets"

	"Ascend-device-plugin/src/plugin/pkg/npu/common"
	"Ascend-device-plugin/src/plugin/pkg/npu/dsmi"
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
		dmgr:     dsmi.NewFakeDeviceManager(),
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
	hps := NewHwPluginServe(fakeHwDevManager, "Ascend910")
	hps.Stop()
	t.Logf("TestStop Run Pass")
}

// TestGetDevByType for test GetDevByType
func TestGetDevByType(t *testing.T) {
	fakeHwDevManager := createFakeDevManager("")
	fakeHwDevManager.runMode = common.RunMode310
	err := fakeHwDevManager.GetNPUs()
	if err != nil {
		t.Fatal(err)
	}
	hps := NewHwPluginServe(fakeHwDevManager, "Ascend310")
	err = hps.GetDevByType()
	if err != nil {
		t.Fatal("TestGetDevByType Run Failed")
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
