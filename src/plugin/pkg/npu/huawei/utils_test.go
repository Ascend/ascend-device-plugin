/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2022. All rights reserved.
 */

package huawei

import (
	"os"
	"syscall"
	"testing"

	"github.com/smartystreets/goconvey/convey"
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

// TestGetDevTypeByTemplateName for test getDevTypeByTemplateName
func TestGetDevTypeByTemplateName(t *testing.T) {
	convey.Convey("TestGetDevTypeByTemplateName", t, func() {
		convey.Convey("devType is default type", func() {
			devType := hiAIAscend310Prefix
			template := ""
			vDevType, exist := getDevTypeByTemplateName(devType, template)
			convey.So(vDevType, convey.ShouldBeEmpty)
			convey.So(exist, convey.ShouldBeFalse)
		})
		convey.Convey("devType is 310P", func() {
			devType := hiAIAscend310PPrefix
			template := "vir04"
			vDevType, exist := getDevTypeByTemplateName(devType, template)
			convey.So(vDevType, convey.ShouldNotBeEmpty)
			convey.So(exist, convey.ShouldBeTrue)
		})
		convey.Convey("devType is 910", func() {
			devType := hiAIAscend910Prefix
			template := "vir04"
			vDevType, exist := getDevTypeByTemplateName(devType, template)
			convey.So(vDevType, convey.ShouldNotBeEmpty)
			convey.So(exist, convey.ShouldBeTrue)
		})
	})
}

// TestGetDeviceType for test getDeviceType
func TestGetDeviceType(t *testing.T) {
	convey.Convey("TestGetDeviceType", t, func() {
		convey.Convey("devType is valid physical device", func() {
			devName := hiAIAscend310Prefix + "-0"
			devType, err := getDeviceType(devName)
			convey.So(err, convey.ShouldBeNil)
			convey.So(devType, convey.ShouldEqual, hiAIAscend310Prefix)
		})
		convey.Convey("devType is invalid physical device", func() {
			devName := "AscendX10-0"
			_, err := getDeviceType(devName)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("devType is valid virtual device", func() {
			validDeviceType := getVirtualDeviceType()
			for vDeviceType := range validDeviceType {
				deviceName := vDeviceType + "-100-0"
				devType, err := getDeviceType(deviceName)
				convey.So(err, convey.ShouldBeNil)
				convey.So(devType, convey.ShouldEqual, vDeviceType)
			}
		})
		convey.Convey("devType is invalid virtual device", func() {
			devName := hiAIAscend310PPrefix + "-8c-100-0"
			_, err := getDeviceType(devName)
			convey.So(err, convey.ShouldNotBeNil)
		})
		convey.Convey("devType is invalid device name", func() {
			devName := hiAIAscend310PPrefix + "-8c-100-0-0"
			_, err := getDeviceType(devName)
			convey.So(err, convey.ShouldNotBeNil)
			devName = hiAIAscend310PPrefix
			_, err = getDeviceType(devName)
			convey.So(err, convey.ShouldNotBeNil)
		})
	})
}
