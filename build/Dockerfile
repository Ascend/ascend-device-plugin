FROM ubuntu:18.04

RUN useradd -d /home/HwHiAiUser -u 1000 -m -s /usr/sbin/nologin HwHiAiUser && \
    usermod root -s /usr/sbin/nologin

ENV USE_ASCEND_DOCKER true

ENV LD_LIBRARY_PATH  /usr/local/Ascend/driver/lib64/driver:/usr/local/Ascend/driver/lib64/common

ENV LD_LIBRARY_PATH $LD_LIBRARY_PATH:/usr/local/Ascend/driver/lib64/:/usr/local/lib

COPY ./device-plugin /usr/local/bin/
COPY ./lib  /usr/local/lib
RUN chmod 550 /usr/local/bin/device-plugin &&\
    chmod 550 /usr/local/bin &&\
    chmod 750 /home/HwHiAiUser &&\
    chmod 550 /usr/local/lib/ &&\
    chmod 500 /usr/local/lib/* &&\
    echo 'umask 027' >> /etc/profile &&\
    echo 'source /etc/profile' >> ~/.bashrc