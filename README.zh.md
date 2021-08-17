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
-   设备分配：支持在Kubernetes系统中分配昇腾设备；

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

    >![](figures/icon-note.gif) **说明：** 
    >GOPROXY代理地址请根据实际选择，可通过在“ascend-device-plugin“目录下执行**go mod download**命令进行检查。若无返回错误信息，则表示代理设置成功。

2.  进入“ascend-device-plugin“目录，执行以下命令，可根据用户需要修改yaml文件（可选）。
    -   通用yaml文件

        **ascendplugin-910.yaml**

    -   MindX DL使用的yaml文件

        **vim ascendplugin-volcano.yaml**

        ```
        ......      
              containers: 
              - image: ascend-k8sdeviceplugin:v2.0.2   #镜像名称及版本号。 
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
                        -volcanoType=true
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
    drwxr-xr-x 2 root root     4096 Jun  8 18:42 ./ 
    drwxr-xr-x 9 root root     4096 Jun  8 17:12 ../ 
    -r-x------. 1 root root 31927176 Jul 26 14:12 device-plugin 
    -r--------. 1 root root     2081 Jul 26 14:12 device-plugin-310-v2.0.2.yaml 
    -r--------. 1 root root     2202 Jul 26 14:12 device-plugin-710-v2.0.2.yaml 
    -r--------. 1 root root     1935 Jul 26 14:12 device-plugin-910-v2.0.2.yaml 
    -r--------. 1 root root     3070 Jul 26 14:12 device-plugin-volcano-v2.0.2.yaml 
    -r--------. 1 root root      469 Jul 26 14:12 Dockerfile
    ```

    >![](figures/icon-note.gif) **说明：** 
    >“ascend-device-plugin“目录下的**ascendplugin-910.yaml**文件在“ascend-device-plugin/output/“下生成的对应文件为**device-plugin-910-v2.0.2.yaml**，作用是更新版本号。

6.  执行以下命令，查看Dockerfile文件，可根据实际情况修改。

    **vi** _/home/test/_**ascend-device-plugin/Dockerfile**

    ```
    #用户根据实际选择基础镜像，可通过docker images命令查询。 
    FROM ubuntu:18.04 as build 
     
    RUN useradd -d /home/HwHiAiUser -u 1000 -m -s /usr/sbin/nologin HwHiAiUser 
    #是否使用昇腾Docker，默认为true，请修改为false。 
    ENV USE_ASCEND_DOCKER true 
     
    ENV LD_LIBRARY_PATH  /usr/local/Ascend/driver/lib64/driver:/usr/local/Ascend/driver/lib64/common 
     
    ENV LD_LIBRARY_PATH $LD_LIBRARY_PATH:/usr/local/Ascend/driver/lib64/ 
    #确保device-plugin二进制文件存在以进行拷贝 
    COPY ./output/device-plugin /usr/local/bin/ 
     
    RUN chmod 550 /usr/local/bin/device-plugin &&\ 
    echo 'umask 027' >> /etc/profile &&\ 
    echo 'source /etc/profile' >> ~/.bashrc
    ```


<h2 id="创建DaemonSet.md">创建DaemonSet</h2>

## 操作步骤<a name="zh-cn_topic_0269670254_section2036324211563"></a>

>![](figures/icon-note.gif) **说明：** 
>以下操作以ARM平台下构建并分发镜像为例。

1.  获取Ascend Device Plugin软件包。如果是自己编译得到的软件包则跳过该步骤。
    1.  登录[MindX DL](https://support.huawei.com/enterprise/zh/ascend-computing/mindx-pid-252501207/software)。
    2.  打开目标版本，单击软件包后面的下载，可获取软件包和数字签名文件。
    3.  下载完后请进行软件包完整性校验。

2.  将软件包上传至服务器任意目录并解压（如：“/home/ascend-device-plugin“）。
3.  进入解压目录，执行以下命令构建Ascend Device Plugin的镜像。

    ```
    docker build -t ascend-k8sdeviceplugin:v2.0.2 .
    ```

    当出现“Successfully built xxx”表示镜像构建成功，注意不要遗漏命令结尾的“.”。这里的“.”代表当前目录。

    使用**docker images**命令查看，可以看见名为ascend-k8sdeviceplugin，tag为v2.0.2的镜像。

4.  执行如下命令将编译好的镜像打包并压缩，便于在各个服务器之间传输。

    >![](figures/icon-note.gif) **说明：** 
    >\{arch\}表示系统架构，不同架构之间的镜像不能通用。

    ```
    docker save -o ascend-k8sdeviceplugin:v2.0.2 | gzip > Ascend-K8sDevicePlugin-v2.0.2-{arch}-Docker.tar.gz
    ```

    或者使用不带压缩功能的命令：

    ```
    docker save -o Ascend-K8sDevicePlugin-v2.0.2-{arch}-Docker.tar ascend-k8sdeviceplugin:v2.0.2
    ```

5.  （可选）集群场景下将打包好的镜像（比如存放路径为：“/home/ascend-device-plugin/”）分发到拥有昇腾系列AI处理器的计算节点上。

    ```
    cd /home/ascend-device-plugin
    scp Ascend-K8sDevicePlugin-v2.0.2-{arch}-Docker.tar.gz root@{节点IP地址}:/home/ascend-device-plugin
    ```

6.  执行以下命令加载镜像。以打包压缩后的文件为例，没有选择压缩则需要修改相应的文件名。
    -   ARM架构

        ```
        docker load -i Ascend-K8sDevicePlugin-v2.0.2-arm64-Docker.tar.gz
        ```

    -   x86架构

        ```
        docker load -i Ascend-K8sDevicePlugin-v2.0.2-amd64-Docker.tar.gz
        ```

7.  执行如下命令，给带有Ascend 910（或含有Ascend 310、Ascend 710）的节点打标签。

    ```
    kubectl label nodes localhost.localdomain accelerator=huawei-Ascend910
    ```

    localhost.localdomain为有Ascend 910（或含有Ascend 310、Ascend 710）的节点名称，可通过**kubectl get node**命令查看。

    标签名称需要和软件包中yaml文件里的nodeSelector标签名称保持一致。

8.  进入“ascend-device-plugin/output“执行以下命令，部署DaemonSet。

    -   昇腾310 AI处理器节点。

        ```
        kubectl apply -f device-plugin-310-v2.0.2.yaml
        ```

    -   昇腾910 AI处理器节点，协同Volcano工作。

        ```
        kubectl apply -f device-plugin-volcano-v2.0.2.yaml
        ```

    -   昇腾710 AI处理器节点。

        ```
        kubectl apply -f device-plugin-710-v2.0.2.yaml
        ```
    
    - 昇腾910 AI处理器节点，Ascend Device Plugin独立工作，不协同Volcano。
    
      ```
      kubectl apply -f device-plugin-910-v2.0.2.yaml
      ```
    
    部署完成后需要等待几分钟，才能看到节点设备部署信息。yaml文件的中参数信息请参考[表1](#table1286935610129)。
    
    **表 1**  Ascend Device Plugin启动参数
    
    <a name="table1286935610129"></a>
    
    <table><thead align="left"><tr id="zh-cn_topic_0000001125023832_row78661456101218"><th class="cellrowborder" valign="top" width="25%" id="mcps1.2.5.1.1"><p id="zh-cn_topic_0000001125023832_p14866185661214"><a name="zh-cn_topic_0000001125023832_p14866185661214"></a><a name="zh-cn_topic_0000001125023832_p14866185661214"></a>参数</p>
    </th>
    <th class="cellrowborder" valign="top" width="15%" id="mcps1.2.5.1.2"><p id="zh-cn_topic_0000001125023832_p58661456151210"><a name="zh-cn_topic_0000001125023832_p58661456151210"></a><a name="zh-cn_topic_0000001125023832_p58661456151210"></a>类型</p>
    </th>
    <th class="cellrowborder" valign="top" width="15%" id="mcps1.2.5.1.3"><p id="zh-cn_topic_0000001125023832_p3866185610124"><a name="zh-cn_topic_0000001125023832_p3866185610124"></a><a name="zh-cn_topic_0000001125023832_p3866185610124"></a>默认值</p>
    </th>
    <th class="cellrowborder" valign="top" width="45%" id="mcps1.2.5.1.4"><p id="zh-cn_topic_0000001125023832_p2866556171219"><a name="zh-cn_topic_0000001125023832_p2866556171219"></a><a name="zh-cn_topic_0000001125023832_p2866556171219"></a>说明</p>
    </th>
    </tr>
    </thead>
    <tbody><tr id="zh-cn_topic_0000001125023832_row178677563121"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000001125023832_p186620564122"><a name="zh-cn_topic_0000001125023832_p186620564122"></a><a name="zh-cn_topic_0000001125023832_p186620564122"></a>-mode</p>
    </td>
    <td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000001125023832_p18666563129"><a name="zh-cn_topic_0000001125023832_p18666563129"></a><a name="zh-cn_topic_0000001125023832_p18666563129"></a>string</p>
    </td>
    <td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000001125023832_p686645671212"><a name="zh-cn_topic_0000001125023832_p686645671212"></a><a name="zh-cn_topic_0000001125023832_p686645671212"></a>无</p>
    </td>
    <td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000001125023832_p10866156121215"><a name="zh-cn_topic_0000001125023832_p10866156121215"></a><a name="zh-cn_topic_0000001125023832_p10866156121215"></a>指定Ascend Device Plugin运行模式，不指定该参数会根据NPU芯片类型自动指定。</p>
    <a name="zh-cn_topic_0000001125023832_ul14191351239"></a><a name="zh-cn_topic_0000001125023832_ul14191351239"></a><ul id="zh-cn_topic_0000001125023832_ul14191351239"><li>ascend310：以<span id="zh-cn_topic_0000001125023832_ph9702132613260"><a name="zh-cn_topic_0000001125023832_ph9702132613260"></a><a name="zh-cn_topic_0000001125023832_ph9702132613260"></a>昇腾310 AI处理器</span>的模式运行</li><li>ascend710：以昇腾710 AI处理器的模式运行</li><li>ascend910：以<span id="zh-cn_topic_0000001125023832_ph10107162619264"><a name="zh-cn_topic_0000001125023832_ph10107162619264"></a><a name="zh-cn_topic_0000001125023832_ph10107162619264"></a>昇腾910 AI处理器</span>的模式运行</li></ul>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001125023832_row48671256151214"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000001125023832_p12867145618121"><a name="zh-cn_topic_0000001125023832_p12867145618121"></a><a name="zh-cn_topic_0000001125023832_p12867145618121"></a>-fdFlag</p>
    </td>
    <td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000001125023832_p17867195661211"><a name="zh-cn_topic_0000001125023832_p17867195661211"></a><a name="zh-cn_topic_0000001125023832_p17867195661211"></a>bool</p>
    </td>
    <td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000001125023832_p2867856131218"><a name="zh-cn_topic_0000001125023832_p2867856131218"></a><a name="zh-cn_topic_0000001125023832_p2867856131218"></a>false</p>
    </td>
    <td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000001125023832_p12867956131212"><a name="zh-cn_topic_0000001125023832_p12867956131212"></a><a name="zh-cn_topic_0000001125023832_p12867956131212"></a>边缘场景标志，是否使用<span id="zh-cn_topic_0000001125023832_ph147289368387"><a name="zh-cn_topic_0000001125023832_ph147289368387"></a><a name="zh-cn_topic_0000001125023832_ph147289368387"></a>FusionDirector</span>系统来管理设备。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001125023832_row198671564128"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000001125023832_p1486717563122"><a name="zh-cn_topic_0000001125023832_p1486717563122"></a><a name="zh-cn_topic_0000001125023832_p1486717563122"></a>-useAscendDocker</p>
    </td>
    <td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000001125023832_p5867856161210"><a name="zh-cn_topic_0000001125023832_p5867856161210"></a><a name="zh-cn_topic_0000001125023832_p5867856161210"></a>bool</p>
    </td>
    <td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000001125023832_p78671956171212"><a name="zh-cn_topic_0000001125023832_p78671956171212"></a><a name="zh-cn_topic_0000001125023832_p78671956171212"></a>true</p>
    </td>
    <td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000001125023832_p8867856151215"><a name="zh-cn_topic_0000001125023832_p8867856151215"></a><a name="zh-cn_topic_0000001125023832_p8867856151215"></a>是否使用Ascend-docker-runtime。</p>
    <div class="note" id="zh-cn_topic_0000001125023832_note58972114252"><a name="zh-cn_topic_0000001125023832_note58972114252"></a><a name="zh-cn_topic_0000001125023832_note58972114252"></a><span class="notetitle"> 说明： </span><div class="notebody"><p id="zh-cn_topic_0000001125023832_p20899219253"><a name="zh-cn_topic_0000001125023832_p20899219253"></a><a name="zh-cn_topic_0000001125023832_p20899219253"></a>开启CPU绑核功能，无论是否使用Ascend-docker-runtime，参数“useAscendDocker”均设置为false。</p>
    </div></div>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001125023832_row13867205691219"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000001125023832_p1086715563123"><a name="zh-cn_topic_0000001125023832_p1086715563123"></a><a name="zh-cn_topic_0000001125023832_p1086715563123"></a>-volcanoType</p>
    </td>
    <td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000001125023832_p2867556181217"><a name="zh-cn_topic_0000001125023832_p2867556181217"></a><a name="zh-cn_topic_0000001125023832_p2867556181217"></a>bool</p>
    </td>
    <td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000001125023832_p138676568125"><a name="zh-cn_topic_0000001125023832_p138676568125"></a><a name="zh-cn_topic_0000001125023832_p138676568125"></a>false</p>
    </td>
    <td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000001125023832_p1867145614122"><a name="zh-cn_topic_0000001125023832_p1867145614122"></a><a name="zh-cn_topic_0000001125023832_p1867145614122"></a>是否使用Volcano进行调度。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001125023832_row188681056111220"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000001125023832_p1486795641210"><a name="zh-cn_topic_0000001125023832_p1486795641210"></a><a name="zh-cn_topic_0000001125023832_p1486795641210"></a>-version</p>
    </td>
    <td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000001125023832_p58671756201215"><a name="zh-cn_topic_0000001125023832_p58671756201215"></a><a name="zh-cn_topic_0000001125023832_p58671756201215"></a>bool</p>
    </td>
    <td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000001125023832_p15868125621216"><a name="zh-cn_topic_0000001125023832_p15868125621216"></a><a name="zh-cn_topic_0000001125023832_p15868125621216"></a>false</p>
    </td>
    <td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000001125023832_p17868135615124"><a name="zh-cn_topic_0000001125023832_p17868135615124"></a><a name="zh-cn_topic_0000001125023832_p17868135615124"></a>查看当前device-plugin的版本号。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001125023832_row11868756151216"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000001125023832_p8868135631213"><a name="zh-cn_topic_0000001125023832_p8868135631213"></a><a name="zh-cn_topic_0000001125023832_p8868135631213"></a>-edgeLogFile</p>
    </td>
    <td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000001125023832_p1868155620126"><a name="zh-cn_topic_0000001125023832_p1868155620126"></a><a name="zh-cn_topic_0000001125023832_p1868155620126"></a>string</p>
    </td>
    <td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000001125023832_p18868145616126"><a name="zh-cn_topic_0000001125023832_p18868145616126"></a><a name="zh-cn_topic_0000001125023832_p18868145616126"></a>/var/alog/AtlasEdge_log/devicePlugin.log</p>
    </td>
    <td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000001125023832_p15868145614121"><a name="zh-cn_topic_0000001125023832_p15868145614121"></a><a name="zh-cn_topic_0000001125023832_p15868145614121"></a>边缘场景日志路径配置。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001125023832_row178281038145918"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000001125023832_p716681713714"><a name="zh-cn_topic_0000001125023832_p716681713714"></a><a name="zh-cn_topic_0000001125023832_p716681713714"></a>-logLevel</p>
    </td>
    <td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000001125023832_p116631743711"><a name="zh-cn_topic_0000001125023832_p116631743711"></a><a name="zh-cn_topic_0000001125023832_p116631743711"></a>int</p>
    </td>
    <td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000001125023832_p10166121714371"><a name="zh-cn_topic_0000001125023832_p10166121714371"></a><a name="zh-cn_topic_0000001125023832_p10166121714371"></a>0</p>
    </td>
    <td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000001125023832_p7166161783714"><a name="zh-cn_topic_0000001125023832_p7166161783714"></a><a name="zh-cn_topic_0000001125023832_p7166161783714"></a>日志级别：</p>
    <a name="zh-cn_topic_0000001125023832_ul086611015481"></a><a name="zh-cn_topic_0000001125023832_ul086611015481"></a><ul id="zh-cn_topic_0000001125023832_ul086611015481"><li>-1：debug</li><li>0：info</li><li>1：warning</li><li>2：error</li><li>3：dpanic</li><li>4：panic</li><li>5：fatal</li></ul>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001125023832_row12828338105911"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000001125023832_p962414206374"><a name="zh-cn_topic_0000001125023832_p962414206374"></a><a name="zh-cn_topic_0000001125023832_p962414206374"></a>-maxAge</p>
    </td>
    <td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000001125023832_p196241420153715"><a name="zh-cn_topic_0000001125023832_p196241420153715"></a><a name="zh-cn_topic_0000001125023832_p196241420153715"></a>int</p>
    </td>
    <td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000001125023832_p5624720113715"><a name="zh-cn_topic_0000001125023832_p5624720113715"></a><a name="zh-cn_topic_0000001125023832_p5624720113715"></a>7</p>
    </td>
    <td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000001125023832_p2062492033716"><a name="zh-cn_topic_0000001125023832_p2062492033716"></a><a name="zh-cn_topic_0000001125023832_p2062492033716"></a>日志备份时间限制，最少为7天，单位：天。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001125023832_row19827193825914"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000001125023832_p8601162318374"><a name="zh-cn_topic_0000001125023832_p8601162318374"></a><a name="zh-cn_topic_0000001125023832_p8601162318374"></a>-isCompress</p>
    </td>
    <td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000001125023832_p360112233376"><a name="zh-cn_topic_0000001125023832_p360112233376"></a><a name="zh-cn_topic_0000001125023832_p360112233376"></a>bool</p>
    </td>
    <td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000001125023832_p36029231378"><a name="zh-cn_topic_0000001125023832_p36029231378"></a><a name="zh-cn_topic_0000001125023832_p36029231378"></a>false</p>
    </td>
    <td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000001125023832_p1602102311375"><a name="zh-cn_topic_0000001125023832_p1602102311375"></a><a name="zh-cn_topic_0000001125023832_p1602102311375"></a>是否自动将备份日志文件进行压缩。</p>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001125023832_row282710388595"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000001125023832_p172501226103710"><a name="zh-cn_topic_0000001125023832_p172501226103710"></a><a name="zh-cn_topic_0000001125023832_p172501226103710"></a>-logFile</p>
    </td>
    <td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000001125023832_p13250626193718"><a name="zh-cn_topic_0000001125023832_p13250626193718"></a><a name="zh-cn_topic_0000001125023832_p13250626193718"></a>string</p>
    </td>
    <td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000001125023832_p225052653711"><a name="zh-cn_topic_0000001125023832_p225052653711"></a><a name="zh-cn_topic_0000001125023832_p225052653711"></a>/var/log/mindx-dl/devicePlugin/devicePlugin.log</p>
    </td>
    <td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000001125023832_p1525042673718"><a name="zh-cn_topic_0000001125023832_p1525042673718"></a><a name="zh-cn_topic_0000001125023832_p1525042673718"></a>日志文件。</p>
    <div class="note" id="zh-cn_topic_0000001125023832_note82654194514"><a name="zh-cn_topic_0000001125023832_note82654194514"></a><a name="zh-cn_topic_0000001125023832_note82654194514"></a><span class="notetitle"> 说明： </span><div class="notebody"><p id="zh-cn_topic_0000001125023832_p13265151916514"><a name="zh-cn_topic_0000001125023832_p13265151916514"></a><a name="zh-cn_topic_0000001125023832_p13265151916514"></a>单个日志文件超过20 MB时会触发自动转储功能，文件大小上限不支持修改。</p>
    </div></div>
    </td>
    </tr>
    <tr id="zh-cn_topic_0000001125023832_row15827113819595"><td class="cellrowborder" valign="top" width="25%" headers="mcps1.2.5.1.1 "><p id="zh-cn_topic_0000001125023832_p12517346377"><a name="zh-cn_topic_0000001125023832_p12517346377"></a><a name="zh-cn_topic_0000001125023832_p12517346377"></a>-maxBackups</p>
    </td>
    <td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.2 "><p id="zh-cn_topic_0000001125023832_p225133416374"><a name="zh-cn_topic_0000001125023832_p225133416374"></a><a name="zh-cn_topic_0000001125023832_p225133416374"></a>int</p>
    </td>
    <td class="cellrowborder" valign="top" width="15%" headers="mcps1.2.5.1.3 "><p id="zh-cn_topic_0000001125023832_p18251173413378"><a name="zh-cn_topic_0000001125023832_p18251173413378"></a><a name="zh-cn_topic_0000001125023832_p18251173413378"></a>30</p>
    </td>
    <td class="cellrowborder" valign="top" width="45%" headers="mcps1.2.5.1.4 "><p id="zh-cn_topic_0000001125023832_p225183419372"><a name="zh-cn_topic_0000001125023832_p225183419372"></a><a name="zh-cn_topic_0000001125023832_p225183419372"></a>转储后日志文件保留个数上限，范围：(0，30]，单位：个。</p>
    </td>
    </tr>
    </tbody>
    </table>
    
9. 执行如下命令，查看节点设备部署信息。

   ```
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

    -   huawei.com/Ascend310: 2，表示分配2颗昇腾310 AI处理器。
    -   huawei.com/Ascend710: 1，表示分配1颗昇腾710 AI处理器。
    -   huawei.com/Ascend910: 4，表示分配4颗昇腾910 AI处理器。
    -   huawei.com/Ascend910-16c: 1，表示分配1颗算力为16核的虚拟设备，只支持单卡单容器任务（即数值只能填1）。支持调度2c、4c、8c、16c四种AI core数量的虚拟设备。

2.  执行如下命令，创建Pod。

    ```
    kubectl apply -f ascend.yaml
    ```

    >![](figures/icon-note.gif) **说明：** 
    >如需删除请执行以下命令：
    >**kubectl delete -f** **ascend.yaml**

3.  分别执行以下命令，进入Pod查看分配信息。

    ```
    kubectl exec -it pod名称 bash
    ```

    Pod名称为[1](#zh-cn_topic_0269670251_zh-cn_topic_0249483204_li104071617503)中配置的Pod名称。

    ```
    ls /dev/
    ```

    如下类似回显信息中可以看到davinci3和davinci4即为分配的Pod。

    ```
    core davinci3 davinci4 davinci_manager devmm_svm fd full hisi_hdc mqueue null ptmx
    ```


<h2 id="目录结构.md">目录结构</h2>

```
├── build                                             # 编译脚本   
│   └── build.sh
├── output                                           # 编译结果目录
├── src                                              # 源代码目录
│   └── plugin
│   │    ├── cmd/ascendplugin
│   │    │   └── ascend_plugin.go    
│   │    └── pkg/npu/huawei
├── test                                             # 测试目录
├── Dockerfile                                       # 镜像文件
├── LICENSE                                          
├── Open Source Software Notice.md                   
├── README.ZH.md
├── README.EN.md
├── ascend.yaml                                      # sample运行任务yaml
├── ascendplugin-310.yaml                            # 310推理卡部署yaml
├── ascendplugin-710.yaml                            # 710推理卡部署yaml
├── ascendplugin-volcano.yaml                        # 910搭配volcano实现亲和性调度部署yaml
├── ascendplugin-910.yaml                            # 910未使用volcano部署yaml
├── go.mod                                           
└── go.sum                                           
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
<tbody><tr id="row7293189122012"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.1 "><p id="p9235101416201"><a name="p9235101416201"></a><a name="p9235101416201"></a>v2.0.2</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.2 "><p id="p1523518145208"><a name="p1523518145208"></a><a name="p1523518145208"></a>2021-07-15</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.3 "><a name="ul162558202525"></a><a name="ul162558202525"></a><ul id="ul162558202525"><li>昇腾910 AI处理器支持算力拆分。</li></ul>
</td>
</tr>
<tr id="row13392135841913"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.1 "><p id="p14450752204"><a name="p14450752204"></a><a name="p14450752204"></a>v2.0.1</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.2 "><p id="p1445019582013"><a name="p1445019582013"></a><a name="p1445019582013"></a>2021-04-20</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.3 "><a name="ul194113318201"></a><a name="ul194113318201"></a><ul id="ul194113318201"><li>适配昇腾710 AI处理器。</li><li>处理器信息上报由逻辑ID修改为物理ID。</li><li>处理器一般告警修改为不主动隔离。</li></ul>
</td>
</tr>
<tr id="row137501013384"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.1 "><p id="p137501613585"><a name="p137501613585"></a><a name="p137501613585"></a>v20.2.0</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.2 "><p id="p1675010135811"><a name="p1675010135811"></a><a name="p1675010135811"></a>2021-01-08</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.3 "><p id="p3750813889"><a name="p3750813889"></a><a name="p3750813889"></a>优化“创建DaemonSet”描述。</p>
</td>
</tr>
<tr id="zh-cn_topic_0280467800_row118567425441"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.1 "><p id="zh-cn_topic_0280467800_p08571442174415"><a name="zh-cn_topic_0280467800_p08571442174415"></a><a name="zh-cn_topic_0280467800_p08571442174415"></a>v20.2.0</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.2 "><p id="zh-cn_topic_0280467800_p38571542154414"><a name="zh-cn_topic_0280467800_p38571542154414"></a><a name="zh-cn_topic_0280467800_p38571542154414"></a>2020-11-18</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.3 "><p id="zh-cn_topic_0280467800_p5857142154415"><a name="zh-cn_topic_0280467800_p5857142154415"></a><a name="zh-cn_topic_0280467800_p5857142154415"></a>第一次正式发布。</p>
</td>
</tr>
</tbody>
</table>

