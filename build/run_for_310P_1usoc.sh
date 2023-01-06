#!/bin/bash
set -e

# create soft link for ubuntu image
os="$(cat /etc/*release* | grep -i "ubuntu")"
if [[ "$os" != "" ]]
then
    echo -e "[INFO]\t $(date +"%F %T:%N")\t use ubuntu image, so create soft link \"/lib64\" for \"/lib\""
    ln -s /lib /lib64 2>&1 >> /dev/null
fi

umask 027

echo -e "[INFO]\t $(date +"%F %T:%N")\t create driver's related directory"
mkdir -m 750 /var/driver -m 750 /var/dmp -m 750 /usr/slog -p -m 750 /home/drv/hdc_ppc

echo -e "[INFO]\t $(date +"%F %T:%N")\t modify owner and permission"
chown HwDmUser:HwDmUser /var/dmp
chown HwHiAiUser:HwHiAiUser /var/driver
chown HwHiAiUser:HwHiAiUser /home/drv/hdc_ppc
chown HwHiAiUser:HwHiAiUser /usr/slog

# log process run in background
echo -e "[INFO]\t $(date +"%F %T:%N")\t start slogd server in background"
su - HwHiAiUser -c "export LD_LIBRARY_PATH=/usr/local/Ascend/driver/lib64/ && /var/slogd &"
echo -e "[INFO]\t $(date +"%F %T:%N")\t start dmp_daemon server in background"
# dcmi interface process run in background
su - HwDmUser -c "export LD_LIBRARY_PATH=/usr/local/Ascend/driver/lib64/ && /var/dmp_daemon -I -M -U 8087 &"

export LD_LIBRARY_PATH=/usr/local/lib:/usr/local/Ascend/driver/lib64/driver:/usr/local/Ascend/driver/lib64/common:/usr/local/Ascend/add-ons:/usr/local/Ascend/driver/lib64:/usr/local/dcmi
echo -e "[INFO]\t $(date +"%F %T:%N")\t start ascend device plugin server"
/usr/local/bin/device-plugin -useAscendDocker=false -volcanoType=true -presetVirtualDevice=true -logFile=/var/log/mindx-dl/devicePlugin/devicePlugin.log -logLevel=0

