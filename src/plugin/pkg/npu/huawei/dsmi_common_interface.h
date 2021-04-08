/*
 * Copyright (c) Huawei Technologies Co., Ltd. 2012-2019. All rights reserved.
 * Description:
 * Author: huawei
 * Create: 2019-10-15
 */
#ifndef __DSMI_COMMON_INTERFACE_H__
#define __DSMI_COMMON_INTERFACE_H__
#ifdef __cplusplus
extern "C" {
#endif

typedef enum rdfx_detect_result {
    RDFX_DETECT_OK = 0,
    RDFX_DETECT_SOCK_FAIL = 1,
    RDFX_DETECT_RECV_TIMEOUT = 2,
    RDFX_DETECT_UNREACH = 3,
    RDFX_DETECT_TIME_EXCEEDED = 4,
    RDFX_DETECT_FAULT = 5,
    RDFX_DETECT_INIT = 6,
    RDFX_DETECT_MAX
} DSMI_NET_HEALTH_STATUS;

struct dsmi_power_info_stru {
    unsigned short power;
};
struct dsmi_memory_info_stru {
    unsigned long memory_size;
    unsigned int freq;
    unsigned int utiliza;
};

struct dsmi_hbm_info_stru {
    unsigned long memory_size;  /**< HBM total size, KB */
    unsigned int freq;          /**< HBM freq, MHZ */
    unsigned long memory_usage; /**< HBM memory_usage, KB */
    int temp;                   /**< HBM temperature */
    unsigned int bandwith_util_rate;
};

#define MAX_CHIP_NAME 32
#define MAX_DEVICE_COUNT 64

struct dsmi_chip_info_stru {
    unsigned char chip_type[MAX_CHIP_NAME];
    unsigned char chip_name[MAX_CHIP_NAME];
    unsigned char chip_ver[MAX_CHIP_NAME];
};

#define DSMI_VNIC_PORT 0
#define DSMI_ROCE_PORT 1

enum ip_addr_type {
    IPADDR_TYPE_V4 = 0U,    /**< IPv4 */
    IPADDR_TYPE_V6 = 1U,    /**< IPv6 */
    IPADDR_TYPE_ANY = 2U
};

#define DSMI_ARRAY_IPV4_NUM 4
#define DSMI_ARRAY_IPV6_NUM 16

typedef struct ip_addr {
    union {
        unsigned char ip6[DSMI_ARRAY_IPV6_NUM];
        unsigned char ip4[DSMI_ARRAY_IPV4_NUM];
    } u_addr;
    enum ip_addr_type ip_type;
} ip_addr_t;

#define DSMI_MAX_VDEV_NUM 16
#define DSMI_MAX_SPEC_RESERVE 8

struct dsmi_vdev_spec_info {
    unsigned char core_num;                         /**< aicore num for virtual device */
    unsigned char reservesd[DSMI_MAX_SPEC_RESERVE]; /**reserved */
};

// Dsmi each virtual device info
struct dsmi_sub_vdev_info {
    unsigned int status;                            /**< whether the vdevice used by container */
    unsigned int vdevid;                            /**< id number of vdevice */
    unsigned int vfid;
    unsigned long int cid;                           /**< container id */
    struct dsmi_vdev_spec_info spec;                /**< specification of vdevice */
};

// Dsmi physical device split info
struct dsmi_vdev_info {
    unsigned int vdev_num;                          /**< number of vdevice the devid had created */
    struct dsmi_vdev_spec_info spec_unused;         /**< resource the devid unallocated */
    struct dsmi_sub_vdev_info vdev[DSMI_MAX_VDEV_NUM];
};

/**
* @ingroup driver
* @brief Get the number of devices
* @attention NULL
* @param [out] device_count  The space requested by the user is used to store the number of returned devices
* @return  0 for success, others for fail
* @note Support:Ascend310,Ascend910
*/
int dsmi_get_device_count(int *device_count);

/**
* @ingroup driver
* @brief Get the id of all devices
* @attention NULL
* @param [out] device_id_list[] The space requested by the user is used to store the id of all returned devices
* @param [in] count Number of equipment
* @return  0 for success, others for fail
* @note Support:Ascend310,Ascend910
*/
int dsmi_list_device(int device_id_list[], int count);



/**
* @ingroup driver
* @brief Convert the logical ID of the device to a physical ID
* @attention NULL
* @param [in] logicid logic id
* @param [out] phyid physic id
* @return  0 for success, others for fail
* @note Support:Ascend310,Ascend910
*/
int dsmi_get_phyid_from_logicid(unsigned int logicid, unsigned int *phyid);

/**
* @ingroup driver
* @brief Convert the physical ID of the device to a logical ID
* @attention NULL
* @param [in] phyid   physical id
* @param [out] logicid logic id
* @return  0 for success, others for fail
* @note Support:Ascend310,Ascend910
*/
int dsmi_get_logicid_from_phyid(unsigned int phyid, unsigned int *logicid);

/**
* @ingroup driver
* @brief Query the overall health status of the device, support AI Server
* @attention NULL
* @param [in] device_id  The device id
* @param [out] phealth  The pointer of the overall health status of the device only represents this component,
                        and does not include other components that have a logical relationship with this component.
* @return  0 for success, others for fail
* @note Support:Ascend310,Ascend910
*/
int dsmi_get_device_health(int device_id, unsigned int *phealth);

/**
* @ingroup driver
* @brief get the ip address and mask address.
* @attention NULL
* @param [in] device_id  The device id
* @param [in] port_type  Specify the network port type
* @param [in] port_id  Specify the network port number, reserved field
* @param [out] ip_address  return ip address info
* @param [out] mask_address  return mask address info
* @return  0 for success, others for fail
* @note Support:Ascend310,Ascend910
*/
int dsmi_get_device_ip_address(int device_id, int port_type, int port_id, ip_addr_t *ip_address,
    ip_addr_t *mask_address);

/**
* @ingroup driver
* @brief Relevant information about the HiSilicon SOC of the AI ??processor, including chip_type, chip_name,
         chip_ver version number
* @attention NULL
* @param [in] device_id  The device id
* @param [out] chip_info  Get the relevant information of ascend AI processor Hisilicon SOC
* @return  0 for success, others for fail
* @note Support:Ascend310,Ascend910
*/
int dsmi_get_chip_info(int device_id, struct dsmi_chip_info_stru *chip_info);

/**
* @ingroup driver
* @brief Query the connectivity status of the RoCE network card's IP address
* @attention NULL
* @param [in] device_id The device id
* @param [out] presult return the result wants to query
* @return  0 for success, others for fail
* @note Support:Ascend910
*/
int dsmi_get_network_health(int device_id, DSMI_NET_HEALTH_STATUS *presult);

/**
* @ingroup driver
* @brief Query the cvirtual device info by device_id(logicID)
* @attention NULL
* @param [in] devid The device id
* @param [out] result return the virtual device info wants to query
* @return  0 for success, DRV_ERROR_NOT_SUPPORT: not support function, others for fail
* @note Support:Ascend910
*/
int dsmi_get_vdevice_info(unsigned int devid, struct dsmi_vdev_info *info);

#ifdef __cplusplus
}
#endif
#endif