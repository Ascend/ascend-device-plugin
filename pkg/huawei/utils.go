/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2020-2022. All rights reserved.
 */
// Package huawei utils
package huawei

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"path"
	"strings"

	"github.com/fsnotify/fsnotify"
	"huawei.com/npu-exporter/hwlog"

	"Ascend-device-plugin/pkg/common"
)

const (
	phyDevTypeIndex = 0
	virDevTypeIndex = 1
)

// FileWatch is used to watch sock file
type FileWatch struct {
	fileWatcher *fsnotify.Watcher
}

// NewFileWatch is used to watch socket file
func NewFileWatch() *FileWatch {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil
	}
	return &FileWatch{
		fileWatcher: watcher,
	}
}

func (fw *FileWatch) watchFile(fileName string) error {
	_, err := os.Stat(fileName)
	if err != nil {
		return err
	}
	err = fw.fileWatcher.Add(fileName)
	if err != nil {
		return err
	}
	return nil
}

func newSignWatcher(osSigns ...os.Signal) chan os.Signal {
	// create signs chan
	signChan := make(chan os.Signal, 1)
	for _, sign := range osSigns {
		signal.Notify(signChan, sign)
	}

	return signChan
}

func createNetListen(pluginSocketPath string) (net.Listener, error) {
	if _, err := os.Stat(pluginSocketPath); err == nil {
		hwlog.RunLog.Infof("Found exist sock file, sockName is: %s, now remove it.", path.Base(pluginSocketPath))
		if err = os.Remove(pluginSocketPath); err != nil {
			return nil, err
		}
	}
	netListen, err := net.Listen("unix", pluginSocketPath)
	if err != nil {
		hwlog.RunLog.Errorf("device plugin start failed, err: %s", err.Error())
		return nil, err
	}
	err = os.Chmod(pluginSocketPath, common.SocketChmod)
	if err != nil {
		hwlog.RunLog.Errorf("change file: %s mode error", path.Base(pluginSocketPath))
	}
	return netListen, err
}

func get910TemplateName2DeviceTypeMap() map[string]string {
	return map[string]string{
		"vir16": chip910Core16C,
		"vir08": chip910Core8C,
		"vir04": chip910Core4C,
		"vir02": chip910Core2C,
	}
}

func get310PTemplateName2DeviceTypeMap() map[string]string {
	return map[string]string{
		"vir04":    chip310PCore4C,
		"vir04_3c": chip310PCore4C3Cpu,
		"vir02":    chip310PCore2C,
		"vir02_1c": chip310PCore2C1Cpu,
		"vir01":    chip310PCore1C,
	}
}

func getDevTypeByTemplateName(devType, template string) (string, bool) {
	templateList := map[string]string{}
	switch devType {
	case hiAIAscend910Prefix:
		templateList = get910TemplateName2DeviceTypeMap()
	case hiAIAscend310PPrefix:
		templateList = get310PTemplateName2DeviceTypeMap()
	default:
		return "", false
	}
	temp, exist := templateList[template]
	return temp, exist
}

func getVirtualDeviceType() map[string]struct{} {
	return map[string]struct{}{
		chip310PCore1C:     {},
		chip310PCore2C:     {},
		chip310PCore4C:     {},
		chip310PCore4C3Cpu: {},
		chip310PCore2C1Cpu: {},
		chip910Core2C:      {},
		chip910Core4C:      {},
		chip910Core8C:      {},
		chip910Core16C:     {},
	}
}

func getDeviceType(deviceName string) (string, error) {
	idSplit := strings.Split(deviceName, "-")
	// phy device like Ascend310P-0
	if len(idSplit) == common.PhyDeviceLen {
		switch idSplit[phyDevTypeIndex] {
		case hiAIAscend310Prefix, hiAIAscend910Prefix, hiAIAscend310PPrefix:
			return idSplit[phyDevTypeIndex], nil
		default:
		}
		return "", fmt.Errorf("id: %s is invalid", deviceName)
	}
	// virtual device like Ascend310P-2c.1cpu-103-0
	if len(idSplit) != common.VirDeviceLen {
		return "", fmt.Errorf("id: %s is invalid", deviceName)
	}
	devType := fmt.Sprintf("%s-%s", idSplit[phyDevTypeIndex], idSplit[virDevTypeIndex])
	if _, exist := getVirtualDeviceType()[devType]; exist {
		return devType, nil
	}
	return "", fmt.Errorf("id: %s is invalid", deviceName)
}
