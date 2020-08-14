FROM golang:1.13.11-buster as build

ENV USE_ASCEND_DOCKER true

ENV GOPATH /usr/app/

ENV GO111MODULE off

ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

ENV LD_LIBRARY_PATH  /usr/local/Ascend/driver/lib64/driver:/usr/local/Ascend/driver/lib64/common

ENV  LD_LIBRARY_PATH $LD_LIBRARY_PATH:/usr/local/Ascend/driver/lib64/

RUN mkdir -p /usr/app/src/ascend-device-plugin

COPY . /usr/app/src/Ascend-device-plugin

WORKDIR /usr/app/src/Ascend-device-plugin