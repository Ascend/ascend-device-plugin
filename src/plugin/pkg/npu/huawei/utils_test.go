/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2022. All rights reserved.
 */

package huawei

import (
	"os"
	"syscall"
	"testing"
)

// TestCreateNetListen for createNetListen
func TestCreateNetListen(t *testing.T) {
	sockPath := "file not exist"
	if _, err := createNetListen(sockPath); err != nil {
		t.Errorf("netListen err %v", err)
	}

	sockPath = "/tmp/Ascend.sock"
	if _, err := createNetListen(sockPath); err != nil {
		t.Errorf("netListen err %v", err)
	}
	if _, err := os.Stat(sockPath); err != nil {
		t.Logf("fail to create sock %v", err)
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

// TestWatchFile for test watchFile
func TestWatchFile(t *testing.T) {
	watcher := NewFileWatch()
	if watcher == nil {
		t.Errorf("TestNewFileWatch is failed")
	}
	fileName := "file not exist"
	if err := watcher.watchFile(fileName); err != nil {
		t.Logf("watchFile failed")
	}

	fileName = "./watch_file"
	f, err := os.Create(fileName)
	if err != nil {
		t.Fatal("TestSignalWatch Run FAiled, reason is failed to create sock file")
	}
	defer f.Close()
	if err := watcher.watchFile(fileName); err != nil {
		t.Logf("watchFile failed")
	}
	t.Logf("TestNewFileWatch Run Pass")
}
