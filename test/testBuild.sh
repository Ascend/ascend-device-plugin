#!/bin/sh
set -x
CUR_DIR=$(dirname $(readlink -f $0))
TOP_DIR=$(realpath "${CUR_DIR}"/..)
export GO111MODULE="on"
export PATH=$GOPATH/bin:$PATH
CONFIGDIR=${TOP_DIR}/src/plugin/config/config_310

DRIVER_FILE="310driver"
DOWN_DRIVER_FILE="platform/Tuscany"
PC_File="ascend_device_plugin.pc"
SODIR=${TOP_DIR}/${DRIVER_FILE}/driver/lib64/
export LD_LIBRARY_PATH=${SODIR}:${LD_LIBRARY_PATH}
export PKG_CONFIG_PATH=${CONFIGDIR}:$PKG_CONFIG_PATH
ls ${TOP_DIR}/${DOWN_DRIVER_FILE}
plateform=$(arch)
chmod 550 ${TOP_DIR}/${DOWN_DRIVER_FILE}/Ascend310-driver-*.${plateform}.run

mkdir -p /var/lib/kubelet/device-plugins
${TOP_DIR}/${DOWN_DRIVER_FILE}/Ascend310-driver-*${osname}*.${plateform}*.run \
--noexec --extract=${TOP_DIR}/${DRIVER_FILE}
sed -i "/^prefix=/c prefix=${TOP_DIR}/${DRIVER_FILE}" ${CONFIGDIR}/${PC_File}
ldd ${SODIR}/libdrvdsmi_host.so
go get github.com/golang/mock/mockgen
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

file_input='testDeviceplugin.txt'
file_detail_output='DevicepluginCoverageReport.html'

echo "************************************* Start LLT Test *************************************"
mkdir -p "${TOP_DIR}"/test/
cd "${TOP_DIR}"/test/
rm -rf $file_detail_output $file_input

go test -v -race -coverprofile cov.out ${TOP_DIR}/src/plugin/pkg/npu/huawei/ >./$file_input

if [ $? != 0 ]; then
  echo '****** go test cases error! ******'
  echo 'Failed' >$file_input
else
  echo ${file_detail_output}
  gocov convert cov.out | gocov-html >${file_detail_output}
fi

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