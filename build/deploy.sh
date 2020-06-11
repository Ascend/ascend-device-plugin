#!/bin/bash

unset http_proxy https_proxy

CURRENT_PATH=$(cd "$(dirname "$0")"; pwd)
APP_NAME="ascendplugin"
SERVICENAME="deviceplugin.service"
SERVICE_PATH=/etc/systemd/system
TARGET_DIR=/usr/local/bin
lograte_path=/etc/logrotate.d

logrotate_file="k8s-devicePlugin"
log_bak=/var/log/devicePlugin/deploy_old

cron_path=/etc/cron.daily
cron_file="k8s-deploy"

TMPFILE=/tmp/deploy.sh.tmp

SCRIP_NAME="$0"
ARGS="$1"


target_kubelet_version=v1.13
target_go_version=go1.11

HOSTNAME=$(hostname)


# assign log file 
mkdir -p  /var/log/devicePlugin
logfile=/var/log/devicePlugin/deploy.log


function log_info()
{
	echo "[`date +%Y-%m-%d-%H:%M:%S`] [INFO] $1"  | tee -a $logfile
}

function log_error()
{
	echo "[`date +%Y-%m-%d-%H:%M:%S`] [ERROR] $1"  | tee -a $logfile
}

function lograte_setting()
{
	mkdir -p ${log_bak}

	if [ -e ${lograte_path}/${logrotate_file} ]
	then
		rm -f ${lograte_path}/${logrotate_file}
	fi
	cat > ${lograte_path}/${logrotate_file} <<EOF

/var/log/devicePlugin/deploy.log {
        su root root
        daily
        size=5M
        rotate 20
        missingok
        nocompress
        olddir   /var/log/devicePlugin/deploy_old
        copytruncate
    }

EOF

cron_setting

}


function cron_setting()
{
	if [ -e ${cron_path}/${cron_file} ]
	then
		rm -f ${cron_path}/${cron_file}
	fi

	cat > ${cron_path}/${cron_file} <<EOF
#!/usr/bin/sh
/usr/sbin/logrotate  ${lograte_path}/${logrotate_file}

EOF

chmod 777 ${cron_path}/${cron_file}

}


function install_plugin()
{

	if [ -e ${TARGET_DIR}/${APP_NAME} ]
	then
		log_error "old version ${APP_NAME} is installed, use '${APP_NAME} --version'  command to check version"
		log_error  "please use upgrade command or uninstall it first "
		exit 1
	fi

	log_info "${APP_NAME} install"


	check_env


	# check app exist

    if [ -e ${CURRENT_PATH}/${APP_NAME} ]
    then
    	log_info "install package is ready, installing"
	else
		log_error "fail to find ${APP_NAME} package in ${CURRENT_PATH},install failed"
		exit 1
	fi

	cp ${CURRENT_PATH}/${APP_NAME} ${TARGET_DIR}

	dp_config_file

	if [ -e ${TARGET_DIR}/${APP_NAME} ]
	then
		log_info "${APP_NAME} install success"
		check_version
		read -p "run ${APP_NAME} now an set it to start on startup [y/n]?" CONT
		if [ "$CONT" = "y" ]; then
			device_plugin_service_start
		else
  			echo "exit"
  			exit 0
		fi
	else
		log_error "fail to find ${APP_NAME} in target dir,install failed"
		exit 1
	fi

}

function upgrade_plugin()
{

    read -p "During upgrade, new jobs that use Ascend910 allocate will be failed, Do you want continue [y/n]?" CONTN
	if [ "$CONTN" = "y" ]; then
		uninstall_plugin
		install_plugin
	else
		echo "exit"
		exit 0
	fi

}

function uninstall_plugin()
{
	log_info "uninstall old ${APP_NAME} version"

	if [ -e ${TARGET_DIR}/${APP_NAME} ]
    then
    	log_info "old ${APP_NAME} version: "
    	check_version
    	read -p "Do you want continue [y/n]?" CONT
		if [ "$CONT" = "y" ]; then
			check_and_kill_process ${APP_NAME}
			clean_service_config
			rm -f ${TARGET_DIR}/${APP_NAME}

		else
  			echo "exit"
  			exit 0
		fi
    else
    	log_info "${APP_NAME} has not installed"
	fi


	if [ -e ${TARGET_DIR}/${APP_NAME} ]
    then
    	log_error "${APP_NAME} uninstall failed"
    else
    	log_info "${APP_NAME} uninstall success"
	fi
}

function clean_service_config()
{
	if [ -e ${SERVICE_PATH}/${SERVICENAME} ]
    then
    	rm -f ${SERVICE_PATH}/${SERVICENAME}
    	log_info "clean ${SERVICENAME} file"
    fi


}

function check_env()
{
	check_kubelet
	check_go_version
}

function check_version()
{
    if [ -e ${TARGET_DIR}/${APP_NAME} ]
    then
        ${APP_NAME} --version | tee -a $logfile
    else
    	log_error "${APP_NAME} has not install yet,please check!"
        exit 1
    fi
}

function version_ge() { test "$(echo "$@" | tr " " "\n" | sort -rV | head -n 1)" == "$1"; }


function check_kubelet()
{
	check_kubelet_install
	log_info "check kubelet version, recommend kubelet version is >= ${target_kubelet_version}"

	kubelet_version=$(kubectl get node | awk  'NR==2  {print $5}')


	if version_ge "$kubelet_version" $target_kubelet_version; then
		log_info "kubelet version is ${kubelet_version}, kubelet env ok"
	else
		log_error "${APP_NAME} install failed"
		log_error "kubelet version $kubelet_version is less than $target_kubelet_version, please upgrade your kubenetes !"
	    exit 1

	fi
}

function check_go_version()
{
	check_golang_install

	log_info "check golang version, recommend golang version is >= ${target_go_version} "

	go_version=$(go version |awk '{print $3}')

	if version_ge "$go_version" $target_go_version; then
		log_info "golang version is ${go_version}, golang env ok"
	else
		log_error "${APP_NAME} install failed"
		log_error "golang version $go_version is less than $target_go_version, please upgrade your golang !"
	    exit 1
	fi

}

function check_kubelet_install()
{

	is_ready=$(kubectl get node |grep $(hostname) |awk '{print $2}')

	if [[ "${is_ready}" != "Ready" ]];then
		log_error "${APP_NAME} install failed"
		log_error "kubernetes is not installed,Recommended version >= ${target_kubelet_version}"
	    exit 1
	fi
}

function check_golang_install()
{
	log_info "check golang env"

	if ! [ -x "$(command -v go)" ]; then
		log_error "${APP_NAME} install failed"
		log_error "golang is not installed,Recommended version >= ${target_go_version}"
	    exit 1
	fi
}


function dp_config_file()
{
cat > ${SERVICENAME} <<EOF
[Unit]
Description=ascendplugin: The Ascend910 k8s device plugin
Documentation=https://kubernetes.io/docs/
After=kubelet.service

[Service]
ExecStart=/usr/local/bin/ascendplugin --mode=ascend310 --fdFlag=true
ExecReload=/bin/kill -s HUP $MAINPID
ExecStop=/bin/kill -s QUIT $MAINPID
Restart=no
StartLimitInterval=0
RestartSec=10
KillMode=process

[Install]
WantedBy=multi-user.target

EOF
mv  -f ${SERVICENAME}  ${SERVICE_PATH}

}

function device_plugin_service_start()
{
	if [ -e ${TARGET_DIR}/${APP_NAME} ]
    then
        log_info "start ${APP_NAME} "
    else
    	log_error "${APP_NAME} has not install yet,please check!"
        exit 1
    fi

    if [ -e ${SERVICE_PATH}/${SERVICENAME} ]
    then
        systemctl enable deviceplugin.service
        systemctl daemon-reload &&  systemctl start deviceplugin.service
        check_device_plugin_status
    else
    	log_error "${SERVICE_PATH}/${SERVICENAME} not found, device plugin service start failed"
        exit 1
    fi

}

function check_device_plugin_status()
{
	log_info "check ${APP_NAME} status "  

	deviceplugin_state=$(systemctl status ${SERVICENAME} |grep Active | awk  '{print $2}')

	if [[ ${deviceplugin_state} == "active" ]];then
		log_info "${APP_NAME} runing success"  
	else
		log_error "${APP_NAME} is runing failed, status check command: 'systemctl status ${SERVICENAME}'."  
	fi

}

check_and_kill_process(){

    if [ "$1" = "" ];
    then
        return 
    fi

    if pgrep $1 2>/dev/null; then
    	log_info "Terminating old $1 process"
    	systemctl stop ${SERVICENAME}
    fi
}

check_deploy_process()
{
	if [ -e ${TMPFILE} ]
    then
    	log_error "The deployment program is alreay running, exit !"
    	exit 1
    else
    	touch ${TMPFILE}
    	chmod 600 $TMPFILE
    fi

	trap "rm -f ${TMPFILE}; exit"  0 1 2 3 9 15
}


function help()
{
	echo " ${SCRIP_NAME} --install  usage:  install k8s-device-plugin"
	echo " ${SCRIP_NAME} --upgrade  usage:  upgrade k8s-device-plugin"
	echo " ${SCRIP_NAME} --uninstall  usage:  uninstall k8s-device-plugin"

	echo "'systemctl start ${SERVICENAME}' 	usage:  start k8s-device-plugin"
	echo "'systemctl stop ${SERVICENAME}' 	usage:  stop k8s-device-plugin"
	echo "'systemctl restart ${SERVICENAME}' 	usage:  restart k8s-device-plugin"
	echo "'systemctl status ${SERVICENAME}' 	usage:  check status of  k8s-device-plugin"
	echo "'systemctl enable ${SERVICENAME}' 	usage:  enable k8s-device-plugin start on startup "
	echo "'systemctl disable ${SERVICENAME}' 	usage:  disable k8s-device-plugin start on startup"

}

function main()
{

	log_info "***********************************devicePlugin deploy start***************************************" 
	log_info "deploy log path: ${logfile}" 
	check_deploy_process

	lograte_setting
	

	if [[ ${ARGS} == "--install" ]];then
		install_plugin
	elif [[ ${ARGS} == "--upgrade" ]];then
		upgrade_plugin
	elif [[ ${ARGS} == "--uninstall" ]];then
		uninstall_plugin
	elif [[ ${ARGS} == "--help" ]];then
		help
	else
		echo "command not support !"
	fi
}

main


