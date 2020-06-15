#!/bin/bash

build_version="1.0.8"
build_time=$(date +'%Y-%m-%d_%T')
SODIR=${TOP_DIR}/${DRIVER_FILE}/driver/lib64/
OUTPUT_NAME="ascendplugin"
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
  ostype="arm64"
else
  ostype="x86"
fi
PKGNAME="Ascend-k8s_device-plugin-${build_version}-linux.run"
TARNAME="Ascend-K8sDevicePlugin-${build_version}-${ostype}-Linux.tar.gz"
docker_zip_name="Ascend-K8sDevicePlugin-${build_version}-${ostype}-Docker.tar.gz"
# export so library path
export LD_LIBRARY_PATH=${SODIR}:${LD_LIBRARY_PATH}
export PKG_CONFIG_PATH=$PKG_CONFIG_PATH:${CONFIGDIR}


function clear_env() {
    rm -rf ${TOP_DIR}/output/*
    if [ ! -d "${TOP_DIR}/makerunout" ]; then
        mkdir -p ${TOP_DIR}/makerunout
    fi
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

    cp ${TOP_DIR}/src/plugin/cmd/ascendplugin/${OUTPUT_NAME}   ${TOP_DIR}/output
    dos2unix ${TOP_DIR}/build/${DEPLOYNAME}
    chmod 500 ${TOP_DIR}/build/${DEPLOYNAME}
    cp ${TOP_DIR}/build/${DEPLOYNAME}     ${TOP_DIR}/output

}

function copy2runpackage() {
    mv ${TOP_DIR}/src/plugin/cmd/ascendplugin/${OUTPUT_NAME}   ${TOP_DIR}/makerunout
    dos2unix ${TOP_DIR}/build/${DEPLOYNAME}
    chmod 500 ${TOP_DIR}/build/${DEPLOYNAME}
    cp ${TOP_DIR}/build/${DEPLOYNAME}     ${TOP_DIR}/makerunout/
}

function zip_file(){
    cd ${TOP_DIR}/output
    tar -zcvf ${TARNAME}  ${OUTPUT_NAME}  ${DEPLOYNAME}
    rm -f ${OUTPUT_NAME}  ${DEPLOYNAME}
}

function make_run_package() {
    chmod +x  ${CUR_DIR}/script/makepackgeinstall.sh
    dos2unix  ${CUR_DIR}/script/makepackgeinstall.sh
    cp ${CUR_DIR}/script/makepackgeinstall.sh  ${TOP_DIR}/makerunout
    dirname="${ostype}"
    if [ ! -d "${TOP_DIR}/output/${dirname}" ]; then
        mkdir -p "${TOP_DIR}/output/${dirname}"
    fi
    if [ -d "${TOP_DIR}/tools/makeself-release-2.4.0" ]; then
        rm -rf ${TOP_DIR}/tools/makeself-release-2.4.0
    fi
    cd ${TOP_DIR}/tools || retrun
    unzip makeself-release-2.4.0.zip
    cd ${TOP_DIR}/tools/makeself-release-2.4.0 || return
    cp makeself.sh ${CUR_DIR}/script
    cp makeself-header.sh ${CUR_DIR}/script
    cd ${CUR_DIR}/script || retrun
    patch -p0 < mkselfmodify.patch
    cd ${TOP_DIR}/output/${dirname} || return
    sh ${CUR_DIR}/script/makeself.sh --nomd5 --nocrc --header ${CUR_DIR}/script/makeself-header.sh  --help-header \
    ${CUR_DIR}/script/help.info ${TOP_DIR}/makerunout "${PKGNAME}" ascendplugin ./makepackgeinstall.sh
    rm -rf ${TOP_DIR}/makerunout
    rm -f ${TOP_DIR}/output/${OUTPUT_NAME}  ${TOP_DIR}/output/${DEPLOYNAME}
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

