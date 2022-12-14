/* Copyright(C) 2022. Huawei Technologies Co.,Ltd. All rights reserved.
   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

// Package server holds the implementation of registration to kubelet, k8s pod resource interface.
package server

import (
	"context"
	"fmt"
	"time"

	"huawei.com/mindx/common/hwlog"
	"k8s.io/kubernetes/pkg/kubelet/apis/podresources"
	"k8s.io/kubernetes/pkg/kubelet/apis/podresources/v1alpha1"

	"Ascend-device-plugin/pkg/common"
)

const (
	socketPath                 = "/var/lib/kubelet/pod-resources/kubelet.sock"
	defaultPodResourcesMaxSize = 1024 * 1024 * 16
	callTimeout                = 2 * time.Second
)

// Start starts the gRPC server, registers the pod resource with the Kubelet
func (pr *PodResource) Start(socketWatcher *common.FileWatch) error {
	if pr == nil {
		return fmt.Errorf("invalid interface receiver")
	}
	pr.Stop()
	realKubeletSockPath, isOk := common.VerifyPathAndPermission(socketPath)
	if !isOk {
		return fmt.Errorf("check kubelet socket file path failed")
	}
	var err error
	if err = socketWatcher.WatchFile(realKubeletSockPath); err != nil {
		hwlog.RunLog.Errorf("failed to create file watcher, err: %#v", err)
		return err
	}
	if pr.client, pr.conn, err = podresources.GetClient("unix://"+realKubeletSockPath, callTimeout,
		defaultPodResourcesMaxSize); err != nil {
		hwlog.RunLog.Errorf("get pod resource client failed, %#v", err)
		return err
	}
	hwlog.RunLog.Info("pod resource client init success.")
	return nil
}

func (pr *PodResource) getContainerResource(containerResource *v1alpha1.ContainerResources) (string, []string, error) {
	if containerResource == nil {
		return "", nil, fmt.Errorf("invalid container resource")
	}
	if len(containerResource.Devices) > common.MaxDevicesNum {
		return "", nil, fmt.Errorf("the number of container device type %d exceeds the upper limit",
			len(containerResource.Devices))
	}
	var deviceIds []string
	resourceName := ""
	for _, containerDevice := range containerResource.Devices {
		if containerDevice == nil {
			hwlog.RunLog.Warn("invalid container device")
			continue
		}
		if _, exist := common.GetAllDeviceInfoTypeList()[containerDevice.ResourceName]; !exist {
			continue
		}
		if len(containerDevice.DeviceIds) > common.MaxDevicesNum || len(containerDevice.DeviceIds) == 0 {
			return "", nil, fmt.Errorf("container device num %d exceeds the upper limit", len(containerDevice.DeviceIds))
		}
		if resourceName == "" {
			resourceName = containerDevice.ResourceName
		}
		for _, id := range containerDevice.DeviceIds {
			if len(id) > common.MaxDeviceNameLen {
				return "", nil, fmt.Errorf("length of device name %d is invalid", len(id))
			}
			deviceIds = append(deviceIds, id)
		}
	}
	return resourceName, deviceIds, nil
}

func (pr *PodResource) getDeviceFromPod(podResources *v1alpha1.PodResources) (string, []string, error) {
	if podResources == nil {
		return "", nil, fmt.Errorf("invalid podReousrces")
	}
	if len(podResources.Containers) > common.MaxContainerLimit {
		return "", nil, fmt.Errorf("the number of containers %d exceeds the upper limit", len(podResources.Containers))
	}
	var podDevice []string
	var resourceName string
	total := 0
	for _, containerResource := range podResources.Containers {
		containerResourceName, containerDevices, err := pr.getContainerResource(containerResource)
		if err != nil {
			return "", nil, err
		}
		if resourceName != "" && containerResourceName != resourceName {
			return "", nil, fmt.Errorf("only support one device type in a pod")
		}
		if resourceName == "" {
			resourceName = containerResourceName
		}
		total += len(containerDevices)
		if total > common.MaxDevicesNum {
			return "", nil, fmt.Errorf("pod device num exceeds the upper limit")
		}
		podDevice = append(podDevice, containerDevices...)
	}
	return resourceName, podDevice, nil
}

// GetPodResource call pod resource List interface, get pod resource info
func (pr *PodResource) GetPodResource() (map[string]PodDevice, error) {
	if pr == nil {
		return nil, fmt.Errorf("invalid interface receiver")
	}
	if pr.conn == nil || pr.client == nil {
		return nil, fmt.Errorf("client not init")
	}
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()
	resp, err := pr.client.List(ctx, &v1alpha1.ListPodResourcesRequest{})
	if err != nil {
		return nil, fmt.Errorf("list pod resource failed, err: %#v", err)
	}
	if resp == nil {
		return nil, fmt.Errorf("invalid list response")
	}
	if len(resp.PodResources) > common.MaxPodLimit {
		return nil, fmt.Errorf("the number of pods %d exceeds the upper limit", len(resp.PodResources))
	}
	device := make(map[string]PodDevice, 1)
	for _, pod := range resp.PodResources {
		if pod == nil {
			hwlog.RunLog.Warn("invalid pod")
			continue
		}
		if err := common.CheckPodNameAndSpace(pod.Name, common.PodNameMaxLength); err != nil {
			hwlog.RunLog.Warnf("pod name syntax illegal, err: %#v", err)
			continue
		}
		if err := common.CheckPodNameAndSpace(pod.Namespace, common.PodNameSpaceMaxLength); err != nil {
			hwlog.RunLog.Warnf("pod namespace syntax illegal, err: %#v", err)
			continue
		}
		resourceName, podDevice, err := pr.getDeviceFromPod(pod)
		if err != nil || resourceName == "" || len(podDevice) == 0 {
			continue
		}
		device[pod.Namespace+common.UnderLine+pod.Name] = PodDevice{
			ResourceName: resourceName,
			DeviceIds:    podDevice,
		}
	}
	return device, nil
}

// Stop the connection
func (pr *PodResource) Stop() {
	if pr == nil {
		hwlog.RunLog.Error("invalid interface receiver")
		return
	}
	if pr.conn != nil {
		if err := pr.conn.Close(); err != nil {
			hwlog.RunLog.Errorf("stop connect failed, err: %#v", err)
		}
		pr.conn = nil
		pr.client = nil
	}
}

// GetRestartFlag get restart flag
func (pr *PodResource) GetRestartFlag() bool {
	if pr == nil {
		hwlog.RunLog.Error("invalid interface receiver")
		return false
	}
	return pr.restart
}

// SetRestartFlag set restart flag
func (pr *PodResource) SetRestartFlag(flag bool) {
	if pr == nil {
		hwlog.RunLog.Error("invalid interface receiver")
		return
	}
	pr.restart = flag
}

// NewPodResource returns an initialized PodResource
func NewPodResource() *PodResource {
	return &PodResource{restart: true}
}
