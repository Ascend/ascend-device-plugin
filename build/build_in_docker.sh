#!/bin/bash
CUR_DIR=$(dirname $(readlink -f $0))
TOP_DIR=$(realpath ${CUR_DIR}/..)

OUTPUT_NAME="ascendplugin"

function main() {
   cp ${TOP_DIR}/output/${OUTPUT_NAME}   /usr/local/bin/
}
main
