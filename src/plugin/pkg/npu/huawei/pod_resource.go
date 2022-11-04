// Copyright(C) 2022. Huawei Technologies Co.,Ltd. All rights reserved.

package huawei

import (
	"context"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kubernetes/pkg/kubelet/apis/podresources"
	"k8s.io/kubernetes/pkg/kubelet/apis/podresources/v1alpha1"

	"Ascend-device-plugin/src/plugin/pkg/npu/hwlog"
)

// PodDevice define device info in pod
type PodDevice struct {
	ResourceName string
	DeviceIds    []string
}

// PodResource implements the get pod resource info
type PodResource struct {
	conn    *grpc.ClientConn
	client  v1alpha1.PodResourcesListerClient
	restart bool
}

const (
	socketPath                 = "/var/lib/kubelet/pod-resources/kubelet.sock"
	defaultPodResourcesMaxSize = 1024 * 1024 * 16
	callTimeout                = time.Second
)

func (s *pluginAPI) getContainerResource(containerResource *v1alpha1.ContainerResources) (string, []string, error) {
	if containerResource == nil {
		return "", nil, fmt.Errorf("invalid container resource")
	}
	if len(containerResource.Devices) > maxDevicesNum {
		return "", nil, fmt.Errorf("the number of container device type %d exceeds the upper limit",
			len(containerResource.Devices))
	}
	var deviceIds []string
	resourceName := ""
	for _, containerDevice := range containerResource.Devices {
		if containerDevice == nil {
			hwlog.Warn("invalid container device")
			continue
		}
		if !strings.HasPrefix(containerDevice.ResourceName, resourceNamePrefix) {
			continue
		}
		if len(containerDevice.DeviceIds) > maxDevicesNum || len(containerDevice.DeviceIds) == 0 {
			return "", nil, fmt.Errorf("container device num %d exceeds the upper limit",
				len(containerDevice.DeviceIds))
		}
		if resourceName == "" {
			resourceName = containerDevice.ResourceName
		}
		for _, id := range containerDevice.DeviceIds {
			if len(id) > maxDeviceNameLen {
				return "", nil, fmt.Errorf("length of device name %d is invalid", len(id))
			}
			deviceIds = append(deviceIds, id)
		}
	}
	return resourceName, deviceIds, nil
}

func (s *pluginAPI) getDeviceFromPod(podResources *v1alpha1.PodResources) (string, []string, error) {
	if podResources == nil {
		return "", nil, fmt.Errorf("invalid podResources")
	}
	if len(podResources.Containers) > maxContainerLimit {
		return "", nil, fmt.Errorf("the number of containers %d exceeds the upper limit",
			len(podResources.Containers))
	}
	var podDevice []string
	var resourceName string
	total := 0
	for _, containerResource := range podResources.Containers {
		containerResourceName, containerDevices, err := s.getContainerResource(containerResource)
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
		if total > maxDevicesNum {
			return "", nil, fmt.Errorf("pod device num exceeds the upper limit")
		}
		podDevice = append(podDevice, containerDevices...)
	}
	return resourceName, podDevice, nil
}

func (s *pluginAPI) getPodResource() (map[string]PodDevice, error) {
	if !VerifyPath(socketPath) {
		return nil, fmt.Errorf("socket path verify failed")
	}
	client, conn, err := podresources.GetClient("unix://"+socketPath, callTimeout, defaultPodResourcesMaxSize)
	if err != nil {
		hwlog.Errorf("get pod resource client failed, %s", err.Error())
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), callTimeout)
	defer cancel()
	resp, err := client.List(ctx, &v1alpha1.ListPodResourcesRequest{})
	if err != nil {
		return nil, fmt.Errorf("list pod resource failed: %s", err.Error())
	}
	if len(resp.PodResources) > maxPodLimit {
		return nil, fmt.Errorf("the number of pods %d exceeds the upper limit", len(resp.PodResources))
	}
	device := make(map[string]PodDevice, 1)
	for _, pod := range resp.PodResources {
		if err = s.checkPodNameAndSpace(pod.Name, podNameMaxLength); err != nil {
			hwlog.Errorf("pod name syntax illegal, %s", err.Error())
			continue
		}
		if err = s.checkPodNameAndSpace(pod.Namespace, podNameSpaceMaxLength); err != nil {
			hwlog.Errorf("pod namespace syntax illegal, %s", err.Error())
			continue
		}
		resourceName, podDevice, err := s.getDeviceFromPod(pod)
		if err != nil || resourceName == "" || len(podDevice) == 0 {
			continue
		}
		device[pod.Namespace+"_"+pod.Name] = PodDevice{
			ResourceName: resourceName,
			DeviceIds:    podDevice,
		}
	}
	if err = conn.Close(); err != nil {
		hwlog.Errorf("stop connect failed, %s", err.Error())
	}
	return device, nil
}

func (s *pluginAPI) updatePodConfiguration() error {
	kubeClient := s.hps.kubeInteractor.clientset
	if kubeClient == nil {
		return fmt.Errorf("invalid kubeclient")
	}
	podDevice, err := s.getPodResource()
	if err != nil {
		return err
	}
	selector := fields.SelectorFromSet(fields.Set{"spec.nodeName": s.hps.kubeInteractor.nodeName,
		"status.phase": string(v1.PodRunning)})
	podList, err := kubeClient.CoreV1().Pods(v1.NamespaceAll).List(metav1.ListOptions{
		FieldSelector: selector.String()})
	if err != nil {
		return fmt.Errorf("list pod failed, err: %#v", err)
	}
	for _, pod := range podList.Items {
		if _, exist := pod.Annotations[podDeviceKey]; exist {
			continue
		}
		podKey := pod.Namespace + "_" + pod.Name
		podResource, exist := podDevice[podKey]
		if !exist {
			hwlog.Debugf("get %s klt device list failed, not in pod resource", podKey)
			continue
		}
		allocateDevice := sets.NewString()
		for _, id := range podResource.DeviceIds {
			allocateDevice.Insert(id)
		}
		ascendVisibleDevices := make(map[string]string, MaxVirtualDevNum)
		if err = s.getAscendVisiDevsWithVolcano(allocateDevice, &ascendVisibleDevices); err != nil {
			hwlog.Errorf("get ascend device ip failed, err: %#v", err)
			continue
		}
		if err = s.updatePodAnnotations(&pod, ascendVisibleDevices); err != nil {
			hwlog.Errorf("update pod annotation failed, err: %#v", err)
			continue
		}
		hwlog.Infof("update pod %s annotation success", podKey)
	}
	return nil
}
