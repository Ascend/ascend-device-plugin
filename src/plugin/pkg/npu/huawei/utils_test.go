/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2020. All rights reserved.
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
