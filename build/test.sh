#!/bin/bash
# Perform  test ascend-device-plugin
# Copyright(C) Huawei Technologies Co.,Ltd. 2020-2022. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
# ============================================================================

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
