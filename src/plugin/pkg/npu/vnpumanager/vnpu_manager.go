// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package vnpumanager, using for create and destroy device
package vnpumanager

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/kubernetes"

	"huawei.com/npu-exporter/hwlog"

	"Ascend-device-plugin/src/plugin/pkg/npu/common"
	"Ascend-device-plugin/src/plugin/pkg/npu/dsmi"
)

const (
	// maxRetryCount try to create or destroy virtual device
	maxRetryCount      = 2
	resourceNamePrefix = "huawei.com/"

	// For 710
	chip710c1 = resourceNamePrefix + "Ascend710-1c"
	chip710c2 = resourceNamePrefix + "Ascend710-2c"
	chip710c4 = resourceNamePrefix + "Ascend710-4c"

	// For 910
	chip910c2  = resourceNamePrefix + "Ascend910-2c"
	chip910c4  = resourceNamePrefix + "Ascend910-4c"
	chip910c8  = resourceNamePrefix + "Ascend910-8c"
	chip910c16 = resourceNamePrefix + "Ascend910-16c"
)

// DestroyVirtualDev destroy virtual devices
func DestroyVirtualDev(dmgr dsmi.DeviceMgrInterface, dcmiDevices []common.NpuDevice, cardVNPUs []CardVNPUs,
	runMode string, kubeClient kubernetes.Interface) {
	hwlog.RunLog.Infof("starting get virtual device which need to be destroy")
	needToBeDel := getNeedDestroyDev(dcmiDevices, cardVNPUs, runMode, kubeClient)
	for deviceID, virIDList := range needToBeDel {
		phyID, err := strconv.Atoi(deviceID)
		if err != nil {
			hwlog.RunLog.Errorf("deviceID %s convert to integer failed, err: %v\n", deviceID, err)
			continue
		}
		for _, virID := range virIDList {
			if err := destroyRetry(dmgr, phyID, virID); err != nil {
				hwlog.RunLog.Errorf("destroy virtual %s device failed, err: %v\n", virID, err)
				continue
			}
		}
	}
	hwlog.RunLog.Infof("free virtual device which need to be destroy success")
}

func destroyRetry(dmgr dsmi.DeviceMgrInterface, phyID int, virID string) error {
	retryCount := 0
	for {
		if retryCount > maxRetryCount {
			return fmt.Errorf("exceeded maximum number of retries")
		}
		logicID, err := dmgr.GetLogicID(uint32(phyID))
		if err != nil {
			return err
		}
		virIDCode, err := strconv.Atoi(virID)
		if err != nil {
			retryCount++
			hwlog.RunLog.Errorf("format virID failed, err: %v\n", err)
			continue
		}
		if err := dmgr.DestroyVirtualDevice(logicID, uint32(virIDCode)); err != nil {
			retryCount++
			hwlog.RunLog.Errorf("destroy virtual device %d failed, err: %v\n", virIDCode, err)
			continue
		}
		return nil
	}
}

// CreateVirtualDev create virtual devices
func CreateVirtualDev(dmgr dsmi.DeviceMgrInterface, cardVNPUs []CardVNPUs, runMode string,
	kubeClient kubernetes.Interface) {
	hwlog.RunLog.Infof("starting create virtual device which is cm adding")
	for _, cardVNPU := range cardVNPUs {
		// it's necessary, otherwise frequent calls to create interface may fail
		time.Sleep(time.Second)
		phyIDStr, virID, err := common.GetDeviceID(cardVNPU.CardName, "")
		if err != nil || virID != "" {
			hwlog.RunLog.Errorf("current card name invalid, err: %v", err)
			continue
		}
		if err := createRetry(dmgr, phyIDStr, runMode, cardVNPU, kubeClient); err != nil {
			hwlog.RunLog.Errorf("current card name invalid, err: %v", err)
			continue
		}
	}
	hwlog.RunLog.Infof("create virtual device which is cm added success")
}

func createRetry(dmgr dsmi.DeviceMgrInterface, phyIDStr, runMode string, cardVNPU CardVNPUs,
	kubeClient kubernetes.Interface) error {
	retryCount := 0
	for {
		if retryCount > maxRetryCount {
			return fmt.Errorf("exceeded maximum number of retries")
		}
		phyID, err := strconv.Atoi(phyIDStr)
		if err != nil {
			retryCount++
			hwlog.RunLog.Errorf("current card name change to int failed, err: %v", err)
			continue
		}
		logicID, err := dmgr.GetLogicID(uint32(phyID))
		if err != nil {
			retryCount++
			hwlog.RunLog.Errorf("get logic id failed, err: %v", err)
			continue
		}
		if err := dmgr.CreateVirtualDevice(logicID, runMode, getNeedCreateDev(cardVNPU, kubeClient, runMode,
			phyIDStr)); err != nil {
			retryCount++
			hwlog.RunLog.Errorf("create virtual device failed, err: %v", err)
			continue
		}
		return nil
	}
}

func getNeedCreateDev(cardVNPU CardVNPUs, kubeClient kubernetes.Interface, runMode, phyIDStr string) []string {
	pattern := `\d+c`
	reg := regexp.MustCompile(pattern)
	var reqDevs, allocDevs, createList []string
	for _, reqDev := range cardVNPU.Req {
		reqDevs = append(reqDevs, reg.FindString(reqDev))
	}
	annotateDevs, err := getAnnotationFromNode(kubeClient, runMode, phyIDStr)
	cutDevs := convertToSets(cardVNPU.Alloc).Union(convertToSets(annotateDevs))
	if err != nil {
		hwlog.RunLog.Warnf("query node annotation info failed, err: %v\n", err)
	}
	for allocDev := range cutDevs {
		allocDevs = append(allocDevs, reg.FindString(allocDev))
	}
	reqMap := getCoreAndCount(reqDevs)
	allocMap := getCoreAndCount(allocDevs)
	for devCore, count := range reqMap {
		allocCount, ok := allocMap[devCore]
		if !ok {
			createList = append(createList, getData(allocCount, devCore)...)
		}
		diff := count - allocCount
		if diff <= 0 {
			continue
		}
		createList = append(createList, getData(diff, devCore)...)
	}
	return createList
}

func getData(count int, devCore string) []string {
	var res []string
	for i := 0; i < count; i++ {
		res = append(res, devCore)
	}
	return res
}

func getCoreAndCount(devCoreList []string) map[string]int {
	coreAndCount := make(map[string]int, 1)
	for _, devCore := range devCoreList {
		coreAndCount[devCore]++
	}
	return coreAndCount
}

func getAnnotationFromNode(kubeClient kubernetes.Interface, runMode, phyIDStr string) ([]string, error) {
	nodeName := os.Getenv("NODE_NAME")
	if err := common.CheckNodeName(nodeName); err != nil {
		return nil, fmt.Errorf("check node name failed: %v", err)
	}
	node, err := kubeClient.CoreV1().Nodes().Get(context.Background(), nodeName, v1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("get node failed: %v", err)
	}
	return getAnnotation(node.Annotations, runMode, phyIDStr), nil
}

func getAnnotation(annotation map[string]string, runMode, phyIDStr string) []string {
	annotationTags := []string{chip910c2, chip910c4, chip910c8, chip910c16}
	if runMode == common.RunMode710 {
		annotationTags = []string{chip710c1, chip710c2, chip710c4}
	}
	var annotationCutted []string
	for _, annotationTag := range annotationTags {
		for _, devName := range strings.Split(annotation[annotationTag], ",") {
			phyID, _, err := common.GetDeviceID(devName, common.VirtualDev)
			if err != nil || phyIDStr != phyID {
				continue
			}
			annotationCutted = append(annotationCutted, devName)
		}
	}
	return annotationCutted
}

func convertToSets(devices []string) sets.String {
	devSet := sets.String{}
	for _, devType := range devices {
		devSet.Insert(devType)
	}
	return devSet
}

func getNeedDestroyDev(dcmiDevices []common.NpuDevice, cardVNPUs []CardVNPUs, runMode string,
	kubeClient kubernetes.Interface) map[string][]string {
	var needToBeDel = make(map[string][]string, 1)
	for _, npuDev := range dcmiDevices {
		deviceID, virID, err := common.GetDeviceID(npuDev.ID, common.VirtualDev)
		// if not is virtual device, do nothing
		if err != nil || virID == "" {
			continue
		}
		annotateDevs, err := getAnnotationFromNode(kubeClient, runMode, deviceID)
		if !isInVNpuCfg(npuDev.ID, cardVNPUs, getSpecCoreDevCount(dcmiDevices, deviceID), deviceID, annotateDevs) {
			// not found in configMap, means need to be deleted
			needToBeDel[deviceID] = append(needToBeDel[deviceID], virID)
		}
	}
	return needToBeDel
}

func getSpecCoreDevCount(dcmiDevices []common.NpuDevice, deviceID string) int {
	var count = 0
	for _, devName := range dcmiDevices {
		phyID, _, err := common.GetDeviceID(devName.ID, common.VirtualDev)
		if err != nil {
			continue
		}
		if phyID == deviceID {
			count++
		}
	}
	return count
}

func isInVNpuCfg(devName string, cardVNPUs []CardVNPUs, dcmiDevCount int, deviceID string, annotateDevs []string) bool {
	// deviceName format is "Ascend910-8c-101-0"
	// cardVNPUs format is "Cards": [{
	//              "CardName": "huawei.com/Ascend710-0",
	//              "Req": ["huawei.com/Ascend710-1c"],
	//              "Alloc": []
	//            }]
	for _, cardVPU := range cardVNPUs {
		if strings.Split(cardVPU.CardName, "-")[1] != deviceID {
			continue
		}
		if dcmiDevCount <= len(cardVPU.Req) {
			return true
		}
		usingDevs := convertToSets(cardVPU.Alloc).Union(convertToSets(annotateDevs))

		for usingDev := range usingDevs {
			if strings.Replace(usingDev, resourceNamePrefix, "", -1) == devName {
				return true
			}
		}
	}
	return false
}
