package huawei

import (
	"os"
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
	t.Logf("Run Pass")
}
