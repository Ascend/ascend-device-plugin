#!/bin/bash
# Perform  build k8s-device-plugin
# Copyright(C) Huawei Technologies Co.,Ltd. 2020-2021. All rights reserved.
set -e
CUR_DIR=$(dirname "$(readlink -f "$0")")
TOP_DIR=$(realpath "${CUR_DIR}"/..)
build_version="v2.0.3"
version_file="${TOP_DIR}"/service_config.ini
if  [ -f "$version_file" ]; then
  line=$(sed -n '4p' "$version_file" 2>&1)
  #cut the chars after ':'
  build_version=${line#*:}
fi

output_name="device-plugin"
os_type=$(arch)
build_type=build

if [ "$1" == "ci" ] || [ "$2" == "ci" ]; then
    export GO111MODULE="on"
    export GOPROXY="http://mirrors.tools.huawei.com/goproxy/"
    export GONOSUMDB="*"
    build_type=ci
fi

function clean() {
    rm -rf "${TOP_DIR}"/output/*
}

function build_plugin() {
    cd "${TOP_DIR}"/src/plugin/cmd/ascendplugin
    export CGO_ENABLED=1
    export CGO_CFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv"
    export CGO_CPPFLAGS="-fstack-protector-strong -D_FORTIFY_SOURCE=2 -O2 -fPIC -ftrapv"
    go build -mod=mod -buildmode=pie -ldflags "-X main.BuildName=${output_name} \
            -X main.BuildVersion=${build_version}_linux-${os_type} \
            -buildid none     \
            -s   \
            -extldflags=-Wl,-z,relro,-z,now,-z,noexecstack" \
            -o "${output_name}"  \
            -trimpath
    ls "${output_name}"
    if [ $? -ne 0 ]; then
        echo "fail to find device-plugin"
        exit 1
    fi
}

function mv_file() {
    mv "${TOP_DIR}/src/plugin/cmd/ascendplugin/${output_name}"   "${TOP_DIR}"/output
}

function change_mod() {
    chmod 400 "$TOP_DIR"/output/*
    chmod 500 "${TOP_DIR}/output/${output_name}"
}

function modify_version() {
    cd "${TOP_DIR}"
    sed -i "s/ascend-k8sdeviceplugin:.*/ascend-k8sdeviceplugin:${build_version}/" "$TOP_DIR"/ascendplugin-910.yaml
    sed -i "s/ascend-k8sdeviceplugin:.*/ascend-k8sdeviceplugin:${build_version}/" "$TOP_DIR"/ascendplugin-volcano.yaml
    sed -i "s/ascend-k8sdeviceplugin:.*/ascend-k8sdeviceplugin:${build_version}/" "$TOP_DIR"/ascendplugin-310.yaml
    sed -i "s/ascend-k8sdeviceplugin:.*/ascend-k8sdeviceplugin:${build_version}/" "$TOP_DIR"/ascendplugin-310-volcano.yaml
    sed -i "s/ascend-k8sdeviceplugin:.*/ascend-k8sdeviceplugin:${build_version}/" "$TOP_DIR"/ascendplugin-710.yaml
    sed -i "s/ascend-k8sdeviceplugin:.*/ascend-k8sdeviceplugin:${build_version}/" "$TOP_DIR"/ascendplugin-710-volcano.yaml
    cp "$TOP_DIR"/Dockerfile "$TOP_DIR"/output/
    cp "$TOP_DIR"/ascendplugin-910.yaml "$TOP_DIR"/output/device-plugin-910-"${build_version}".yaml
    cp "$TOP_DIR"/ascendplugin-volcano.yaml "$TOP_DIR"/output/device-plugin-volcano-"${build_version}".yaml
    cp "$TOP_DIR"/ascendplugin-310.yaml "$TOP_DIR"/output/device-plugin-310-"${build_version}".yaml
    cp "$TOP_DIR"/ascendplugin-310-volcano.yaml "$TOP_DIR"/output/device-plugin-310-volcano-"${build_version}".yaml
    cp "$TOP_DIR"/ascendplugin-710.yaml "$TOP_DIR"/output/device-plugin-710-"${build_version}".yaml
    cp "$TOP_DIR"/ascendplugin-710-volcano.yaml "$TOP_DIR"/output/device-plugin-710-volcano-"${build_version}".yaml

    sed -i "s#output/device-plugin#device-plugin#" "$TOP_DIR"/output/Dockerfile
}


function main() {
  clean
  build_plugin
  mv_file
  modify_version
  change_mod
}


main
