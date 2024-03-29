#!/bin/bash
# Perform  build ascend-device-plugin
# Copyright(C) Huawei Technologies Co.,Ltd. 2020-2023. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ============================================================================

set -e

CUR_DIR=$(dirname "$(readlink -f "$0")")
TOP_DIR=$(realpath "${CUR_DIR}"/..)

build_version="v5.0.RC3"
version_file="${TOP_DIR}"/service_config.ini
if  [ -f "$version_file" ]; then
  line=$(sed -n '1p' "$version_file" 2>&1)
  #cut the chars after ':' and add char 'v', the final example is v3.0.0
  build_version="v"${line#*=}
fi

output_name="device-plugin"
build_scene="center"
os_type=$(arch)
build_type=build

if [ "$1" == "ci" ] || [ "$2" == "ci" ]; then
    export GO111MODULE="on"
    export GONOSUMDB="*"
    build_type=ci
fi

if [ "$1" == "edge" ]; then
   build_scene="edge"
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
            -X main.BuildScene=${build_scene} \
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
    mv "${TOP_DIR}/${output_name}"   "${TOP_DIR}"/output
}

function change_mod() {
    chmod 400 "$TOP_DIR"/output/*
    chmod 500 "${TOP_DIR}/output/${output_name}"
}

function modify_version() {
    if [ $build_scene == "edge" ]; then
      return
    fi
    cd "${TOP_DIR}"
    sed -i "s/ascend-k8sdeviceplugin:.*/ascend-k8sdeviceplugin:${build_version}/" "$CUR_DIR"/ascendplugin-910.yaml
    sed -i "s/ascend-k8sdeviceplugin:.*/ascend-k8sdeviceplugin:${build_version}/" "$CUR_DIR"/ascendplugin-volcano.yaml
    sed -i "s/ascend-k8sdeviceplugin:.*/ascend-k8sdeviceplugin:${build_version}/" "$CUR_DIR"/ascendplugin-310.yaml
    sed -i "s/ascend-k8sdeviceplugin:.*/ascend-k8sdeviceplugin:${build_version}/" "$CUR_DIR"/ascendplugin-310-volcano.yaml
    sed -i "s/ascend-k8sdeviceplugin:.*/ascend-k8sdeviceplugin:${build_version}/" "$CUR_DIR"/ascendplugin-310P.yaml
    sed -i "s/ascend-k8sdeviceplugin:.*/ascend-k8sdeviceplugin:${build_version}/" "$CUR_DIR"/ascendplugin-310P-volcano.yaml
    sed -i "s/ascend-k8sdeviceplugin:.*/ascend-k8sdeviceplugin:${build_version}/" "$CUR_DIR"/ascendplugin-310P-1usoc-volcano.yaml
    sed -i "s/ascend-k8sdeviceplugin:.*/ascend-k8sdeviceplugin:${build_version}/" "$CUR_DIR"/ascendplugin-310P-1usoc.yaml
    cp "$CUR_DIR"/Dockerfile "$TOP_DIR"/output/
    cp "$CUR_DIR"/Dockerfile-310P-1usoc "$TOP_DIR"/output/Dockerfile-310P-1usoc
    cp "$CUR_DIR"/run_for_310P_1usoc.sh "$TOP_DIR"/output/run_for_310P_1usoc.sh
    cp "$CUR_DIR"/ascendplugin-910.yaml "$TOP_DIR"/output/device-plugin-910-"${build_version}".yaml
    cp "$CUR_DIR"/ascendplugin-volcano.yaml "$TOP_DIR"/output/device-plugin-volcano-"${build_version}".yaml
    cp "$CUR_DIR"/ascendplugin-310.yaml "$TOP_DIR"/output/device-plugin-310-"${build_version}".yaml
    cp "$CUR_DIR"/ascendplugin-310-volcano.yaml "$TOP_DIR"/output/device-plugin-310-volcano-"${build_version}".yaml
    cp "$CUR_DIR"/ascendplugin-310P.yaml "$TOP_DIR"/output/device-plugin-310P-"${build_version}".yaml
    cp "$CUR_DIR"/ascendplugin-310P-volcano.yaml "$TOP_DIR"/output/device-plugin-310P-volcano-"${build_version}".yaml
    cp "$CUR_DIR"/ascendplugin-310P-1usoc.yaml "$TOP_DIR"/output/device-plugin-310P-1usoc-"${build_version}".yaml
    cp "$CUR_DIR"/ascendplugin-310P-1usoc-volcano.yaml "$TOP_DIR"/output/device-plugin-310P-1usoc-volcano-"${build_version}".yaml

    cp "$CUR_DIR"/faultCode.json "$TOP_DIR"/output/faultCode.json
    cp "$CUR_DIR"/faultCustomization.json "$TOP_DIR"/output/faultCustomization.json

    sed -i "s#output/device-plugin#device-plugin#" "$TOP_DIR"/output/Dockerfile
}

function main() {
  clean
  build_plugin
  mv_file
  modify_version
  change_mod
}


main $1
