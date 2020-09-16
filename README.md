# Ascend Device Plugin
-   [Ascend Device Plugin](#ascend-device-plugin.md)
    -   [Description](#description.md)
    -   [Compiling the Ascend Device Plugin](#compiling-the-ascend-device-plugin.md)
        -   [Quickly Compiling the Ascend Device Plugin](#quickly-compiling-the-ascend-device-plugin.md)
        -   [Compiling the Ascend Device Plugin](#compiling-the-ascend-device-plugin-0.md)
    -   [Creating DaemonSet.](#creating-daemonset.md)
    -   [Creating a Service Container](#creating-a-service-container.md)
-   [Environment Dependencies](#environment-dependencies.md)
-   [Directory Structure](#directory-structure.md)
-   [Version Updates](#version-updates.md)
<h2 id="ascend-device-plugin.md">Ascend Device Plugin</h2>

-   **[Description](#description.md)**  

-   **[Compiling the Ascend Device Plugin](#compiling-the-ascend-device-plugin.md)**  

-   **[Creating DaemonSet.](#creating-daemonset.md)**  

-   **[Creating a Service Container](#creating-a-service-container.md)**  


<h2 id="description.md">Description</h2>

The device management plug-in provides the following functions:

-   Device discovery: The number of discovered devices can be obtained from the Ascend device driver and reported to the Kubernetes system.
-   Health check: The health status of Ascend devices can be detected. When a device is unhealthy, the device is reported to the Kubernetes system and is removed.
-   Device allocation: Ascend devices can be allocated in the Kubernetes system.

<h2 id="compiling-the-ascend-device-plugin.md">Compiling the Ascend Device Plugin</h2>

-   **[Quickly Compiling the Ascend Device Plugin](#quickly-compiling-the-ascend-device-plugin.md)**  
You can modify the configuration parameters during compilation by running a shell script. You only need to modify the parameters in the script to quickly complete the compilation.
-   **[Compiling the Ascend Device Plugin](#compiling-the-ascend-device-plugin-0.md)**  


<h2 id="quickly-compiling-the-ascend-device-plugin.md">Quickly Compiling the Ascend Device Plugin</h2>

You can modify the configuration parameters during compilation by running a shell script. You only need to modify the parameters in the script to quickly complete the compilation.

## Procedure<a name="section125457120293"></a>

1.  Run the following command to install the latest pkg-config tool:

    **apt-get install -y pkg-config**

2.  Run the following commands to set environment variables:

    **export GO111MODULE=on**

    **export GOPROXY=**_Proxy address_

    **export GONOSUMDB=\***

    >![](figures/icon-note.gif) **NOTE:** 
    >Use the actual GOPROXY proxy address. You can run the  **go mod download**  command in the  **ascend-device-plugin**  directory to check the address.

3.  Create and execute the shell file in  **./build/**.

    ```
      #!/bin/bash
      ASCEND_TYPE=910 #Select 310 or 910 based on the processor model.
      ASCNED_INSTALL_PATH=/usr/local/Ascend  #Driver installation path. Change it as required.
      USE_ASCEND_DOCKER=false  #whether to use Ascend Docker.
    
    
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

4.  Run the following commands to generate a binary file and image file \(use the actual script name\):

    Select  **build\_910.sh**  for  Ascend 910  and select  **build\_310.sh**  for  Ascend 310.

    **cd** _/home/test/_ascend-device-plugin**/build**/

    **chmod +x** _build\_XXX.sh_

    **dos2unix** _build\_XXX.sh_

    **./**_build\_XXX.sh_ **dockerimages**

5.  Run the following command to view the generated software package:

    **ll** _/home/test/_ascend-device-plugin**/output**

    The software package name for the x86 environment and that for the ARM environment are different. The following uses the ARM environment as an example.

    >![](figures/icon-note.gif) **NOTE:** 
    >-   **Ascend-K8sDevicePlugin-**_xxx_**-arm64-Docker.tar.gz**: K8s device plugin image.
    >-   **Ascend-K8sDevicePlugin-**_xxx_**-arm64-Linux.tar.gz**: binary installation package of the K8s device plugin.

    ```
    drwxr-xr-x 2 root root     4096 Jun  8 18:42 ./
    drwxr-xr-x 9 root root     4096 Jun  8 17:12 ../
    -rw-r--r-- 1 root root 29584705 Jun  9 10:37 Ascend-K8sDevicePlugin-xxx-arm64-Docker.tar.gz
    -rwxr-xr-x 1 root root  6721073 Jun  9 16:20 Ascend-K8sDevicePlugin-xxx-arm64-Linux.tar.gz
    ```


<h2 id="compiling-the-ascend-device-plugin-0.md">Compiling the Ascend Device Plugin</h2>

## Procedure<a name="section112101632152317"></a>

1.  Run the following command to install the latest pkg-config tool:

    **apt-get install -y pkg-config**

2.  Run the following commands to set environment variables:

    **export GO111MODULE=on**

    **export GOPROXY=**_Proxy address_

    **export GONOSUMDB=\***

    >![](figures/icon-note.gif) **NOTE:** 
    >Use the actual GOPROXY proxy address. You can run the  **go mod download**  command in the  **ascend-device-plugin**  directory to check the address.

3.  Go to the  **ascend-device-plugin**  directory and run the following command to modify the YAML file:
    -   Common YAML file

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
                - key: huawei.com/Ascend910 #Resource name. Set the value based on the chip type.
                  operator: Exists
                  effect: NoSchedule
                - key: "ascendplugin"
                  operator: "Equal"
                  value: "v2"
                  effect: NoSchedule
              priorityClassName: "system-node-critical"
              nodeSelector:
                accelerator: huawei-Ascend910 #Set the label name based on the chip type.
              containers:
              - image: ascend-device-plugin:v1.0.1  #Image name and version, which must be the same as the settings in build_common.sh.
                name: device-plugin-01
                resources:
                  requests:
                    memory: 500Mi
                    cpu: 500m
                  limits:
                    memory: 500Mi
                    cpu: 500m
                command: [ "/bin/bash", "-c", "--"]
                args: [ "./build/build_in_docker.sh;ascendplugin  --useAscendDocker=${USE_ASCEND_DOCKER}" ] #Add --mode=ascend310 if Ascend310 is used.
                securityContext:
                  privileged: true
                imagePullPolicy: Never
                volumeMounts:
                  - name: device-plugin
                    mountPath: /var/lib/kubelet/device-plugins
                  - name: hiai-driver
                    mountPath: /usr/local/Ascend/driver  #Set the value to the actual driver installation directory.
                  - name: log-path
                    mountPath: /var/log/devicePlugin
              volumes:
                - name: device-plugin
                  hostPath:
                    path: /var/lib/kubelet/device-plugins
                - name: hiai-driver
                  hostPath:
                    path: /usr/local/Ascend/driver  #Set the value to the actual driver installation directory.
                - name: log-path
                  hostPath:
                    path: /var/log/devicePlugin
        
        ```

    -   The YAML file used by  MindX DL

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
              - image: ascend-k8sdeviceplugin:V20.1.0  #Image name and version, which must be the same as the settings in build_common.sh.
                name: device-plugin-01
                resources:
                  requests:
                    memory: 500Mi
                    cpu: 500m
                  limits:
                    memory: 500Mi
                    cpu: 500m
                command: [ "/bin/bash", "-c", "--"]
                args: [ "./build/build_in_docker.sh;ascendplugin  --useAscendDocker=${USE_ASCEND_DOCKER} --volcanoType=true" ] #Add --mode=ascend310 if Ascend310 is used.
                securityContext:
                  privileged: true
                imagePullPolicy: Never
                volumeMounts:
                  - name: device-plugin
                    mountPath: /var/lib/kubelet/device-plugins
                  - name: hiai-driver
                    mountPath: /usr/local/Ascend/driver  #Set the value to the actual driver installation directory.
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
                    path: /usr/local/Ascend/driver  #Set the value to the actual driver installation directory.
                - name: log-path
                  hostPath:
                    path: /var/log/devicePlugin
        
        ```

4.  Run the following command to edit the  **Dockerfile**  file and change the image name and version to the obtained values:

    **vi **_/home/test/_ascend-device-plugin**/Dockerfile**

    ```
    #Select the basic image with go compilation. You can run the docker images command to query the basic image.
    FROM golang:1.13.11-buster as build
    
    #Specify whether to use Ascend Docker. The default value is true. Change it to false.
    ENV USE_ASCEND_DOCKER false
    
    ENV GOPATH /usr/app/
    
    ENV GO111MODULE off
    
    ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH
    #Directory where libdrvdsmi_host.so is located. The directories of Ascend 310 and Ascend 910 are different.
    ENV LD_LIBRARY_PATH  /usr/local/Ascend/driver/lib64/driver:/usr/local/Ascend/driver/lib64/common
    
    RUN mkdir -p /usr/app/src/ascend-device-plugin
    
    COPY . /usr/app/src/Ascend-device-plugin
    
    WORKDIR /usr/app/src/Ascend-device-plugin
    ```

5.  Go to the directory where the  **ascend\_device\_plugin.pc**  file is stored and run the following command to check whether the following paths are correct.

    -   Ascend 310  path:  **ascend-device-plugin/src/plugin/config/config\_310**
    -   Ascend 910  path:  **ascend-device-plugin/src/plugin/config/config\_910**

    **vi ascend\_device\_plugin.pc**

    ```
    #Package Information for pkg-config
    #Set the value to the actual driver installation directory.
    prefix=/usr/local/Ascend
    #Change the value to the actual DSMI dynamic library address.
    libdriver=${prefix}/driver/lib64
    #Directory of the DSMI driver header file dsmi_common_interface.h.
    includedir=${prefix}/driver/kernel/inc/driver/
    Name: ascend_docker_plugin
    Description: Ascend device plugin
    Version: 0.0.1
    Libs: -L${libdriver}/    -ldrvdsmi_host
    Cflags: -I${includedir}
    ```

    >![](figures/icon-note.gif) **NOTE:** 
    >You can change the value of  **docker\_images\_name**  in  **build\_common.sh**  in the  **build**  directory to change the plugin image name. Ensure that the value is the same as the setting in  **ascendplugin.yaml**.

6.  Run the following commands to generate a binary file and image file \(use the actual script name\):

    Select  **build\_910.sh**  for  Ascend 910  and select  **build\_310.sh**  for  Ascend 310.

    **cd** _/home/test/_ascend-device-plugin**/build**/

    **chmod +x** _build\_XXX.sh_

    **dos2unix** _build\_XXX.sh_

    **./**_build\_XXX.sh_ **dockerimages**


1.  Run the following command to view the generated software package:

    **ll** _/home/test/_ascend-device-plugin**/output**

    The software package name for the x86 environment and that for the ARM environment are different. The following uses the ARM environment as an example.

    >![](figures/icon-note.gif) **NOTE:** 
    >-   **Ascend-K8sDevicePlugin-**_xxx_**-arm64-Docker.tar.gz**: K8s device plugin image.
    >-   **Ascend-K8sDevicePlugin-**_xxx_**-arm64-Linux.tar.gz**: binary installation package of the K8s device plugin.

    ```
    drwxr-xr-x 2 root root     4096 Jun  8 18:42 ./
    drwxr-xr-x 9 root root     4096 Jun  8 17:12 ../
    -rw-r--r-- 1 root root 29584705 Jun  9 10:37 Ascend-K8sDevicePlugin-xxx-arm64-Docker.tar.gz
    -rwxr-xr-x 1 root root  6721073 Jun  9 16:20 Ascend-K8sDevicePlugin-xxx-arm64-Linux.tar.gz
    ```


<h2 id="creating-daemonset.md">Creating DaemonSet.</h2>

## Procedure<a name="en-us_topic_0269670254_section2036324211563"></a>

>![](figures/icon-note.gif) **NOTE:** 
>The following uses the tar.gz file generated on the ARM platform as an example.

1.  Run the following command to check whether the Docker software package is successfully imported:

    **docker images**

    -   If yes, go to  [3](#en-us_topic_0269670254_li26268471380).
    -   If no, perform  [2](#en-us_topic_0269670254_li1372334715567)  to import the file again.

2.  <a name="en-us_topic_0269670254_li1372334715567"></a>Go to the directory where the Docker software package is stored and run the following command to import the Docker image.

    **cd** _/home/test/_**ascend-device-plugin/output**

    **docker load < **_Ascend-K8sDevicePlugin-xxx-arm64-Docker.tar.gz_

3.  <a name="en-us_topic_0269670254_li26268471380"></a>Run the following command to label the node with Ascend 910 or  Ascend 310:

    **kubectl label nodes **_localhost.localdomain_** accelerator=**_huawei-Ascend910_

    **localhost.localdomain**  is the name of the node with Ascend 910 or  Ascend 310. You can run the  **kubectl get node**  command to view the node name.

    The label name must be the same as the name specified by  **nodeSelector**  in  [3](#compiling-the-ascend-device-plugin-0.md#en-us_topic_0252775101_li8538035183714).

    >![](figures/icon-note.gif) **NOTE:** 
    >You can perform  [2](#en-us_topic_0269670254_li1372334715567)  to  [3](#en-us_topic_0269670254_li26268471380)  to add new nodes to the cluster.

4.  Run the following commands to deploy DaemonSet:

    **cd** _/home/test/_**ascend-device-plugin**

    **kubectl apply -f  ascendplugin.yaml**

    >![](figures/icon-note.gif) **NOTE:** 
    >To view the node deployment information, you need to wait for several minutes after the deployment is complete.


1.  Run the following command to view the node device deployment information:

    **kubectl describe node**

    If the label and number of nodes are correct, the deployment is successful, as shown in the following figure.

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


<h2 id="creating-a-service-container.md">Creating a Service Container</h2>

## Procedure<a name="en-us_topic_0269670251_section28051148174119"></a>

1.  <a name="en-us_topic_0269670251_en-us_topic_0249483204_li104071617503"></a>Go to the  **ascend-device-plugin**  directory and run the following command to edit the pod configuration file:

    **cd**_ /home/test/_**ascend-device-plugin**

    **vi ascend.yaml**

    ```
    apiVersion: v1 #Specifies the API version. This value must be included in kubectl apiversion.
    kind: Pod #Role or type of the resource to be created
    metadata:
      name: rest502 #Pod name, which must be unique in the same namespace.
    spec:  #Detailed definition of a container in a pod.
      containers: #Containers in the pod.
      - name: rest502 #Container name in the pod.
        image: centos_arm64_resnet50:7.8 #Address of the inference or training service image used by the container in the pod.
        imagePullPolicy: Never
        resources:
          limits: #Resource limits
            huawei.com/Ascend310: 2 #Change the value based on the actual resource type.
        volumeMounts:
          - name: joblog
            mountPath: /home/log/ #Path of the container internal log. Change the value based on the task requirements.
          - name: model
            mountPath: /home/app/model #Container internal model path. Change the value based on the task requirements.
          - name: slog-path
            mountPath: /var/log/npu/conf/slog/slog.conf
          - name: ascend-driver-path
            mountPath: /usr/local/Ascend/driver #Change the value based on the actual driver path.
      volumes:
        - name: joblog
          hostPath:
            path: /home/test/docker_log    #Log path mounted to the host. Change the value based on the task requirements.
        - name: model
          hostPath:
            path: /home/test/docker_model/  #Model path mounted to the host. Change the value based on the task requirements.
        - name: slog-path
          hostPath:
            path: /var/log/npu/conf/slog/slog.conf  
        - name: ascend-driver-path
          hostPath:
            path: /usr/local/Ascend/driver #Change the value based on the actual driver path.
    ```


1.  Run the following command to create a pod:

    **kubectl apply -f ascend.yaml**

    >![](figures/icon-note.gif) **NOTE:** 
    >To delete the pod, run the following command:
    >**kubectl delete -f** **ascend.yaml**


1.  Run the following commands to access the pod and view the allocation information:

    **kubectl exec -it **_Pod name_** bash**

    The pod name is the one configured in  [1](#en-us_topic_0269670251_en-us_topic_0249483204_li104071617503).

    **ls /dev/**

    In the command output similar to the following,  **davinci3**  and  **davinci4**  are the allocated pods.

    ```
    core davinci3 davinci4 davinci_manager devmm_svm fd full hisi_hdc mqueue null ptmx
    ```


<h2 id="environment-dependencies.md">Environment Dependencies</h2>

**Table  1**  Environment dependencies

<a name="en-us_topic_0252788324_table171211952105718"></a>
<table><thead align="left"><tr id="en-us_topic_0269670261_en-us_topic_0252788324_row51223524573"><th class="cellrowborder" valign="top" width="30%" id="mcps1.2.3.1.1"><p id="en-us_topic_0269670261_en-us_topic_0252788324_p15122175218576"><a name="en-us_topic_0269670261_en-us_topic_0252788324_p15122175218576"></a><a name="en-us_topic_0269670261_en-us_topic_0252788324_p15122175218576"></a>Check Item</p>
</th>
<th class="cellrowborder" valign="top" width="70%" id="mcps1.2.3.1.2"><p id="en-us_topic_0269670261_en-us_topic_0252788324_p1712211526578"><a name="en-us_topic_0269670261_en-us_topic_0252788324_p1712211526578"></a><a name="en-us_topic_0269670261_en-us_topic_0252788324_p1712211526578"></a>Requirement</p>
</th>
</tr>
</thead>
<tbody><tr id="en-us_topic_0269670261_row1985835314489"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.3.1.1 "><p id="en-us_topic_0269670261_p1925915619412"><a name="en-us_topic_0269670261_p1925915619412"></a><a name="en-us_topic_0269670261_p1925915619412"></a>dos2unix</p>
</td>
<td class="cellrowborder" valign="top" width="70%" headers="mcps1.2.3.1.2 "><p id="en-us_topic_0269670261_p1025985634111"><a name="en-us_topic_0269670261_p1025985634111"></a><a name="en-us_topic_0269670261_p1025985634111"></a>Run the <strong id="en-us_topic_0269670261_b026181053915"><a name="en-us_topic_0269670261_b026181053915"></a><a name="en-us_topic_0269670261_b026181053915"></a>dos2unix --version</strong> command to check that the software has been installed. There is no requirement on the version.</p>
</td>
</tr>
<tr id="en-us_topic_0269670261_row16906451114817"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.3.1.1 "><p id="en-us_topic_0269670261_p212295212575"><a name="en-us_topic_0269670261_p212295212575"></a><a name="en-us_topic_0269670261_p212295212575"></a>Driver version of the RUN package</p>
</td>
<td class="cellrowborder" valign="top" width="70%" headers="mcps1.2.3.1.2 "><p id="en-us_topic_0269670261_p31997012111"><a name="en-us_topic_0269670261_p31997012111"></a><a name="en-us_topic_0269670261_p31997012111"></a>Go to the directory of the driver (for example, <strong id="en-us_topic_0269670261_b1580411415488"><a name="en-us_topic_0269670261_b1580411415488"></a><a name="en-us_topic_0269670261_b1580411415488"></a>/usr/local/Ascend/driver</strong>) and run the <strong id="en-us_topic_0269670261_b1954711304617"><a name="en-us_topic_0269670261_b1954711304617"></a><a name="en-us_topic_0269670261_b1954711304617"></a>cat version.info</strong> command to confirm that the driver version is 1.73 or later.</p>
</td>
</tr>
<tr id="en-us_topic_0269670261_row12226135012483"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.3.1.1 "><p id="en-us_topic_0269670261_p3124195265717"><a name="en-us_topic_0269670261_p3124195265717"></a><a name="en-us_topic_0269670261_p3124195265717"></a>Go language environment</p>
</td>
<td class="cellrowborder" valign="top" width="70%" headers="mcps1.2.3.1.2 "><p id="en-us_topic_0269670261_p012435218578"><a name="en-us_topic_0269670261_p012435218578"></a><a name="en-us_topic_0269670261_p012435218578"></a>Run the <strong id="en-us_topic_0269670261_b432482213508"><a name="en-us_topic_0269670261_b432482213508"></a><a name="en-us_topic_0269670261_b432482213508"></a>go version</strong> command to check that the version is 1.14.3 or later.</p>
</td>
</tr>
<tr id="en-us_topic_0269670261_row05615595485"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.3.1.1 "><p id="en-us_topic_0269670261_p2124252115719"><a name="en-us_topic_0269670261_p2124252115719"></a><a name="en-us_topic_0269670261_p2124252115719"></a>gcc version</p>
</td>
<td class="cellrowborder" valign="top" width="70%" headers="mcps1.2.3.1.2 "><p id="en-us_topic_0269670261_p512445215576"><a name="en-us_topic_0269670261_p512445215576"></a><a name="en-us_topic_0269670261_p512445215576"></a>Run the <strong id="en-us_topic_0269670261_b81442049153212"><a name="en-us_topic_0269670261_b81442049153212"></a><a name="en-us_topic_0269670261_b81442049153212"></a>gcc --version</strong> command to check that the version is 7.3.0 or later.</p>
</td>
</tr>
<tr id="en-us_topic_0269670261_row11826547124816"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.3.1.1 "><p id="en-us_topic_0269670261_p151241522577"><a name="en-us_topic_0269670261_p151241522577"></a><a name="en-us_topic_0269670261_p151241522577"></a>Kubernetes version</p>
</td>
<td class="cellrowborder" valign="top" width="70%" headers="mcps1.2.3.1.2 "><p id="en-us_topic_0269670261_p1124115285720"><a name="en-us_topic_0269670261_p1124115285720"></a><a name="en-us_topic_0269670261_p1124115285720"></a>Run the <strong id="en-us_topic_0269670261_b75481128165112"><a name="en-us_topic_0269670261_b75481128165112"></a><a name="en-us_topic_0269670261_b75481128165112"></a>kubectl version</strong> command to check that the version is 1.13.0 or later.</p>
</td>
</tr>
<tr id="en-us_topic_0269670261_en-us_topic_0252788324_row11244529577"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.3.1.1 "><p id="en-us_topic_0269670261_en-us_topic_0252788324_p16191917113619"><a name="en-us_topic_0269670261_en-us_topic_0252788324_p16191917113619"></a><a name="en-us_topic_0269670261_en-us_topic_0252788324_p16191917113619"></a>Docker environment</p>
</td>
<td class="cellrowborder" valign="top" width="70%" headers="mcps1.2.3.1.2 "><p id="en-us_topic_0269670261_en-us_topic_0252788324_p461711733616"><a name="en-us_topic_0269670261_en-us_topic_0252788324_p461711733616"></a><a name="en-us_topic_0269670261_en-us_topic_0252788324_p461711733616"></a>Run the <strong id="en-us_topic_0269670261_b2079634755111"><a name="en-us_topic_0269670261_b2079634755111"></a><a name="en-us_topic_0269670261_b2079634755111"></a>docker info</strong> command to check that Docker has been installed.</p>
</td>
</tr>
<tr id="en-us_topic_0269670261_row34271613113113"><td class="cellrowborder" valign="top" width="30%" headers="mcps1.2.3.1.1 "><p id="en-us_topic_0269670261_p1942971303117"><a name="en-us_topic_0269670261_p1942971303117"></a><a name="en-us_topic_0269670261_p1942971303117"></a>root user permission</p>
</td>
<td class="cellrowborder" valign="top" width="70%" headers="mcps1.2.3.1.2 "><p id="en-us_topic_0269670261_p8429113133117"><a name="en-us_topic_0269670261_p8429113133117"></a><a name="en-us_topic_0269670261_p8429113133117"></a>Check that the root user permission of the BMS is available.</p>
</td>
</tr>
</tbody>
</table>

<h2 id="directory-structure.md">Directory Structure</h2>

```
├── build                                             # Compilation scripts
│   ├── build_310.sh
│   ├── build_910.sh
│   ├── build_common.sh
│   ├── build_in_docker.sh
│   ├── build.sh
│   ├── deploy.sh
│   └── sample_check.sh
├── output                                           # Compilation result directory.
├── src                                              # Source code directory.
│   └── plugin
│   │    ├── cmd/ascendplugin
│   │    │   └── ascend_plugin.go
│   │    ├── config
│   │    │   ├── config_310
│   │    │   │   └── ascend_device_plugin.pc
│   │    │   └── config_910
│   │    │       └── ascend_device_plugin.pc
│   │    └── pkg/npu/huawei
├── test                                             # Test directory.
├── Dockerfile                                       # Image file.
├── LICENSE                                          
├── Open Source Software Notice.md                   
├── README.zh.md
├── ascend.yaml                                      # YAML file of the sample running task 
├── ascendplugin-310.yaml                            # YAML file for deploying the inference card
├── ascendplugin-volcano.yaml                        # YAML file for implementing affinity scheduling and deployment with Volcano.
├──ascendplugin.yaml                                 # YAML file for deploying the inference card
├── docker_run.sh                                    # Docker running command
├── go.mod                                           
└── go.sum                                           
```

<h2 id="version-updates.md">Version Updates</h2>

<a name="table7854542104414"></a>
<table><thead align="left"><tr id="row785512423445"><th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.1.4.1.1"><p id="p19856144274419"><a name="p19856144274419"></a><a name="p19856144274419"></a>Version</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.1.4.1.2"><p id="p3856134219446"><a name="p3856134219446"></a><a name="p3856134219446"></a>Date</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.1.4.1.3"><p id="p585634218445"><a name="p585634218445"></a><a name="p585634218445"></a>Description</p>
</th>
</tr>
</thead>
<tbody><tr id="row118567425441"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.1 "><p id="p08571442174415"><a name="p08571442174415"></a><a name="p08571442174415"></a>V20.1.0</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.2 "><p id="p38571542154414"><a name="p38571542154414"></a><a name="p38571542154414"></a>2020-09-30</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.3 "><p id="p5857142154415"><a name="p5857142154415"></a><a name="p5857142154415"></a>This issue is the first official release.</p>
</td>
</tr>
</tbody>
</table>

