/*
* Copyright(C) Huawei Technologies Co.,Ltd. 2022. All rights reserved.
 */

// Package dcmi is driver dcmi interface related
package dcmi

// #cgo LDFLAGS: -ldl
/*
#include <stddef.h>
#include <dlfcn.h>
#include <stdlib.h>
#include <stdio.h>

#include "dcmi_interface_api.h"

void *dcmiHandle;
#define SO_NOT_FOUND  -99999
#define FUNCTION_NOT_FOUND  -99998
#define SUCCESS  0
#define ERROR_UNKNOWN  -99997
#define	CALL_FUNC(name,...) if(name##_func==NULL){return FUNCTION_NOT_FOUND;}return name##_func(__VA_ARGS__);

// dcmi
int (*dcmi_init_func)();
int dcmi_init(){
	CALL_FUNC(dcmi_init)
}

int (*dcmi_get_card_num_list_func)(int *card_num,int *card_list,int list_length);
int dcmi_get_card_num_list(int *card_num,int *card_list,int list_length){
	CALL_FUNC(dcmi_get_card_num_list,card_num,card_list,list_length)
}

int (*dcmi_get_device_num_in_card_func)(int card_id,int *device_num);
int dcmi_get_device_num_in_card(int card_id,int *device_num){
	CALL_FUNC(dcmi_get_device_num_in_card,card_id,device_num)
}

int (*dcmi_get_device_logic_id_func)(int *device_logic_id,int card_id,int device_id);
int dcmi_get_device_logic_id(int *device_logic_id,int card_id,int device_id){
	CALL_FUNC(dcmi_get_device_logic_id,device_logic_id,card_id,device_id)
}

int (*dcmi_create_vdevice_func)(int card_id, int device_id, int vdev_id, const char *template_name,
	struct dcmi_create_vdev_out *out);
int dcmi_create_vdevice(int card_id, int device_id, int vdev_id, const char *template_name,
	struct dcmi_create_vdev_out *out){
	CALL_FUNC(dcmi_create_vdevice,card_id,device_id,vdev_id,template_name,out)
}

int (*dcmi_get_device_info_func)(int card_id, int device_id, enum dcmi_main_cmd main_cmd, unsigned int sub_cmd,
	void *buf, unsigned int *size);
int dcmi_get_device_info(int card_id, int device_id, enum dcmi_main_cmd main_cmd, unsigned int sub_cmd, void *buf,
	unsigned int *size){
	CALL_FUNC(dcmi_get_device_info,card_id,device_id,main_cmd,sub_cmd,buf,size)
}

int (*dcmi_set_destroy_vdevice_func)(int card_id,int device_id, unsigned int VDevid);
int dcmi_set_destroy_vdevice(int card_id,int device_id, unsigned int VDevid){
	CALL_FUNC(dcmi_set_destroy_vdevice,card_id,device_id,VDevid)
}

// load .so files and functions
int dcmiInit_dl(void){
	dcmiHandle = dlopen("libdcmi.so",RTLD_LAZY | RTLD_GLOBAL);
	if (dcmiHandle == NULL){
		fprintf (stderr,"%s\n",dlerror());
		return SO_NOT_FOUND;
	}

	dcmi_init_func = dlsym(dcmiHandle,"dcmi_init");

	dcmi_get_card_num_list_func = dlsym(dcmiHandle,"dcmi_get_card_num_list");

	dcmi_get_device_num_in_card_func = dlsym(dcmiHandle,"dcmi_get_device_num_in_card");

	dcmi_get_device_logic_id_func = dlsym(dcmiHandle,"dcmi_get_device_logic_id");

	dcmi_create_vdevice_func = dlsym(dcmiHandle,"dcmi_create_vdevice");

	dcmi_get_device_info_func = dlsym(dcmiHandle,"dcmi_get_device_info");

	dcmi_set_destroy_vdevice_func = dlsym(dcmiHandle,"dcmi_set_destroy_vdevice");

	return SUCCESS;
}

int dcmiShutDown(void){
	if (dcmiHandle == NULL) {
		return SUCCESS;
	}
	return (dlclose(dcmiHandle) ? ERROR_UNKNOWN : SUCCESS);
}
*/
import "C"
import (
	"fmt"
	"math"
	"unsafe"

	"huawei.com/npu-exporter/hwlog"
)

// MainCmd main command enum
type MainCmd uint32

// VDevMngSubCmd virtual device manager sub command
type VDevMngSubCmd uint32

const (
	// strError return string error
	strError = "error"

	// retError return error when the function failed
	retError = -1

	// unRetError return unsigned int error
	unretError = 100

	// dcmiMaxVdevNum is max number of vdevice, value is from driver specification
	dcmiMaxVdevNum = 32

	hiAIMaxCardNum = 16

	dcmiVDevResNameLen = 16

	vDeviceCreateTemplateNamePrefix = "vir0"

	// MainCmdVDevMng virtual device manager
	MainCmdVDevMng MainCmd = 52

	// VmngSubCmdGetVDevResource get virtual device resource info
	VmngSubCmdGetVDevResource VDevMngSubCmd = 0
	// VmngSubCmdGetTotalResource get total resource info
	VmngSubCmdGetTotalResource VDevMngSubCmd = 1
	// VmngSubCmdGetFreeResource get free resource info
	VmngSubCmdGetFreeResource VDevMngSubCmd = 2
)

// CgoDcmiCreateVDevOut create virtual device info
type CgoDcmiCreateVDevOut struct {
	VDevID     uint32
	PcieBus    uint32
	PcieDevice uint32
	PcieFunc   uint32
	VfgID      uint32
	Reserved   []uint8
}

// CgoDcmiBaseResource base resource info
type CgoDcmiBaseResource struct {
	token       uint64
	tokenMax    uint64
	taskTimeout uint64
	vfgID       uint32
	vipMode     uint8
	reserved    []uint8
}

// CgoDcmiComputingResource compute resource info
type CgoDcmiComputingResource struct {
	// accelator resource
	aic     float32
	aiv     float32
	dsa     uint16
	rtsq    uint16
	acsq    uint16
	cdqm    uint16
	cCore   uint16
	ffts    uint16
	sdma    uint16
	pcieDma uint16

	// memory resource, MB as unit
	memorySize uint64

	// id resource
	eventID  uint32
	notifyID uint32
	streamID uint32
	modelID  uint32

	// cpu resource
	topicScheduleAicpu uint16
	hostCtrlCPU        uint16
	hostAicpu          uint16
	deviceAicpu        uint16
	topicCtrlCPUSlot   uint16

	reserved []uint8
}

// CgoDcmiMediaResource media resource info
type CgoDcmiMediaResource struct {
	jpegd    float32
	jpege    float32
	vpc      float32
	vdec     float32
	pngd     float32
	venc     float32
	reserved []uint8
}

// CgoVDevQueryInfo virtual resource special info
type CgoVDevQueryInfo struct {
	name            string
	status          uint32
	isContainerUsed uint32
	vfid            uint32
	vfgID           uint32
	containerID     uint64
	base            CgoDcmiBaseResource
	computing       CgoDcmiComputingResource
	media           CgoDcmiMediaResource
}

// CgoVDevQueryStru virtual resource info
type CgoVDevQueryStru struct {
	vDevID    uint32
	queryInfo CgoVDevQueryInfo
}

// CgoDcmiSocFreeResource soc free resource info
type CgoDcmiSocFreeResource struct {
	vfgNum    uint32
	vfgBitmap uint32
	base      CgoDcmiBaseResource
	computing CgoDcmiComputingResource
	media     CgoDcmiMediaResource
}

// CgoDcmiSocTotalResource soc total resource info
type CgoDcmiSocTotalResource struct {
	vDevNum   uint32
	vDevID    []uint32
	vfgNum    uint32
	vfgBitmap uint32
	base      CgoDcmiBaseResource
	computing CgoDcmiComputingResource
	media     CgoDcmiMediaResource
}

// CgoVDevInfo virtual device infos
type CgoVDevInfo struct {
	VDevNum       uint32    // number of virtual devices
	CoreNumUnused float32   // number of unused cores
	Status        []uint32  // status of vitrual devices
	VDevID        []uint32  // id of virtual devices
	VfID          []uint32  // id
	CID           []uint64  // container id
	CoreNum       []float32 // aicore num for virtual device
}

// Init load symbol and initialize dcmi
func Init() {
	if err := C.dcmiInit_dl(); err != C.SUCCESS {
		fmt.Printf("dcmi lib load failed, error code: %d\n", int32(err))
		return
	}
	if err := C.dcmi_init(); err != C.SUCCESS {
		fmt.Printf("dcmi init failed, error code: %d\n", int32(err))
	}
}

// ShutDown clean the dynamically loaded resource
func ShutDown() {
	if err := C.dcmiShutDown(); err != C.SUCCESS {
		hwlog.RunLog.Error("dcmi shut down failed, error code: %d\n", int32(err))
	}
}

// GetCardList get card list
func GetCardList() (int32, []int32, error) {
	var ids [hiAIMaxCardNum]C.int
	var cNum C.int
	if err := C.dcmi_get_card_num_list(&cNum, &ids[0], hiAIMaxCardNum); err != 0 {
		return retError, nil, fmt.Errorf("get card list failed, error code: %d", int32(err))
	}
	// checking card's quantity
	if cNum <= 0 || cNum > hiAIMaxCardNum {
		return retError, nil, fmt.Errorf("get error card quantity: %d", int32(cNum))
	}
	var cardNum = int32(cNum)
	var i int32
	var cardIDList []int32
	for i = 0; i < cardNum; i++ {
		cardID := int32(ids[i])
		if cardID < 0 {
			errInfo := fmt.Errorf("get invalid card ID: %d", cardID)
			hwlog.RunLog.Error(errInfo)
			continue
		}
		cardIDList = append(cardIDList, cardID)
	}
	return cardNum, cardIDList, nil
}

// GetDeviceNumInCard get device number in the npu card
func GetDeviceNumInCard(cardID int32) (int32, error) {
	var deviceNum C.int
	if err := C.dcmi_get_device_num_in_card(C.int(cardID), &deviceNum); err != 0 {
		return retError, fmt.Errorf("get device count on the card failed, error code: %d", int32(err))
	}
	if deviceNum <= 0 {
		return retError, fmt.Errorf("get error device quantity: %d", int32(deviceNum))
	}
	return int32(deviceNum), nil
}

// GetDeviceLogicID get device logicID
func GetDeviceLogicID(cardID, deviceID int32) (uint32, error) {
	var logicID C.int
	if err := C.dcmi_get_device_logic_id(&logicID, C.int(cardID), C.int(deviceID)); err != 0 {
		return unretError, fmt.Errorf("get logicID failed, error code: %d", int32(err))
	}

	// check whether logicID is invalid
	if logicID < 0 || uint32(logicID) > uint32(math.MaxInt8) {
		return unretError, fmt.Errorf("get invalid logicID: %d", int32(logicID))
	}
	return uint32(logicID), nil
}

// SetDestroyVirtualDevice destroy virtual device
func SetDestroyVirtualDevice(cardID, deviceID int32, vDevID uint32) error {
	if err := C.dcmi_set_destroy_vdevice(C.int(cardID), C.int(deviceID), C.uint(vDevID)); err != 0 {
		return fmt.Errorf("destroy virtual device failed, error code: %d", int32(err))
	}
	return nil
}

func convertCreateVDevOut(cCreateVDevOut C.struct_dcmi_create_vdev_out) CgoDcmiCreateVDevOut {
	cgoCreateVDevOut := CgoDcmiCreateVDevOut{
		VDevID:     uint32(cCreateVDevOut.vdev_id),
		PcieBus:    uint32(cCreateVDevOut.pcie_bus),
		PcieDevice: uint32(cCreateVDevOut.pcie_device),
		PcieFunc:   uint32(cCreateVDevOut.pcie_func),
		VfgID:      uint32(cCreateVDevOut.vfg_id),
	}
	return cgoCreateVDevOut
}

// CreateVirtualDevice create virtual device
func CreateVirtualDevice(cardID, deviceID, vDevID int32, aiCore uint32) (CgoDcmiCreateVDevOut, error) {
	templateName := fmt.Sprintf("%s%d", vDeviceCreateTemplateNamePrefix, aiCore)
	cTemplateName := C.CString(templateName)
	defer C.free(unsafe.Pointer(cTemplateName))

	var createVDevOut C.struct_dcmi_create_vdev_out
	if err := C.dcmi_create_vdevice(C.int(cardID), C.int(deviceID), C.int(vDevID), (*C.char)(cTemplateName),
		&createVDevOut); err != 0 {
		return CgoDcmiCreateVDevOut{}, fmt.Errorf("create vdevice failed, error is: %d", int32(err))
	}

	return convertCreateVDevOut(createVDevOut), nil
}

func convertToCharArr(charArr []rune, cgoArr [dcmiVDevResNameLen]C.char) []rune {
	for _, v := range cgoArr {
		if v != 0 {
			charArr = append(charArr, rune(v))
		}
	}
	return charArr
}

func convertBaseResource(cBaseResource C.struct_dcmi_base_resource) CgoDcmiBaseResource {
	baseResource := CgoDcmiBaseResource{
		token:       uint64(cBaseResource.token),
		tokenMax:    uint64(cBaseResource.token_max),
		taskTimeout: uint64(cBaseResource.task_timeout),
		vfgID:       uint32(cBaseResource.vfg_id),
		vipMode:     uint8(cBaseResource.vip_mode),
	}
	return baseResource
}

func convertComputingResource(cComputingResource C.struct_dcmi_computing_resource) CgoDcmiComputingResource {
	computingResource := CgoDcmiComputingResource{
		aic:                float32(cComputingResource.aic),
		aiv:                float32(cComputingResource.aiv),
		dsa:                uint16(cComputingResource.dsa),
		rtsq:               uint16(cComputingResource.rtsq),
		acsq:               uint16(cComputingResource.acsq),
		cdqm:               uint16(cComputingResource.cdqm),
		cCore:              uint16(cComputingResource.c_core),
		ffts:               uint16(cComputingResource.ffts),
		sdma:               uint16(cComputingResource.sdma),
		pcieDma:            uint16(cComputingResource.pcie_dma),
		memorySize:         uint64(cComputingResource.memory_size),
		eventID:            uint32(cComputingResource.event_id),
		notifyID:           uint32(cComputingResource.notify_id),
		streamID:           uint32(cComputingResource.stream_id),
		modelID:            uint32(cComputingResource.model_id),
		topicScheduleAicpu: uint16(cComputingResource.topic_schedule_aicpu),
		hostCtrlCPU:        uint16(cComputingResource.host_ctrl_cpu),
		hostAicpu:          uint16(cComputingResource.host_aicpu),
		deviceAicpu:        uint16(cComputingResource.device_aicpu),
		topicCtrlCPUSlot:   uint16(cComputingResource.topic_ctrl_cpu_slot),
	}
	return computingResource
}

func convertMediaResource(cMediaResource C.struct_dcmi_media_resource) CgoDcmiMediaResource {
	mediaResource := CgoDcmiMediaResource{
		jpegd: float32(cMediaResource.jpegd),
		jpege: float32(cMediaResource.jpege),
		vpc:   float32(cMediaResource.vpc),
		vdec:  float32(cMediaResource.vdec),
		pngd:  float32(cMediaResource.pngd),
		venc:  float32(cMediaResource.venc),
	}
	return mediaResource
}

func convertVDevQuertyInfo(cVDevQueryInfo C.struct_dcmi_vdev_query_info) CgoVDevQueryInfo {
	var name []rune
	name = convertToCharArr(name, cVDevQueryInfo.name)

	vDevQueryInfo := CgoVDevQueryInfo{
		name:            string(name),
		status:          uint32(cVDevQueryInfo.status),
		isContainerUsed: uint32(cVDevQueryInfo.is_container_used),
		vfid:            uint32(cVDevQueryInfo.vfid),
		vfgID:           uint32(cVDevQueryInfo.vfg_id),
		containerID:     uint64(cVDevQueryInfo.container_id),
		base:            convertBaseResource(cVDevQueryInfo.base),
		computing:       convertComputingResource(cVDevQueryInfo.computing),
		media:           convertMediaResource(cVDevQueryInfo.media),
	}
	return vDevQueryInfo
}

func convertVDevQueryStru(cVDevQueryStru C.struct_dcmi_vdev_query_stru) CgoVDevQueryStru {
	vDevQueryStru := CgoVDevQueryStru{
		vDevID:    uint32(cVDevQueryStru.vdev_id),
		queryInfo: convertVDevQuertyInfo(cVDevQueryStru.query_info),
	}
	return vDevQueryStru
}

// GetDeviceVDevResource get virtual device resource info
func GetDeviceVDevResource(cardID, deviceID int32, vDevID uint32) (CgoVDevQueryStru, error) {
	var cMainCmd = C.enum_dcmi_main_cmd(MainCmdVDevMng)
	subCmd := VmngSubCmdGetVDevResource
	var vDevResource C.struct_dcmi_vdev_query_stru
	size := C.uint(unsafe.Sizeof(vDevResource))
	vDevResource.vdev_id = C.uint(vDevID)
	if err := C.dcmi_get_device_info(C.int(cardID), C.int(deviceID), cMainCmd, C.uint(subCmd),
		unsafe.Pointer(&vDevResource), &size); err != 0 {
		return CgoVDevQueryStru{}, fmt.Errorf("get device info failed, error is: %d", int32(err))
	}
	return convertVDevQueryStru(vDevResource), nil
}

func convertSocTotalResource(cSocTotalResource C.struct_dcmi_soc_total_resource) CgoDcmiSocTotalResource {
	socTotalResource := CgoDcmiSocTotalResource{
		vDevNum:   uint32(cSocTotalResource.vdev_num),
		vfgNum:    uint32(cSocTotalResource.vfg_num),
		vfgBitmap: uint32(cSocTotalResource.vfg_bitmap),
		base:      convertBaseResource(cSocTotalResource.base),
		computing: convertComputingResource(cSocTotalResource.computing),
		media:     convertMediaResource(cSocTotalResource.media),
	}
	for i := uint32(0); i < uint32(cSocTotalResource.vdev_num) && i < dcmiMaxVdevNum; i++ {
		socTotalResource.vDevID = append(socTotalResource.vDevID, uint32(cSocTotalResource.vdev_id[i]))
	}
	return socTotalResource
}

// GetDeviceTotalResource get device total resource info
func GetDeviceTotalResource(cardID, deviceID int32) (CgoDcmiSocTotalResource, error) {
	var cMainCmd = C.enum_dcmi_main_cmd(MainCmdVDevMng)
	subCmd := VmngSubCmdGetTotalResource
	var totalResource C.struct_dcmi_soc_total_resource
	size := C.uint(unsafe.Sizeof(totalResource))
	if err := C.dcmi_get_device_info(C.int(cardID), C.int(deviceID), cMainCmd, C.uint(subCmd),
		unsafe.Pointer(&totalResource), &size); err != 0 {
		return CgoDcmiSocTotalResource{}, fmt.Errorf("get device info failed, error is: %d", int32(err))
	}
	if uint32(totalResource.vdev_num) > dcmiMaxVdevNum {
		return CgoDcmiSocTotalResource{}, fmt.Errorf("get error virtual quantity: %d", uint32(totalResource.vdev_num))
	}

	return convertSocTotalResource(totalResource), nil
}

func convertSocFreeResource(cSocFreeResource C.struct_dcmi_soc_free_resource) CgoDcmiSocFreeResource {
	socFreeResource := CgoDcmiSocFreeResource{
		vfgNum:    uint32(cSocFreeResource.vfg_num),
		vfgBitmap: uint32(cSocFreeResource.vfg_bitmap),
		base:      convertBaseResource(cSocFreeResource.base),
		computing: convertComputingResource(cSocFreeResource.computing),
		media:     convertMediaResource(cSocFreeResource.media),
	}
	return socFreeResource
}

// GetDeviceFreeResource get device free resource info
func GetDeviceFreeResource(cardID, deviceID int32) (CgoDcmiSocFreeResource, error) {
	var cMainCmd = C.enum_dcmi_main_cmd(MainCmdVDevMng)
	subCmd := VmngSubCmdGetFreeResource
	var freeResource C.struct_dcmi_soc_free_resource
	size := C.uint(unsafe.Sizeof(freeResource))
	if err := C.dcmi_get_device_info(C.int(cardID), C.int(deviceID), cMainCmd, C.uint(subCmd),
		unsafe.Pointer(&freeResource), &size); err != 0 {
		return CgoDcmiSocFreeResource{}, fmt.Errorf("get device info failed, error is: %d", int32(err))
	}
	return convertSocFreeResource(freeResource), nil
}

// GetDeviceInfo get device resource info
func GetDeviceInfo(cardID, deviceID int32) (CgoVDevInfo, error) {
	cgoDcmiSocTotalResource, err := GetDeviceTotalResource(cardID, deviceID)
	if err != nil {
		return CgoVDevInfo{}, fmt.Errorf("get device tatal resource failed, error is: %v", err)
	}

	cgoDcmiSocFreeResource, err := GetDeviceFreeResource(cardID, deviceID)
	if err != nil {
		return CgoVDevInfo{}, fmt.Errorf("get device free resource failed, error is: %v", err)
	}

	dcmiVDevInfo := CgoVDevInfo{
		VDevNum:       cgoDcmiSocTotalResource.vDevNum,
		CoreNumUnused: cgoDcmiSocFreeResource.computing.aic,
	}
	for i := 0; i < len(cgoDcmiSocTotalResource.vDevID); i++ {
		dcmiVDevInfo.VDevID = append(dcmiVDevInfo.VDevID, cgoDcmiSocTotalResource.vDevID[i])
	}
	for _, vDevID := range cgoDcmiSocTotalResource.vDevID {
		cgoVDevQueryStru, err := GetDeviceVDevResource(cardID, deviceID, vDevID)
		if err != nil {
			return CgoVDevInfo{}, fmt.Errorf("get device vitrual resource failed, error is: %v", err)
		}
		dcmiVDevInfo.Status = append(dcmiVDevInfo.Status, cgoVDevQueryStru.queryInfo.status)
		dcmiVDevInfo.VfID = append(dcmiVDevInfo.VfID, cgoVDevQueryStru.queryInfo.vfid)
		dcmiVDevInfo.CID = append(dcmiVDevInfo.CID, cgoVDevQueryStru.queryInfo.containerID)
		dcmiVDevInfo.CoreNum = append(dcmiVDevInfo.CoreNum, cgoVDevQueryStru.queryInfo.computing.aic)
	}
	return dcmiVDevInfo, nil
}

// GetCardIDDeviceID get card id and device id from logic id
func GetCardIDDeviceID(logicID uint32) (int32, int32, error) {
	if logicID > uint32(math.MaxInt8) {
		return retError, retError, fmt.Errorf("input invalid logicID: %d", logicID)
	}

	_, cards, err := GetCardList()
	if err != nil {
		return retError, retError, fmt.Errorf("get card list failed, error is: %v", err)
	}

	for _, cardID := range cards {
		deviceNum, err := GetDeviceNumInCard(cardID)
		if err != nil {
			hwlog.RunLog.Errorf("get device num in card failed, error is: %v", err)
			continue
		}
		for deviceID := int32(0); deviceID < deviceNum; deviceID++ {
			logicIDGet, err := GetDeviceLogicID(cardID, deviceID)
			if err != nil {
				hwlog.RunLog.Errorf("get device logic id failed, error is: %v", err)
				continue
			}
			if logicID == logicIDGet {
				return cardID, deviceID, nil
			}
		}
	}
	errInfo := fmt.Errorf("the card id and device id corresponding to the logic id are not found")
	return retError, retError, errInfo
}

// CreateVDevice create virtual device by logic id
func CreateVDevice(logicID uint32, aiCore uint32) (uint32, error) {
	cardID, deviceID, err := GetCardIDDeviceID(logicID)
	if err != nil {
		return unretError, fmt.Errorf("get card id and device id failed, error is: %v", err)
	}

	cgoDcmiSocFreeResource, err := GetDeviceFreeResource(cardID, deviceID)
	if err != nil {
		return unretError, fmt.Errorf("get virtual device info failed, error is: %v", err)
	}

	if cgoDcmiSocFreeResource.computing.aic < float32(aiCore) {
		return unretError, fmt.Errorf("the remaining core resource is insufficient, free core: %f",
			cgoDcmiSocFreeResource.computing.aic)
	}

	var vDevID int32
	createVDevOut, err := CreateVirtualDevice(cardID, deviceID, vDevID, aiCore)
	if err != nil {
		return unretError, fmt.Errorf("create virtual device failed, error is: %v", err)
	}
	return createVDevOut.VDevID, nil
}

// GetVDeviceInfo get virtual device info by logic id
func GetVDeviceInfo(logicID uint32) (CgoVDevInfo, error) {
	cardID, deviceID, err := GetCardIDDeviceID(logicID)
	if err != nil {
		return CgoVDevInfo{}, fmt.Errorf("get card id and device id failed, error is: %v", err)
	}

	dcmiVDevInfo, err := GetDeviceInfo(cardID, deviceID)
	if err != nil {
		return CgoVDevInfo{}, fmt.Errorf("get virtual device info failed, error is: %v", err)
	}
	return dcmiVDevInfo, nil
}

// DestroyVDevice destroy spec virtual device by logic id
func DestroyVDevice(logicID, vDevID uint32) error {
	cardID, deviceID, err := GetCardIDDeviceID(logicID)
	if err != nil {
		return fmt.Errorf("get card id and device id failed, error is: %v", err)
	}

	if err = SetDestroyVirtualDevice(cardID, deviceID, vDevID); err != nil {
		return fmt.Errorf("destroy virtual device failed, error is: %v", err)
	}
	return nil
}
