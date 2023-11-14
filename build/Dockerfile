FROM ubuntu:18.04

RUN useradd -d /home/HwHiAiUser -u 1000 -m -s /usr/sbin/nologin HwHiAiUser && \
    useradd -d /home/hwMindX -u 9000 -m -s /usr/sbin/nologin hwMindX &&\
    usermod root -s /usr/sbin/nologin

ENV LD_LIBRARY_PATH /usr/local/Ascend/driver/lib64:/usr/local/Ascend/driver/lib64/driver:/usr/local/Ascend/driver/lib64/common

COPY ./device-plugin /usr/local/bin/
COPY ./faultCode.json /usr/local/
RUN chmod 550 /usr/local/bin/device-plugin &&\
    chmod 550 /usr/local/bin &&\
    chmod 440 /usr/local/faultCode.json &&\
    chmod 750 /home/HwHiAiUser &&\
    chmod 750 /home/hwMindX &&\
    echo 'umask 027' >> /etc/profile &&\
    echo 'source /etc/profile' >> ~/.bashrc