#!/bin/bash
set -x
CUR_DIR=$(dirname $(readlink -f $0))
TOP_DIR=$(realpath ${CUR_DIR}/..)

build_version="1.0.8"
build_time=$(date +'%Y-%m-%d_%T')

DOWN_DRIVER_FILE="platform/Tuscany"
DRIVER_FILE="910driver"

OUTPUT_NAME="ascendplugin"
SODIR=${TOP_DIR}/${DRIVER_FILE}/driver/lib64/driver
CONFIGDIR=${TOP_DIR}/src/plugin/config/config_910

PKGNAME="Ascend-K8sDevicePlugin-20.0.0-ARM64-Linux.run"

DEPLOYNAME="deploy.sh"
DOCKER_FILE_NAME="Dockerfile"
PC_File="ascend_device_plugin.pc"
docker_zip_name="Ascend-K8sDevicePlugin-20.0.0-ARM64-Docker.tar.gz"
docker_images_name="Ascend-K8sDevicePlugin:latest"


# export so library path
export LD_LIBRARY_PATH=${SODIR}
export PKG_CONFIG_PATH=$PKG_CONFIG_PATH:${CONFIGDIR}


function clear_env() {
    rm -rf ${TOP_DIR}/output/*
}

function make_lib() {
    ls ${TOP_DIR}/${DOWN_DRIVER_FILE}
    chmod +x  ${TOP_DIR}/${DOWN_DRIVER_FILE}/Ascend310-driver-*.aarch64.run
    ${TOP_DIR}/${DOWN_DRIVER_FILE}/Ascend910-driver-*.aarch64.run  --noexec --extract=${TOP_DIR}/${DRIVER_FILE}
    sed -i "1i\prefix=${TOP_DIR}/${DRIVER_FILE}" ${CONFIGDIR}/${PC_File}
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
