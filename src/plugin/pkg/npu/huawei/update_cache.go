// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package huawei update cache for hps.devices
package huawei

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"go.uber.org/atomic"
	"k8s.io/apimachinery/pkg/apis/meta/v1"

	"huawei.com/npu-exporter/hwlog"
	"k8s.io/client-go/kubernetes"

	"Ascend-device-plugin/src/plugin/pkg/npu/common"
	"Ascend-device-plugin/src/plugin/pkg/npu/vnpumanager"
)

// ListenAnnotation ListenAnnotation
var ListenAnnotation = NewListenAnnotation()

// ListenAnnotation is ListenAnnotation
type ListenAnnotations struct {
	// WaitUpdateAnnotation is the annotation from patch result
	WaitUpdateAnnotation map[string]string
	// IsUpdateComplete is annotation update complete
	IsUpdateComplete *atomic.Bool
	// IsUpdateComplete is time task update complete or not
	IsTimingComplete *atomic.Bool
}

// NewListenAnnotation NewListenAnnotation
func NewListenAnnotation() *ListenAnnotations {
	return &ListenAnnotations{
		WaitUpdateAnnotation: make(map[string]string, 1),
		IsUpdateComplete:     atomic.NewBool(false),
		IsTimingComplete:     atomic.NewBool(false),
	}
}

// UpdateVNpuDevice to create and destroy virtual device
func UpdateVNpuDevice(hdm *HwDevManager, stopCh <-chan struct{}) {
	go func(hdm *HwDevManager) {
		for {
			if stopCh == nil {
				return
			}
			client, err := common.NewKubeClient(kubeConfig)
			if err != nil {
				fmt.Errorf("failed to create kube client: %v", err)
				return
			}
			isExecTimingUpdate(client)
			if err := TimingUpdate(hdm, client); err != nil {
				hwlog.RunLog.Errorf("current timing update failed, waiting for next time, err: %v", err)
			}
			time.Sleep(time.Minute)
		}
	}(hdm)
	go func() {
		for {
			if stopCh == nil {
				return
			}
			InformerCmUpdate(hdm)
		}
	}()
}

// TimingUpdate each minute exec update function
func TimingUpdate(hdm *HwDevManager, client *kubernetes.Clientset) error {
	if !ListenAnnotation.IsUpdateComplete.Load() {
		return nil
	}
	ListenAnnotation.IsUpdateComplete.Store(false)
	ListenAnnotation.IsTimingComplete.Store(true)
	defer ListenAnnotation.IsTimingComplete.Store(false)
	hwlog.RunLog.Infof("starting configMap timing update task")
	m.Lock()
	defer m.Unlock()
	var dcmiDevices []common.NpuDevice
	var dcmiDeviceTypes []string
	if err := hdm.manager.GetNPUs(&dcmiDevices, &dcmiDeviceTypes, hdm.manager.GetMatchingDeviType()); err != nil {
		return err
	}
	cardVNPUs, err := vnpumanager.GetVNpuCfg(client)
	if err != nil {
		return err
	}
	vnpumanager.DestroyVirtualDev(hdm.dmgr, dcmiDevices, cardVNPUs)
	vnpumanager.CreateVirtualDev(hdm.dmgr, cardVNPUs, hdm.runMode, client)
	updateHpsCache(hdm)
	hwlog.RunLog.Infof("configMap timing update task complete")
	return nil
}

// InformerCmUpdate update vnpu by configMap informer
func InformerCmUpdate(hdm *HwDevManager) {
	client, err := common.NewKubeClient(kubeConfig)
	if err != nil {
		fmt.Errorf("failed to create kube client: %v", err)
		return
	}
	NewConfigMapAgent(client, hdm)
}

func isExecTimingUpdate(client kubernetes.Interface) {
	for annotationTag, patchAnnotations := range ListenAnnotation.WaitUpdateAnnotation {
		if isSpecDev(annotationTag) && len(patchAnnotations) != 0 {
			nodeAnnotations, err := getAnnotationFromNode(client)
			if err != nil {
				hwlog.RunLog.Errorf("get annotation from node failed, err: %v", err)
				return
			}
			if isSortListEqual(patchAnnotations, nodeAnnotations[annotationTag]) {
				return
			}
		}
	}
	ListenAnnotation.IsUpdateComplete.Store(true)
	ListenAnnotation.WaitUpdateAnnotation = make(map[string]string, 1)
}

func isSortListEqual(patchAnnotations, nodeAnnotations string) bool {
	patchAnnotationsList := strings.Split(patchAnnotations, ",")
	nodeAnnotationsList := strings.Split(nodeAnnotations, ",")
	sort.SliceStable(patchAnnotationsList, func(i, j int) bool {
		return patchAnnotationsList[i] < patchAnnotationsList[j]
	})
	sort.SliceStable(nodeAnnotationsList, func(i, j int) bool {
		return nodeAnnotationsList[i] < nodeAnnotationsList[j]
	})
	return strings.Join(patchAnnotationsList, ",") != strings.Join(nodeAnnotationsList, ",")
}

func getAnnotationFromNode(kubeClient kubernetes.Interface) (map[string]string, error) {
	node, err := kubeClient.CoreV1().Nodes().Get(context.Background(), common.NodeName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("get node failed: %v", err)
	}
	return node.Annotations, nil
}

func isSpecDev(annotationTag string) bool {
	pwrSuffix := []string{hiAIAscend910Prefix, pwr2CSuffix, pwr4CSuffix, pwr8CSuffix, pwr16CSuffix,
		hiAIAscend710Prefix, chip710Core1C, chip710Core2C, chip710Core4C}
	for _, devType := range pwrSuffix {
		if annotationTag == fmt.Sprintf("%s%s", resourceNamePrefix, devType) {
			return true
		}
	}
	return false
}
