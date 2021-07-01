FROM ubuntu:18.04 as build

RUN useradd -d /home/HwHiAiUser -u 1000 -m -s /usr/sbin/nologin HwHiAiUser

ENV USE_ASCEND_DOCKER true

ENV LD_LIBRARY_PATH  /usr/local/Ascend/driver/lib64/driver:/usr/local/Ascend/driver/lib64/common

ENV LD_LIBRARY_PATH $LD_LIBRARY_PATH:/usr/local/Ascend/driver/lib64/

COPY ./output/deviceplugin /usr/local/bin/

RUN chmod 550 /usr/local/bin/deviceplugin

RUN echo 'umask 027' >> /etc/profile && \
    echo 'source /etc/profile' >> ~/.bashrc