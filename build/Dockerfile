FROM ubuntu:18.04 as build

ENV USE_ASCEND_DOCKER true

ENV GOPATH /home/gopath

ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

ENV LD_LIBRARY_PATH  /usr/local/Ascend/driver/lib64/driver:/usr/local/Ascend/driver/lib64/common

RUN rm -f /usr/local/bin/ascend*

COPY ./ascendplugin  /usr/local/bin/


