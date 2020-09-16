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
	"os"
	"syscall"
	"testing"
)

// TestCreateNetListen for createNetListen
func TestCreateNetListen(t *testing.T) {
	sockPath := "/tmp/Ascend.sock"
	_, err := createNetListen(sockPath)
	if err != nil {
		t.Errorf("netListen err %v", err)
	}
	if _, err := os.Stat(sockPath); err != nil {
		t.Errorf("fail to create sock %v", err)
	}
	t.Logf("TestCreateNetListen Run Pass")
}

// TestNewSignWatcher for create NewSignWatcher
func TestNewSignWatcher(t *testing.T) {
	osSignChan := newSignWatcher(syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	if osSignChan == nil {
		t.Errorf("TestNewSignWatcher is failed")
	}
	t.Logf("TestNewSignWatcher Run Pass")
}

// TestNewFileWatch for test FileWatch
func TestNewFileWatch(t *testing.T) {
	watcher := NewFileWatch()
	if watcher == nil {
		t.Errorf("TestNewFileWatch is failed")
	}
	t.Logf("TestNewFileWatch Run Pass")
}
