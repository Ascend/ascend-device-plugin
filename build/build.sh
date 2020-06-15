#!/bin/bash
CUR_DIR=$(dirname $(readlink -f $0))
set -x
dos2unix build_310.sh
chmod +x build_310.sh
${CUR_DIR}/build_310.sh ci
