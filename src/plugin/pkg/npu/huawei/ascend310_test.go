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

//
import (
	"testing"

	"github.com/stretchr/testify/assert"

	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"
)

func TestAscend310(t *testing.T) {
	testManager := NewHwDevManager("ascend310", "/var/dlog")
	as := assert.New(t)
	as.NotNil(testManager)

	testManager.manager = NewHwPCIManager()

	testManager.allDevTypes = append(testManager.allDevTypes, "davinci-mini")
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
