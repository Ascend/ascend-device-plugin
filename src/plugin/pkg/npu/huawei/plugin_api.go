/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2019-2024. All rights reserved.
 * Description: plugin_api.go
 * Create: 19-11-20 下午9:06
 */

package huawei

import (
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"log"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"
)

type pluginAPI struct {
	hps *HwPluginServe
}

// Instance is for annoation
type Instance struct { // Instance
	PodName  string   `json:"pod_name"`  // pod Name
	ServerID string   `json:"server_id"` // serverdId
	Devices  []Device `json:"devices"`   // dev
}

// Device id for Instcance
type Device struct { // Device
	DeviceID string `json:"device_id"` // device id
	DeviceIP string `json:"device_ip"` // device ip
}

var ip string

// Register function is use to register k8s devicePlugin to kubelet.
func Register(k8sSocketPath, pluginSocket, resourceName string) error {
	conn, err := grpc.Dial(k8sSocketPath, grpc.WithInsecure(),
		grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
			return net.DialTimeout("unix", addr, timeout)
		}))
	if err != nil {
		logger.Error("connect to kubelet failed.", zap.String("err", err.Error()))
		return fmt.Errorf("connect to kubelet fail: %v", err)
	}
	defer conn.Close()
	client := pluginapi.NewRegistrationClient(conn)

	reqt := &pluginapi.RegisterRequest{
		Version:      pluginapi.Version,
		Endpoint:     pluginSocket,
		ResourceName: resourceName,
	}
	logger.Info("the device plugin api version is:", zap.String("apiVersion", pluginapi.Version))
	if _, err = client.Register(context.Background(), reqt); err != nil {
		return fmt.Errorf("register to kubelet fail: %v", err)
	}
	return nil
}

// GetDevicePluginOptions is Standard interface to kubelet.
func (s *pluginAPI) GetDevicePluginOptions(ctx context.Context, e *pluginapi.Empty) (*pluginapi.DevicePluginOptions,
	error) {
	return &pluginapi.DevicePluginOptions{}, nil
}

// ListAndWatch: if the server get stop signal ,the ListAndWatch should  stop,to be fix
func (s *pluginAPI) ListAndWatch(emtpy *pluginapi.Empty, stream pluginapi.DevicePlugin_ListAndWatchServer) error {

	logger.Info("device-plugin: ListAndWatch start")
	resp := new(pluginapi.ListAndWatchResponse)
	for _, dev := range s.hps.devices {
		resp.Devices = append(resp.Devices, &pluginapi.Device{ID: dev.ID, Health: dev.Health})
	}

	if err := sendDevToKubulet(resp, stream); err != nil {
		logger.Error("listAndWatch: send device info failed, please check kubelet status.",
			zap.String("err", err.Error()))
	}

	outs := false
	for {
		if outs {
			break
		}
		time.Sleep(sleepTime * time.Second)
		resp := new(pluginapi.ListAndWatchResponse)
		stated := false
		for id, dev := range s.hps.devices {
			state := s.hps.hdm.manager.GetDevState(id)
			if dev.Health != state {
				stated = true
				dev.Health = state
				s.hps.devices[id] = dev
			}
		}
		if !stated {
			// close log print
			logFlag = false
		}
		if stated {
			// turn on log print
			logFlag = true
			for _, dev := range s.hps.devices {
				resp.Devices = append(resp.Devices, &pluginapi.Device{ID: dev.ID, Health: dev.Health})
			}

			if err := sendDevToKubulet(resp, stream); err != nil {
				logger.Error("listAndWatch: send device info failed, please check kubelet status.",
					zap.String("err", err.Error()))
			}
		}
	}

	return nil
}

// Allocate is called by kubelet to mount device to k8s pod.
func (s *pluginAPI) Allocate(ctx context.Context, requests *pluginapi.AllocateRequest) (*pluginapi.AllocateResponse,
	error) {

	resps := new(pluginapi.AllocateResponse)
	logger.Info("allocate request:", zap.String("request", requests.String()))

	for _, rqt := range requests.ContainerRequests {
		resp := new(pluginapi.ContainerAllocateResponse)

		var devID []string
		var dlogMountPath string

		var majorID string
		var minorID string
		ascendVisibleDevices := ""

		// 从kubelet获取id
		for _, id := range rqt.DevicesIDs {
			err := getAscendDeviceID(id, &majorID, &minorID)
			if err != nil {
				logger.Error("getAscendDeviceID", zap.Error(err))
			}
			ascendVisibleDevices += majorID + ","
		}
		ascendVisibleDevices = addEnv(ascendVisibleDevices, &resp)
		if !UseAscendDocker {
			// 从kubelet获取id
			for _, id := range rqt.DevicesIDs {
				AddDev(id, s, resp, devID)
			}
			// mount default devices
			for _, d := range s.hps.defaultDevs {
				resp.Devices = append(resp.Devices, &pluginapi.DeviceSpec{
					HostPath:      d,
					ContainerPath: d,
					Permissions:   "mrw",
				})
			}
			// allocate device
			if err := AllocateAscendDev(devID, resp); err != nil {
				logger.Error("AllocateAscendDev failed", zap.String("err", err.Error()))
				return nil, fmt.Errorf("AllocateAscendDev failed, %s", err)
			}
		}
		if s.hps.hdm.runMode == "ascend910" {
			if err := s.hps.hdm.manager.GetLogPath(devID, s.hps.hdm.dlogPath, &dlogMountPath); err != nil {
				logger.Error("get logPath failed.", zap.String("err", err.Error()))
				dlogMountPath = s.hps.hdm.dlogPath
			}
			timeStr := time.Now().Format("20060102150405")
			rankID := "" + timeStr + "-0"
			s.mountfile(resp, dlogMountPath, rankID)
			addAnnotation(resp, ascendVisibleDevices)
		}
		resps.ContainerResponses = append(resps.ContainerResponses, resp)
		logger.Info("allocate responses:", zap.String("request", resps.String()))
	}
	return resps, nil
}

func addEnv(ascendVisibleDevices string, resp **pluginapi.ContainerAllocateResponse) string {
	// add env
	envs := make(map[string]string)
	ascendVisibleDevices = strings.TrimSuffix(ascendVisibleDevices, ",")
	envs[ascendVisibleDevicesEnv] = ascendVisibleDevices
	(*resp).Envs = envs
	return ascendVisibleDevices
}

func addAnnotation(resp *pluginapi.ContainerAllocateResponse, devices string) {
	// Annotations
	annotation := make(map[string]string)
	var instance Instance
	instance.PodName = "cloud-localhost-"

	instance.ServerID = "127.0.0.1"

	err := setDevices(&instance, devices)
	if err != nil {
		logger.Error("Add annotation failed", zap.String("error", err.Error()))
		return
	}
	instanceByte, err := json.Marshal(instance)
	if err != nil {
		logger.Error("Transfor marshal failed", zap.String("error", err.Error()))
		return
	}
	instanceInfo := string(instanceByte)
	annotation[podDeviceKey] = instanceInfo
	resp.Annotations = annotation
}

func setDevices(instance *Instance, devices string) error {
	idSplit := strings.Split(devices, ",")
	for _, deviceID := range idSplit {
		logicID64, err := strconv.ParseInt(deviceID, 10, 32)
		if err != nil {
			logger.Error(" Device id trasnsform failes ", zap.String("DeviceName", deviceID))
			return err
		}
		logicID := int32(logicID64)
		deviceIP, err := getDeviceIP(logicID)
		if err != nil {
			logger.Error("Get device ip failed:->", zap.Int("deviceId", int(logicID)), zap.Error(err))
			return err
		}
		var device Device
		device.DeviceID = deviceID
		device.DeviceIP = deviceIP
		instance.Devices = append(instance.Devices, device)
	}
	return nil
}

// PreStartContainer is Standard interface to kubelet with empty implement.
func (s *pluginAPI) PreStartContainer(ctx context.Context,
	r *pluginapi.PreStartContainerRequest) (*pluginapi.PreStartContainerResponse, error) {
	logger.Info("PreStart just call in UT.")
	return &pluginapi.PreStartContainerResponse{}, nil
}

func (s *pluginAPI) mountfile(resp *pluginapi.ContainerAllocateResponse, dlogMountPath string, randID string) {
	slogConfigPath := GetSlogConfigFilePath()
	resp.Mounts = append(resp.Mounts, &pluginapi.Mount{
		ContainerPath: slogConfigPath,
		HostPath:      slogConfigPath,
		ReadOnly:      true,
	})
	resp.Mounts = append(resp.Mounts, &pluginapi.Mount{
		ContainerPath: "/var/dlog",
		HostPath:      dlogMountPath,
		ReadOnly:      false,
	})
	// mount log format

	logpath := "/var/log/npu"
	hostLogPath := logpath + "/slog/container/" + randID
	resp.Mounts = append(resp.Mounts, &pluginapi.Mount{
		ContainerPath: logpath + "/slog",
		HostPath:      hostLogPath,
		ReadOnly:      false,
	})

	hostProfilingPath := logpath + "/profiling/container/" + randID
	resp.Mounts = append(resp.Mounts, &pluginapi.Mount{
		ContainerPath: logpath + "/profiling",
		HostPath:      hostProfilingPath,
		ReadOnly:      false,
	})

	hostDumpPath := logpath + "/dump/container/" + randID
	resp.Mounts = append(resp.Mounts, &pluginapi.Mount{
		ContainerPath: logpath + "/dump",
		HostPath:      hostDumpPath,
		ReadOnly:      false,
	})

	hostDockerSlogPath := logpath + "/docker_slog_" + randID
	resp.Mounts = append(resp.Mounts, &pluginapi.Mount{
		ContainerPath: "/usr/slog",
		HostPath:      hostDockerSlogPath,
		ReadOnly:      false,
	})
}

func sendDevToKubulet(resp *pluginapi.ListAndWatchResponse, stream pluginapi.DevicePlugin_ListAndWatchServer) error {
	logger.Info("ListAndWatch: send devices.", zap.String("resp", resp.String()))
	if err := stream.Send(resp); err != nil {
		return err
	}
	return nil
}

// AllocateAscendDev function is used to return para to kubelet when kubelet call allocate
func AllocateAscendDev(devID []string, resp *pluginapi.ContainerAllocateResponse) error {
	var repeat bool
	var devices []*pluginapi.DeviceSpec
	fmt.Printf("resp device info:%v \n", resp.Devices)
	for _, item := range resp.Devices {
		repeat = false

		for _, exact := range devices {
			if exact.HostPath == item.HostPath {
				repeat = true
				break
			}
		}

		if !repeat {
			devices = append(devices, item)
		}
	}

	for _, exact := range devices {
		match, err := regexp.MatchString("/dev/davinci*", exact.HostPath)
		if err != nil {
			logger.Error("match device id failed.", zap.String("err", err.Error()))
			continue
		}
		size := len(hiAIManagerDevice)
		if match && len(exact.HostPath) != size {
			phyID, err := recordPhyID(exact.HostPath)
			if err != nil {
				logger.Error("get phyID failed", zap.String("err", err.Error()))
			}
			logger.Debug("allocate ascend devices phyID", zap.Int32("phyID", phyID))

		}

		logger.Info("allocate ascend devices host path:", zap.String("hostPath", exact.HostPath))
	}

	resp.Devices = devices

	return nil
}

// AddDev is used to  specifies a host device to mount into a container
func AddDev(id string, s *pluginAPI, resp *pluginapi.ContainerAllocateResponse, devID []string) {
	// check if the device is in hps device map
	dev, ok := s.hps.devices[id]
	if !ok {
		log.Printf("plugin doesn't have device %s", id)
		return
	}

	if dev.Health != pluginapi.Healthy {
		logger.Warn("device is unhealthy", zap.String("deviceID", id))
		return
	}

	hostPathTmp := ""
	containerPathTmp := ""
	if err := s.hps.hdm.manager.GetDevPath(id, &hostPathTmp, &containerPathTmp); err != nil {
		logger.Error("invalid host and container path with legal device", zap.String("deviceID", id))
		return
	}

	resp.Devices = append(resp.Devices, &pluginapi.DeviceSpec{
		HostPath:      hostPathTmp,
		ContainerPath: containerPathTmp,
		Permissions:   "mrw",
	})

	devID = append(devID, dev.ID)
}

// GetSlogConfigFilePath is used to get slog path
func GetSlogConfigFilePath() string {
	return hiAISlogdConfig
}

func recordPhyID(pathString string) (int32, error) {

	logicID := pathString[len(hiAIDavinciPrefix):]
	tmpLogicID, err := strconv.Atoi(logicID)
	if err != nil {
		return -1, err
	}
	phyID, err := getPhyID(uint32(tmpLogicID))
	if err != nil {
		return -1, err
	}

	return int32(phyID), nil
}
