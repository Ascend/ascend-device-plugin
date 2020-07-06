#!/bin/bash

build_version="0.0.1"
build_time=$(date +'%Y-%m-%d_%T')
OUTPUT_NAME="ascendplugin"
DEPLOYNAME="deploy.sh"
DOCKER_FILE_NAME="Dockerfile"
PC_File="ascend_device_plugin.pc"
docker_images_name="ascend-k8sdeviceplugin:v0.0.1"

osname=$(grep -i ^id= /etc/os-release| cut -d"=" -f2 | sed 's/"//g');
ostype=$(arch)
if [ "${ostype}" = "aarch64" ]; then
  ostype="arm64"
else
  ostype="x86"
fi
PKGNAME="Ascend-k8s_device_plugin-${build_version}-${ostype}-linux.run"
TARNAME="Ascend-K8sDevicePlugin-${build_version}-${ostype}-Linux.tar.gz"
docker_zip_name="Ascend-K8sDevicePlugin-${build_version}-${ostype}-Docker.tar.gz"
# export so library path
export LD_LIBRARY_PATH=${SODIR}:${LD_LIBRARY_PATH}
export PKG_CONFIG_PATH=${CONFIGDIR}:$PKG_CONFIG_PATH


function clear_env() {
    rm -rf ${TOP_DIR}/output/*
    rm -rf ~/.cache/go-build
    if [ ! -d "${TOP_DIR}/makerunout" ]; then
        mkdir -p ${TOP_DIR}/makerunout
        chmod 750 ${TOP_DIR}/makerunout
    fi
}



function build_plugin() {

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
}

function mv_file() {

    mv ${TOP_DIR}/src/plugin/cmd/ascendplugin/${OUTPUT_NAME}   ${TOP_DIR}/output
    dos2unix ${TOP_DIR}/build/${DEPLOYNAME}
    chmod 550 ${TOP_DIR}/build/${DEPLOYNAME}
    cp ${TOP_DIR}/build/${DEPLOYNAME}     ${TOP_DIR}/output

}

function copy2runpackage() {
    mv ${TOP_DIR}/src/plugin/cmd/ascendplugin/${OUTPUT_NAME}   ${TOP_DIR}/makerunout
    cp ${TOP_DIR}/build/${DEPLOYNAME}     ${TOP_DIR}/makerunout/
    if [ ! -d "${TOP_DIR}/makerunout/script" ]; then
        mkdir -p ${TOP_DIR}/makerunout/script
        chmod 750 ${TOP_DIR}/makerunout/script
    fi
    chmod 550 ${TOP_DIR}/build/script/uninstall.sh
    cp ${TOP_DIR}/build/script/uninstall.sh ${TOP_DIR}/makerunout/script/
}

function zip_file(){
    cd ${TOP_DIR}/output
    tar -zcvf ${TARNAME}  ${OUTPUT_NAME}  ${DEPLOYNAME}
    if [ $? == 0 ]; then
        echo "build device plugin success"
    fi
    rm -f ${OUTPUT_NAME}  ${DEPLOYNAME}
}

function make_run_package() {
    chmod 550  ${CUR_DIR}/script/makepackgeinstall.sh
    dos2unix  ${CUR_DIR}/script/makepackgeinstall.sh
    cp ${CUR_DIR}/script/makepackgeinstall.sh  ${TOP_DIR}/makerunout
    dirname="${ostype}-$(get_os_name)$(get_os_version)"
    if [ ! -d "${TOP_DIR}/output/${dirname}" ]; then
        mkdir -p "${TOP_DIR}/output/${dirname}"
        chmod 750 ${TOP_DIR}/output/${dirname}
    fi
    if [ -d "${TOP_DIR}/tools/makeself-release-2.4.0" ]; then
        rm -rf ${TOP_DIR}/tools/makeself-release-2.4.0
    fi
    cd ${TOP_DIR}/tools || retrun
    unzip makeself-release-2.4.0.zip
    cd ${TOP_DIR}/tools/makeself-release-2.4.0 || return
    cp makeself.sh ${CUR_DIR}/script
    cp makeself-header.sh ${CUR_DIR}/script
    cd ${CUR_DIR}/script || retrun
    patch -p0 < mkselfmodify.patch

    ./makeself.sh --nomd5 --nocrc --header ./makeself-header.sh  --help-header \
    ./help.info ../../makerunout "${PKGNAME}" ascendplugin ./makepackgeinstall.sh
    mv ${PKGNAME} ${TOP_DIR}/output/${dirname}
    rm -rf ${TOP_DIR}/makerunout
    rm -f ${TOP_DIR}/output/${OUTPUT_NAME}  ${TOP_DIR}/output/${DEPLOYNAME}
}
function build_docker_images()
{
    cd ${TOP_DIR}
    docker rmi ${docker_images_name}
    docker build -t ${docker_images_name} .
    docker save ${docker_images_name} | gzip > ${docker_zip_name}
    mv ${docker_zip_name} ./output/
}

function get_os_name() {
    lsb_release -i | awk '{print $3}' | tr 'A-Z' 'a-z'
}

function get_os_version() {
    local os_name=$(get_os_name)
    declare -A os_version=(["ubuntu"]="18.04" ["centos"]="7.6" ["euleros"]="2.8" ["debian"]="9.9")
    for key in "${!os_version[@]}"; do
        if [ $key == $os_name ]; then
            echo "${os_version[$key]}"
            return 0
        fi
    done
    exit 1
}

function getVendorMode() {
    go mod download
    go mod vendor
}