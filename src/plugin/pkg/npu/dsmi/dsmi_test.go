/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2020-2021. All rights reserved.
 */

package dsmi

import (
	"Ascend-device-plugin/src/plugin/pkg/npu/common"
	"Ascend-device-plugin/src/plugin/pkg/npu/huawei"
	"os"
	"testing"
)

// EnableContainerService for enableContainerService
func (d *fakeDeviceManager) EnableContainerService() error {
	return nil
}

// TestUnhealthyState for UnhealthyState
func TestUnhealthyState(t *testing.T) {
	err := huawei.UnhealthyState(1, uint32(3), "healthState", NewFakeDeviceManager())
	if err != nil {
		t.Errorf("TestUnhealthyState Run Failed")
	}
	t.Logf("TestUnhealthyState Run Pass")
}

// TestGetPhyIDByName for PhyIDByName
func TestGetPhyIDByName(t *testing.T) {
	phyID, err := huawei.GetPhyIDByName("Ascend310-3")
	if err != nil || unHealthyTestLogicID != phyID {
		t.Errorf("TestGetLogicIDByName Run Failed")
	}

	_, err = huawei.GetPhyIDByName("Ascend310-1000")
	if err == nil {
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

	if _, err := os.Stat(common.HiAISVMDevice); err == nil {
		if err = createFile(common.HiAISVMDevice); err != nil {
			t.Fatal("TestGetDefaultDevices Run Failed")
		}
	}
	var defaultDeivces []string
	err := huawei.GetDefaultDevices(&defaultDeivces)
	if err != nil {
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
		_, ok := defaultMap[str]
		if !ok {
			t.Errorf("TestGetDefaultDevices Run Failed")
		}
	}
	t.Logf("TestGetDefaultDevices Run Pass")
}

func createFile(filePath string) error {
	f, err := os.Create(filePath)
	defer f.Close()
	if err != nil {
		return err
	}
	if err := f.Chmod(common.SocketChmod); err != nil {
		return err
	}
	return nil
}
