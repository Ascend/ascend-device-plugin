#!/bin/bash
# Perform  build k8s-device-plugin
# Copyright @ Huawei Technologies CO., Ltd. 2020-2020. All rights reserved
set -e
CUR_DIR=$(dirname $(readlink -f "$0"))
TOP_DIR=$(realpath "${CUR_DIR}"/..)
build_version="V20.1.0"
output_name="ascendplugin"
deploy_name="deploy.sh"
docker_images_name="ascend-k8sdeviceplugin:V20.1.0"
os_type=$(arch)
if [ "${os_type}" = "aarch64" ]; then
  os_type="arm64"
else
  os_type="x86"
fi
tar_name="Ascend-K8sDevicePlugin-${build_version}-${os_type}-Linux.tar.gz"
docker_zip_name="Ascend-K8sDevicePlugin-${build_version}-${os_type}-Docker.tar.gz"
build_type=build
docker_type=nodocker

if [ "$1" == "ci" ] || [ "$2" == "ci" ]; then
    export GO111MODULE="on"
    export GOPROXY="http://mirrors.tools.huawei.com/goproxy/"
    export GONOSUMDB="*"
    build_type=ci
fi
if [ "$1" == "dockerimages" ] || [ "$2" == "dockerimages" ]; then
    docker_type=dockerimages
fi

function clear_env() {
    rm -rf "${TOP_DIR}"/output/*
}

function build_plugin() {
    cd "${TOP_DIR}"/src/plugin/cmd/ascendplugin
    export CGO_ENABLED=1 
    export CGO_CFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv"
    export CGO_CPPFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv" 
    go build -buildmode=pie -ldflags "-X main.BuildName=${output_name} \
            -X main.BuildVersion=${build_version} \
            -buildid none     \
            -s   \
            -extldflags=-Wl,-z,relro,-z,now,-z,noexecstack" \
            -o "${output_name}"  \
            -trimpath
    ls "${output_name}"
    if [ $? -ne 0 ]; then
        echo "fail to find ascendplugin"
        exit 1
    fi
}

function mv_file() {
    mv "${TOP_DIR}/src/plugin/cmd/ascendplugin/${output_name}"   "${TOP_DIR}"/output
    dos2unix "${TOP_DIR}/other/${deploy_name}"
    chmod 550 "${TOP_DIR}/other/${deploy_name}"
    cp "${TOP_DIR}/other/${deploy_name}"     "${TOP_DIR}"/output
}

function zip_file(){
    cd "${TOP_DIR}/output"
    tar -zcvf "${tar_name}"  "${output_name}"  "${deploy_name}"
    rm -f "${output_name}"  "${deploy_name}"
}

function build_docker_images(){
    cd "${TOP_DIR}"
    docker rmi "${docker_images_name}" || true
    docker build -t "${docker_images_name}" .
    docker save "${docker_images_name}" | gzip > "${docker_zip_name}"
    mv "${docker_zip_name}" ./output/
}

function main() {
  clear_env
  build_plugin
  mv_file
  if [ "${docker_type}" == "dockerimages" ]; then
      build_docker_images
  fi
  zip_file
}

main
