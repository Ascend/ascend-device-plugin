/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2020-2021. All rights reserved.
 */

package huawei

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
)

const (
	unHealthyTestLogicID = 3
	serverSockFd         = "/var/lib/kubelet/device-plugins/davinci-mini.sock"
	serverSock310        = "/var/lib/kubelet/device-plugins/Ascend310.sock"
	maxAiCoreNum         = 32
	testAiCoreNum        = 20
	defaultVDevNum       = 0
	testVDevNum          = 2
	testComputeCoreNum   = 4

	// testPhyDevID use for ut, represent device id: 0,2,4,5,6,7
	testPhyDevID = "0234567"
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

//  GetDeviceNetworkHealth get device network health by id
func (d *fakeDeviceManager) GetDeviceNetworkHealth(logicID int32) (uint32, error) {
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
func (d *fakeDeviceManager) GetChipInfo(logicID int32) (string, error) {
	return "310", nil
}

// GetDeviceIP get deviceIP
func (d *fakeDeviceManager) GetDeviceIP(logicID int32) (string, error) {
	retIPAddress := fmt.Sprintf("%d.%d.%d.%d", 0, 0, 0, logicID)
	return retIPAddress, nil
}

// GetVDevicesInfo for fakeDeviceManager
func (d *fakeDeviceManager) GetVDevicesInfo(logicID uint32) (CgoDsmiVDevInfo, error) {
	var cgoDsmiVDevInfos CgoDsmiVDevInfo
	if strings.Contains(testPhyDevID, strconv.Itoa(int(logicID))) {
		cgoDsmiVDevInfos = CgoDsmiVDevInfo{
			vDevNum:       uint32(defaultVDevNum),
			coreNumUnused: uint32(maxAiCoreNum),
		}
		return cgoDsmiVDevInfos, nil
	}
	cgoDsmiVDevInfos = CgoDsmiVDevInfo{
		vDevNum:       uint32(testVDevNum),
		coreNumUnused: uint32(testAiCoreNum),
	}
	for i := 0; i < 2; i++ {
		coreNum := fmt.Sprintf("%d", testComputeCoreNum*(i+1))
		cgoDsmiVDevInfos.cgoDsmiSubVDevInfos = append(cgoDsmiVDevInfos.cgoDsmiSubVDevInfos, CgoDsmiSubVDevInfo{
			status: uint32(0),
			vdevid: uint32(int(logicID) + i),
			vfid:   uint32(int(logicID) + i),
			cid:    uint64(i),
			spec: CgoDsmiVdevSpecInfo{
				coreNum: coreNum,
			},
		})
	}
	return cgoDsmiVDevInfos, nil
}

// GetDeviceErrorCode get device error code
func (d *fakeDeviceManager) GetDeviceErrorCode(logicID uint32) error {
	return nil
}

// TestUnhealthyState for UnhealthyState
func TestUnhealthyState(t *testing.T) {
	err := unhealthyState(1, uint32(3), "healthState", newFakeDeviceManager())
	if err != nil {
		t.Errorf("TestUnhealthyState Run Failed")
	}
	t.Logf("TestUnhealthyState Run Pass")
}

// TestGetPhyIDByName for PhyIDByName
func TestGetPhyIDByName(t *testing.T) {
	phyID, err := getPhyIDByName("Ascend310-3")
	if err != nil || unHealthyTestLogicID != phyID {
		t.Errorf("TestGetLogicIDByName Run Failed")
	}

	_, err = getPhyIDByName("Ascend310-1000")
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
	defaultMap := make(map[string]string)
	defaultMap[hiAIHDCDevice] = ""
	defaultMap[hiAIManagerDevice] = ""
	defaultMap[hiAISVMDevice] = ""
	defaultMap[hiAi200RCEventSched] = ""
	defaultMap[hiAi200RCHiDvpp] = ""
	defaultMap[hiAi200RCLog] = ""
	defaultMap[hiAi200RCMemoryBandwidth] = ""
	defaultMap[hiAi200RCSVM0] = ""
	defaultMap[hiAi200RCTsAisle] = ""
	defaultMap[hiAi200RCUpgrade] = ""

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
	if err := f.Chmod(socketChmod); err != nil {
		return err
	}
	return nil
}
