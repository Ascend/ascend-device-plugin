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
	"github.com/golang/protobuf/ptypes/empty"
	"os"
	"testing"
)

const (
	unHealthyTestLogicID = 3
	serverSockFd         = "/var/lib/kubelet/device-plugins/davinci-mini.sock"
	serverSock310        = "/var/lib/kubelet/device-plugins/Ascend310.sock"
)

type fakeDeviceManager struct{}

func newFakeDeviceManager() *fakeDeviceManager {
	return &fakeDeviceManager{}
}

// EnableContainerService for enableContainerService
func (d *fakeDeviceManager) EnableContainerService() error {
	return nil
}

// GetDeviceCount get ascend910 device quantity
func (d *fakeDeviceManager) GetDeviceCount() (int32, error) {
	return int32(npuTestNum), nil
}

// GetDeviceList device get list
func (d *fakeDeviceManager) GetDeviceList(devices *[hiAIMaxDeviceNum]uint32) (int32, error) {
	devNum, err := d.GetDeviceCount()
	if err != nil {
		return devNum, err
	}
	// transfer device list
	var i int32
	for i = 0; i < devNum; i++ {
		(*devices)[i] = uint32(i)
	}

	return devNum, nil
}

//  GetDeviceHealth get device health by id
func (d *fakeDeviceManager) GetDeviceHealth(logicID int32) (uint32, error) {
	if logicID == unHealthyTestLogicID {
		return uint32(unHealthyTestLogicID), nil
	}
	return uint32(0), nil
}

// GetPhyID get physic id form logic id
func (d *fakeDeviceManager) GetPhyID(logicID uint32) (uint32, error) {
	return logicID, nil
}

// GetLogicID get logic id form physic id
func (d *fakeDeviceManager) GetLogicID(phyID uint32) (uint32, error) {
	return phyID, nil

}

// ShutDown the function
func (d *fakeDeviceManager) ShutDown() {
	fmt.Printf("use fake DeviceManager function ShutDown")
}

// GetChipInfo for fakeDeviceManager
func (d *fakeDeviceManager) GetChipInfo(logicID int32) (*ChipInfo, error) {
	chip := &ChipInfo{
		ChipName: "310",
		ChipType: "ASCEND",
		ChipVer:  "",
	}
	return chip, nil
}

// GetDeviceIP get deviceIP
func (d *fakeDeviceManager) GetDeviceIP(logicID int32) (string, error) {
	retIPAddress := fmt.Sprintf("%d.%d.%d.%d", 0, 0, 0, logicID)
	return retIPAddress, nil
}

// TestUnhealthyState for UnhealthyState
func TestUnhealthyState(t *testing.T) {
	err := unhealthyState(1, uint32(3), "healthState", newFakeDeviceManager())
	if err != nil {
		t.Errorf("TestUnhealthyState Run Failed")
	}
	t.Logf("TestUnhealthyState Run Pass")
}

// TestGetLogicIDByName for LogicIDByName
func TestGetLogicIDByName(t *testing.T) {
	var logicID int32
	err := getLogicIDByName("Ascend310-3", &logicID)
	if err != nil || unHealthyTestLogicID != logicID {
		t.Errorf("TestGetLogicIDByName Run Failed")
	}

	err = getLogicIDByName("Ascend310-1000", &logicID)
	if err == nil {
		t.Errorf("TestGetLogicIDByName Run Failed")
	}
	t.Logf("TestGetLogicIDByName Run Pass")
}

// TestGetDefaultDevices for GetDefaultDevices
func TestGetDefaultDevices(t *testing.T) {
	if _, err := os.Stat(hiAIHDCDevice); err != nil {
		if err = createFile(hiAIHDCDevice); err != nil {
			t.Fatal("TestGetDefaultDevices Run Failed")
		}
	}

	if _, err := os.Stat(hiAIManagerDevice); err != nil {
		if err = createFile(hiAIManagerDevice); err != nil {
			t.Fatal("TestGetDefaultDevices Run Failed")
		}
	}

	if _, err := os.Stat(hiAISVMDevice); err == nil {
		if err = createFile(hiAISVMDevice); err != nil {
			t.Fatal("TestGetDefaultDevices Run Failed")
		}
	}
	var defaultDeivces []string
	err := getDefaultDevices(&defaultDeivces)
	if err != nil {
		t.Errorf("TestGetDefaultDevices Run Failed")
	}
	defaultMap := make(map[string]empty.Empty)
	defaultMap[hiAIHDCDevice] = empty.Empty{}
	defaultMap[hiAIManagerDevice] = empty.Empty{}
	defaultMap[hiAISVMDevice] = empty.Empty{}
	for _, str := range defaultDeivces {
		_, ok := defaultMap[str]
		if !ok {
			t.Errorf("TestGetDefaultDevices Run Failed")
		}
	}
	t.Logf("TestGetDefaultDevices Run Pass")
}

func createFile(filePath string) error {
	f, err := os.Create(filePath)
	f.Chmod(logChmod)
	f.Close()
	return err
}
