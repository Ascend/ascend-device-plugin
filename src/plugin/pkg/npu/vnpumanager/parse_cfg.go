// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package vnpumanager for parse configMap
package vnpumanager

import (
	"context"
	"encoding/json"
	"fmt"
	"huawei.com/npu-exporter/hwlog"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"Ascend-device-plugin/src/plugin/pkg/npu/common"
)

// CardVNPUs record virtual NPU Chip allocation information in one card.
type CardVNPUs struct {
	// NPU1
	CardName string
	// VNPU-8C
	Req []string
	// VNPU-8C-100-1
	Alloc []string
}

// NodeVNPUs record virtual NPU Chip allocation information in one node. Will be solidified into cm.
type NodeVNPUs struct {
	// NodeName preallocate node name.
	NodeName string
	// VNPUs record VNPUs Cutting and requirements in k8s.
	Cards []CardVNPUs
}

// VNPUCM record virtual NPU Chip allocation information in k8s. Will be solidified into cm.
type VNPUCM struct {
	// VNPUs record VNPUs Cutting and requirements in k8s.
	Nodes []NodeVNPUs
	// UpdateTime data update time.
	UpdateTime int64
	// CheckCode
	CheckCode uint32
}

const (
	// CfgMapName is configMap name
	CfgMapName = "mindx-dl-vnpu-manager"


	// CfgMapNamespace is VNPU configMap namespace
	CfgMapNamespace = "volcano-system"

	// CheckCodeZeroError is the error of VNPU configMap check code is zero
	CheckCodeZeroError = "check code is 0, do nothing"

	// NodeNameNotFoundError is the error when not found node name
	NodeNameNotFoundError = "not found node"
)

// GetVNpuCfg get vnpu info, for create virtual device
func GetVNpuCfg(client *kubernetes.Clientset) (string, []CardVNPUs, error) {
	cm, err := getVNpuCMFromK8s(client, CfgMapNamespace, CfgMapName)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get vnpu configMap, err: %v\n", err)
		return "", nil, err
	}
	data, ok := cm.Data[common.VNpuCfgKey]
	if !ok {
		return "", nil, fmt.Errorf("configMap not exist")
	}
	nodeName, cardVNPUs, err := GetCfgContent(data)
	if err != nil || len(cardVNPUs) == 0 {
		hwlog.RunLog.Warnf("failed to parse vnpu configMap or cm is nil, err: %v\n", err)
		return nodeName, nil, err
	}
	return nodeName, cardVNPUs, nil
}

func getVNpuCMFromK8s(client kubernetes.Interface, namespace, cmName string) (*v1.ConfigMap, error) {
	return client.CoreV1().ConfigMaps(namespace).Get(context.TODO(), cmName, metav1.GetOptions{})
}

// GetCfgContent get configMap
func GetCfgContent(data string) (string, []CardVNPUs, error) {
	var vNpuCfg VNPUCM
	if err := json.Unmarshal([]byte(data), &vNpuCfg); err != nil {
		hwlog.RunLog.Errorf("ummarshal configMap data failed, err: %v", err)
		return "", nil, err
	}
	if vNpuCfg.CheckCode == 0 {
		return "", nil, fmt.Errorf("%s", CheckCodeZeroError)
	}
	for _, vNpuCtn := range vNpuCfg.Nodes {
		cardVNPUs, isOk := getCurNodeCfg(vNpuCtn, common.NodeName)
		if isOk {
			return vNpuCtn.NodeName, cardVNPUs, nil
		}
	}
	return "", nil, fmt.Errorf("%s", NodeNameNotFoundError)
}

func getCurNodeCfg(vNpuCtn NodeVNPUs, nodeName string) ([]CardVNPUs, bool) {
	if nodeName == vNpuCtn.NodeName {
		return vNpuCtn.Cards, true
	}
	return nil, false
}