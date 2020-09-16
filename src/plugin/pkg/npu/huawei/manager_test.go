package huawei

import (
	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
	"os"
	"syscall"
	"testing"
	"time"
)

func TestSignalWatch(t *testing.T) {

	os.Create(serverSockfd)
	watcher := NewFileWatch()
	err := watcher.watchFile(pluginapi.DevicePluginPath)
	if err != nil {
		t.Errorf("failed to create file watcher. %v", err)
	}
	defer watcher.fileWatcher.Close()
	logger.Info("Starting OS signs watcher.")
	osSignChan := newSignWatcher(syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	hdm := HwDevManager{}
	hps := NewHwPluginServe(&hdm, "", "")
	var restart bool
	go deleteServerSocket(serverSockfd)
	restart = hdm.signalWatch(watcher.fileWatcher, osSignChan, restart, hps)
	if true == restart {
		t.Errorf("TestSignalWatch fales ")
	}
	t.Logf("TestSignalWatch Run Pass")

}

/*func TestPreStart(t *testing.T)  {
	hdm := HwDevManager{}
	hps := NewHwPluginServe(&hdm, "", "")
	preStart(hps,"")
	t.Logf("TestPreStart Run Pass")
}*/

func deleteServerSocket(serverSocket string) {
	time.Sleep(5 * time.Second)
	os.Remove(serverSocket)
}
