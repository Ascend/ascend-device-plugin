#!/bin/bash
# Perform  build k8s-device-plugin
# Copyright(C) Huawei Technologies Co.,Ltd. 2020-2022. All rights reserved.
set -e

CUR_DIR=$(dirname "$(readlink -f "$0")")
TOP_DIR=$(realpath "${CUR_DIR}"/..)

build_version="v3.0.0"
version_file="${TOP_DIR}"/service_config.ini
if  [ -f "$version_file" ]; then
  line=$(sed -n '4p' "$version_file" 2>&1)
  #cut the chars after ':'
  build_version=${line#*:}
fi
npu_exporter_folder="${TOP_DIR}/npu-exporter"

output_name="device-plugin"
os_type=$(arch)
build_type=build

if [ "$1" == "ci" ] || [ "$2" == "ci" ]; then
    export GO111MODULE="on"
    export GONOSUMDB="*"
    build_type=ci
fi

function clean() {
    rm -rf "${TOP_DIR}"/output/
    mkdir -p "${TOP_DIR}"/output
}

function build_plugin() {
    cd "${TOP_DIR}"
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

function copy_kmc_files() {
    cp -rf "${npu_exporter_folder}/lib" "${TOP_DIR}"/output
    cp -rf "${npu_exporter_folder}/cert-importer" "${TOP_DIR}"/output
    chmod 550 "${TOP_DIR}"/output/lib
    chmod 500 "${TOP_DIR}"/output/lib/*
    chmod 500 "${TOP_DIR}/output/cert-importer"
}

function mv_file() {
    mv "${TOP_DIR}/${output_name}"   "${TOP_DIR}"/output
}

function change_mod() {
    chmod 400 "$TOP_DIR"/output/*
    chmod 500 "${TOP_DIR}/output/${output_name}"
}

function main() {
  clean
  build_plugin
  mv_file
  change_mod
  if [ "$1" != nokmc ]; then
   copy_kmc_files
  fi
}


main $1
