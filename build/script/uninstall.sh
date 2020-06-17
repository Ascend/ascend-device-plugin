#!/bin/bash

cd ..
cd ..
PROJECT_DIR=`pwd`
INSTALL_DIR="k8s_device_plguin"
# 卸载判断
INSTALL_DERECT_PATH=${PROJECT_DIR}/${INSTALL_DIR}
if [ ! -d "${INSTALL_DERECT_PATH}" ]||[ ! -d "${INSTALL_DERECT_PATH}/script" ]; then
		echo "error : uninstall failed, Directory structure is broken"
else
    rm -rf "${INSTALL_DERECT_PATH}"
  	echo "Uninstalled successfully"
fi
