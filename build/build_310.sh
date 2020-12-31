#!/bin/bash

CUR_DIR=$(dirname $(readlink -f $0))
TOP_DIR=$(realpath ${CUR_DIR}/..)
DOWN_DRIVER_FILE="platform/Tuscany"
DRIVER_FILE="310driver"
CONFIGDIR=${TOP_DIR}/src/plugin/config/config_310
SODIR=/usr/local/Ascend/driver/lib64
BUILD_TYPE=build
DOCKER_TYPE=nodockerimages
if [ "$1" == "ci" ] || [ "$2" == "ci" ]; then
    export GO111MODULE="on"
    export GOPROXY="http://mirrors.tools.huawei.com/goproxy/"
    export GONOSUMDB="*"
    BUILD_TYPE=ci
    SODIR=${TOP_DIR}/${DRIVER_FILE}/driver/lib64/
fi

if [ "$1" == "dockerimages" ] || [ "$2" == "dockerimages" ]; then
    DOCKER_TYPE=dockerimages
fi

chmod 550 build_common.sh
dos2unix build_common.sh
source build_common.sh

function make_lib() {
    ls ${TOP_DIR}/${DOWN_DRIVER_FILE}
    plateform=$(arch)
    chmod 550  ${TOP_DIR}/${DOWN_DRIVER_FILE}/Ascend310-driver-*.${plateform}.run

    ${TOP_DIR}/${DOWN_DRIVER_FILE}/Ascend310-driver-*${osname}*.${plateform}*.run \
    --noexec --extract=${TOP_DIR}/${DRIVER_FILE}
    sed -i "/^prefix=/c prefix=${TOP_DIR}/${DRIVER_FILE}" ${CONFIGDIR}/${PC_File}
}

function main() {
    clear_env
    if [ "${BUILD_TYPE}" = "ci" ]; then
      make_lib
    fi
    build_plugin
    mv_file
    if [ "${DOCKER_TYPE}" == "dockerimages" ]; then
        dos2unix ${CUR_DIR}/build_in_docker.sh
        chmod 550 ${CUR_DIR}/build_in_docker.sh
        build_docker_images
    fi
    zip_file
}
main
