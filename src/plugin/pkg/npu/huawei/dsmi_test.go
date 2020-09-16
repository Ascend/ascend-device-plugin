package huawei

import (
	"github.com/golang/protobuf/ptypes/empty"
	"os"
	"testing"
)

type fakeDeviceManager struct{}

func newFakeDeviceManager() *fakeDeviceManager {
	return &fakeDeviceManager{}
}

func (d *fakeDeviceManager) EnableContainerService() error {
	return nil
}

// get ascend910 device quantity
func (d *fakeDeviceManager) GetDeviceCount() (int32, error) {
	return int32(8), nil
}

// device get list
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

// get device health by id
func (d *fakeDeviceManager) GetDeviceHealth(logicID int32) (uint32, error) {
	if logicID == 3 {
		return uint32(3), nil
	}
	return uint32(0), nil

}

// get physic id form logic id
func (d *fakeDeviceManager) GetPhyID(logicID uint32) (uint32, error) {
	return logicID, nil
}

// get logic id form physic id
func (d *fakeDeviceManager) GetLogicID(phyID uint32) (uint32, error) {
	return phyID, nil

}

func (d *fakeDeviceManager) GetChipInfo(logicID int32) (*ChipInfo, error) {
	chip := &ChipInfo{
		ChipName: "310",
		ChipType: "ASCEND",
		ChipVer:  "",
	}
	return chip, nil
}

func TestUnhealthyState(t *testing.T) {
	err := unhealthyState(1, uint32(1), "healthState", newFakeDeviceManager())
	if err != nil {
		t.Errorf("TestUnhealthyState Run Failed")
	}
	t.Logf("TestUnhealthyState Run Pass")
}

func TestGetLogicIDByName(t *testing.T) {
	var logicID int32
	err := getLogicIDByName("Ascend310-1", &logicID)
	if err != nil || 1 != logicID {
		t.Errorf("TestGetLogicIDByName Run Failed")
	}

	err = getLogicIDByName("Ascend310-1000", &logicID)
	if err == nil {
		t.Errorf("TestGetLogicIDByName Run Failed")
	}
	t.Logf("TestGetLogicIDByName Run Pass")
}

func TestGetDefaultDevices(t *testing.T) {
	if _, err := os.Stat(hiAIHDCDevice); err != nil {
		os.Create(hiAIHDCDevice)
	}

	if _, err := os.Stat(hiAIManagerDevice); err != nil {
		os.Create(hiAIManagerDevice)
	}

	if _, err := os.Stat(hiAISVMDevice); err == nil {
		os.Create(hiAISVMDevice)
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
