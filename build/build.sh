#!/bin/bash
# Perform  build k8s-device-plugin
# Copyright @ Huawei Technologies CO., Ltd. 2020-2021. All rights reserved
set -e
CUR_DIR=$(dirname $(readlink -f "$0"))
TOP_DIR=$(realpath "${CUR_DIR}"/..)
build_version="v2.0.2"
output_name="ascendplugin"
docker_images_name="ascend-k8sdeviceplugin:v2.0.2"
os_type=$(arch)
if [ "${os_type}" = "aarch64" ]; then
  os_type="arm64"
else
  os_type="x86"
fi
build_type=build

if [ "$1" == "ci" ] || [ "$2" == "ci" ]; then
    export GO111MODULE="on"
    export GOPROXY="http://mirrors.tools.huawei.com/goproxy/"
    export GONOSUMDB="*"
    build_type=ci
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
    chmod 500 "${TOP_DIR}/output/${output_name}"
}

function modify_version() {
    cd "${TOP_DIR}"
    sed -i "s/ascend-k8sdeviceplugin:.*/ascend-k8sdeviceplugin:${version}/" "$TOP_DIR"/ascendplugin.yaml
    sed -i "s/ascend-k8sdeviceplugin:.*/ascend-k8sdeviceplugin:${version}/" "$TOP_DIR"/ascendplugin-volcano.yaml
    sed -i "s/ascend-k8sdeviceplugin:.*/ascend-k8sdeviceplugin:${version}/" "$TOP_DIR"/ascendplugin-310.yaml
    sed -i "s/ascend-k8sdeviceplugin:.*/ascend-k8sdeviceplugin:${version}/" "$TOP_DIR"/ascendplugin-710.yaml

    cp "$TOP_DIR"/Dockerfile "$TOP_DIR"/output/
    cp "$TOP_DIR"/ascendplugin.yaml "$TOP_DIR"/output/ascendplugin-"${version}".yaml
    cp "$TOP_DIR"/ascendplugin-volcano.yaml "$TOP_DIR"/output/ascendplugin-volcano-"${version}".yaml
    cp "$TOP_DIR"/ascendplugin-310.yaml "$TOP_DIR"/output/ascendplugin-310-"${version}".yaml
    cp "$TOP_DIR"/ascendplugin-710.yaml "$TOP_DIR"/output/ascendplugin-710-"${version}".yaml

    sed -i "s#output/ascendplugin#ascendplugin#" "$TOP_DIR"/output/Dockerfile
}

function parse_version() {
    version_file="${TOP_DIR}"/service_config.ini
    version=${build_version}
    if  [ -f "$version_file" ]; then
      line=$(sed -n '4p' "$version_file" 2>&1)
      #cut the chars after ':'
      version=${line#*:}
      build_version=${version}
    fi
}

function main() {
  clear_env
  parse_version
  build_plugin
  mv_file
  modify_version
}

main
