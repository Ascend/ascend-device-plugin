#!/bin/bash
# Perform build k8s-device-plugin
# Copyright @ Huawei Technologies CO., Ltd. 2021-2021. All rights reserved

set -e
set -x
CUR_DIR="$(dirname "$(readlink -f "$0")")"
TOP_DIR=$(realpath "${CUR_DIR}"/../..)

mkdir -p "$TOP_DIR"/ascend-device-plugin/build
mkdir -p "$TOP_DIR"/ascend-device-plugin/output
cd "$TOP_DIR"/ascend-device-plugin/

# 文件格式转换
dos2unix "$TOP_DIR"/ascend-device-plugin/build/build.sh
chmod 550 "$TOP_DIR"/ascend-device-plugin/build/*
cd "$TOP_DIR"/ascend-device-plugin/build/

version="v2.0.1"

# 修改build.sh文件
sed -i "s/build_verison=.*/build_verison=\"${version}\"/" "$TOP_DIR"/ascend-device-plugin/build/build.sh
sed -i "s/docker_images_name=.*/docker_images_name=\"ascend-k8sdeviceplugin:${version}\"/" "$TOP_DIR"/ascend-device-plugin/build/build.sh

sed -i "s/x86/amd64/" "$TOP_DIR"/ascend-device-plugin/build/build.sh

# 修改三个yaml的镜像
sed -i "s/ascend-k8sdeviceplugin:.*/ascend-k8sdeviceplugin:${version}/" "$TOP_DIR"/ascend-device-plugin/ascendplugin.yaml
sed -i "s/ascend-k8sdeviceplugin:.*/ascend-k8sdeviceplugin:${version}/" "$TOP_DIR"/ascend-device-plugin/ascendplugin-volcano.yaml
sed -i "s/ascend-k8sdeviceplugin:.*/ascend-k8sdeviceplugin:${version}/" "$TOP_DIR"/ascend-device-plugin/ascendplugin-310.yaml
sed -i "s/ascend-k8sdeviceplugin:.*/ascend-k8sdeviceplugin:${version}/" "$TOP_DIR"/ascend-device-plugin/ascendplugin-710.yaml
# 执行构建
bash -x "$TOP_DIR"/ascend-device-plugin/build/build.sh dockerimages

# 拷贝出修改后的yaml镜像
cp "$TOP_DIR"/ascend-device-plugin/ascendplugin.yaml "$TOP_DIR"/ascend-device-plugin/output/ascendplugin-"${version}".yaml
cp "$TOP_DIR"/ascend-device-plugin/ascendplugin-volcano.yaml "$TOP_DIR"/ascend-device-plugin/output/ascendplugin-volcano-"${version}".yaml
cp "$TOP_DIR"/ascend-device-plugin/ascendplugin-310.yaml "$TOP_DIR"/ascend-device-plugin/output/ascendplugin-310-"${version}".yaml
cp "$TOP_DIR"/ascend-device-plugin/ascendplugin-710.yaml "$TOP_DIR"/ascend-device-plugin/output/ascendplugin-710-"${version}".yaml