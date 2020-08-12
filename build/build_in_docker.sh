#!/bin/bash
CUR_DIR=$(dirname $(readlink -f $0))
TOP_DIR=$(realpath ${CUR_DIR}/..)
CONFIGDIR=${TOP_DIR}/src/plugin/config/config_910

OUTPUT_NAME="ascendplugin"
export PKG_CONFIG_PATH=${CONFIGDIR}:$PKG_CONFIG_PATH
function main() {
   cp ${TOP_DIR}/output/${OUTPUT_NAME}   /usr/local/bin/
}
main
