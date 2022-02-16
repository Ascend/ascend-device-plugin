/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2020-2021. All rights reserved.
 */

package huawei

import (
	"github.com/fsnotify/fsnotify"
	"huawei.com/npu-exporter/hwlog"
	"net"
	"os"
	"os/signal"
	"path"
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
	err = os.Chmod(pluginSocketPath, socketChmod)
	if err != nil {
		hwlog.RunLog.Errorf("change file: %s mode error", path.Base(pluginSocketPath))
	}
	return netListen, err
}
