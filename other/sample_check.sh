#!/bin/bash
ASCNED_TYPE=910
ASCNED_INSTALL_PATH=/usr/local/Ascend
USE_ASCEND_DOCKER=false


CUR_DIR=$(dirname $(readlink -f $0))
TOP_DIR=$(realpath ${CUR_DIR}/..)


LD_LIBRARY_PATH_PARA1=${ASCNED_INSTALL_PATH}/driver/lib64/driver
LD_LIBRARY_PATH_PARA2=${ASCNED_INSTALL_PATH}/driver/lib64
TYPE=Ascend910
PKG_PATH=${TOP_DIR}/src/plugin/config/config_910
PKG_PATH_STRING=\$\{TOP_DIR\}/src/plugin/config/config_910
LIBDRIVER="driver/lib64/driver"
if [ ${ASCNED_TYPE} == "310"  ]; then
  TYPE=Ascend310
  LD_LIBRARY_PATH_PARA1=${ASCNED_INSTALL_PATH}/driver/lib64
  PKG_PATH=${TOP_DIR}/src/plugin/config/config_310
  PKG_PATH_STRING=\$\{TOP_DIR\}/src/plugin/config/config_310
  sed -i "s#device-plugin  --useAscendDocker=\${USE_ASCEND_DOCKER}#device-plugin --mode=ascend310 --useAscendDocker=${USE_ASCEND_DOCKER}#g" ${TOP_DIR}/ascend-plugin.yaml
  LIBDRIVER="/driver/lib64"
fi
sed -i "s/Ascend[0-9]\{3\}/${TYPE}/g" ${TOP_DIR}/device-plugin.yaml
sed -i "s#ath: /usr/local/Ascend/driver#ath: ${ASCNED_INSTALL_PATH}/driver#g" ${TOP_DIR}/device-plugin.yaml
sed -i "/^ENV USE_ASCEND_DOCKER /c ENV USE_ASCEND_DOCKER ${USE_ASCEND_DOCKER}" ${TOP_DIR}/Dockerfile
sed -i "/^libdriver=/c libdriver=$\{prefix\}/${LIBDRIVER}" ${PKG_PATH}/ascend_device_plugin.pc
sed -i "/^prefix=/c prefix=${ASCNED_INSTALL_PATH}" ${PKG_PATH}/ascend_device_plugin.pc
sed -i "/^CONFIGDIR=/c CONFIGDIR=${PKG_PATH_STRING}" ${CUR_DIR}/build_in_docker.sh
