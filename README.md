# Ascend Device Plugin.zh
-   [Ascend Device Plugin](#Ascend-Device-Plugin.md)
    -   [组件介绍](#组件介绍.md)
    -   [环境依赖](#环境依赖.md)
    -   [编译Ascend Device Plugin](#编译Ascend-Device-Plugin.md)
    -   [创建DaemonSet](#创建DaemonSet.md)
    -   [创建业务容器](#创建业务容器.md)
-   [目录结构](#目录结构.md)
-   [版本更新记录](#版本更新记录.md)
<h2 id="Ascend-Device-Plugin.md">Ascend Device Plugin</h2>

-   **[组件介绍](#组件介绍.md)**  

-   **[环境依赖](#环境依赖.md)**  

-   **[编译Ascend Device Plugin](#编译Ascend-Device-Plugin.md)**  

-   **[创建DaemonSet](#创建DaemonSet.md)**  

-   **[创建业务容器](#创建业务容器.md)**  


<h2 id="组件介绍.md">组件介绍</h2>

设备管理插件拥有以下功能：

-   设备发现：支持从昇腾设备驱动中发现设备个数，将其发现的设备个数上报到Kubernetes系统中。支持发现拆分物理设备得到的虚拟设备，虚拟设备需要提前完成拆分。
-   健康检查：支持检测昇腾设备的健康状态，当设备处于不健康状态时，上报到Kubernetes系统中，将不健康的昇腾设备从Kubernetes系统中剔除。虚拟设备健康状态由拆分这些虚拟设备的物理设备决定。
-   设备分配：支持在Kubernetes系统中分配昇腾设备；支持NPU设备重调度功能，设备故障后会自动拉起新容器，挂载健康设备，并重建训练任务。

<h2 id="环境依赖.md">环境依赖</h2>

**表 1**  环境依赖

<a name="zh-cn_topic_0252788324_table171211952105718"></a>
<table><thead align="left"><tr id="zh-cn_topic_0269670261_zh-cn_topic_0252788324_row51223524573"><th class="cellrowborder" valign="top" width="48%" id="mcps1.2.3.1.1"><p id="zh-cn_topic_0269670261_zh-cn_topic_0252788324_p15122175218576"><a name="zh-cn_topic_0269670261_zh-cn_topic_0252788324_p15122175218576"></a><a name="zh-cn_topic_0269670261_zh-cn_topic_0252788324_p15122175218576"></a>检查项</p>
</th>
<th class="cellrowborder" valign="top" width="52%" id="mcps1.2.3.1.2"><p id="zh-cn_topic_0269670261_zh-cn_topic_0252788324_p1712211526578"><a name="zh-cn_topic_0269670261_zh-cn_topic_0252788324_p1712211526578"></a><a name="zh-cn_topic_0269670261_zh-cn_topic_0252788324_p1712211526578"></a>要求</p>
</th>
</tr>
</thead>
<tbody><tr id="zh-cn_topic_0269670261_row1985835314489"><td class="cellrowborder" valign="top" width="48%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0269670261_p1925915619412"><a name="zh-cn_topic_0269670261_p1925915619412"></a><a name="zh-cn_topic_0269670261_p1925915619412"></a>dos2unix</p>
</td>
<td class="cellrowborder" valign="top" width="52%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0269670261_p1025985634111"><a name="zh-cn_topic_0269670261_p1025985634111"></a><a name="zh-cn_topic_0269670261_p1025985634111"></a>已安装（无版本要求），执行<strong id="zh-cn_topic_0269670261_b026181053915"><a name="zh-cn_topic_0269670261_b026181053915"></a><a name="zh-cn_topic_0269670261_b026181053915"></a>dos2unix --version</strong>命令查看。</p>
</td>
</tr>
<tr id="zh-cn_topic_0269670261_row16906451114817"><td class="cellrowborder" valign="top" width="48%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0269670261_p212295212575"><a name="zh-cn_topic_0269670261_p212295212575"></a><a name="zh-cn_topic_0269670261_p212295212575"></a>run包的驱动版本</p>
</td>
<td class="cellrowborder" valign="top" width="52%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0269670261_p31997012111"><a name="zh-cn_topic_0269670261_p31997012111"></a><a name="zh-cn_topic_0269670261_p31997012111"></a>大于等于1.73，进入驱动所在路径（如<span class="filepath" id="zh-cn_topic_0269670261_filepath15286102081119"><a name="zh-cn_topic_0269670261_filepath15286102081119"></a><a name="zh-cn_topic_0269670261_filepath15286102081119"></a>“/usr/local/Ascend/driver”</span>），执行<strong id="zh-cn_topic_0269670261_b133711055171113"><a name="zh-cn_topic_0269670261_b133711055171113"></a><a name="zh-cn_topic_0269670261_b133711055171113"></a>cat version.info</strong>命令查看。</p>
</td>
</tr>
<tr id="zh-cn_topic_0269670261_row12226135012483"><td class="cellrowborder" valign="top" width="48%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0269670261_p3124195265717"><a name="zh-cn_topic_0269670261_p3124195265717"></a><a name="zh-cn_topic_0269670261_p3124195265717"></a>Go语言环境版本</p>
</td>
<td class="cellrowborder" valign="top" width="52%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0269670261_p012435218578"><a name="zh-cn_topic_0269670261_p012435218578"></a><a name="zh-cn_topic_0269670261_p012435218578"></a>大于等于1.14.3，执行<strong id="zh-cn_topic_0269670261_b15724113573315"><a name="zh-cn_topic_0269670261_b15724113573315"></a><a name="zh-cn_topic_0269670261_b15724113573315"></a>go version</strong>命令查看。</p>
</td>
</tr>
<tr id="zh-cn_topic_0269670261_row05615595485"><td class="cellrowborder" valign="top" width="48%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0269670261_p2124252115719"><a name="zh-cn_topic_0269670261_p2124252115719"></a><a name="zh-cn_topic_0269670261_p2124252115719"></a>gcc版本</p>
</td>
<td class="cellrowborder" valign="top" width="52%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0269670261_p512445215576"><a name="zh-cn_topic_0269670261_p512445215576"></a><a name="zh-cn_topic_0269670261_p512445215576"></a>大于等于7.3.0，执行<strong id="zh-cn_topic_0269670261_b1019441317397"><a name="zh-cn_topic_0269670261_b1019441317397"></a><a name="zh-cn_topic_0269670261_b1019441317397"></a>gcc --version</strong>命令查看。</p>
</td>
</tr>
<tr id="zh-cn_topic_0269670261_row11826547124816"><td class="cellrowborder" valign="top" width="48%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0269670261_p151241522577"><a name="zh-cn_topic_0269670261_p151241522577"></a><a name="zh-cn_topic_0269670261_p151241522577"></a>Kubernetes版本</p>
</td>
<td class="cellrowborder" valign="top" width="52%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0269670261_p89141115124714"><a name="zh-cn_topic_0269670261_p89141115124714"></a><a name="zh-cn_topic_0269670261_p89141115124714"></a>1.17.x，建议选择最新的bugfix版本。</p>
<p id="zh-cn_topic_0269670261_p1124115285720"><a name="zh-cn_topic_0269670261_p1124115285720"></a><a name="zh-cn_topic_0269670261_p1124115285720"></a>执行<strong id="zh-cn_topic_0269670261_b11575194924412"><a name="zh-cn_topic_0269670261_b11575194924412"></a><a name="zh-cn_topic_0269670261_b11575194924412"></a>kubectl version</strong>命令查看。</p>
</td>
</tr>
<tr id="zh-cn_topic_0269670261_zh-cn_topic_0252788324_row11244529577"><td class="cellrowborder" valign="top" width="48%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0269670261_zh-cn_topic_0252788324_p16191917113619"><a name="zh-cn_topic_0269670261_zh-cn_topic_0252788324_p16191917113619"></a><a name="zh-cn_topic_0269670261_zh-cn_topic_0252788324_p16191917113619"></a>Docker环境</p>
</td>
<td class="cellrowborder" valign="top" width="52%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0269670261_zh-cn_topic_0252788324_p461711733616"><a name="zh-cn_topic_0269670261_zh-cn_topic_0252788324_p461711733616"></a><a name="zh-cn_topic_0269670261_zh-cn_topic_0252788324_p461711733616"></a>已安装Docker，执行<strong id="zh-cn_topic_0269670261_b1210311189413"><a name="zh-cn_topic_0269670261_b1210311189413"></a><a name="zh-cn_topic_0269670261_b1210311189413"></a>docker info</strong>命令查看。</p>
</td>
</tr>
<tr id="zh-cn_topic_0269670261_row34271613113113"><td class="cellrowborder" valign="top" width="48%" headers="mcps1.2.3.1.1 "><p id="zh-cn_topic_0269670261_p1942971303117"><a name="zh-cn_topic_0269670261_p1942971303117"></a><a name="zh-cn_topic_0269670261_p1942971303117"></a>root用户</p>
</td>
<td class="cellrowborder" valign="top" width="52%" headers="mcps1.2.3.1.2 "><p id="zh-cn_topic_0269670261_p8429113133117"><a name="zh-cn_topic_0269670261_p8429113133117"></a><a name="zh-cn_topic_0269670261_p8429113133117"></a>裸机拥有root用户权限。</p>
</td>
</tr>
</tbody>
</table>

<h2 id="编译Ascend-Device-Plugin.md">编译Ascend Device Plugin</h2>

## 操作步骤<a name="section1719544174917"></a>

1.  执行以下命令，设置环境变量。

    **export GO111MODULE=on**

    **export GOPROXY=**_代理地址_

    **export GONOSUMDB=\\\***

    >![](doc/figures/icon-note.gif) **说明：** 
    >GOPROXY代理地址请根据实际选择，可通过在“ascend-device-plugin“目录下执行**go mod download**命令进行检查。若无返回错误信息，则表示代理设置成功。

2.  进入“ascend-device-plugin“目录，执行以下命令，可根据用户需要修改yaml文件（可选）。
    -   通用yaml文件

        **ascendplugin-910.yaml**

    -   MindX DL使用的yaml文件

        **vim ascendplugin-volcano.yaml**

        ```
        ......      
              containers: 
              - image: ascend-k8sdeviceplugin:v3.0.0   #镜像名称及版本号。 
                name: device-plugin-01 
                resources: 
                  requests: 
                    memory: 500Mi 
                    cpu: 500m 
                  limits: 
                    memory: 500Mi 
                    cpu: 500m 
                command: [ "/bin/bash", "-c", "--"] 
                args: [ "ascendplugin  -useAscendDocker=true 
                        -volcanoType=true                    # 重调度场景下必须使用volcano 
                        -autoStowing=true                    # 是否开启自动纳管开关，默认为true；设置为false代表关闭自动纳管，当芯片健康状态由unhealth变为health后，不会自动加入到可调度资源池中                  
                        -listWatchPeriod=5                   # 健康状态检查周期，范围[3,60]；默认5秒                                                                             
                        -logFile=/var/log/mindx-dl/devicePlugin/devicePlugin.log  
                        -logLevel=0" ]  
        ......
        ```


3.  设置“useAscendDocker”参数。
    -   如果安装了Ascend-docker-runtime，则设置useAscendDocker=true。默认场景，推荐用户使用。
    -   如果未安装Ascend-docker-runtime，则设置useAscendDocker=false。
    -   开启CPU绑核功能后，无论是否安装Ascend-docker-runtime，都设置useAscendDocker=false。

4.  执行以下命令，进入构建目录，执行构建脚本，在“output“目录下生成二进制device-plugin、yaml文件和Dockerfile。

    **cd **_/home/test/_**ascend-device-plugin/build/**

    **chmod +x build.sh**

    **./build.sh**

5.  执行以下命令，查看生成的软件包。

    **ll **_/home/test/_**ascend-device-plugin/output**

    ```
    drwxr-xr-x  2 root root     4096 Jan 18 17:04 ./
    drwxr-xr-x 12 root root     4096 Jan 18 17:04 ../
    -r-x------  1 root root 36058664 Jan 18 17:04 device-plugin
    -r--------  1 root root     2478 Jan 18 17:04 device-plugin-310P-1usoc-v3.0.0.yaml
    -r--------  1 root root     3756 Jan 18 17:04 device-plugin-310P-1usoc-volcano-v3.0.0.yaml
    -r--------  1 root root     2478 Jan 18 17:04 device-plugin-310P-v3.0.0.yaml
    -r--------  1 root root     3756 Jan 18 17:04 device-plugin-310P-volcano-v3.0.0.yaml
    -r--------  1 root root     2131 Jan 18 17:04 device-plugin-310-v3.0.0.yaml
    -r--------  1 root root     3431 Jan 18 17:04 device-plugin-310-volcano-v3.0.0.yaml
    -r--------  1 root root     2130 Jan 18 17:04 device-plugin-910-v3.0.0.yaml
    -r--------  1 root root     3447 Jan 18 17:04 device-plugin-volcano-v3.0.0.yaml
    -r--------  1 root root      654 Jan 18 17:04 Dockerfile
    -r--------  1 root root     1199 Jan 18 17:04 Dockerfile-310P-1usoc
    -r--------  1 root root     1537 Jan 18 17:04 run_for_310P_1usoc.sh
    ```

    >![](doc/figures/icon-note.gif) **说明：** 
    >“ascend-device-plugin“目录下的**ascendplugin-910.yaml**文件在“ascend-device-plugin/output/“下生成的对应文件为**device-plugin-910-v3.0.0.yaml**，作用是更新版本号。

6.  执行以下命令，查看Dockerfile文件，可根据实际情况修改。

    **vi** _/home/test/_**ascend-device-plugin/Dockerfile**

    ```
    FROM ubuntu:18.04
    
    RUN useradd -d /home/HwHiAiUser -u 1000 -m -s /usr/sbin/nologin HwHiAiUser && \
        usermod root -s /usr/sbin/nologin
    
    ENV USE_ASCEND_DOCKER true
    
    ENV LD_LIBRARY_PATH  /usr/local/Ascend/driver/lib64/driver:/usr/local/Ascend/driver/lib64/common
    
    ENV LD_LIBRARY_PATH $LD_LIBRARY_PATH:/usr/local/Ascend/driver/lib64/:/usr/local/lib
    
    COPY ./device-plugin /usr/local/bin/
    RUN chmod 550 /usr/local/bin/device-plugin &&\
        chmod 550 /usr/local/bin &&\
        chmod 750 /home/HwHiAiUser &&\
        chmod 550 /usr/local/lib/ &&\
        chmod 500 /usr/local/lib/* &&\
        echo 'umask 027' >> /etc/profile &&\
        echo 'source /etc/profile' >> ~/.bashrc
    ```


<h2 id="创建DaemonSet.md">创建DaemonSet</h2>

## 操作步骤<a name="zh-cn_topic_0269670254_section2036324211563"></a>

>![](doc/figures/icon-note.gif) **说明：** 
>以下操作以ARM平台下构建并分发镜像为例。

1.  获取Ascend Device Plugin源码包
2.  进入output目录，执行以下命令构建Ascend Device Plugin的镜像。

    ```
    docker build -t ascend-k8sdeviceplugin:v3.0.0 .
    ```

    当出现“Successfully built xxx”表示镜像构建成功，注意不要遗漏命令结尾的“.”。这里的“.”代表当前目录。

    使用**docker images**命令查看，可以看见名为ascend-k8sdeviceplugin，tag为v3.0.0的镜像。

4.  执行如下命令将编译好的镜像打包并压缩，便于在各个服务器之间传输。

    >![](doc/figures/icon-note.gif) **说明：** 
    >\{arch\}表示系统架构，不同架构之间的镜像不能通用。

    ```
    docker save -o ascend-k8sdeviceplugin:v3.0.0 | gzip > Ascend-K8sDevicePlugin-v3.0.0-{arch}-Docker.tar.gz
    ```

    或者使用不带压缩功能的命令：

    ```
    docker save -o Ascend-K8sDevicePlugin-v3.0.0-{arch}-Docker.tar ascend-k8sdeviceplugin:v3.0.0
    ```

5.  （可选）集群场景下将打包好的镜像（比如存放路径为：“/home/ascend-device-plugin/”）分发到拥有昇腾系列AI处理器的计算节点上。

    ```
    cd /home/ascend-device-plugin
    scp Ascend-K8sDevicePlugin-v3.0.0-{arch}-Docker.tar.gz root@{节点IP地址}:/home/ascend-device-plugin
    ```

6.  执行以下命令加载镜像。以打包压缩后的文件为例，没有选择压缩则需要修改相应的文件名。
    -   ARM架构

        ```
        docker load -i Ascend-K8sDevicePlugin-v3.0.0-arm64-Docker.tar.gz
        ```

    -   x86架构

        ```
        docker load -i Ascend-K8sDevicePlugin-v3.0.0-amd64-Docker.tar.gz
        ```


7.  执行如下命令，给带有Ascend 910（或含有Ascend 310、Ascend 310P）的节点打标签。

    ```
    kubectl label nodes localhost.localdomain accelerator=huawei-Ascend910
    ```

    localhost.localdomain为有Ascend 910（或含有Ascend 310、Ascend 310P）的节点名称，可通过**kubectl get node**命令查看。

    标签名称需要和软件包中yaml文件里的nodeSelector标签名称保持一致。

    针对断点续训功能：用户可以使用以下命令，将健康状态由unhealth恢复为health的芯片重新放入资源池。

    ```
    kubelet label nodes node_name huawei.com/Ascend910-Recover-
    ```

    执行该命令后会删除“**huawei.com/Ascend910-Recover**”标签，该标签中的芯片会重新放入Node Annotation中供程序调度。注意，该命令仅做清除recover标签信息使用，请不要用于添加标签。

8.  进入“ascend-device-plugin/output“执行以下命令，部署DaemonSet。

    -   昇腾310 AI处理器节点。

        ```
        kubectl apply -f device-plugin-310-v3.0.0.yaml
        ```

    -   昇腾910 AI处理器节点，协同Volcano工作。

        ```
        kubectl apply -f device-plugin-volcano-v3.0.0.yaml
        ```

    -   昇腾310P AI处理器节点。

        ```
        kubectl apply -f device-plugin-310P-v3.0.0.yaml
        ```
    
    - 昇腾910 AI处理器节点，Ascend Device Plugin独立工作，不协同Volcano。
    
      ```
      kubectl apply -f device-plugin-910-v3.0.0.yaml
      ```
    
    部署完成后需要等待几分钟，才能看到节点设备部署信息。yaml文件的中参数信息请参考[表1](#table1286935610129)。

9. 执行如下命令，查看节点设备部署信息。

   ```
   kubectl describe node nodeName
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


<h2 id="创建业务容器.md">创建业务容器</h2>

## 操作步骤<a name="zh-cn_topic_0269670251_section28051148174119"></a>

1.  <a name="zh-cn_topic_0269670251_zh-cn_topic_0249483204_li104071617503"></a>进入“ascend-device-plugin“目录，执行如下命令编辑Pod的配置文件，根据文件模板编写配置文件。

    ```
    cd /home/ascend-device-plugin
    vi ascend.yaml
    ```

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
            huawei.com/Ascend310: 2 #根据实际修改资源类型。支持资源类型见下文。
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

    支持的资源类型：

    -   huawei.com/Ascend310: 2，表示分配2颗昇腾310 AI处理器芯片。
    -   huawei.com/Ascend310P: 1，表示分配1颗昇腾310P AI处理器芯片。
    -   huawei.com/Ascend910: 4，表示分配4颗昇腾910 AI处理器芯片。
    -   huawei.com/Ascend910-16c: 1，表示分配1颗算力为16核的虚拟设备，只支持单卡单容器任务（即数值只能填1）。支持调度2c、4c、8c、16c四种AI core数量的虚拟设备。

2.  执行如下命令，创建Pod。

    ```
    kubectl apply -f ascend.yaml
    ```

    ![](doc/figures/icon-note.gif) **说明：** 如需删除请执行以下命令：
    
        kubectl delete -f ascend.yaml

3.  分别执行以下命令，进入Pod查看分配信息。

    ```
    kubectl exec -it podName bash
    ```

    Pod名称为[1](#zh-cn_topic_0269670251_zh-cn_topic_0249483204_li104071617503)中配置的Pod名称。

    ```
    ls /dev/
    ```

    如下类似回显信息中可以看到davinci3和davinci4即为分配的Pod。

    ```
    davinci3 davinci4 davinci_manager devmm_svm fd full hisi_hdc mqueue null ptmx
    ```


<h2 id="目录结构.md">目录结构</h2>

```
├── build                                             # 构建目录   
│   ├── ascend.yaml                                   # sample运行任务yaml
│   ├── ascendplugin-310.yaml                         # 310推理卡部署yaml             
│   ├── ascendplugin-310-volcano.yaml                 # 310搭配volcano实现亲和性调度部署yaml
│   ├── ascendplugin-310P.yaml                        # 310P推理卡部署yaml
│   ├── ascendplugin-310P-volcano.yaml                # 310P搭配volcano实现亲和性调度部署yaml
│   ├── ascendplugin-volcano.yaml                     # 910搭配volcano实现亲和性调度部署yaml
│   ├── ascendplugin-910.yaml                         # 910未使用volcano部署yaml
│   ├── ascendplugin-310P-1usoc.yaml                  # 1usoc专用yaml
│   ├── ascendplugin-310P-1usoc-volcano.yaml          
│   ├── Dockerfile                                    # 镜像文件
│   ├── Dockerfile-310P-1usoc                         # 1usoc专用镜像文件
│   ├── test.sh                                       # UT shell
│   ├── run_for_310P_1usoc.sh                         # 1usoc专用脚本
│   └── build.sh
├── doc                                       
│   └── figures
│       ├── icon-caution.gif
│       ├── icon-danger.gif
│       ├── icon-note.gif
│       ├── icon-notice.gif
│       ├── icon-tip.gif
│       └── icon-warning.gif
├── pkg                                              # 源代码目录
│   └── common
│        ├── atomic_bool.go
│        ├── atomic_bool_test.go
│        ├── common.go
│        ├── common_test.go
│        ├── constants.go
│        ├── device.go
│        ├── device_test.go
│        └── proto.go
│   └── device
│        ├── ascend310.go
│        ├── ascend310_test.go
│        ├── ascend310p.go
│        ├── ascend310p_test.go
│        ├── ascend910.go
│        ├── ascend910_test.go
│        ├── ascendcommon.go
│        └── ascendcommon_test.go
│   └── kubeclient
│        ├── client_server.go
│        ├── client_server_test.go
│        └── kubeclient.go
│   └── server
│        ├── manager.go
│        ├── manager_test.go
│        ├── plugin.go
│        ├── plugin_test.go
│        ├── pod_resource.go
│        ├── pod_resource_test.go
│        ├── server.go
│        ├── server_test.go
│        └── type.go
├── go.mod                                           
│    └── go.sum 
├── LICENSE
├── main.go
└── README.md                                        
```

<h2 id="版本更新记录.md">版本更新记录</h2>

<a name="table7854542104414"></a>
<table><thead align="left"><tr id="zh-cn_topic_0280467800_row785512423445"><th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.1.4.1.1"><p id="zh-cn_topic_0280467800_p19856144274419"><a name="zh-cn_topic_0280467800_p19856144274419"></a><a name="zh-cn_topic_0280467800_p19856144274419"></a>版本</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.1.4.1.2"><p id="zh-cn_topic_0280467800_p3856134219446"><a name="zh-cn_topic_0280467800_p3856134219446"></a><a name="zh-cn_topic_0280467800_p3856134219446"></a>发布日期</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.1.4.1.3"><p id="zh-cn_topic_0280467800_p585634218445"><a name="zh-cn_topic_0280467800_p585634218445"></a><a name="zh-cn_topic_0280467800_p585634218445"></a>修改说明</p>
</th>
</tr>
</thead>
<tbody><tr id="row7293189122012"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.1 "><p id="p9235101416201"><a name="p9235101416201"></a><a name="p9235101416201"></a>v3.0.0</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.2 "><p id="p1523518145208"><a name="p1523518145208"></a><a name="p1523518145208"></a>2022-1230</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.3 "><a name="ul162558202525"></a><a name="ul162558202525"></a><ul id="ul162558202525"><li>首次发布</li></ul>
</td>
</tr>
</tbody>
</table>

