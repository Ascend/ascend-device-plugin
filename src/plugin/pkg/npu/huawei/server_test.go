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
	"fmt"
	"go.uber.org/atomic"
	"google.golang.org/grpc"
	"huawei.com/npu-exporter/hwlog"
	"k8s.io/apimachinery/pkg/util/sets"
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
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
	pluginSocket := "Ascend910.sock"
	pluginSocketPath := "/var/lib/kubelet/device-plugins/" + pluginSocket
	hps := NewHwPluginServe(fakeHwDevManager, "Ascend910", pluginSocketPath)
	err := hps.Start(pluginSocket, pluginSocketPath)
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
func (hps *fakeHwPluginServe) Start(pluginSocket, pluginSocketPath string) error {
	hwlog.Infof("device plugin start serving.")
	// Registers To Kubelet.
	resourceName := fmt.Sprintf("%s%s", resourceNamePrefix, hps.devType)
	k8sSocketPath := pluginapi.KubeletSocket
	err := hps.Register(k8sSocketPath, pluginSocket, resourceName)
	if err == nil {
		hwlog.Infof("register to kubelet success.")
		return nil
	}
	hwlog.Errorf("register to kubelet failed, err: %s", err.Error())
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
func (hps *fakeHwPluginServe) Register(k8sSocketPath, pluginSocket, resourceName string) error {
	return nil
}
