#!/bin/bash
UNINSTALL_PATH="./script/uninstall.sh"
INSTALL_DIR="k8s_device_plguin"
OUTPUT_NAME="./ascendplugin"
DEPLOYNAME="./deploy.sh"

install_tool()
{
    inst=y
  	if [ -z $install_path ]; then
        echo "error :installpath is empty"
        inst=n
    else
        echo ${install_path}
  	    install_ascend_path="${install_path}/${INSTALL_DIR}"
        if [ ! -d "${install_ascend_path}" ]; then
            chmod 750 ${install_path}
            mkdir -p "${install_ascend_path}"
            chmod 750 ${install_ascend_path}
        fi
        if [ ! -d ${install_ascend_path}/script ]; then
            mkdir ${install_ascend_path}/script
            chmod 750 ${install_ascend_path}/script
        fi
        cp ${UNINSTALL_PATH} ${install_ascend_path}/script
        chmod 550 ${install_ascend_path}/script/${UNINSTALL_PATH}
        cp ${OUTPUT_NAME} ${install_ascend_path}
        chmod 550 ${install_ascend_path}/${OUTPUT_NAME}
        cp ${DEPLOYNAME} ${install_ascend_path}/
        chmod 550  ${install_ascend_path}/${DEPLOYNAME}
      if [ "${inst}" != "n" ]&&[ -d "${install_ascend_path}/script" ]; then

          echo "-----------------------------------------------------------------------------------------------------"
          echo "INFO: your install path is ${install_path}"
          echo "INFO: install is success"
          echo "-----------------------------------------------------------------------------------------------------"
      else
	        echo "error : install is failed"
      fi
    fi
}

while true
do
    case "$3" in
        --upgrade)
           is_upgrade=y
            shift
            ;;
        --uninstall)
             is_uninstall=y
            shift
            ;;
        --install-path=*)
            is_install_path=y
            echo ${3}
			      check_null=${3}
            install_path=`echo $3 | cut -d"=" -f2 `
            # 去除指定安装目录后所有的 "/"
            install_path=`echo $install_path | sed "s/\/*$//g"`
            shift
            ;;
        --run)
             is_install=y
            shift
            ;;
        --full)
            is_install=y
            shift
            ;;
    		*)
      			break
      			;;
		esac
done

mkdir -p ${install_path}/${INSTALL_DIR}
chmod 750 ${install_path}/${INSTALL_DIR}
install_bin_path=${install_path}/${INSTALL_DIR}


# 安装为相对路径时报错
if [ "${install_path}" == ".." ]||[ "${install_path}" == "." ]; then
    is_install_path=n
	  echo "Error :please follow the installation directory after the --install-path=<Absolute path>"
fi

# 命令行安装
# 单纯只有--install-path=且输入地址为空的判定处理
if [ "${is_install_path}" == "y" ]&&[ -z ${check_null} ]&&[ x"${is_install}" == "x" ]; then
	  echo "Error :installpath is empty,Please follow the installation directory after the [--install-path=]."
fi

# 单纯只有--install-path=且输入地址不为空的处理安装
if [ "${is_install_path}" == "y" ]&&[ ! -z ${check_null} ]&&[ x"${is_install}" == "x" ]; then
	  install_tool
fi


# 单纯只有--install选项的判定处理
if [ "${is_install}" == "y" ]&&[ x"${is_install_path}" == "x" ]; then
	  echo "Error :Only the <--run> or <--full> command can't tell me you hit the installation directory. "
	  echo "Please enter the <--install-path=> command to tell me the directory where you want to install."

fi

# install和install_path都有的情况下的安装
if [ "${is_install_path}" == "y" ]&&[ "${is_install}" == "y" ]; then
	  install_tool
fi

# 判断卸载情况然后执行动作1
if [ "${is_uninstall}" == "y" ]&&[ "${is_install_path}" == "y" ]; then
  	if [ ! -d "${install_path}/${INSTALL_DIR}" ]||[ ! -d "${install_path}/${INSTALL_DIR}/script" ]; then
    		echo "Error :uninstall failed, Incorrect directory or incomplete command the <--uninstall> command needs to be used."
    		echo "You should use <--install-path=> command when you want uninstall. "

  	else
  		  rm -rf "${install_path}/${INSTALL_DIR}"
  		  echo "Uninstalled successfully"
  	fi
fi

# 判断卸载情况然后执行动作2
if [ "${is_uninstall}" == "y" ]&&[ "x""${check_null}" == "x" ]; then
  	echo "Error: uninstall failed, Incorrect directory or incomplete command."
  	echo "The <--uninstall> command needs to be used with the <--install-path=>."
fi


# 判断更新命令然后执行动作1
if [ "${is_upgrade}" == "y" ]&&[ "${is_install_path}" == "y" ]; then
   install_ascend_path=${install_path}/${INSTALL_DIR}
  	if [ ! -d "${install_ascend_path}" ]||[ ! -d "${install_ascend_path}" ]; then
    		echo "Error : Update failed, Incorrect directory or incomplete command."
    		echo "The <--upgrade> command needs to be used with the <--install-path=> command."

  	else
        rm -rf "${install_ascend_path}/script"
        rm -rf "${install_ascend_path}"
        mkdir -p  "${install_ascend_path}/script"
        chmod 750 ${install_ascend_path}/script
        cp ${UNINSTALL_PATH} ${install_ascend_path}/script
        cp ${OUTPUT_NAME} ${install_ascend_path}/
        cp ${DEPLOYNAME} ${install_ascend_path}/
        chmod 550  ${install_ascend_path}/script/${UNINSTALL_PATH}
        chmod 550  ${install_ascend_path}/${OUTPUT_NAME}
        chmod 550  ${install_ascend_path}/${DEPLOYNAME}
        echo "Upgrade successfully"
  	fi
fi
# 判断更新命令然后执行动作(未输入路径情况)
if [ "${is_upgrade}" == "y" ]&&[ "x""${check_null}" == "x" ]; then
  	echo "Error:The path of the update tool is empty or the command input is incorrect."
  	echo "The <--upgrade> command needs to be used with the <--install-path=> command. "
fi