package huawei

import (
	"os"
	"syscall"
	"testing"
)

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

func TestNewSignWatcher(t *testing.T) {
	osSignChan := newSignWatcher(syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	if osSignChan == nil {
		t.Errorf("TestNewSignWatcher is failed")
	}
	t.Logf("TestNewSignWatcher Run Pass")
}

func TestNewFileWatch(t *testing.T) {
	watcher := NewFileWatch()
	if watcher == nil {
		t.Errorf("TestNewFileWatch is failed")
	}
	t.Logf("TestNewFileWatch Run Pass")
}
