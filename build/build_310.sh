#!/bin/bash
set -x
CUR_DIR=$(dirname $(readlink -f $0))
TOP_DIR=$(realpath ${CUR_DIR}/..)

build_version="1.0.8"
build_time=$(date +'%Y-%m-%d_%T')

DOWN_DRIVER_FILE="platform/Tuscany"
DRIVER_FILE="310driver"

OUTPUT_NAME="ascendplugin"
SODIR=${TOP_DIR}/${DRIVER_FILE}/driver/lib64/
CONFIGDIR=${TOP_DIR}/src/plugin/config/config_310

DEPLOYNAME="deploy.sh"
DOCKER_FILE_NAME="Dockerfile"
PC_File="ascend_device_plugin.pc"
docker_zip_name="ascend-device-plugin_docker.tar.gz"
docker_images_name="ascend-k8sdeviceplugin:latest"
export GO111MODULE="on"
export GOPROXY="http://mirrors.tools.huawei.com/goproxy/"
export GONOSUMDB="*"

osname=$(grep -i ^id= /etc/os-release| cut -d"=" -f2 | sed 's/"//g');
ostype=$(arch)
if [ "${ostype}" = "aarch64" ]; then
  ostype="ARM64"
else
  ostype="X86"
fi
PKGNAME="Ascend-K8sDevicePlugin-${build_version}-${ostype}-Linux.tar.gz"
docker_zip_name="Ascend-K8sDevicePlugin-${build_version}-${ostype}-Docker.tar.gz"
# export so library path
export LD_LIBRARY_PATH=${SODIR}
export PKG_CONFIG_PATH=$PKG_CONFIG_PATH:${CONFIGDIR}

function clear_env() {
    rm -rf ${TOP_DIR}/output/*
}

function make_lib() {
    ls ${TOP_DIR}/${DOWN_DRIVER_FILE}
    plateform=$(arch)
    chmod +x  ${TOP_DIR}/${DOWN_DRIVER_FILE}/Ascend310-driver-*.${plateform}.run

    ${TOP_DIR}/${DOWN_DRIVER_FILE}/Ascend310-driver-*${osname}*.${plateform}*.run \
    --noexec --extract=${TOP_DIR}/${DRIVER_FILE}
    sed -i "/^prefix=/c prefix=${TOP_DIR}/${DRIVER_FILE}" ${CONFIGDIR}/${PC_File}
}

function build_plugin() {

    cd ${TOP_DIR}/src/plugin/cmd/ascendplugin
    go build -ldflags "-X main.BuildName=${OUTPUT_NAME} \
            -X main.BuildVersion=${build_version} \
            -X main.BuildTime=${build_time}"  \
            -o ${OUTPUT_NAME}

    ls ${OUTPUT_NAME}
    if [ $? -ne 0 ]; then
        echo "fail to find ascendplugin"
        exit 1
    fi
}

function mv_file() {

    mv ${TOP_DIR}/src/plugin/cmd/ascendplugin/${OUTPUT_NAME}   ${TOP_DIR}/output
    chmod 500 ${TOP_DIR}/build/${DEPLOYNAME}
    cp ${TOP_DIR}/build/${DEPLOYNAME}     ${TOP_DIR}/output

}

function zip_file(){
    cd ${TOP_DIR}/output
    tar -zcvf ${PKGNAME}  ${OUTPUT_NAME}  ${DEPLOYNAME} 
    rm -f ${OUTPUT_NAME}  ${DEPLOYNAME}   
}

function build_docker_images()
{
    cp ${TOP_DIR}/build/${DOCKER_FILE_NAME}     ${TOP_DIR}/output
    cd ${TOP_DIR}/output
    docker rmi ${docker_images_name}
    docker build -t ${docker_images_name} .
    docker save ${docker_images_name} | gzip > ${docker_zip_name}
    rm -f ${DOCKER_FILE_NAME}
}

function main() {
    clear_env
    make_lib
    build_plugin
    mv_file
    build_docker_images
    zip_file
}

main
