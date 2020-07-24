# ascend-device-plugin

## 1 环境依赖

| 检查项          | 要求                                                         |
| --------------- | ------------------------------------------------------------ |
| dos2unix        | 已安装。                                                     |
| atlas的驱动版本 | 大于等于1.73.5.0.B050 |
| Go语言环境版本  | 大于等于1.14.3。                                             |
| gcc版本         | 大于等于7.3.0。                                              |
| Kubernetes版本  | 大于等于1.13.0。                                             |
| Docker环境      | 已安装Docker，可以从镜像仓拉取镜像或已有对应操作系统的镜像。 |
## 2 编译

1. 执行以下目录安装最新版本的pkg-config。

   ```shell 
   apt-get install -y pkg-config
   ```

2. 执行以下命令，设置环境变量。

   ``` shell 
   export GO111MODULE=on

   export GOPROXY=https://gocenter.io

   export GONOSUMDB=\
   ```
   GOPROXY代理地址请根据实际选择，可通过go mod download命令进行检查。
   如出现x509错误，需要配置证书。
   
3. 进入ascend-device-plugin目录，执行以下命令，修改yaml文件。

   ```shell 
   vi ascendplugin.yaml
   ```

```yaml
apiVersion: apps/v1
   kind: DaemonSet
   metadata:
     name: ascend-device-plugin-daemonset
     namespace: kube-system
   spec:
     selector:
       matchLabels:
         name: ascend-device-plugin-ds
     updateStrategy:
       type: RollingUpdate
     template:
       metadata:
         annotations:
           scheduler.alpha.kubernetes.io/critical-pod: ""
         labels:
           name: ascend-device-plugin-ds
       spec:
         tolerations:
           - key: CriticalAddonsOnly
             operator: Exists
           - key: huawei.com/Ascend910  #资源名称，根据芯片类型设置。
             operator: Exists
             effect: NoSchedule
           - key: "ascendplugin"
             operator: "Equal"
             value: "v2"
             effect: NoSchedule
         priorityClassName: "system-node-critical"
         nodeSelector:
           accelerator: huawei-Ascend910  #根据芯片类型设置标签名称。
         containers:
         - image: ascend-k8sdeviceplugin:v0.0.1  #镜像名称及版本号。
           name: device-plugin-01
           command: [ "/bin/bash", "-c", "--"]
           args: [ "./build/build_in_docker.sh;ascendplugin  --useAscendDocker=${USE_ASCEND_DOCKER}" ] #使用Ascend310，则需要增加--mode=ascend310
           securityContext:
             privileged: true
           imagePullPolicy: Never
           volumeMounts:
             - name: device-plugin
               mountPath: /var/lib/kubelet/device-plugins
             - name: hiai-driver
               mountPath: /usr/local/Ascend/driver  #驱动安装目录，用户根据实际填写。
             - name: log-path
               mountPath: /var/log/devicePlugin
         volumes:
           - name: device-plugin
             hostPath:
               path: /var/lib/kubelet/device-plugins
           - name: hiai-driver
             hostPath:
               path: /usr/local/Ascend/driver  #驱动安装目录，用户根据实际填写。
           - name: log-path
             hostPath:
               path: /var/log/devicePlugin
```

4. 执行以下命令，编辑Dockerfile文件，将镜像修改为查询的镜像名及版本号。

   ```shell
   vi /home/test/ascend-device-plugin/build/Dockerfile
   ```
#用户根据实际选择需要使用的带go编译的基础镜像，可通过docker images命令查询。
   ``` yaml
   FROM golang:1.13.11-buster as build
   
   #是否使用昇腾Docker，true表示使用，false表示不使用（将会使用原生Docker）。
   ENV USE_ASCEND_DOCKER true
   
   ENV GOPATH /usr/app/
   
   ENV GO111MODULE off
   
   ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
   #libdrvdsmi_host.so所在目录，Ascend 310和Ascend 910目录不同。
   ENV LD_LIBRARY_PATH  /usr/local/Ascend/driver/lib64/driver:/usr/local/Ascend/driver/lib64/common
   
   RUN mkdir -p /usr/app/src/ascend-device-plugin
   
   COPY . /usr/app/src/Ascend-device-plugin
   
   WORKDIR /usr/app/src/Ascend-device-plugin
   ```
   
5. 进入ascend_device_plugin.pc文件所在目录，执行以下命令，查看以下路径是否正确，根据实际修改。

   - Ascend 310目录：ascend-device-plugin/src/plugin/config/config_310
   - Ascend 910目录：ascend-device-plugin/src/plugin/config/config_910
   ```shell
   vi ascend_device_plugin.pc
   ```
   
   ```pkg-config
   #Package Information for pkg-config
   #驱动安装目录，根据实际填写。
   prefix=/usr/local/Ascend
   #dsmi动态库地址，根据实际修改。
   libdriver=${prefix}/driver/lib64
   #dsmi驱动头文件dsmi_common_interface.h所在目录。
   includedir=${prefix}/driver/kernel/inc/driver/
   Name: ascend_docker_plugin
   Description: Ascend device plugin
   Version: 0.0.1
   Libs: -L${libdriver}/    -ldrvdsmi_host
   Cflags: -I${includedir}
   ```
   
   支持修改插件镜像的名称，build目录下build_common.sh中修改“docker_images_name”即可。

6. 进入“/ascend-device-plugin/build”目录，执行以下命令，查看CONFIGDIR是否正确。
   ``` shell
   vi build_in_docker.sh
   ```
   
   ``` shell
   #!/bin/bash
   set -x
   CUR_DIR=$(dirname $(readlink -f $0))
   TOP_DIR=$(realpath ${CUR_DIR}/..)
   CONFIGDIR=${TOP_DIR}/src/plugin/config/config_910 #默认使用config_910，    使用Ascend 310请改为config_310。
   
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
       cp ${TOP_DIR}/src/plugin/cmd/ascendplugin/${OUTPUT_NAME}              /usr/local/bin/
   }
   main
   ```
   #### 简易步骤
   为了简化在3，4，5，6步骤，提供了一个shell脚本进行快速修改：
   在./build/中创建并执行shell文件
   ```shell
     #!/bin/bash
     ASCNED_TYPE=910 #根据芯片类型选择310或910。
     ASCNED_INSTALL_PATH=/usr/local/Ascend  #驱动安装路径，根据实际修改。
     USE_ASCEND_DOCKER=false  #是否使用昇腾Docker。
     
     
     CUR_DIR=$(dirname $(readlink -f $0))
     TOP_DIR=$(realpath ${CUR_DIR}/..)
     LD_LIBRARY_PATH_PARA1=${ASCNED_INSTALL_PATH}/driver/lib64/driver
     LD_LIBRARY_PATH_PARA2=${ASCNED_INSTALL_PATH}/driver/lib64
     TYPE=Ascend910
     PKG_PATH=${TOP_DIR}/src/plugin/config/config_910
     PKG_PATH_STRING=\$\{TOP_DIR\}/src/plugin/config/config_910
     LIBDRIVER="driver/lib64/driver"
     if [ ${ASCNED_TYPE} == "310"  ]; then
       TYPE=Ascend310
       LD_LIBRARY_PATH_PARA1=${ASCNED_INSTALL_PATH}/driver/lib64
       PKG_PATH=${TOP_DIR}/src/plugin/config/config_310
       PKG_PATH_STRING=\$\{TOP_DIR\}/src/plugin/config/config_310
       LIBDRIVER="/driver/lib64"
       sed -i "s#ascendplugin  --useAscendDocker=\${USE_ASCEND_DOCKER}#ascendplugin --mode=ascend310 --useAscendDocker=${USE_ASCEND_DOCKER}#g" ${TOP_DIR}/ascendplugin.yaml
     fi
     sed -i "s/Ascend[0-9]\{3\}/${TYPE}/g" ${TOP_DIR}/ascendplugin.yaml
     sed -i "s#ath: /usr/local/Ascend/driver#ath: ${ASCNED_INSTALL_PATH}/driver#g" ${TOP_DIR}/ascendplugin.yaml
     sed -i "/^ENV LD_LIBRARY_PATH /c ENV LD_LIBRARY_PATH ${LD_LIBRARY_PATH_PARA1}:${LD_LIBRARY_PATH_PARA2}/common" ${TOP_DIR}/Dockerfile
     sed -i "/^ENV USE_ASCEND_DOCKER /c ENV USE_ASCEND_DOCKER ${USE_ASCEND_DOCKER}" ${TOP_DIR}/Dockerfile
     sed -i "/^libdriver=/c libdriver=$\{prefix\}/${LIBDRIVER}" ${PKG_PATH}/ascend_device_plugin.pc
     sed -i "/^prefix=/c prefix=${ASCNED_INSTALL_PATH}" ${PKG_PATH}/ascend_device_plugin.pc
     sed -i "/^CONFIGDIR=/c CONFIGDIR=${PKG_PATH_STRING}" ${CUR_DIR}/build_in_docker.sh
   ```
7. 执行以下命令，根据实际选择执行的脚本，生成二进制和镜像文件。

   Ascend 910请选择build910.sh，Ascend 310请选择build_310.sh。
   ```shell
   cd /home/test/ascend-device-plugin/build/
   chmod +x build_910.sh
   ./build_910.sh dockerimages
   ```
8. 执行以下命令，查看生成的软件包。
   ``` shell
   ll /home/test/ascend-device-plugin/output
   ```
   X86和ARM生成的软件包名不同，以下示例为ARM环境：
   **Ascend-K8sDevicePlugin-xxx**-arm64-Docker.tar.gz：K8S设备插件镜像。
   **Ascend-K8sDevicePlugin-xxx**-arm64-Linux.tar.gz：K8S设备插件二进制安装包。
   ```
   drwxr-xr-x 2 root root     4096 Jun  8 18:42 ./
   drwxr-xr-x 9 root root     4096 Jun  8 17:12 ../
   -rw-r--r-- 1 root root 29584705 Jun  9 10:37 Ascend-K8sDevicePlugin-xxx-arm64-Docker.tar.gz
   -rwxr-xr-x 1 root root  6721073 Jun  9 16:20 Ascend-K8sDevicePlugin-xxx-arm64-Linux.tar.gz
   ```
## 3 使用ascend-deviceplugin镜像创建Daemonset

#### 操作步骤

以下操作以ARM平台下生成的tar.gz文件为例。

1. 进入生成的Docker软件包所在目录，执行以下命令，导入Docker镜像。
   ``` shell
   cd /home/test/ascend-device-plugin/output

   docker load <Ascend-K8sDevicePlugin-xxx-arm64-Docker.tar.gz
   ```
2. 执行如下命令，给带有Ascend 910（或Ascend 310）的节点打标签。
   ```shell
   kubectl label nodes  localhost.localdomain accelerator=huawei-Ascend910
   ```
   localhost.localdomain为有Ascend 910（或Ascend 310）的节点名称，可通过**kubectl get node**命令查看。

3. 执行以下命令，部署DaemonSet。
   ``` shell
   cd /home/test/ascend-device-plugin

   kubectl apply -f  ascendplugin.yaml
   ```
4. 执行如下命令，查看节点设备部署信息。
   ```shell
   kubectl describe node
   ```
  
	如下所示，字段中对应标签及节点数量正确说明部署成功。
   ```
   Capacity:
     cpu:                   128
     ephemeral-storage:     3842380928Ki
     huawei.com/Ascend910:  8
     hugepages-2Mi:         0
     memory:                263865068Ki
     pods:                  110
   Allocatable:
     cpu:                   128
     ephemeral-storage:     3541138257382
     huawei.com/Ascend910:  8
     hugepages-2Mi:         0
     memory:                263762668Ki
     pods:                  110
   ```
   ## 4 使用yaml创建带Ascend芯片的任务容器
   ### 4.1 编写任务yaml，在resources中指定任务容器需要的芯片类型和个数
   ``` shell
     vi ascend.yaml
   ```
   
   ``` yaml
    apiVersion: v1  #指定api版本，此值必须在kubectl apiversion中
    kind: Pod #指定创建资源的角色/类型
    metadata:
      name: rest502 #资源的名字，在同一个namespace中必须唯一
    spec:
      containers:
      - name: rest502 #容器的名字
        image: ubuntu_arm64_resnet50:18.04 #容器使用的镜像地址
        imagePullPolicy: Never
        resources:
          limits: #资源限制
            huawei.com/Ascend910: 2  #使用芯片类型和个数，如为310芯片，修改为huawei.com/Ascend310
        volumeMounts:
          - name: joblog
            mountPath: /home/log/  #容器内部日志路径，根据任务需要修改。
          - name: model
            mountPath: /home/app/model #容器内部模型路径，根据任务需要修改。
      volumes:
        - name: joblog
          hostPath:
            path: /home/test/docker_log    #宿主机挂载日志路径，根据任务需要修改。
        - name: model
          hostPath:
            path: /home/test/docker_model/  #宿主机挂载模型路径，根据任务需要修改。
   ```
   ### 4.2 执行命令启动Pod
   ```shell
   kubectl apply -f ascend.yaml
   ```
   ### 4.3 分别执行以下命令，进入pod查看分配信息。
   ```shell
   kubectl exec -it pod名称 bash
   ```
   pod名称为yaml中的资源名称。
   ```shell
   ls /dev/
   ```
   如下类似回显信息中可以看到davinci3和davinci4即为分配的pod。
   ```
   core davinci3 davinci4 davinci_manager devmm_svm fd full hisi_hdc   mqueue null ptmx
   ```