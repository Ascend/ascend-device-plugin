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
		hwlog.Infof("Found exist sock file, sockName is: %s, now remove it.", path.Base(pluginSocketPath))
		os.Remove(pluginSocketPath)
	}
	netListen, err := net.Listen("unix", pluginSocketPath)
	if err != nil {
		hwlog.Errorf("device plugin start failed, err: %s", err.Error())
		return nil, err
	}
	err = os.Chmod(pluginSocketPath, socketChmod)
	if err != nil {
		hwlog.Errorf("change file: %s mode error", path.Base(pluginSocketPath))
	}
	return netListen, err
}
