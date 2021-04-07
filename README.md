# Ascend Device Plugin.en
-   [Ascend Device Plugin](#ascend-device-plugin.md)
    -   [Description](#description.md)
    -   [Compiling the Ascend Device Plugin](#compiling-the-ascend-device-plugin.md)
    -   [Creating DaemonSet](#creating-daemonset.md)
    -   [Creating a Service Container](#creating-a-service-container.md)
-   [Environment Dependencies](#environment-dependencies.md)
-   [Directory Structure](#directory-structure.md)
-   [Version Updates](#version-updates.md)
<h2 id="ascend-device-plugin.md">Ascend Device Plugin</h2>

-   **[Description](#description.md)**  

-   **[Compiling the Ascend Device Plugin](#compiling-the-ascend-device-plugin.md)**  

-   **[Creating DaemonSet](#creating-daemonset.md)**  

-   **[Creating a Service Container](#creating-a-service-container.md)**  


<h2 id="description.md">Description</h2>

The device management plug-in provides the following functions:

-   Device discovery: The number of discovered devices can be obtained from the Ascend device driver and reported to the Kubernetes system.
-   Health check: The health status of Ascend devices can be detected. When a device is unhealthy, the device is reported to the Kubernetes system and is removed.
-   Device allocation: Ascend devices can be allocated in the Kubernetes system.

<h2 id="compiling-the-ascend-device-plugin.md">Compiling the Ascend Device Plugin</h2>

## Procedure<a name="section112101632152317"></a>

1.  Set environment variables.

    **export GO111MODULE=on**

    **export GOPROXY=**_Proxy address_

    **export GONOSUMDB=\***

    >![](figures/icon-note.gif) **NOTE:** 
    >-   Use the actual GOPROXY proxy address. You can run the  **go mod download**  command in the  **ascend-device-plugin**  directory to check the address.
    >-   If no error information is displayed, the proxy is set successfully.

2.  Go to the  **ascend-device-plugin**  directory and run the following command to modify the YAML file:
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
                - key: huawei.com/Ascend910 #Resource name. Set the value based on the processor type.
                  operator: Exists
                  effect: NoSchedule
                - key: "ascendplugin"
                  operator: "Equal"
                  value: "v2"
                  effect: NoSchedule
              priorityClassName: "system-node-critical"
              nodeSelector:
                accelerator: huawei-Ascend910 #Set the label name based on the processor type.
              containers:
              - image: ascend-device-plugin:v1.0.1  #Image name and version
                name: device-plugin-01
                resources:
                  requests:
                    memory: 500Mi
                    cpu: 500m
                  limits:
                    memory: 500Mi
                    cpu: 500m
                command: [ "/bin/bash", "-c", "--"]
                args: [ "ascendplugin  --useAscendDocker=${USE_ASCEND_DOCKER}" ] 
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

    -   YAML file of MindX DL

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
              - image: ascend-k8sdeviceplugin:v0.0.1   #Image name and version
                name: device-plugin-01
                resources:
                  requests:
                    memory: 500Mi
                    cpu: 500m
                  limits:
                    memory: 500Mi
                    cpu: 500m
                command: [ "/bin/bash", "-c", "--"]
                args: [ "ascendplugin  --useAscendDocker=${USE_ASCEND_DOCKER} --volcanoType=true" ] 
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


3.  Run the following command to edit the  **Dockerfile**  file and change the image name and version to the obtained values:

    **vi** _/home/test/_ascend-device-plugin**/Dockerfile**

    ```
    #Select the basic image as required. You can run the docker images command to query the basic image.
    FROM ubuntu:18.04 as build
    #Specify whether to use Ascend Docker. The default value is true. Change it to false.
    ENV USE_ASCEND_DOCKER true
    
    ENV LD_LIBRARY_PATH  /usr/local/Ascend/driver/lib64/driver:/usr/local/Ascend/driver/lib64/common
    
    ENV  LD_LIBRARY_PATH $LD_LIBRARY_PATH:/usr/local/Ascend/driver/lib64/
    
    COPY ./output/ascendplugin /usr/local/bin/
    
    ```

4.  Run the following commands to generate a binary file and image file \(use the actual script name\):

    **cd** _/home/test/_ascend-device-plugin**/build**/

    **chmod +x build.sh**

    **dos2unix build.sh**

    **./build.sh dockerimages**

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
    -rw-r--r-- 1 root root  6721073 Jun  9 16:20 Ascend-K8sDevicePlugin-xxx-arm64-Linux.tar.gz
    ```


<h2 id="creating-daemonset.md">Creating DaemonSet</h2>

## Procedure<a name="en-us_topic_0269670254_section2036324211563"></a>

>![](figures/icon-note.gif) **NOTE:** 
>The following uses the tar.gz file generated on the ARM platform as an example.

1.  Run the following command to check whether the Docker software package is successfully imported:

    **docker images**

    -   If yes, go to  [3](#en-us_topic_0269670254_li26268471380).
    -   If no, perform  [2](#en-us_topic_0269670254_li1372334715567)  to import the file again.

2.  <a name="en-us_topic_0269670254_li1372334715567"></a>Go to the directory where the Docker software package is stored and run the following command to import the Docker image.

    **cd** _/home/test/_**ascend-device-plugin/output**

    **docker load** **-i** _Ascend-K8sDevicePlugin-xxx-arm64-Docker.tar.gz_

3.  <a name="en-us_topic_0269670254_li26268471380"></a>Run the following command to label the node with Ascend 910 or  Ascend 310:

    **kubectl label nodes** _localhost.localdomain_ **accelerator=**_huawei-Ascend910_

    **localhost.localdomain**  is the name of the node with Ascend 910 or  Ascend 310. You can run the  **kubectl get node**  command to view the node name.

    The label name must be the same as the  **nodeSelector**  label name in the YAML file in "Compiling the Ascend Device Plugin."

    >![](figures/icon-note.gif) **NOTE:** 
    >If the K8s plugin needs to be deployed on a new node, perform  [2](#en-us_topic_0269670254_li1372334715567)  to  [3](#en-us_topic_0269670254_li26268471380).

4.  Run the following commands to deploy DaemonSet:

    **cd** _/home/test/_**ascend-device-plugin**

    **kubectl apply -f  ascendplugin.yaml**

    >![](figures/icon-note.gif) **NOTE:** 
    >To view the node deployment information, you need to wait for several minutes after the deployment is complete.

5.  Run the following command to view the node device deployment information:

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

    **cd** _/home/test/_**ascend-device-plugin**

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

2.  Run the following command to create a pod:

    **kubectl apply -f ascend.yaml**

    >![](figures/icon-note.gif) **NOTE:** 
    >To delete the pod, run the following command:
    >**kubectl delete -f** **ascend.yaml**

3.  Run the following commands to access the pod and view the allocation information:

    **kubectl exec -it** _Pod name_ **bash**

    The pod name is the one configured in  [1](#en-us_topic_0269670251_en-us_topic_0249483204_li104071617503).

    **ls /dev/**

    In the command output similar to the following,  **davinci3**  and  **davinci4**  are the allocated pods.

    ```
    core davinci3 davinci4 davinci_manager devmm_svm fd full hisi_hdc mqueue null ptmx
    ```


<h2 id="environment-dependencies.md">Environment Dependencies</h2>

**Table  1**  Environment Dependencies

<a name="en-us_topic_0252788324_table171211952105718"></a>
<table><thead align="left"><tr id="en-us_topic_0269670261_en-us_topic_0252788324_row51223524573"><th class="cellrowborder" valign="top" width="48%" id="mcps1.2.3.1.1"><p id="en-us_topic_0269670261_en-us_topic_0252788324_p15122175218576"><a name="en-us_topic_0269670261_en-us_topic_0252788324_p15122175218576"></a><a name="en-us_topic_0269670261_en-us_topic_0252788324_p15122175218576"></a>Check Item</p>
</th>
<th class="cellrowborder" valign="top" width="52%" id="mcps1.2.3.1.2"><p id="en-us_topic_0269670261_en-us_topic_0252788324_p1712211526578"><a name="en-us_topic_0269670261_en-us_topic_0252788324_p1712211526578"></a><a name="en-us_topic_0269670261_en-us_topic_0252788324_p1712211526578"></a>Requirement</p>
</th>
</tr>
</thead>
<tbody><tr id="en-us_topic_0269670261_row1985835314489"><td class="cellrowborder" valign="top" width="48%" headers="mcps1.2.3.1.1 "><p id="en-us_topic_0269670261_p1925915619412"><a name="en-us_topic_0269670261_p1925915619412"></a><a name="en-us_topic_0269670261_p1925915619412"></a>dos2unix</p>
</td>
<td class="cellrowborder" valign="top" width="52%" headers="mcps1.2.3.1.2 "><p id="en-us_topic_0269670261_p1025985634111"><a name="en-us_topic_0269670261_p1025985634111"></a><a name="en-us_topic_0269670261_p1025985634111"></a>Run the <strong id="b9360144817168"><a name="b9360144817168"></a><a name="b9360144817168"></a>dos2unix --version</strong> command to check that the software has been installed. There is no requirement on the version.</p>
</td>
</tr>
<tr id="en-us_topic_0269670261_row16906451114817"><td class="cellrowborder" valign="top" width="48%" headers="mcps1.2.3.1.1 "><p id="en-us_topic_0269670261_p212295212575"><a name="en-us_topic_0269670261_p212295212575"></a><a name="en-us_topic_0269670261_p212295212575"></a>Driver version of the RUN package</p>
</td>
<td class="cellrowborder" valign="top" width="52%" headers="mcps1.2.3.1.2 "><p id="en-us_topic_0269670261_p31997012111"><a name="en-us_topic_0269670261_p31997012111"></a><a name="en-us_topic_0269670261_p31997012111"></a>Go to the directory of the driver (for example, <strong id="b950512535169"><a name="b950512535169"></a><a name="b950512535169"></a>/usr/local/Ascend/driver</strong>) and run the <strong id="b155111953101616"><a name="b155111953101616"></a><a name="b155111953101616"></a>cat version.info</strong> command to confirm that the driver version is 1.73 or later.</p>
</td>
</tr>
<tr id="en-us_topic_0269670261_row12226135012483"><td class="cellrowborder" valign="top" width="48%" headers="mcps1.2.3.1.1 "><p id="en-us_topic_0269670261_p3124195265717"><a name="en-us_topic_0269670261_p3124195265717"></a><a name="en-us_topic_0269670261_p3124195265717"></a>Go language environment</p>
</td>
<td class="cellrowborder" valign="top" width="52%" headers="mcps1.2.3.1.2 "><p id="en-us_topic_0269670261_p012435218578"><a name="en-us_topic_0269670261_p012435218578"></a><a name="en-us_topic_0269670261_p012435218578"></a>Run the <strong id="b13444656111615"><a name="b13444656111615"></a><a name="b13444656111615"></a>go version</strong> command to confirm that the version is 1.14.3 or later.</p>
</td>
</tr>
<tr id="en-us_topic_0269670261_row05615595485"><td class="cellrowborder" valign="top" width="48%" headers="mcps1.2.3.1.1 "><p id="en-us_topic_0269670261_p2124252115719"><a name="en-us_topic_0269670261_p2124252115719"></a><a name="en-us_topic_0269670261_p2124252115719"></a>gcc version</p>
</td>
<td class="cellrowborder" valign="top" width="52%" headers="mcps1.2.3.1.2 "><p id="en-us_topic_0269670261_p512445215576"><a name="en-us_topic_0269670261_p512445215576"></a><a name="en-us_topic_0269670261_p512445215576"></a>Run the <strong id="b254345881615"><a name="b254345881615"></a><a name="b254345881615"></a>gcc --version</strong> command to confirm that the version is 7.3.0 or later.</p>
</td>
</tr>
<tr id="en-us_topic_0269670261_row11826547124816"><td class="cellrowborder" valign="top" width="48%" headers="mcps1.2.3.1.1 "><p id="en-us_topic_0269670261_p151241522577"><a name="en-us_topic_0269670261_p151241522577"></a><a name="en-us_topic_0269670261_p151241522577"></a>Kubernetes version</p>
</td>
<td class="cellrowborder" valign="top" width="52%" headers="mcps1.2.3.1.2 "><p id="en-us_topic_0269670261_p89141115124714"><a name="en-us_topic_0269670261_p89141115124714"></a><a name="en-us_topic_0269670261_p89141115124714"></a>1.17.<em id="i19261912181713"><a name="i19261912181713"></a><a name="i19261912181713"></a>x</em>. Select the latest bugfix version.</p>
<p id="en-us_topic_0269670261_p1124115285720"><a name="en-us_topic_0269670261_p1124115285720"></a><a name="en-us_topic_0269670261_p1124115285720"></a>You can run the <strong id="b10594172511175"><a name="b10594172511175"></a><a name="b10594172511175"></a>kubectl version</strong> command to view the version.</p>
</td>
</tr>
<tr id="en-us_topic_0269670261_en-us_topic_0252788324_row11244529577"><td class="cellrowborder" valign="top" width="48%" headers="mcps1.2.3.1.1 "><p id="en-us_topic_0269670261_en-us_topic_0252788324_p16191917113619"><a name="en-us_topic_0269670261_en-us_topic_0252788324_p16191917113619"></a><a name="en-us_topic_0269670261_en-us_topic_0252788324_p16191917113619"></a>Docker environment</p>
</td>
<td class="cellrowborder" valign="top" width="52%" headers="mcps1.2.3.1.2 "><p id="en-us_topic_0269670261_en-us_topic_0252788324_p461711733616"><a name="en-us_topic_0269670261_en-us_topic_0252788324_p461711733616"></a><a name="en-us_topic_0269670261_en-us_topic_0252788324_p461711733616"></a>Run the <strong id="b18504438161717"><a name="b18504438161717"></a><a name="b18504438161717"></a>docker info</strong> command to confirm that Docker has been installed.</p>
</td>
</tr>
<tr id="en-us_topic_0269670261_row34271613113113"><td class="cellrowborder" valign="top" width="48%" headers="mcps1.2.3.1.1 "><p id="en-us_topic_0269670261_p1942971303117"><a name="en-us_topic_0269670261_p1942971303117"></a><a name="en-us_topic_0269670261_p1942971303117"></a>root user permission</p>
</td>
<td class="cellrowborder" valign="top" width="52%" headers="mcps1.2.3.1.2 "><p id="en-us_topic_0269670261_p8429113133117"><a name="en-us_topic_0269670261_p8429113133117"></a><a name="en-us_topic_0269670261_p8429113133117"></a>Check that the root user permission of the BMS is available.</p>
</td>
</tr>
</tbody>
</table>

<h2 id="directory-structure.md">Directory Structure</h2>

```
├── build                                             # Compilation scripts
│   └── build.sh
├── output                                           # Compilation result directory.
├── src                                              # Source code directory.
│   └── plugin
│   │    ├── cmd/ascendplugin
│   │    │   └── ascend_plugin.go    
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
<table><thead align="left"><tr id="en-us_topic_0280467800_row785512423445"><th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.1.4.1.1"><p id="en-us_topic_0280467800_p19856144274419"><a name="en-us_topic_0280467800_p19856144274419"></a><a name="en-us_topic_0280467800_p19856144274419"></a>Version</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.1.4.1.2"><p id="en-us_topic_0280467800_p3856134219446"><a name="en-us_topic_0280467800_p3856134219446"></a><a name="en-us_topic_0280467800_p3856134219446"></a>Date</p>
</th>
<th class="cellrowborder" valign="top" width="33.33333333333333%" id="mcps1.1.4.1.3"><p id="en-us_topic_0280467800_p585634218445"><a name="en-us_topic_0280467800_p585634218445"></a><a name="en-us_topic_0280467800_p585634218445"></a>Description</p>
</th>
</tr>
</thead>
<tbody><tr id="row137501013384"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.1 "><p id="p137501613585"><a name="p137501613585"></a><a name="p137501613585"></a>v20.2.0</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.2 "><p id="p1675010135811"><a name="p1675010135811"></a><a name="p1675010135811"></a>2021-01-08</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.3 "><p id="p3750813889"><a name="p3750813889"></a><a name="p3750813889"></a>Optimized the description in "Creating DaemonSet."</p>
</td>
</tr>
<tr id="en-us_topic_0280467800_row118567425441"><td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.1 "><p id="en-us_topic_0280467800_p08571442174415"><a name="en-us_topic_0280467800_p08571442174415"></a><a name="en-us_topic_0280467800_p08571442174415"></a>v20.2.0</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.2 "><p id="en-us_topic_0280467800_p38571542154414"><a name="en-us_topic_0280467800_p38571542154414"></a><a name="en-us_topic_0280467800_p38571542154414"></a>2020-11-18</p>
</td>
<td class="cellrowborder" valign="top" width="33.33333333333333%" headers="mcps1.1.4.1.3 "><p id="en-us_topic_0280467800_p5857142154415"><a name="en-us_topic_0280467800_p5857142154415"></a><a name="en-us_topic_0280467800_p5857142154415"></a>This is the first official release.</p>
</td>
</tr>
</tbody>
</table>
