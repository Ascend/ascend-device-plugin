#!/bin/bash
CUR_DIR=$(dirname $(readlink -f $0))
TOP_DIR=$(realpath ${CUR_DIR}/..)

chmod +x ${CUR_DIR}/deploy.sh
while true
do
    case "$3" in
        --upgrade)
           ./deploy.sh --upgrade
            shift
            ;;
        --uninstall)
            ./deploy.sh --uninstall
            
            shift
            ;;
        --install)
            ./deploy.sh --install $4 $5 $6
            shift
            ;;
        --run)
             ./deploy.sh --install $4 $5 $6
            shift
            ;;
        --full)
             ./deploy.sh --install $4 $5 $6
            shift
            ;;
        --version)
            ./deploy.sh --version
            shift
            ;;
    		*)
      			break
      			;;
		esac
done