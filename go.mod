module Ascend-device-plugin

go 1.14

require (
	github.com/fsnotify/fsnotify v1.4.8-0.20191012010759-4bf2d1fec783
	github.com/golang/mock v1.4.4
	github.com/prashantv/gostub v1.0.0
	github.com/stretchr/testify v1.4.0
	go.uber.org/zap v1.14.0
	golang.org/x/net v0.0.0-20191004110552-13f9640d40b9
	google.golang.org/grpc v1.28.1
	gopkg.in/natefinch/lumberjack.v2 v2.0.0
	gotest.tools v2.2.0+incompatible
	k8s.io/api v0.17.8
	k8s.io/apimachinery v0.17.8
	k8s.io/client-go v0.17.8
	k8s.io/kubelet v0.17.8
	k8s.io/kubernetes v1.17.8
)

replace (
	k8s.io/api => k8s.io/api v0.17.8
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.17.8
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.8
	k8s.io/apiserver => k8s.io/apiserver v0.17.8
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.17.8
	k8s.io/client-go => k8s.io/client-go v0.17.8
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.17.8
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.17.8
	k8s.io/code-generator => k8s.io/code-generator v0.17.8
	k8s.io/component-base => k8s.io/component-base v0.17.8
	k8s.io/cri-api => k8s.io/cri-api v0.17.8
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.17.8
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.17.8
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.17.8
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.17.8
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.17.8
	k8s.io/kubectl => k8s.io/kubectl v0.17.8
	k8s.io/kubelet => k8s.io/kubelet v0.17.8
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.17.8
	k8s.io/metrics => k8s.io/metrics v0.17.8
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.17.8
)
