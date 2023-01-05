###########################################################################
# SmartIDE - Dev Containers
# Copyright (C) 2023 leansoftX.com

# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# any later version.

# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.

# You should have received a copy of the GNU General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.
###########################################################################

FROM --platform=$TARGETPLATFORM ubuntu:20.04


ENV DEBIAN_FRONTEND=noninteractive
ENV TZ Asia/Shanghai
#git中文乱码问题
ENV LESSCHARSET=utf-8

# sshd
RUN mkdir /var/run/sshd && \
    apt-get update && \
    apt-get -y install --no-install-recommends supervisor net-tools openssh-server curl git wget sudo ca-certificates gosu libsecret-1-dev zip apt-utils vim iputils-ping telnet jq moreutils lsof strace && \
    apt-get clean && \
    apt-get autoremove -y && \
    rm -rf /tmp/* /var/lib/apt/lists/* /var/tmp/*

RUN chmod 777 /var/run && \
    chmod 777 /etc/supervisor

RUN sed -i "s/UsePrivilegeSeparation.*/UsePrivilegeSeparation no/g" /etc/ssh/sshd_config && \
	sed -i "s/UsePAM.*/UsePAM no/g" /etc/ssh/sshd_config && \
	sed -i "s/#PermitRootLogin.*/PermitRootLogin yes/g" /etc/ssh/sshd_config && \
    sed -i "s/AllowTcpForwarding.*/AllowTcpForwarding yes/g" /etc/ssh/sshd_config && \
    sed -i "1i\export LESSCHARSET=utf-8" /etc/profile && \
    sed -i 's/#AllowTcpForwarding yes/AllowTcpForwarding yes/g' /etc/ssh/sshd_config && \
    sed -i 's/#GatewayPorts no/GatewayPorts yes/g' /etc/ssh/sshd_config


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
