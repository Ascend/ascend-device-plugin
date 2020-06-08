#!/bin/bash
set -x
dos2unix build_310.sh
chmod +x build_310.sh
sh -x build_310.sh

dos2unix build_910.sh
chmod +x build_910.sh
sh -x build_910.sh
