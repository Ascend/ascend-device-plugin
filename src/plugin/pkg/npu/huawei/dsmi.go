/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2020-2021. All rights reserved.
 */

package huawei

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
	"huawei.com/npu-exporter/hwlog"
	"strings"
)

const (
	// ERROR return string error
	ERROR = "error"

	// RetError return error when the function failed
	retError = -1
	// UnRetError return error
	unretError = 100

	// dsmiMaxVdevNum is number of vdevice the devid spilt
	dsmiMaxVdevNum = 16

	// MaxErrorCodeCount is the max number of error code
	MaxErrorCodeCount = 128
)

// ChipInfo chip info
type ChipInfo struct {
	ChipType string
	ChipName string
	ChipVer  string
}

// CgoDsmiSubVDevInfo single VDevInfo info
type CgoDsmiSubVDevInfo struct {
	status uint32
	vdevid uint32
	vfid   uint32
	cid    uint64
	spec   CgoDsmiVdevSpecInfo
}

// CgoDsmiVdevSpecInfo is special info
type CgoDsmiVdevSpecInfo struct {
	coreNum  string
	reserved string
}

// CgoDsmiVDevInfo total VDevInfos info
type CgoDsmiVDevInfo struct {
	vDevNum             uint32
	coreNumUnused       uint32
	cgoDsmiSubVDevInfos []CgoDsmiSubVDevInfo
}

// DeviceMgrInterface interface for dsmi
type DeviceMgrInterface interface {
	GetDeviceCount() (int32, error)
	GetDeviceList(*[hiAIMaxDeviceNum]uint32) (int32, error)
	GetDeviceHealth(int32) (uint32, error)
	GetPhyID(uint32) (uint32, error)
	GetLogicID(uint32) (uint32, error)
	GetChipInfo(int32) (*ChipInfo, error)
	GetDeviceIP(int32) (string, error)
	GetVDevicesInfo(uint32) (CgoDsmiVDevInfo, error)
	GetDeviceErrorCode(uint32) error
	GetDeviceNetworkHealth(int32) (uint32, error)
	ShutDown()
}

// DeviceManager struct definition
type DeviceManager struct{}

func init() {
	C.dsmiInit_dl()
}

// NewDeviceManager new DeviceManager instance
func NewDeviceManager() *DeviceManager {
	return &DeviceManager{}
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
	return int32(count), nil
}

// GetDeviceList device get list
func (d *DeviceManager) GetDeviceList(devices *[hiAIMaxDeviceNum]uint32) (int32, error) {
	devNum, err := d.GetDeviceCount()
	if err != nil {
		return devNum, err
	}

	var ids [hiAIMaxDeviceNum]C.int
	if err := C.dsmi_list_device(&ids[0], C.int(devNum)); err != 0 {
		return retError, fmt.Errorf("unable to get device list, return error: %d", int32(err))
	}
	// transfer device list
	var i int32
	for i = 0; i < devNum; i++ {
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
	if uint32(logicID) > uint32(hiAIMaxDeviceNum) {
		return unretError, fmt.Errorf("get invalid logic id: %d", uint32(logicID))
	}

	return uint32(logicID), nil

}

// GetChipInfo get chipInfo
func (d *DeviceManager) GetChipInfo(logicID int32) (*ChipInfo, error) {
	var chipInfo C.struct_dsmi_chip_info_stru
	err := C.dsmi_get_chip_info(C.int(logicID), &chipInfo)
	if err != 0 {
		return nil, fmt.Errorf("get device Chip info failed, error code: %d", int32(err))
	}
	var name []rune
	var cType []rune
	var ver []rune
	name = convertToCharArr(name, chipInfo.chip_name)
	cType = convertToCharArr(cType, chipInfo.chip_type)
	ver = convertToCharArr(ver, chipInfo.chip_ver)
	chip := &ChipInfo{
		ChipName: string(name),
		ChipType: string(cType),
		ChipVer:  string(ver),
	}
	return chip, nil
}

func convertToCharArr(charArr []rune, cgoArr [maxChipName]C.uchar) []rune {
	for _, v := range cgoArr {
		if v != 0 {
			charArr = append(charArr, rune(v))
		}
	}
	return charArr
}

// GetDeviceIP get deviceIP
func (d *DeviceManager) GetDeviceIP(logicID int32) (string, error) {
	var portType C.int = 1
	var portID C.int
	var ipAddress [hiAIMaxDeviceNum]C.ip_addr_t
	var maskAddress [hiAIMaxDeviceNum]C.ip_addr_t
	var deviceIP []string

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
	C.dsmiShutDown()
}

// GetVDevicesInfo get the virtual device info by logicid
func (d *DeviceManager) GetVDevicesInfo(logicID uint32) (CgoDsmiVDevInfo, error) {
	var dsmiVDevInfo C.struct_dsmi_vdev_info
	err := C.dsmi_get_vdevice_info(C.uint(logicID), &dsmiVDevInfo)

	if err != 0 || int(dsmiVDevInfo.vdev_num) < 0 || int(dsmiVDevInfo.vdev_num) > dsmiMaxVdevNum {
		return CgoDsmiVDevInfo{}, fmt.Errorf("get virtual device info failed, error code is: %d "+
			"and vdev num is: %d", int32(err), int32(dsmiVDevInfo.vdev_num))
	}
	cgoDsmiVDevInfos := CgoDsmiVDevInfo{
		vDevNum:       uint32(dsmiVDevInfo.vdev_num),
		coreNumUnused: uint32(dsmiVDevInfo.spec_unused.core_num),
	}

	for i := uint32(0); i < cgoDsmiVDevInfos.vDevNum; i++ {
		coreNum := fmt.Sprintf("%v", dsmiVDevInfo.vdev[i].spec.core_num)
		cgoDsmiVDevInfos.cgoDsmiSubVDevInfos = append(cgoDsmiVDevInfos.cgoDsmiSubVDevInfos, CgoDsmiSubVDevInfo{
			status: uint32(dsmiVDevInfo.vdev[i].status),
			vdevid: uint32(dsmiVDevInfo.vdev[i].vdevid),
			vfid:   uint32(dsmiVDevInfo.vdev[i].vfid),
			cid:    uint64(dsmiVDevInfo.vdev[i].cid),
			spec: CgoDsmiVdevSpecInfo{
				coreNum: coreNum,
			},
		})
	}
	return cgoDsmiVDevInfos, nil
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
