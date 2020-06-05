/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2019-2024. All rights reserved.
 * Description: pci.go
 * Create: 19-11-20 下午8:52
 */

package huawei

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"log"

	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"
)

const (
	vfioDevice = "/dev/vfio"
	// sysPCIDeviceDirectory is a folder contains all pci devices folder
	sysPCIDeviceDirectory = "/sys/bus/pci/devices"
	// SysIOMMUDirectory is contains all iommu_group folder
	sysIOMMUDirectory = "/sys/kernel/iommu_groups"
	// huaweiVendorID    = "8086"
	huaweiVendorID = "19e5"
	unbindFile     = "/sys/bus/pci/devices/%s/driver/unbind"
	newIDFile      = "/sys/bus/pci/drivers/vfio-pci/new_id"
)

var (
	npuTypeDict = map[string]string{
		"d100": "mini",
		"d200": "cloud",
	}
)

// HwPCIManager manages huawei pci devices.
type HwPCIManager struct{}

// NewHwPCIManager pci manager
func NewHwPCIManager() *HwPCIManager {
	return &HwPCIManager{}
}

// GetNPUs Discovers all HUAWEI NPU devices available on the local node by traversing
// /sys/bus/pci/devices directory to find the devices whose VENDOR ID are 19e5
func (hpm *HwPCIManager) GetNPUs(allDevices *[]npuDevice, allDeviceTypes *[]string) error {
	f, err := os.Open(sysPCIDeviceDirectory)
	if err != nil {
		log.Printf("Failed to open pcie devices\n")
		return err
	}

	pciDevices, err := f.Readdir(-1)
	defer f.Close()
	if err != nil {
		log.Printf("Failed to read directory: %+v\n", err)
		return err
	}

	recordTypes := make(map[string]int)

	for _, pciDevice := range pciDevices {
		bdf := pciDevice.Name()

		// Get vendor id
		vendorID, err := getPCIIdentifier(bdf, "vendor")
		if err != nil {
			log.Printf("Failed to get vendor id of %s: %+v\n", bdf, err)
			continue
		}
		log.Printf("+++++++++ vendorID=%s: huaweiVendorID=%s\n", vendorID, huaweiVendorID)
		// not Huawei NPU's
		if huaweiVendorID != vendorID {
			continue
		}

		// Huawei NPU's vendor id is 19e5
		deviceID, err := getPCIIdentifier(bdf, "device")
		if err != nil {
			log.Printf("Failed to get device id of %s: %+v\n", bdf, err)
			continue
		}

		// pciID format: vendorId deviceId
		pciID := fmt.Sprintf("%s %s", vendorID, deviceID)

		deviceType, exist := npuTypeDict[deviceID]
		if !exist {
			log.Printf("Unsupported NPU type, deviceID: %s\n", deviceID)
			continue
		}
		deviceType = fmt.Sprintf("davinci-%s", deviceType)

		device := npuDevice{
			devType: deviceType,
			pciID:   pciID,
			ID:      bdf,
			Health:  pluginapi.Healthy,
		}

		*allDevices = append(*allDevices, device)

		if _, exist := recordTypes[deviceType]; !exist {
			recordTypes[deviceType] = 0
			*allDeviceTypes = append(*allDeviceTypes, deviceType)
		}
	}

	return err
}

// GetDefaultDevs get default dev
func (hpm *HwPCIManager) GetDefaultDevs(defaultDevs *[]string) error {
	log.Printf("---> TODO: GetDefaultDevs\n")

	return nil
}

// GetDevState is used to get dev state
func (hpm *HwPCIManager) GetDevState(DeviceName string) string {
	return pluginapi.Healthy
}

// GetDevPath is used to get dev path
func (hpm *HwPCIManager) GetDevPath(id string, hostPath *string, containerPath *string) error {
	vifoGroupID, err := findVFIOGroupID(id)
	if err != nil {
		log.Printf("allocate device %s failed: %v", id, err)
		return err
	}
	*hostPath = "/dev/vfio/" + vifoGroupID
	*containerPath = ""
	return nil
}

func getPCIIdentifier(bdf, fileName string) (string, error) {
	filePath := filepath.Join(sysPCIDeviceDirectory, bdf, fileName)
	contents, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Printf("Failed to read %s: %+v\n", filePath, err)
		return "", err
	}

	id := strings.TrimSpace(string(contents))
	id = strings.TrimLeft(id, "0x")

	return id, nil
}

// Unbind NPU device from host driver
func unbind(bdf string) error {
	// Write BDF to unbind file
	file := fmt.Sprintf(unbindFile, bdf)
	err := writeFile(file, []byte(bdf))
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

// Bind NPU device to VFIO driver
func bindVFIO(pciID string) error {
	// Write PCI ID to new_id file
	if err := writeFile(newIDFile, []byte(pciID)); err != nil {
		return err
	}

	return nil
}

// Open file in write only mode and write bytes to it
// Maybe I can use ioutil.WriteFile
func writeFile(name string, data []byte) error {
	f, err := os.OpenFile(name, os.O_WRONLY, os.FileMode(0755))
	if err != nil {
		return err
	}

	defer f.Close()

	if _, err := f.Write(data); err != nil {
		return err
	}

	return nil
}

// ReadDirNoStat returns a string of files/directories contained
// in dirname without calling lstat on them.
func readDirNoSort(dirname string) ([]string, error) {
	if dirname == "" {
		dirname = "."
	}

	f, err := os.Open(dirname)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	return f.Readdirnames(-1)
}

func findVFIOGroupID(bdf string) (string, error) {
	log.Printf("---> TODO: FindVFIOGroupID\n")
	return "123", nil
}

// GetLogPath is used to get log patch
func (hpm *HwPCIManager) GetLogPath(devID []string, defaultLogPath string, newLogPath *string) error {
	*newLogPath = defaultLogPath
	log.Printf("log dir: %s.\n", *newLogPath)
	return nil
}

// GetDeviceCardID is used to get log patch
func (hpm *HwPCIManager) GetDeviceCardID(id string, majorID *string, minorID *string) error {
	if flag := strings.HasPrefix(id, hiAIAscend910Prefix); flag != true {
		return fmt.Errorf("id: %s is invalid", id)
	}

	idSplit := strings.Split(id, "-")

	if len(idSplit) < idSplitNum {
		return fmt.Errorf("id: %s is invalid", id)
	}

	*majorID = idSplit[1]

	if len(idSplit) > idSplitNum {
		*minorID = idSplit[2]
	} else {
		*minorID = ""
	}
	return nil
}
