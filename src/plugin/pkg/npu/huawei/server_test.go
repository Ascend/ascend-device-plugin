/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2020-2021. All rights reserved.
 */

package huawei

import (
	"go.uber.org/atomic"
	"google.golang.org/grpc"
	"huawei.com/npu-exporter/hwlog"
	"k8s.io/apimachinery/pkg/util/sets"
	"os"
	"testing"
)

type fakeHwPluginServe struct {
	hdm            *HwDevManager
	devices        map[string]*npuDevice
	grpcServer     *grpc.Server
	devType        string
	runMode        string
	defaultDevs    []string
	socket         string
	kubeInteractor *KubeInteractor
	healthDevice   sets.String
	unHealthDevice sets.String
}

// NewFakeHwPluginServe to create fakePlugin
func NewFakeHwPluginServe(hdm *HwDevManager, devType string, socket string) HwPluginServeInterface {
	return &fakeHwPluginServe{
		devType:        devType,
		hdm:            hdm,
		runMode:        hdm.runMode,
		devices:        make(map[string]*npuDevice),
		socket:         socket,
		healthDevice:   sets.String{},
		unHealthDevice: sets.String{},
	}
}

// TestStart for test Start
func TestStart(t *testing.T) {
	fakeHwDevManager := &HwDevManager{
		runMode:  "ascend910",
		dmgr:     newFakeDeviceManager(),
		stopFlag: atomic.NewBool(false),
	}
	pluginSocket := "Ascend10.sock"
	pluginSocketPath := "/var/lib/kubelet/device-plugins/" + pluginSocket
	hps := NewHwPluginServe(fakeHwDevManager, "Ascend910", pluginSocketPath)
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
	pluginSocket := "Ascend910.sock"
	pluginSocketPath := "/var/lib/kubelet/device-plugins/" + pluginSocket
	hps := NewHwPluginServe(fakeHwDevManager, "Ascend910", pluginSocketPath)
	err := hps.Stop()
	if err != nil {
		t.Fatal("TestStop Run Failed")
	}
	err = hps.cleanSock()
	if err != nil {
		t.Fatal("TestStop Run Failed")
	}
	t.Logf("TestStop Run Pass")
}

// TestGetDevByType for test GetDevByType
func TestGetDevByType(t *testing.T) {
	fakeHwDevManager := createFakeDevManager("")
	err := fakeHwDevManager.GetNPUs()
	if err != nil {
		t.Fatal(err)
	}
	pluginSocket := "Ascend310.sock"
	pluginSocketPath := "/var/lib/kubelet/device-plugins/" + pluginSocket
	hps := NewHwPluginServe(fakeHwDevManager, "Ascend310", pluginSocketPath)
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

func (hps *fakeHwPluginServe) setSocket(pluginSocketPath string) {
	// Registers service.
}

// Stop the gRPC server
func (hps *fakeHwPluginServe) Stop() error {

	return hps.cleanSock()
}

// if device plugin stopped,the socket file should be removed
func (hps *fakeHwPluginServe) cleanSock() error {
	return nil
}

// Register function is use to register k8s devicePlugin to kubelet.
func (hps *fakeHwPluginServe) Register() error {
	return nil
}
