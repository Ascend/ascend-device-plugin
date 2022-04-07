// Copyright (c) 2022. Huawei Technologies Co., Ltd. All rights reserved.

// Package vnpumanager using for create and destroy device
package vnpumanager

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

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
	ascendDevNameLen   = 2
	resourceNamePrefix = "huawei.com/"

	// For 710
	chip710   = "Ascend710"
	chip710c1 = resourceNamePrefix + "Ascend710-1c"
	chip710c2 = resourceNamePrefix + "Ascend710-2c"
	chip710c4 = resourceNamePrefix + "Ascend710-4c"

	// For 910
	chip910    = "Ascend910"
	chip910c2  = resourceNamePrefix + "Ascend910-2c"
	chip910c4  = resourceNamePrefix + "Ascend910-4c"
	chip910c8  = resourceNamePrefix + "Ascend910-8c"
	chip910c16 = resourceNamePrefix + "Ascend910-16c"
)

// DestroyVirtualDev destroy virtual devices
func DestroyVirtualDev(dmgr dsmi.DeviceMgrInterface, dcmiDevices []common.NpuDevice, cardVNPUs []CardVNPUs,
	nodeName string) {
	hwlog.RunLog.Infof("starting get virtual device which need to be destroy")
	needToBeDel := getNeedDestroyDev(dcmiDevices, cardVNPUs, nodeName)
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
			hwlog.RunLog.Errorf("destroy virtual device %d from %d failed, err: %v\n", virIDCode, phyID, err)
			continue
		} else {
			hwlog.RunLog.Infof("destroy virtual device %d from %d success", virIDCode, phyID)
		}
		return nil
	}
}

// CreateVirtualDev create virtual devices
func CreateVirtualDev(dmgr dsmi.DeviceMgrInterface, cardVNPUs []CardVNPUs, runMode string) {
	hwlog.RunLog.Infof("starting create virtual device which is cm adding")
	for _, cardVNPU := range cardVNPUs {
		phyIDStr, virID, err := common.GetDeviceID(cardVNPU.CardName, "")
		if err != nil || virID != "" {
			hwlog.RunLog.Errorf("current card name invalid, err: %v", err)
			continue
		}
		if err := createRetry(dmgr, phyIDStr, runMode, cardVNPU); err != nil {
			hwlog.RunLog.Errorf("phy device %s create virtual dev failed, err: %v", phyIDStr, err)
			continue
		}
	}
	hwlog.RunLog.Infof("create virtual device which is cm added success")
}

func createRetry(dmgr dsmi.DeviceMgrInterface, phyIDStr, runMode string, cardVNPU CardVNPUs) error {
	retryCount := 0
	hwlog.RunLog.Infof("create Req: %v + Alloc: %v on phyIDStr: %v", cardVNPU.Req, cardVNPU.Alloc, phyIDStr)
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
		createList := getNeedCreateDev(cardVNPU, dmgr, logicID)
		if len(createList) == 0 {
			return nil
		}
		if err := dmgr.CreateVirtualDevice(logicID, runMode, createList); err != nil {
			retryCount++
			hwlog.RunLog.Errorf("create virtual device failed, err: %v", err)
			continue
		}
		return nil
	}
}

func getNeedCreateDev(cardVNPU CardVNPUs, dmgr dsmi.DeviceMgrInterface, logicID uint32) []string {
	var reqDevs, dcmiDevs, createList []string
	vDevInfo, err := dmgr.GetVDevicesInfo(logicID)
	if err != nil {
		hwlog.RunLog.Errorf("get vDev list on %d failed", logicID)
		return nil
	}
	for _, vDev := range vDevInfo.CgoDsmiSubVDevInfos {
		dcmiDevs = append(dcmiDevs, fmt.Sprintf("%sc", vDev.Spec.CoreNum))
	}
	pattern := `\d+c`
	reg := regexp.MustCompile(pattern)
	for _, reqDev := range cardVNPU.Req {
		reqDevs = append(reqDevs, reg.FindString(reqDev))
	}
	reqMap := getCoreAndCount(reqDevs)
	allocMap := getCoreAndCount(dcmiDevs)
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
	hwlog.RunLog.Infof("get create list on %d, need create device %v", logicID, createList)
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
	node, err := kubeClient.CoreV1().Nodes().Get(context.Background(), common.NodeName, v1.GetOptions{})
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

func getNeedDestroyDev(dcmiDevices []common.NpuDevice, cardVNPUs []CardVNPUs, nodeName string) map[string][]string {
	var needToBeDel = make(map[string][]string, 1)
	for _, npuDev := range dcmiDevices {
		deviceID, virID, err := common.GetDeviceID(npuDev.ID, common.VirtualDev)
		// if not is virtual device, do nothing
		if err != nil || virID == "" {
			continue
		}
		// npuDev.ID format is "Ascend910-8c-101-0"
		if nodeName == "" || !isInVNpuCfg(npuDev.ID, deviceID, cardVNPUs) {
			// not found in configMap, means need to be deleted
			needToBeDel[deviceID] = append(needToBeDel[deviceID], virID)
		}
	}
	hwlog.RunLog.Infof("using complete, need remove devices: %v", needToBeDel)
	return needToBeDel
}

func isInVNpuCfg(devName, deviceID string, cardVNPUs []CardVNPUs) bool {
	for _, cardVPU := range cardVNPUs {
		nameList := strings.Split(cardVPU.CardName, "-")
		if len(nameList) != ascendDevNameLen {
			continue
		}
		if nameList[1] != deviceID {
			continue
		}
		hwlog.RunLog.Infof("destroy Req: %v + Alloc: %v + devName: %v + !isStable: %v", cardVPU.Req, cardVPU.Alloc,
			devName, !isReqAndAllocStable(cardVPU))
		if len(cardVPU.Req) == 0 {
			return false
		}
		if !isReqAndAllocStable(cardVPU) {
			return true
		}

		for _, usingDev := range cardVPU.Alloc {
			if usingDev == devName {
				return true
			}
		}
	}
	return false
}

func isReqAndAllocStable(cardVPU CardVNPUs) bool {
	for _, vNPU := range cardVPU.Alloc {
		if !common.IsVirtualDev(vNPU) {
			return false
		}
	}
	return len(cardVPU.Alloc) == len(cardVPU.Req)
}
