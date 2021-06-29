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
	"go.uber.org/zap"
	"net"
	"os"
	"os/signal"
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
		logger.Info("Found exist sock file,now remove it.", zap.String("sockName", pluginSocketPath))
		os.Remove(pluginSocketPath)
	}
	netListen, err := net.Listen("unix", pluginSocketPath)
	if err != nil {
		logger.Error("device plugin start failed.", zap.String("err", err.Error()))
		return nil, err
	}
	err = os.Chmod(pluginSocketPath, socketChmod)
	if err != nil {
		logger.Error("chmod error", zap.Error(err))
	}
	return netListen, err
}
