#!/bin/bash
CUR_DIR=$(dirname $(readlink -f $0))
TOP_DIR=$(realpath ${CUR_DIR}/..)
CONFIGDIR=${TOP_DIR}/src/plugin/config/config_910

OUTPUT_NAME="ascendplugin"
export PKG_CONFIG_PATH=${CONFIGDIR}:$PKG_CONFIG_PATH
function main() {
    rm -rf ${TOP_DIR}/output/*
    rm -rf ~/.cache/go-build
    rm -rf /tmp/gobuildplguin
    mkdir -p /tmp/gobuildplguin
    chmod 750 /tmp/gobuildplguin
    cd ${TOP_DIR}/src/plugin/cmd/ascendplugin
    go build -ldflags "-X main.BuildName=${OUTPUT_NAME} \
            -X main.BuildVersion=${build_version} \
            -buildid none     \
            -s   \
            -tmpdir /tmp/gobuildplguin" \
            -o ${OUTPUT_NAME}       \
            -trimpath

    ls ${OUTPUT_NAME}
    if [ $? -ne 0 ]; then
        echo "fail to find ascendplugin"
        exit 1
    fi
    cp ${TOP_DIR}/src/plugin/cmd/ascendplugin/${OUTPUT_NAME}   /usr/local/bin/
}
main
