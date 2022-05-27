/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2020-2022. All rights reserved.
 */

// Package dsmi is driver dsmi interface related
package dsmi

// #cgo LDFLAGS: -ldl
/*
 #include <stddef.h>
#include <dlfcn.h>
#include <stdlib.h>

#include "dsmi_common_interface.h"

// dsmiHandle is the handle for dynamically loaded libdrvdsmi_host.so
void *dsmiHandle;
#define SO_NOT_FOUND  -99999
#define FUNCTION_NOT_FOUND  -99998
#define SUCCESS  0
#define ERROR_UNKNOWN  -99997
#define CALL_FUNC(name,...) if(name##_func==NULL){return FUNCTION_NOT_FOUND;}return name##_func(__VA_ARGS__);

int (*dsmi_get_device_count_func)(int *device_count);
int dsmi_get_device_count(int *device_count){
    CALL_FUNC(dsmi_get_device_count,device_count)
}

int (*dsmi_list_device_func)(int device_id_list[], int count);
int dsmi_list_device(int device_id_list[], int count){
	CALL_FUNC(dsmi_list_device,device_id_list,count)
}

int (*dsmi_get_device_health_func)(int device_id, unsigned int *phealth);
int dsmi_get_device_health(int device_id, unsigned int *phealth){
	CALL_FUNC(dsmi_get_device_health,device_id,phealth)
}

int (*dsmi_get_phyid_from_logicid_func)(unsigned int logicid, unsigned int *phyid);
int dsmi_get_phyid_from_logicid(unsigned int logicid, unsigned int *phyid){
	CALL_FUNC(dsmi_get_phyid_from_logicid,logicid,phyid)
}

int (*dsmi_get_logicid_from_phyid_func)(unsigned int phyid, unsigned int *logicid);
int dsmi_get_logicid_from_phyid(unsigned int phyid, unsigned int *logicid){
	CALL_FUNC(dsmi_get_logicid_from_phyid,phyid,logicid)
}

int (*dsmi_get_chip_info_func)(int device_id, struct dsmi_chip_info_stru *chip_info);
int dsmi_get_chip_info(int device_id, struct dsmi_chip_info_stru *chip_info){
	CALL_FUNC(dsmi_get_chip_info,device_id,chip_info)
}

int (*dsmi_get_device_ip_address_func)(int device_id, int port_type, int port_id, ip_addr_t *ip_address,
    ip_addr_t *mask_address);
int dsmi_get_device_ip_address(int device_id, int port_type, int port_id, ip_addr_t *ip_address,
    ip_addr_t *mask_address){
	CALL_FUNC(dsmi_get_device_ip_address,device_id,port_type,port_id,ip_address,mask_address)
}

int (*dsmi_get_vdevice_info_func)(unsigned int devid, struct dsmi_vdev_info *vdevice_info);
int dsmi_get_vdevice_info(unsigned int devid, struct dsmi_vdev_info *vdevice_info){
	CALL_FUNC(dsmi_get_vdevice_info,devid,vdevice_info)
}

int (*dsmi_get_device_errorcode_func)(int device_id, int *errorcount, unsigned int *perrorcode);
int dsmi_get_device_errorcode(int device_id, int *errorcount, unsigned int *perrorcode){
    CALL_FUNC(dsmi_get_device_errorcode,device_id,errorcount,perrorcode)
}

int (*dsmi_get_network_health_func)(int device_id, DSMI_NET_HEALTH_STATUS *presult);
int dsmi_get_network_health(int device_id, DSMI_NET_HEALTH_STATUS *presult){
    CALL_FUNC(dsmi_get_network_health,device_id,presult)
}

int (*dsmi_destroy_vdevice_func)(unsigned int devid, unsigned int vdevid);
int dsmi_destroy_vdevice(unsigned int devid, unsigned int vdevid){
	CALL_FUNC(dsmi_destroy_vdevice,devid,vdevid)
}

int (*dsmi_create_vdevice_func)(unsigned int devid, unsigned int vdev_id,
struct dsmi_create_vdev_res_stru *vdev_res,struct dsmi_create_vdev_result *vdev_result);
int dsmi_create_vdevice(unsigned int devid, unsigned int vdev_id,
struct dsmi_create_vdev_res_stru *vdev_res,struct dsmi_create_vdev_result *vdev_result){
	CALL_FUNC(dsmi_create_vdevice,devid,vdev_id,vdev_res,vdev_result)
}

// load .so files and functions
int dsmiInit_dl(void){
	dsmiHandle = dlopen("libdrvdsmi_host.so",RTLD_LAZY);
	if (dsmiHandle == NULL) {
		dsmiHandle = dlopen("libdrvdsmi.so",RTLD_LAZY);
	}
	if (dsmiHandle == NULL){
		return SO_NOT_FOUND;
	}

	dsmi_list_device_func = dlsym(dsmiHandle,"dsmi_list_device");

	dsmi_get_device_count_func = dlsym(dsmiHandle,"dsmi_get_device_count");

 	dsmi_get_device_health_func = dlsym(dsmiHandle,"dsmi_get_device_health");

	dsmi_get_phyid_from_logicid_func = dlsym(dsmiHandle,"dsmi_get_phyid_from_logicid");

	dsmi_get_logicid_from_phyid_func = dlsym(dsmiHandle,"dsmi_get_logicid_from_phyid");

	dsmi_get_chip_info_func = dlsym(dsmiHandle,"dsmi_get_chip_info");

	dsmi_get_device_ip_address_func = dlsym(dsmiHandle,"dsmi_get_device_ip_address");

	dsmi_get_vdevice_info_func = dlsym(dsmiHandle,"dsmi_get_vdevice_info");

	dsmi_get_device_errorcode_func = dlsym(dsmiHandle,"dsmi_get_device_errorcode");

	dsmi_get_network_health_func = dlsym(dsmiHandle,"dsmi_get_network_health");

	dsmi_destroy_vdevice_func = dlsym(dsmiHandle,"dsmi_destroy_vdevice");

	dsmi_create_vdevice_func = dlsym(dsmiHandle,"dsmi_create_vdevice");

	return SUCCESS;
}

int dsmiShutDown(void){
	if (dsmiHandle == NULL) {
   	 	return SUCCESS;
  	}
	return (dlclose(dsmiHandle) ? ERROR_UNKNOWN : SUCCESS);
}
*/
import "C"
import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"huawei.com/npu-exporter/hwlog"

	"Ascend-device-plugin/src/plugin/pkg/npu/common"
	"Ascend-device-plugin/src/plugin/pkg/npu/dcmi"
)

const (
	// ERROR return string error
	ERROR = "error"

	// RetError return error when the function failed
	retError = -1
	// UnRetError return error
	unretError = 100

	// dsmiMaxVdevNum is max number of vdevice, value is from driver specification
	dsmiMaxVdevNum = 16

	// dcmiMaxVdevNum is max number of 310P vdevice
	dcmiMaxVdevNum = 8

	// MaxErrorCodeCount is the max number of error code
	MaxErrorCodeCount = 128
)

var (
	driverInitOnce     sync.Once
	driverShutdownOnce sync.Once
)

// CgoDsmiSubVDevInfo single VDevInfo info
type CgoDsmiSubVDevInfo struct {
	Status uint32
	VDevID uint32
	VfID   uint32
	CID    uint64
	Spec   CgoDsmiVdevSpecInfo
}

// CgoDsmiVdevSpecInfo is special info
type CgoDsmiVdevSpecInfo struct {
	CoreNum  string
	Reserved string
}

// CgoDsmiVDevInfo total VDevInfos info
type CgoDsmiVDevInfo struct {
	VDevNum             uint32
	CoreNumUnused       uint32
	CoreCount           uint32
	CgoDsmiSubVDevInfos []CgoDsmiSubVDevInfo
}

// DeviceMgrInterface interface for dsmi
type DeviceMgrInterface interface {
	GetDeviceCount() (int32, error)
	GetDeviceList(*[hiAIMaxDeviceNum]uint32) (int32, error)
	GetDeviceHealth(int32) (uint32, error)
	GetPhyID(uint32) (uint32, error)
	GetLogicID(uint32) (uint32, error)
	GetChipInfo(int32) (string, error)
	GetDeviceIP(int32) (string, error)
	GetVDevicesInfo(uint32) (CgoDsmiVDevInfo, error)
	CreateVirtualDevice(uint32, string, []string) error
	DestroyVirtualDevice(uint32, uint32) error
	GetDeviceErrorCode(uint32) error
	GetDeviceNetworkHealth(int32) (uint32, error)
	ShutDown()
}

// DeviceManager struct definition
type DeviceManager struct {
	driverMgr *dcmi.DriverManager
}

// DriverInit initialize driver
func DriverInit() {
	driverInitOnce.Do(func() {
		C.dsmiInit_dl()
		dcmi.Init()
	})
}

// NewDeviceManager new DeviceManager instance
func NewDeviceManager() *DeviceManager {
	return &DeviceManager{driverMgr: dcmi.NewDriverManager()}
}

// GetDeviceCount get ascend910 device quantity
func (d *DeviceManager) GetDeviceCount() (int32, error) {
	var count C.int

	err := C.dsmi_get_device_count(&count)
	if err != 0 {
		return retError, fmt.Errorf("get device quantity failed, error code: %d", int32(err))
	}
	if int(count) < 0 || int(count) > hiAIMaxDeviceNum {
		return retError, fmt.Errorf("number of deviceis incorrect: %d", int(count))
	}
	if int(count) == 0 {
		return 0, fmt.Errorf("the number of available chips is 0")
	}
	return int32(count), nil
}

// GetDeviceList device get list
func (d *DeviceManager) GetDeviceList(devices *[hiAIMaxDeviceNum]uint32) (int32, error) {
	devNum, err := d.GetDeviceCount()
	if err != nil {
		return devNum, err
	}

	var ids [hiAIMaxDeviceNum]C.int
	// ids is an integer array, but its value cannot be less than 0
	if err := C.dsmi_list_device(&ids[0], C.int(devNum)); err != 0 {
		return retError, fmt.Errorf("unable to get device list, return error: %d", int32(err))
	}
	if devNum > hiAIMaxDeviceNum {
		return retError, fmt.Errorf("invalid device num: %d", devNum)
	}
	// transfer device list
	var i int32
	for i = 0; i < devNum; i++ {
		if int(ids[i]) < 0 {
			hwlog.RunLog.Errorf("device logical ID is less than 0")
			continue
		}
		(*devices)[i] = uint32(ids[i])
	}

	return devNum, nil
}

// GetDeviceHealth get device health by id
func (d *DeviceManager) GetDeviceHealth(logicID int32) (uint32, error) {
	var health C.uint

	err := C.dsmi_get_device_health(C.int(logicID), &health)
	if err != 0 {
		return unretError, fmt.Errorf("get device %d health state failed, error code: %d", logicID, int32(err))
	}

	return uint32(health), nil

}

// GetPhyID get physic id form logic id
func (d *DeviceManager) GetPhyID(logicID uint32) (uint32, error) {
	var phyID C.uint

	err := C.dsmi_get_phyid_from_logicid(C.uint(logicID), &phyID)
	if err != 0 {
		return unretError, fmt.Errorf("get phy id failed ,error code is: %d", int32(err))
	}
	if uint32(phyID) > uint32(hiAIMaxDeviceNum) {
		return unretError, fmt.Errorf("get invalid physical id: %d", uint32(phyID))
	}
	return uint32(phyID), nil
}

// GetLogicID get logic id form physic id
func (d *DeviceManager) GetLogicID(phyID uint32) (uint32, error) {
	var logicID C.uint

	err := C.dsmi_get_logicid_from_phyid(C.uint(phyID), &logicID)
	if err != 0 {
		return unretError, fmt.Errorf("get logic id failed ,error code is : %d", int32(err))
	}
	if uint32(logicID) >= uint32(hiAIMaxDeviceNum) {
		return unretError, fmt.Errorf("get invalid logic id: %d", uint32(logicID))
	}

	return uint32(logicID), nil

}

// GetChipInfo get chipInfo
func (d *DeviceManager) GetChipInfo(logicID int32) (string, error) {
	var chipInfo C.struct_dsmi_chip_info_stru
	if err := C.dsmi_get_chip_info(C.int(logicID), &chipInfo); err != 0 {
		return "", fmt.Errorf("get device Chip info failed, error code: %d", int32(err))
	}
	return convertToString(chipInfo.chip_name), nil
}

func convertToString(cgoArr [maxChipName]C.uchar) string {
	var chipName []rune
	for _, v := range cgoArr {
		if v != 0 {
			chipName = append(chipName, rune(v))
		}
	}
	return string(chipName)
}

// GetDeviceIP get deviceIP
func (d *DeviceManager) GetDeviceIP(logicID int32) (string, error) {
	var portType C.int = 1
	var portID C.int
	var ipAddress [hiAIMaxDeviceNum]C.ip_addr_t
	var maskAddress [hiAIMaxDeviceNum]C.ip_addr_t
	var deviceIP []string
	if logicID >= hiAIMaxDeviceNum {
		return ERROR, fmt.Errorf("getDevice IP address failed, error logicID")
	}
	retCode := C.dsmi_get_device_ip_address(C.int(logicID), portType, portID, &ipAddress[C.int(logicID)],
		&maskAddress[C.int(logicID)])
	if retCode != 0 {
		return ERROR, fmt.Errorf("getDevice IP address failed, error code: %d", int32(retCode))
	}

	unionPara := ipAddress[C.int(logicID)].u_addr
	for i := 0; i < deviceIPLength; i++ {
		deviceIP = append(deviceIP, fmt.Sprintf("%d", uint8(unionPara[i])))
	}
	return strings.Join(deviceIP, "."), nil
}

// ShutDown clean the dynamically loaded resource
func (d *DeviceManager) ShutDown() {
	driverShutdownOnce.Do(func() {
		C.dsmiShutDown()
		d.driverMgr.ShutDown()
	})
}

// GetVDevicesInfo get the virtual device info by logicid
func (d *DeviceManager) GetVDevicesInfo(logicID uint32) (CgoDsmiVDevInfo, error) {
	dcmiVDevInfo, err := d.driverMgr.GetVDeviceInfo(logicID)
	if err != nil {
		hwlog.RunLog.Error(err)
		return CgoDsmiVDevInfo{}, fmt.Errorf("get virtual device info failed, error is: %v "+
			"and vdev num is: %d", err, int32(dcmiVDevInfo.VDevNum))
	}
	cgoDsmiVDevInfos := CgoDsmiVDevInfo{
		VDevNum:       dcmiVDevInfo.VDevNum,
		CoreNumUnused: uint32(dcmiVDevInfo.CoreNumUnused),
	}
	usedCoreCount := uint32(0)
	for i := uint32(0); i < cgoDsmiVDevInfos.VDevNum; i++ {
		usedCoreCount += uint32(dcmiVDevInfo.CoreNum[i])
		cgoDsmiVDevInfos.CgoDsmiSubVDevInfos = append(cgoDsmiVDevInfos.CgoDsmiSubVDevInfos, CgoDsmiSubVDevInfo{
			Status: dcmiVDevInfo.Status[i],
			VDevID: dcmiVDevInfo.VDevID[i],
			VfID:   dcmiVDevInfo.VfID[i],
			CID:    dcmiVDevInfo.CID[i],
			Spec: CgoDsmiVdevSpecInfo{
				CoreNum: fmt.Sprintf("%v", dcmiVDevInfo.CoreNum[i]),
			},
		})
	}
	cgoDsmiVDevInfos.CoreCount = cgoDsmiVDevInfos.CoreNumUnused + usedCoreCount
	return cgoDsmiVDevInfos, nil
}

// CreateVirtualDevice to create a virtual device
func (d *DeviceManager) CreateVirtualDevice(logicID uint32, runMode string, vNPUs []string) error {
	cgoDsmiVDevInfos, err := d.GetVDevicesInfo(logicID)
	if err != nil {
		return err
	}
	switch runMode {
	case common.RunMode310P:
		return d.create310PVirDevice(cgoDsmiVDevInfos, logicID, vNPUs)
	case common.RunMode910:
		return d.create910VirDevice(cgoDsmiVDevInfos, logicID, vNPUs)
	default:
		return fmt.Errorf("not support runMode %s", runMode)
	}
}

func (d *DeviceManager) create910VirDevice(vDevInfos CgoDsmiVDevInfo, logicID uint32, vNPUs []string) error {
	if vDevInfos.CoreNumUnused <= 1 || vDevInfos.VDevNum > dsmiMaxVdevNum {
		return fmt.Errorf("the specification used to create 910 virtual device is error")
	}
	for i := 0; i < len(vNPUs); i++ {
		coreStr := strings.Replace(vNPUs[i], "c", "", -1)
		coreNum, err := strconv.Atoi(coreStr)
		if err != nil {
			continue
		}
		if _, err := d.driverMgr.CreateVDevice(logicID, uint32(coreNum)); err != nil {
			hwlog.RunLog.Error(err)
			return fmt.Errorf("create virtual device info failed, error is: %v", err)
		}
	}
	return nil
}

func (d *DeviceManager) create310PVirDevice(vDevInfos CgoDsmiVDevInfo, logicID uint32, vNPUs []string) error {
	if vDevInfos.CoreNumUnused < 1 || vDevInfos.VDevNum > dcmiMaxVdevNum {
		return fmt.Errorf("the specification used to create 310P virtual device is error")
	}
	for i := 0; i < len(vNPUs); i++ {
		coreStr := strings.Replace(vNPUs[i], "c", "", -1)
		coreNum, err := strconv.Atoi(coreStr)
		if err != nil {
			continue
		}
		if _, err = d.driverMgr.CreateVDevice(logicID, uint32(coreNum)); err != nil {
			hwlog.RunLog.Error(err)
			return fmt.Errorf("from %d create virtual device info failed, error is: %v", logicID, err)
		}
	}
	return nil
}

// DestroyVirtualDevice destroy spec virtual device
func (d *DeviceManager) DestroyVirtualDevice(logicID uint32, vDevID uint32) error {
	if err := d.driverMgr.DestroyVDevice(logicID, vDevID); err != nil {
		hwlog.RunLog.Error(err)
		return fmt.Errorf("destroy virtual device info failed, error is: %v", err)
	}
	return nil
}

// GetDeviceErrorCode get device error code
func (d *DeviceManager) GetDeviceErrorCode(logicID uint32) error {
	var errorCount C.int
	var pErrorCode [MaxErrorCodeCount]C.uint
	if err := C.dsmi_get_device_errorcode(C.int(logicID), &errorCount, &pErrorCode[0]); err != 0 {
		return fmt.Errorf("get device %d errorcode failed, error is: %d", logicID, int32(err))
	}

	if int32(errorCount) < 0 || int32(errorCount) > MaxErrorCodeCount {
		return fmt.Errorf("get wrong errorcode count, device: %d, errorcode count: %d", logicID, int32(errorCount))
	}

	hwlog.RunLog.Infof("get device error code, "+
		"logicID: %d, errorCount: %d, pErrorCode: %d", logicID, int(errorCount), int(pErrorCode[0]))

	return nil
}

// GetDeviceNetworkHealth the return value 'healthCode' is as follows:
// 0 : device network is right
// 6 : the IP address of the detected object may not be configured. We also think the network of devices is correct.
// other : the network of device is error
func (d *DeviceManager) GetDeviceNetworkHealth(logicID int32) (uint32, error) {
	var healthCode C.DSMI_NET_HEALTH_STATUS

	err := C.dsmi_get_network_health(C.int(logicID), &healthCode)
	if err != 0 {
		return unretError, fmt.Errorf("get wrong device network healthCode, device: %d, error is: %d, "+
			"healthCode is : %d", logicID, int32(err), uint32(healthCode))
	}

	return uint32(healthCode), nil
}
