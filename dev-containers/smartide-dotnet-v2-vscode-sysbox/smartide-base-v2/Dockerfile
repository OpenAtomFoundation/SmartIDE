#################################################
# SmartIDE Developer Container Image
# Licensed under GPL v3.0
# Copyright (C) leansoftX.com
#################################################

FROM ubuntu:bionic


ENV DEBIAN_FRONTEND=noninteractive
ENV TZ Asia/Shanghai
#git中文乱码问题
ENV LESSCHARSET=utf-8

# sshd
RUN mkdir /var/run/sshd && \
    apt-get update && \
    apt-get -y install --no-install-recommends systemd net-tools openssh-server curl git wget sudo ca-certificates gosu && \
    apt-get clean && \
    apt-get autoremove -y && \
    rm -rf /tmp/* /var/lib/apt/lists/* /var/tmp/*


RUN sed -i "s/UsePrivilegeSeparation.*/UsePrivilegeSeparation no/g" /etc/ssh/sshd_config && \
	sed -i "s/UsePAM.*/UsePAM no/g" /etc/ssh/sshd_config && \
	sed -i "s/#PermitRootLogin.*/PermitRootLogin yes/g" /etc/ssh/sshd_config && \
    sed -i "s/AllowTcpForwarding.*/AllowTcpForwarding yes/g" /etc/ssh/sshd_config && \
    sed -i "1i\export LESSCHARSET=utf-8" /etc/profile


ENV USERNAME=smartide
ARG USER_UID=1000
ARG USER_GID=1000

RUN groupadd -g $USER_GID $USERNAME \
    && useradd -rm -d /home/$USERNAME -s /bin/bash -u $USER_UID -g $USER_GID $USERNAME \
    && echo $USERNAME ALL=\(root\) NOPASSWD:ALL > /etc/sudoers.d/$USERNAME \
    && chmod 0440 /etc/sudoers.d/$USERNAME \
    && chmod g+rw /home \
    && mkdir -p /home/project \
    && mkdir -p /home/opvscode \
    && mkdir -p /idesh

ENV HOME=/home/$USERNAME

EXPOSE 22
EXPOSE 3000
EXPOSE 8887

COPY entrypoint_base.sh /idesh/entrypoint_base.sh
RUN chmod +x /idesh/entrypoint_base.sh

ENTRYPOINT ["/idesh/entrypoint_base.sh"]
