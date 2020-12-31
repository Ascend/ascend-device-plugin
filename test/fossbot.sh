#!/bin/sh
# Perform  test for  k8s-device-plugin
# Copyright @ Huawei Technologies CO., Ltd. 2020-2020. All rights reserved
export GO111MODULE="on"
export PATH=$GOPATH/bin:$PATH
CUR_DIR=$(dirname $(readlink -f $0))
TOP_DIR=$(realpath ${CUR_DIR}/..)
go get github.com/golang/mock/mockgen
MOCK_TOP=${TOP_DIR}/src/plugin/pkg/npu/huawei
mkdir -p "${MOCK_TOP}/mock_v1"
mkdir -p "${MOCK_TOP}/mock_kubernetes"
mkdir -p "${MOCK_TOP}/mock_kubelet_v1beta1"

mockgen k8s.io/client-go/kubernetes/typed/core/v1 CoreV1Interface >${MOCK_TOP}/mock_v1/corev1_mock.go
mockgen k8s.io/client-go/kubernetes Interface >${MOCK_TOP}/mock_kubernetes/k8s_interface_mock.go
mockgen k8s.io/client-go/kubernetes/typed/core/v1 NodeInterface >${MOCK_TOP}/mock_v1/node_interface_mock.go
mockgen k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1 DevicePlugin_ListAndWatchServer >${MOCK_TOP}/mock_kubelet_v1beta1/deviceplugin_mock.go
mockgen k8s.io/client-go/kubernetes/typed/core/v1 PodInterface >${MOCK_TOP}/mock_v1/pod_interface_mock.go

go mod vendor


