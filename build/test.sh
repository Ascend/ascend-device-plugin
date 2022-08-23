#!/bin/bash
# Copyright(C) Huawei Technologies Co.,Ltd. 2020-2022. All rights reserved.
set -e
CUR_DIR=$(dirname "$(readlink -f $0)")
TOP_DIR=$(realpath "${CUR_DIR}"/..)
export GO111MODULE="on"
export GONOSUMDB="*"
export PATH=$GOPATH/bin:$PATH

function execute_test() {
  if ! (go test  -mod=mod -gcflags=all=-l -v -race -coverprofile cov.out ${TOP_DIR}/pkg/... >./$file_input); then
    cat ./$file_input
    echo '****** go test cases error! ******'
    exit 1
  else
    echo ${file_detail_output}
    gocov convert cov.out | gocov-html >${file_detail_output}
    gotestsum --junitfile unit-tests.xml -- -race -gcflags=all=-l "${TOP_DIR}"/pkg/...
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

exit 0
