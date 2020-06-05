/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2019-2024. All rights reserved.
 * Description: ascend910_test.go
 * Create: 19-11-20 下午8:52
 */

package huawei

import (
	"testing"

	"github.com/stretchr/testify/assert"

	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1alpha"
)

func TestNPU(t *testing.T) {
	testManager := NewHwDevManager("ascend910", "/var/dlog")
	as := assert.New(t)
	as.NotNil(testManager)

	testManager.manager = NewHwPCIManager()

	testManager.allDevTypes = append(testManager.allDevTypes, "davinci-cloud")
	deviceType := testManager.allDevTypes[0]
	as.Equal(deviceType, "davinci-mini")

	device1 := npuDevice{
		devType: deviceType,
		pciID:   "0000",
		ID:      "0000",
		Health:  pluginapi.Healthy,
	}
	device2 := npuDevice{
		devType: deviceType,
		pciID:   "0001",
		ID:      "0001",
		Health:  pluginapi.Healthy,
	}
	device3 := npuDevice{
		devType: deviceType,
		pciID:   "0002",
		ID:      "0002",
		Health:  pluginapi.Healthy,
	}
	testManager.allDevs = append(testManager.allDevs, device1, device2, device3)

}
