// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package vnpumanager for parse configMap
package vnpumanager

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

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
)

// GetVNpuCfg get vnpu info, for create virtual device
func GetVNpuCfg(client *kubernetes.Clientset) ([]CardVNPUs, error) {
	cm, err := getVNpuCMFromK8s(client, CfgMapNamespace, CfgMapName)
	if err != nil {
		hwlog.RunLog.Errorf("failed to get vnpu configMap, err: %v\n", err)
		return nil, err
	}
	data, ok := cm.Data[common.VNpuCfgKey]
	if !ok {
		return nil, fmt.Errorf("configMap not exist")
	}
	cardVNPUs, err := GetCfgContent(data)
	if err != nil || len(cardVNPUs) == 0 {
		hwlog.RunLog.Errorf("failed to parse vnpu configMap or cm is nil, err: %v\n", err)
		return nil, err
	}
	return cardVNPUs, nil
}

func getVNpuCMFromK8s(client kubernetes.Interface, namespace, cmName string) (*v1.ConfigMap, error) {
	return client.CoreV1().ConfigMaps(namespace).Get(context.TODO(), cmName, metav1.GetOptions{})
}

// GetCfgContent get configMap
func GetCfgContent(data string) ([]CardVNPUs, error) {
	var vNpuCfg VNPUCM
	if err := json.Unmarshal([]byte(data), &vNpuCfg); err != nil {
		hwlog.RunLog.Errorf("ummarshal configMap data failed, err: %v", err)
		return nil, err
	}
	for _, vNpuCtn := range vNpuCfg.Nodes {
		cardVNPUs, isOk := getCurNodeCfg(vNpuCtn, common.NodeName)
		if isOk {
			return cardVNPUs, nil
		}
	}
	return nil, fmt.Errorf("not found node")
}

func getCurNodeCfg(vNpuCtn NodeVNPUs, nodeName string) ([]CardVNPUs, bool) {
	if nodeName == vNpuCtn.NodeName {
		return vNpuCtn.Cards, true
	}
	return nil, false
}

// ConvertCMToStruct convert configMap to struct
func ConvertCMToStruct(mtaObj metav1.Object) []CardVNPUs {
	mtaConfigMap, ok := mtaObj.(*v1.ConfigMap)
	if !ok {
		hwlog.RunLog.Errorf("convert meta data to configMap failed")
		return nil
	}
	if mtaConfigMap.Name != CfgMapName || mtaConfigMap.Namespace != CfgMapNamespace {
		return nil
	}
	if len(mtaConfigMap.Data) == 0 {
		hwlog.RunLog.Errorf("failed to find vnpu configMap data")
		return nil
	}
	cmData, ok := mtaConfigMap.Data[common.VNpuCfgKey]
	if !ok {
		hwlog.RunLog.Errorf("failed to find configMap VNPUCfg")
		return nil
	}
	cardVNPUs, err := GetCfgContent(cmData)
	if err != nil {
		hwlog.RunLog.Errorf("failed to parse vnpu configMap, err: %v\n", err)
		return nil
	}
	return cardVNPUs
}

// IsConfigMapChange is configMap change
func IsConfigMapChange(newCardNPUs, oldCardNPUs []CardVNPUs) bool {
	sort.SliceStable(newCardNPUs, func(i, j int) bool {
		return newCardNPUs[i].CardName < newCardNPUs[j].CardName
	})
	sort.SliceStable(oldCardNPUs, func(i, j int) bool {
		return oldCardNPUs[i].CardName < oldCardNPUs[j].CardName
	})
	for i := 0; i < len(newCardNPUs); i++ {
		if newCardNPUs[i].CardName != oldCardNPUs[i].CardName {
			return true
		}
		if isStringListEqual(newCardNPUs[i].Req, oldCardNPUs[i].Req) {
			return true
		}
		if isStringListEqual(newCardNPUs[i].Alloc, oldCardNPUs[i].Alloc) {
			return true
		}
	}
	return false
}

func isStringListEqual(newCard, oldCard []string) bool {
	if len(newCard) != len(oldCard) {
		return true
	}
	sort.SliceStable(newCard, func(i, j int) bool {
		return newCard[i] < newCard[j]
	})
	sort.SliceStable(oldCard, func(i, j int) bool {
		return oldCard[i] < oldCard[j]
	})
	for i := 0; i < len(newCard); i++ {
		if newCard[i] != oldCard[i] {
			return true
		}
	}
	return false
}
