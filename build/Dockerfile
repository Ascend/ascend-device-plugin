FROM golang:1.13.11-buster as build

ENV GOPATH /home/gopath

ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

ENV LD_LIBRARY_PATH  /usr/local/Ascend/driver/lib64/driver:/usr/local/Ascend/driver/lib64/common

RUN rm -f /usr/local/bin/ascend*

RUN mkdir -p /home/gopath

RUN rm -rf /home/miniD

RUN ln -sf /usr/share/zoneinfo/Asia/Shanghai/etc/localtime

RUN cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

COPY ./ascendplugin  /usr/local/bin/


