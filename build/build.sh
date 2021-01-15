#!/bin/bash
# Perform  build k8s-device-plugin
# Copyright @ Huawei Technologies CO., Ltd. 2020-2021. All rights reserved
set -e
CUR_DIR=$(dirname $(readlink -f "$0"))
TOP_DIR=$(realpath "${CUR_DIR}"/..)

build_version="v20.2.0"
output_name="ascendplugin"
deploy_name="deploy.sh"
docker_images_name="ascend-k8sdeviceplugin:v20.2.0"
ostype=$(arch)
if [ "${ostype}" = "aarch64" ]; then
  ostype="arm64"
else
  ostype="amd64"
fi
tar_name="Ascend-K8sDevicePlugin-${build_version}-${ostype}-Linux.tar.gz"
docker_zip_name="Ascend-K8sDevicePlugin-${build_version}-${ostype}-Docker.tar.gz"
docker_type=nodocker
if [ "$1" == "dockerimages" ] || [ "$2" == "dockerimages" ]; then
    DOCKER_TYPE=dockerimages
fi

function clear_env() {
    rm -rf ${TOP_DIR}/output/*
}

function build_plugin() {
    cd ${TOP_DIR}/src/plugin/cmd/ascendplugin
    go build -ldflags "-X main.BuildName=${output_name} \
            -X main.BuildVersion=${build_version} \
            -buildid none     \
            -s   \
            -extldflags=-Wl,-z,relro,-z,now,-z,noexecstack" \
            -o "${output_name}"       \
            -trimpath

    ls ${output_name}
    if [ $? -ne 0 ]; then
        echo "fail to find ascendplugin"
        exit 1
    fi
}

function mv_file() {
    mv ${TOP_DIR}/src/plugin/cmd/ascendplugin/${output_name}   ${TOP_DIR}/output/
    cp ${TOP_DIR}/other/${deploy_name}     ${TOP_DIR}/output/
}

function zip_file(){
    cd ${TOP_DIR}/output
    tar -zcvf ${tar_name}  ${output_name}  ${deploy_name}
    if [ $? == 0 ]; then
        echo "build device plugin success"
    fi
    rm -f ${output_name}  ${deploy_name}
}

function build_docker_images()
{
    cd ${TOP_DIR}
    docker rmi ${docker_images_name} || true
    docker build -t ${docker_images_name} .
    docker save ${docker_images_name} | gzip > ${docker_zip_name}
    mv ${docker_zip_name} ./output/
}

function main() {
    clear_env
    build_plugin
    mv_file
    if [ "${DOCKER_TYPE}" == "dockerimages" ]; then
        build_docker_images
    fi
    zip_file
}
main