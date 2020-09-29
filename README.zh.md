# Ascend Device Plugin
-   [Ascend Device Plugin](#Ascend-Device-Plugin.md)
    -   [组件介绍](#组件介绍.md)
    -   [编译Ascend Device Plugin](#编译Ascend-Device-Plugin.md)
        -   [快速编译Ascend Device Plugin](#快速编译Ascend-Device-Plugin.md)
        -   [编译Ascend Device Plugin](#编译Ascend-Device-Plugin-0.md)
    -   [创建DaemonSet](#创建DaemonSet.md)
    -   [创建业务容器](#创建业务容器.md)
-   [环境依赖](#环境依赖.md)
-   [目录结构](#目录结构.md)
-   [版本更新记录](#版本更新记录.md)
<h2 id="Ascend-Device-Plugin.md">Ascend Device Plugin</h2>

-   **[组件介绍](#组件介绍.md)**  

-   **[编译Ascend Device Plugin](#编译Ascend-Device-Plugin.md)**  

-   **[创建DaemonSet](#创建DaemonSet.md)**  

-   **[创建业务容器](#创建业务容器.md)**  


<h2 id="组件介绍.md">组件介绍</h2>

设备管理插件拥有以下功能：

-   设备发现：支持从昇腾设备驱动中发现设备个数，将其发现的设备个数上报到Kubernetes系统中。
-   健康检查：支持检测昇腾设备的健康状态，当设备处于不健康状态时，上报到Kubernetes系统中，将不健康的昇腾设备从Kubernetes系统中剔除。
-   设备分配：支持在Kubernetes系统中分配昇腾设备。

<h2 id="编译Ascend-Device-Plugin.md">编译Ascend Device Plugin</h2>

-   **[快速编译Ascend Device Plugin](#快速编译Ascend-Device-Plugin.md)**  
将修改编译过程中的配置参数通过执行一个shell脚本来完成，用户只需要修改脚本中的参数，就能快速完成编译。
-   **[编译Ascend Device Plugin](#编译Ascend-Device-Plugin-0.md)**  


<h2 id="快速编译Ascend-Device-Plugin.md">快速编译Ascend Device Plugin</h2>

将修改编译过程中的配置参数通过执行一个shell脚本来完成，用户只需要修改脚本中的参数，就能快速完成编译。

## 操作步骤<a name="section125457120293"></a>

1.  执行以下命令安装最新版本的pkg-config。

    **apt-get install -y pkg-config**

2.  执行以下命令，设置环境变量。

    **export GO111MODULE=on**

    **export GOPROXY=**_代理地址_

    **export GONOSUMDB=\***

    >![](figures/icon-note.gif) **说明：** 
    >GOPROXY代理地址请根据实际选择，可通过在ascend-device-plugin目录下执行**go mod download**命令进行检查。

3.  在“./build/”中创建并执行shell文件。

    ```
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

4.  执行以下命令，根据实际选择执行的脚本，生成二进制和镜像文件。

    Ascend 910请选择build\_910.sh，Ascend 310请选择build\_310.sh。

    **cd** _/home/test/_ascend-device-plugin**/build**/

    **chmod +x** _build\_XXX.sh_

    **dos2unix** _build\_XXX.sh_

    **./**_build\_XXX.sh_ **dockerimages**

5.  执行以下命令，查看生成的软件包。

    **ll** _/home/test/_ascend-device-plugin**/output**

    x86和ARM生成的软件包名不同，以下示例为ARM环境：

    >![](figures/icon-note.gif) **说明：** 
    >-   **Ascend-K8sDevicePlugin-**_xxx_**-arm64-Docker.tar.gz**：K8s设备插件镜像。
    >-   **Ascend-K8sDevicePlugin-**_xxx_**-arm64-Linux.tar.gz**：K8s设备插件二进制安装包。

    ```
    drwxr-xr-x 2 root root     4096 Jun  8 18:42 ./
    drwxr-xr-x 9 root root     4096 Jun  8 17:12 ../
    -rw-r--r-- 1 root root 29584705 Jun  9 10:37 Ascend-K8sDevicePlugin-xxx-arm64-Docker.tar.gz
    -rwxr-xr-x 1 root root  6721073 Jun  9 16:20 Ascend-K8sDevicePlugin-xxx-arm64-Linux.tar.gz
    ```


<h2 id="编译Ascend-Device-Plugin-0.md">编译Ascend Device Plugin</h2>

## 操作步骤<a name="section112101632152317"></a>

1.  执行以下命令安装最新版本的pkg-config。

    **apt-get install -y pkg-config**

2.  执行以下命令，设置环境变量。

    **export GO111MODULE=on**

    **export GOPROXY=**_代理地址_

    **export GONOSUMDB=\***

    >![](figures/icon-note.gif) **说明：** 
    >GOPROXY代理地址请根据实际选择，可通过在ascend-device-plugin目录下执行**go mod download**命令进行检查。

3.  进入ascend-device-plugin目录，执行以下命令，修改yaml文件。
    -   通用yaml文件。

        **vi ascendplugin.yaml**

        ```
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
              - image: ascend-device-plugin:v1.0.1  #镜像名称及版本号，需要和build_common.sh中保持一致。
                name: device-plugin-01
                resources:
                  requests:
                    memory: 500Mi
                    cpu: 500m
                  limits:
                    memory: 500Mi
                    cpu: 500m
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

    -   Atlas深度学习组件使用yaml文件。

        **ascendplugin-volcano.yaml**

        ```
        kind: ClusterRoleBinding
        apiVersion: rbac.authorization.k8s.io/v1
        metadata:
          name: pods-device-plugin
        subjects:
          - kind: ServiceAccount
            name: default
            namespace: kube-system
        roleRef:
          kind: ClusterRole
          name: cluster-admin
          apiGroup: rbac.authorization.k8s.io
        ---
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
                - key: huawei.com/Ascend910
                  operator: Exists
                  effect: NoSchedule
                - key: "ascendplugin"
                  operator: "Equal"
                  value: "v2"
                  effect: NoSchedule
              priorityClassName: "system-node-critical"
              nodeSelector:
                accelerator: huawei-Ascend910
              containers:
              - image: ascend-k8sdeviceplugin:V20.1.0   #镜像名称及版本号，需要和build_common.sh中保持一致。
                name: device-plugin-01
                resources:
                  requests:
                    memory: 500Mi
                    cpu: 500m
                  limits:
                    memory: 500Mi
                    cpu: 500m
                command: [ "/bin/bash", "-c", "--"]
                args: [ "./build/build_in_docker.sh;ascendplugin  --useAscendDocker=${USE_ASCEND_DOCKER} --volcanoType=true" ] #使用Ascend310，则需要增加--mode=ascend310
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
                env:
                  - name: NODE_NAME
                    valueFrom:
                      fieldRef:
                        fieldPath: spec.nodeName
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

4.  执行以下命令，编辑Dockerfile文件，将镜像修改为查询的镜像名及版本号。

    **vi **_/home/test/_ascend-device-plugin**/Dockerfile**

    ```
    #用户根据实际选择需要使用的带Go编译的基础镜像，可通过docker images命令查询。
    FROM golang:1.13.11-buster as build
    
    #是否使用昇腾Docker，默认为true，请修改为false。
    ENV USE_ASCEND_DOCKER false
    
    ENV GOPATH /usr/app/
    
    ENV GO111MODULE off
    
    ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
    #libdrvdsmi_host.so所在目录，Ascend 310和Ascend 910目录不同。
    ENV LD_LIBRARY_PATH  /usr/local/Ascend/driver/lib64/driver:/usr/local/Ascend/driver/lib64/common
    
    RUN mkdir -p /usr/app/src/ascend-device-plugin
    
    COPY . /usr/app/src/Ascend-device-plugin
    
    WORKDIR /usr/app/src/Ascend-device-plugin
    ```

5.  进入ascend\_device\_plugin.pc文件所在目录，执行以下命令，查看以下路径是否正确，根据实际修改。

    -   Ascend 310目录：ascend-device-plugin/src/plugin/config/config\_310
    -   Ascend 910目录：ascend-device-plugin/src/plugin/config/config\_910

    **vi ascend\_device\_plugin.pc**

    ```
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

    >![](figures/icon-note.gif) **说明：** 
    >支持修改插件镜像的名称，build目录下build\_common.sh中修改“docker\_images\_name”即可，需要和ascendplugin.yaml中保持一致。

6.  执行以下命令，根据实际选择执行的脚本，生成二进制和镜像文件。

    Ascend 910请选择build\_910.sh，Ascend 310请选择build\_310.sh。

    **cd** _/home/test/_ascend-device-plugin**/build**/

    **chmod +x** _build\_XXX.sh_

    **dos2unix** _build\_XXX.sh_

    **./**_build\_XXX.sh_ **dockerimages**


1.  执行以下命令，查看生成的软件包。

    **ll** _/home/test/_ascend-device-plugin**/output**

    x86和ARM生成的软件包名不同，以下示例为ARM环境：

    >![](figures/icon-note.gif) **说明：** 
    >-   **Ascend-K8sDevicePlugin-**_xxx_**-arm64-Docker.tar.gz**：K8s设备插件镜像。
    >-   **Ascend-K8sDevicePlugin-**_xxx_**-arm64-Linux.tar.gz**：K8s设备插件二进制安装包。

    ```
    drwxr-xr-x 2 root root     4096 Jun  8 18:42 ./
    drwxr-xr-x 9 root root     4096 Jun  8 17:12 ../
    -rw-r--r-- 1 root root 29584705 Jun  9 10:37 Ascend-K8sDevicePlugin-xxx-arm64-Docker.tar.gz
    -rwxr-xr-x 1 root root  6721073 Jun  9 16:20 Ascend-K8sDevicePlugin-xxx-arm64-Linux.tar.gz
    ```


<h2 id="创建DaemonSet.md">创建DaemonSet</h2>

## 操作步骤<a name="zh-cn_topic_0269670254_section2036324211563"></a>

>![](figures/icon-note.gif) **说明：** 
>以下操作以ARM平台下生成的tar.gz文件为例。

1.  执行以下命令，查看Docker软件包是否导入成功。

    **docker images**

    -   是，请执行[3](#zh-cn_topic_0269670254_li26268471380)。
    -   否，请执行[2](#zh-cn_topic_0269670254_li1372334715567)重新导入。

2.  <a name="zh-cn_topic_0269670254_li1372334715567"></a>进入生成的Docker软件包所在目录，执行以下命令，导入Docker镜像。

    **cd** _/home/test/_**ascend-device-plugin/output**

    **docker load < **_Ascend-K8sDevicePlugin-xxx-arm64-Docker.tar.gz_

3.  <a name="zh-cn_topic_0269670254_li26268471380"></a>执行如下命令，给带有Ascend 910（或Ascend 310）的节点打标签。

    **kubectl label nodes **_localhost.localdomain_** accelerator=**_huawei-Ascend910_

    localhost.localdomain为有Ascend 910（或Ascend 310）的节点名称，可通过**kubectl get node**命令查看。

    标签名称需要和[3](#编译Ascend-Device-Plugin-0.md#zh-cn_topic_0252775101_li8538035183714)中的nodeSelector标签名称保持一致。

    >![](figures/icon-note.gif) **说明：** 
    >如需扩容集群节点，请参考[2](#zh-cn_topic_0269670254_li1372334715567)\~[3](#zh-cn_topic_0269670254_li26268471380)操作将新节点加入集群。

4.  执行以下命令，部署DaemonSet。

    **cd** _/home/test/_**ascend-device-plugin**

    **kubectl apply -f  ascendplugin.yaml**

    >![](figures/icon-note.gif) **说明：** 
    >部署完成后需要等待几分钟，才能看到节点设备部署信息。


1.  执行如下命令，查看节点设备部署信息。

    **kubectl describe node**

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


<h2 id="创建业务容器.md">创建业务容器</h2>

## 操作步骤<a name="zh-cn_topic_0269670251_section28051148174119"></a>

1.  <a name="zh-cn_topic_0269670251_zh-cn_topic_0249483204_li104071617503"></a>进入ascend-device-plugin目录，执行如下命令编辑Pod的配置文件，根据文件模板编写配置文件。

    **cd**_ /home/test/_**ascend-device-plugin**

    **vi ascend.yaml**

    ```
    apiVersion: v1  #指定API版本，此值必须在kubectl apiversion中
    kind: Pod #指定创建资源的角色/类型
    metadata:
      name: rest502 #Pod名称，在同一个namespace中必须唯一。
    spec:  #Pod中容器的详细定义。
      containers: #Pod中容器列表。
      - name: rest502 #Pod中容器名称。
        image: centos_arm64_resnet50:7.8 #Pod中容器使用的推理或训练业务镜像地址。
        imagePullPolicy: Never
        resources:
          limits: #资源限制
            huawei.com/Ascend310: 2 #根据实际修改资源类型。
        volumeMounts:
          - name: joblog
            mountPath: /home/log/  #容器内部日志路径，根据任务需要修改。
          - name: model
            mountPath: /home/app/model #容器内部模型路径，根据任务需要修改。
          - name: slog-path
            mountPath: /var/log/npu/conf/slog/slog.conf
          - name: ascend-driver-path
            mountPath: /usr/local/Ascend/driver #根据Driver实际所在路径修改。
      volumes:
        - name: joblog
          hostPath:
            path: /home/test/docker_log    #宿主机挂载日志路径，根据任务需要修改。
        - name: model
          hostPath:
            path: /home/test/docker_model/  #宿主机挂载模型路径，根据任务需要修改。
        - name: slog-path
          hostPath:
            path: /var/log/npu/conf/slog/slog.conf  
        - name: ascend-driver-path
          hostPath:
            path: /usr/local/Ascend/driver #根据Driver实际所在路径修改。
    ```


1.  执行如下命令，创建Pod。

    **kubectl apply -f ascend.yaml**

    >![](figures/icon-note.gif) **说明：** 
    >如需删除请执行以下命令：
    >**kubectl delete -f** **ascend.yaml**


1.  分别执行以下命令，进入Pod查看分配信息。

    **kubectl exec -it **_pod名称_** bash**

    Pod名称为[1](#zh-cn_topic_0269670251_zh-cn_topic_0249483204_li104071617503)中配置的Pod名称。

    **ls /dev/**

    如下类似回显信息中可以看到davinci3和davinci4即为分配的Pod。

    ```
    core davinci3 davinci4 davinci_manager devmm_svm fd full hisi_hdc mqueue null ptmx
    ```


<h2 id="环境依赖.md">环境依赖</h2>

**表 1**  环境依赖

<a name="zh-cn_topic_0252788324_table171211952105718"></a>
<table><thead align="left"><tr id="zh-cn_topic_0269670261_zh-cn_topic_0252788324_row51223524573"><th class="cellrowborder" valign="top" width="30%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0269670261_zh-cn_topic_0252788324_p15122175218576"><a name="zh-cn_topic_0269670261_zh-cn_topic_0252788324_p15122175218576"></a><a name="zh-cn_topic_0269670261_zh-cn_topic_0252788324_p15122175218576"></a>检查项</p>
</th>
<th class="cellrowborder" valign="top" width="70%" id="mcps1.2.3.1.2"><p id="zh-cn_topic_0269670261_zh-cn_topic_0252788324_p1712211526578"><a name="zh-cn_topic_0269670261_zh-cn_topic_0252788324_p1712211526578"></a><a name="zh-cn_topic_0269670261_zh-cn_topic_0252788324_p1712211526578"></a>要求</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0269670261_row1985835314489"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0269670261_p1925915619412"><a name="zh-cn_topic_0269670261_p1925915619412"></a><a name="zh-cn_topic_0269670261_p1925915619412"></a>dos2unix</p>
</td>
<td class="cellrowborder" valign="top" width="70%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0269670261_p1025985634111"><a name="zh-cn_topic_0269670261_p1025985634111"></a><a name="zh-cn_topic_0269670261_p1025985634111"></a>已安装（无版本要求），执行<strong id="zh-cn_topic_0269670261_b026181053915"><a name="zh-cn_topic_0269670261_b026181053915"></a><a name="zh-cn_topic_0269670261_b026181053915"></a>dos2unix --version</strong>命令查看。</p>
</td>
</tr>
<tr id="zh-cn_topic_0269670261_row16906451114817"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0269670261_p212295212575"><a name="zh-cn_topic_0269670261_p212295212575"></a><a name="zh-cn_topic_0269670261_p212295212575"></a>run包的驱动版本</p>
</td>
<td class="cellrowborder" valign="top" width="70%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0269670261_p31997012111"><a name="zh-cn_topic_0269670261_p31997012111"></a><a name="zh-cn_topic_0269670261_p31997012111"></a>大于等于1.73，进入驱动所在路径（如<span class="filepath" id="zh-cn_topic_0269670261_filepath15286102081119"><a name="zh-cn_topic_0269670261_filepath15286102081119"></a><a name="zh-cn_topic_0269670261_filepath15286102081119"></a>“/usr/local/Ascend/driver”</span>），执行<strong id="zh-cn_topic_0269670261_b133711055171113"><a name="zh-cn_topic_0269670261_b133711055171113"></a><a name="zh-cn_topic_0269670261_b133711055171113"></a>cat version.info</strong>命令查看。</p>
</td>
</tr>
<tr id="zh-cn_topic_0269670261_row12226135012483"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0269670261_p3124195265717"><a name="zh-cn_topic_0269670261_p3124195265717"></a><a name="zh-cn_topic_0269670261_p3124195265717"></a>Go语言环境版本</p>
</td>
<td class="cellrowborder" valign="top" width="70%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0269670261_p012435218578"><a name="zh-cn_topic_0269670261_p012435218578"></a><a name="zh-cn_topic_0269670261_p012435218578"></a>大于等于1.14.3，执行<strong id="zh-cn_topic_0269670261_b15724113573315"><a name="zh-cn_topic_0269670261_b15724113573315"></a><a name="zh-cn_topic_0269670261_b15724113573315"></a>go version</strong>命令查看。</p>
</td>
</tr>
<tr id="zh-cn_topic_0269670261_row05615595485"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0269670261_p2124252115719"><a name="zh-cn_topic_0269670261_p2124252115719"></a><a name="zh-cn_topic_0269670261_p2124252115719"></a>gcc版本</p>
</td>
<td class="cellrowborder" valign="top" width="70%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0269670261_p512445215576"><a name="zh-cn_topic_0269670261_p512445215576"></a><a name="zh-cn_topic_0269670261_p512445215576"></a>大于等于7.3.0，执行<strong id="zh-cn_topic_0269670261_b1019441317397"><a name="zh-cn_topic_0269670261_b1019441317397"></a><a name="zh-cn_topic_0269670261_b1019441317397"></a>gcc --version</strong>命令查看。</p>
</td>
</tr>
<tr id="zh-cn_topic_0269670261_row11826547124816"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0269670261_p151241522577"><a name="zh-cn_topic_0269670261_p151241522577"></a><a name="zh-cn_topic_0269670261_p151241522577"></a>Kubernetes版本</p>
</td>
<td class="cellrowborder" valign="top" width="70%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0269670261_p1124115285720"><a name="zh-cn_topic_0269670261_p1124115285720"></a><a name="zh-cn_topic_0269670261_p1124115285720"></a>大于等于1.13.0，执行<strong id="zh-cn_topic_0269670261_b11575194924412"><a name="zh-cn_topic_0269670261_b11575194924412"></a><a name="zh-cn_topic_0269670261_b11575194924412"></a>kubectl version</strong>命令查看。</p>
</td>
</tr>
<tr id="zh-cn_topic_0269670261_zh-cn_topic_0252788324_row11244529577"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0269670261_zh-cn_topic_0252788324_p16191917113619"><a name="zh-cn_topic_0269670261_zh-cn_topic_0252788324_p16191917113619"></a><a name="zh-cn_topic_0269670261_zh-cn_topic_0252788324_p16191917113619"></a>Docker环境</p>
</td>
<td class="cellrowborder" valign="top" width="70%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0269670261_zh-cn_topic_0252788324_p461711733616"><a name="zh-cn_topic_0269670261_zh-cn_topic_0252788324_p461711733616"></a><a name="zh-cn_topic_0269670261_zh-cn_topic_0252788324_p461711733616"></a>已安装Docker，执行<strong id="zh-cn_topic_0269670261_b1210311189413"><a name="zh-cn_topic_0269670261_b1210311189413"></a><a name="zh-cn_topic_0269670261_b1210311189413"></a>docker info</strong>命令查看。</p>
</td>
</tr>
<tr id="zh-cn_topic_0269670261_row34271613113113"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0269670261_p1942971303117"><a name="zh-cn_topic_0269670261_p1942971303117"></a><a name="zh-cn_topic_0269670261_p1942971303117"></a>root用户</p>
</td>
<td class="cellrowborder" valign="top" width="70%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0269670261_p8429113133117"><a name="zh-cn_topic_0269670261_p8429113133117"></a><a name="zh-cn_topic_0269670261_p8429113133117"></a>裸机拥有root用户权限。</p>
</td>
</tr>
</tbody>
</table>

<h2 id="目录结构.md">目录结构</h2>

```
├── build                                             # 编译脚本
│   ├── build_310.sh
│   ├── build_910.sh
│   ├── build_common.sh
│   ├── build_in_docker.sh
│   ├── build.sh
│   ├── deploy.sh
│   └── sample_check.sh
├── output                                           # 编译结果目录
├── src                                              # 源代码目录
│   └── plugin
│   │    ├── cmd/ascendplugin
│   │    │   └── ascend_plugin.go
│   │    ├── config
│   │    │   ├── config_310
│   │    │   │   └── ascend_device_plugin.pc
│   │    │   └── config_910
│   │    │       └── ascend_device_plugin.pc
│   │    └── pkg/npu/huawei
├── test                                             # 测试目录
├── Dockerfile                                       # 镜像文件
├── LICENSE                                          
├── Open Source Software Notice.md                   
├── README.zh.md
├── ascend.yaml                                      # sample运行任务yaml
├── ascendplugin-310.yaml                            # 推理卡部署yaml
├── ascendplugin-volcano.yaml                        # 和volcano实现亲和性调度部署yaml
├── ascendplugin.yaml                                # 推理卡部署yaml
├── docker_run.sh                                    # docker运行命令
├── go.mod                                           
└── go.sum                                           
```

<h2 id="版本更新记录.md">版本更新记录</h2>

<a name="table7854542104414"></a>
<table><thead align="left"><tr id="row785512423445"><th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.1.4.1.1"><p id="p19856144274419"><a name="p19856144274419"></a><a name="p19856144274419"></a>版本</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.1.4.1.2"><p id="p3856134219446"><a name="p3856134219446"></a><a name="p3856134219446"></a>发布日期</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.1.4.1.3"><p id="p585634218445"><a name="p585634218445"></a><a name="p585634218445"></a>修改说明</p>
</th>
</tr>
</thead>
<tbody><tr id="row118567425441"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.1 "><p id="p08571442174415"><a name="p08571442174415"></a><a name="p08571442174415"></a>V20.1.0</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.2 "><p id="p38571542154414"><a name="p38571542154414"></a><a name="p38571542154414"></a>2020-09-30</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.3 "><p id="p5857142154415"><a name="p5857142154415"></a><a name="p5857142154415"></a>第一次正式发布。</p>
</td>
</tr>
</tbody>
</table>

