module Ascend-device-plugin

go 1.16

require (
	github.com/agiledragon/gomonkey/v2 v2.8.0
	github.com/fsnotify/fsnotify v1.5.1
	github.com/smartystreets/goconvey v1.7.2
	google.golang.org/grpc v1.41.0
	huawei.com/mindx/common/hwlog v0.0.0
	huawei.com/mindx/common/k8stool v0.0.0
	huawei.com/mindx/common/utils v0.0.0
	huawei.com/mindx/common/x509 v0.0.0
	huawei.com/npu-exporter v0.2.17
	k8s.io/api v0.19.4
	k8s.io/apimachinery v0.19.4
	k8s.io/client-go v0.19.4
	k8s.io/kubelet v0.19.4
	k8s.io/kubernetes v1.19.4
)

replace (
	github.com/containernetworking/cni => github.com/containernetworking/cni v0.8.1
	github.com/gorilla/websocket => github.com/gorilla/websocket v1.4.2
	huawei.com/mindx/common/cache => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/cache v0.0.1
	huawei.com/mindx/common/hwlog => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/hwlog v0.0.4
	huawei.com/mindx/common/k8stool => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/k8stool v0.0.3
	huawei.com/mindx/common/kmc => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/kmc v0.0.4
	huawei.com/mindx/common/limiter => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/limiter v0.0.4
	huawei.com/mindx/common/rand => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/rand v0.0.1
	huawei.com/mindx/common/terminal => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/terminal v0.0.1
	huawei.com/mindx/common/tls => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/tls v0.0.1
	huawei.com/mindx/common/utils => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/utils v0.0.5
	huawei.com/mindx/common/x509 => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/common-utils.git/x509 v0.0.5
	huawei.com/npu-exporter => codehub-dg-y.huawei.com/MindX_DL/AtlasEnableWarehouse/npu-exporter.git v0.2.17
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
