#!/bin/bash
# Copyright @ Huawei Technologies CO., Ltd. 2020-2021. All rights reserved
set -e
CUR_DIR=$(dirname $(readlink -f $0))
TOP_DIR=$(realpath "${CUR_DIR}"/..)
export GO111MODULE="on"
export PATH=$GOPATH/bin:$PATH

go get github.com/golang/mock/mockgen
go get golang.org/x/net
go get golang.org/x/term
go get golang.org/x/text
go get github.com/golang/protobuf/ptypes/empty@v1.3.2
MOCK_TOP=${TOP_DIR}/src/plugin/pkg/npu/huawei
mkdir -p "${MOCK_TOP}/mock_v1"
mkdir -p "${MOCK_TOP}/mock_kubernetes"
mkdir -p "${MOCK_TOP}/mock_kubelet_v1beta1"

mockgen k8s.io/client-go/kubernetes/typed/core/v1 CoreV1Interface >${MOCK_TOP}/mock_v1/corev1_mock.go
mockgen k8s.io/client-go/kubernetes Interface >${MOCK_TOP}/mock_kubernetes/k8s_interface_mock.go
mockgen k8s.io/client-go/kubernetes/typed/core/v1 NodeInterface >${MOCK_TOP}/mock_v1/node_interface_mock.go
mockgen k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1 DevicePlugin_ListAndWatchServer >${MOCK_TOP}/mock_kubelet_v1beta1/deviceplugin_mock.go
mockgen k8s.io/client-go/kubernetes/typed/core/v1 PodInterface >${MOCK_TOP}/mock_v1/pod_interface_mock.go

export PKG_CONFIG_PATH=${TOP_DIR}/src/plugin/config/config_310/:$PKG_CONFIG_PATH

function execute_test() {
  if ! (go test -v -race -coverprofile cov.out ${TOP_DIR}/src/plugin/pkg/npu/huawei/ >./$file_input); then
    echo '****** go test cases error! ******'
    echo 'Failed' >$file_input
    exit 1
  else
    echo ${file_detail_output}
    gocov convert cov.out | gocov-html >${file_detail_output}
    gotestsum --junitfile unit-tests.xml "${TOP_DIR}"/src/plugin/pkg/npu/huawei/...
  fi
}

file_input='testDevicePlugin.txt'
file_detail_output='api.html'

echo "************************************* Start LLT Test *************************************"
mkdir -p "${TOP_DIR}"/test/
cd "${TOP_DIR}"/test/
if [ -f "$file_detail_output" ]; then
  rm -rf $file_detail_output
fi
if [ -f "$file_input" ]; then
  rm -rf $file_input
fi
execute_test
echo "<html<body><h1>==================================================</h1><table border="2">" >>./$file_detail_output
echo "<html<body><h1>DevicePlugin testCase</h1><table border="1">" >>./$file_detail_output
echo "<html<body><h1>==================================================</h1><table border="2">" >>./$file_detail_output
while read line; do
  echo -e "<tr>
   $(echo $line | awk 'BEGIN{FS="|"}''{i=1;while(i<=NF) {print "<td>"$i"</td>";i++}}')
  </tr>" >>$file_detail_output
done <$file_input
echo "</table></body></html>" >>./$file_detail_output

echo "************************************* End   LLT Test *************************************"

rm -rf ${MOCK_TOP}/mock_v1
rm -rf ${MOCK_TOP}/mock_kubernetes
rm -rf ${MOCK_TOP}/mock_kubelet_v1beta1
exit 0

}