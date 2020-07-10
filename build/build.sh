#!/bin/bash
CUR_DIR=$(dirname $(readlink -f $0))
TOP_DIR=$(realpath ${CUR_DIR}/..)
set -x
dos2unix build_310.sh
chmod 550 build_310.sh
bash -x ${CUR_DIR}/build_310.sh ci
rm -rf  ${TOP_DIR}/output/*
cd  ${TOP_DIR}
tar -zcvf  ascend-device-plugin.tar.gz ./build ./output ./src  \
 ./ascend.yaml ./ascendplugin.yaml ./docker_run.sh ./Dockerfile ./go.mod \
 ./README.zh.md ./'Open Source Software Notice.md' ./LICENSE.md
mv ascend-device-plugin.tar.gz ${TOP_DIR}/output/