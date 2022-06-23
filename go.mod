module Ascend-device-plugin

go 1.16

require (
	github.com/agiledragon/gomonkey/v2 v2.2.0
	github.com/fsnotify/fsnotify v1.4.9
	github.com/golang/mock v1.6.0
	github.com/smartystreets/goconvey v1.6.4
	go.uber.org/atomic v1.7.0
	golang.org/x/net v0.0.0-20211112202133-69e39bad7dc2
	google.golang.org/grpc v1.28.0
	huawei.com/npu-exporter v0.2.8
	k8s.io/api v0.19.4
	k8s.io/apimachinery v0.19.4
	k8s.io/client-go v0.19.4
	k8s.io/kubelet v0.19.4
	k8s.io/kubernetes v1.19.4
)

replace (
	github.com/containernetworking/cni => github.com/containernetworking/cni v0.8.1
	github.com/gorilla/websocket => github.com/gorilla/websocket v1.4.2
	go.uber.org/atomic => go.uber.org/atomic v1.6.0
	golang.org/x/net v0.0.0-20210405180319-a5a99cb37ef4 => golang.org/x/net v0.0.0-20210226172049-e18ecbb05110
	huawei.com/kmc => codehub-dg-y.huawei.com/it-edge-native/edge-native-core/coastguard.git v1.0.6
	huawei.com/npu-exporter => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/npu-exporter.git v0.2.8
	k8s.io/api => codehub-dg-y.huawei.com/OpenSourceCenter/kubernetes.git/staging/src/k8s.io/api v1.19.4-h4
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.19.4
	k8s.io/apimachinery => codehub-dg-y.huawei.com/OpenSourceCenter/kubernetes.git/staging/src/k8s.io/apimachinery v1.19.4-h4
	k8s.io/apiserver => codehub-dg-y.huawei.com/OpenSourceCenter/kubernetes.git/staging/src/k8s.io/apiserver v1.19.4-h4
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.19.4
	k8s.io/client-go => codehub-dg-y.huawei.com/OpenSourceCenter/kubernetes.git/staging/src/k8s.io/client-go v1.19.4-h4
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.19.4
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.19.4
	k8s.io/code-generator => k8s.io/code-generator v0.19.4
	k8s.io/component-base => k8s.io/component-base v0.19.4
	k8s.io/cri-api => k8s.io/cri-api v0.19.4
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.19.4
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.19.4
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.19.4
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.19.4
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.19.4
	k8s.io/kubectl => k8s.io/kubectl v0.19.4
	k8s.io/kubelet => codehub-dg-y.huawei.com/OpenSourceCenter/kubernetes.git/staging/src/k8s.io/kubelet v1.19.4-h4
	k8s.io/kubernetes => codehub-dg-y.huawei.com/OpenSourceCenter/kubernetes.git v1.19.4-h4
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.19.4
	k8s.io/metrics => k8s.io/metrics v0.19.4
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.19.4
)
