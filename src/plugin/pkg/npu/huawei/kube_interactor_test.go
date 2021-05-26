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
	"Ascend-device-plugin/src/plugin/pkg/npu/huawei/mock_kubernetes"
	"Ascend-device-plugin/src/plugin/pkg/npu/huawei/mock_v1"
	"github.com/golang/mock/gomock"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"testing"
)

const nodeRunTime = 2

// TestPatchAnnotationOnNode for patch Annonation
func TestPatchAnnotationOnNode(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	node1 := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{Annotations: make(map[string]string)},
	}
	node1.Annotations[huaweiAscend910] = "Ascend910-1,Ascend910-2"
	mockK8s := mock_kubernetes.NewMockInterface(ctrl)
	mockV1 := mock_v1.NewMockCoreV1Interface(ctrl)
	mockNode := mock_v1.NewMockNodeInterface(ctrl)
	mockNode.EXPECT().Get(gomock.Any(), metav1.GetOptions{}).Return(node1, nil)
	mockNode.EXPECT().Patch(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(node1, nil)
	mockV1.EXPECT().Nodes().Return(mockNode).Times(nodeRunTime)
	mockK8s.EXPECT().CoreV1().Return(mockV1).Times(nodeRunTime)
	freeDevices := sets.NewString()
	freeDevices.Insert("Ascend910-1")
	freeDevices.Insert("Ascend910-5")
	fakeKubeInteractor := &KubeInteractor{
		clientset: mockK8s,
		nodeName:  "NODE_NAME",
	}
	err := fakeKubeInteractor.patchAnnotationOnNode(freeDevices, "")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("TestPatchAnnotationOnNode Run Pass")
}
