#!/bin/bash


docker_name="k8sDevicePlugin"

chattr -i /usr/local/HiAI


docker stop ${docker_name}
docker rm ${docker_name}


sudo docker run --net=host --user=root:root   -it --privileged \
-v /var/log/devicePlugin:/var/log/devicePlugin   \
-v /var/lib/kubelet/device-plugins:/var/lib/kubelet/device-plugins \
-v /usr/local/Ascend/driver:/usr/local/Ascend/driver  \
-v /opt/deviceplugin:/opt/deviceplugin \
--name ${docker_name}  ascend-device-plugin:latest   /bin/bash
  
 
