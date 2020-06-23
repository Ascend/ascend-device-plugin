#!/bin/bash
CUR_DIR=$(dirname $(readlink -f $0))
TOP_DIR=$(realpath ${CUR_DIR}/..)
set -x
dos2unix build_310.sh
chmod +x build_310.sh
cd ${TOP_DIR}/output/
rm -rf *
tar -zcvf  ascend-device-plugin.tar.gz ../src/* ../go.mod ../ascendplugin.yaml ../Dockerfile