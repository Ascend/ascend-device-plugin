#!/bin/bash
set -x
CUR_DIR=$(dirname $(readlink -f $0))
TOP_DIR=$(realpath ${CUR_DIR}/..)

build_version="1.0.8"
build_time=$(date +'%Y-%m-%d_%T')

OUTPUT_NAME="ascendplugin"
SODIR=/usr/local/Ascend/driver/lib64/
CONFIGDIR=${TOP_DIR}/src/plugin/config/config_310
PKGNAME="K8sDevicePlugin.tar.gz"

DEPLOYNAME="deploy.sh"
DOCKER_FILE_NAME="Dockerfile"

docker_zip_name="ascend-device-plugin_docker.tar.gz"
docker_images_name="ascend-device-plugin:latest"


# export so library path
export LD_LIBRARY_PATH=${SODIR}
export PKG_CONFIG_PATH=$PKG_CONFIG_PATH:${CONFIGDIR}
#export GOPATH=${TOP_DIR}/:/home/gopath
#export GOROOT=/opt/buildtools/go
#export PATH=$PATH:$GOROOT/bin:$GOPATH/bin


function clearEnv() {
    rm -rf ${TOP_DIR}/output/*
}

function buildPlugin() {

    cd ${TOP_DIR}/src/plugin/cmd/ascendplugin


    go build -ldflags "-X main.BuildName=${OUTPUT_NAME} \
            -X main.BuildVersion=${build_version} \
            -X main.BuildTime=${build_time}"  \
            -o ${OUTPUT_NAME}

    ls ${OUTPUT_NAME}
    if [ $? -ne 0 ]; then
        echo "fail to find ascendplugin"
        exit -1
    fi
}

function mvFile() {

    mv ${TOP_DIR}/src/plugin/cmd/ascendplugin/${OUTPUT_NAME}   ${TOP_DIR}/output
    chmod 500 ${TOP_DIR}/build/${DEPLOYNAME}
    cp ${TOP_DIR}/build/${DEPLOYNAME}     ${TOP_DIR}/output

}

function zipFile(){
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
    clearEnv
    buildPlugin
    mvFile
    #build_docker_images
    zipFile
}

main
