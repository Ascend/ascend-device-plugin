#!/bin/bash
CUR_DIR=$(dirname $(readlink -f $0))
TOP_DIR=$(realpath ${CUR_DIR}/..)
set -x
dos2unix build_310.sh
chmod 550 build_310.sh
bash -x ${CUR_DIR}/build_310.sh ci
cd ${TOP_DIR}/output/
rm -rf *
tar -zcvf  ascend-device-plugin.tar.gz ../build ../output ../src  \
 ../ascend.yaml ../ascendplugin.yaml ../docker_run.sh ../Dockerfile ../go.mod \
 ../README.md