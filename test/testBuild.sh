#!/bin/sh
set -x
CUR_DIR=$(dirname $(readlink -f $0))
TOP_DIR=$(realpath ${CUR_DIR}/..)
export PKG_CONFIG_PATH=${TOP_DIR}/src/plugin/config/config_910/:$PKG_CONFIG_PATH

file_input='testDeviceplugin.txt'
file_detail_output='DevicepluginCoverageReport.html'

echo "************************************* Start LLT Test *************************************"
mkdir -p "${TOP_DIR}"/test/
cd "${TOP_DIR}"/test/
rm -rf $file_detail_output $file_input

go test -v -race -coverprofile cov.out ${TOP_DIR}/src/plugin/pkg/npu/huawei/ > ./$file_input

if [ $? != 0 ]
then
  echo '****** go test cases error! ******'
  echo 'Failed' > $file_input
else
  echo ${file_detail_output}
  gocov convert cov.out | gocov-html > ${file_detail_output}
fi

echo "<html<body><h1>==================================================</h1><table border="2">" >> ./$file_detail_output
echo "<html<body><h1>DevicePlugin testCase</h1><table border="1">" >> ./$file_detail_output
echo "<html<body><h1>==================================================</h1><table border="2">" >> ./$file_detail_output
while read line
do
  echo -e "<tr>
   `echo $line | awk 'BEGIN{FS="|"}''{i=1;while(i<=NF) {print "<td>"$i"</td>";i++}}'`
  </tr>" >> $file_detail_output
done < $file_input
echo "</table></body></html>" >> ./$file_detail_output

echo "************************************* End   LLT Test *************************************"