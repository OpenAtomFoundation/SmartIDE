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
    apt-get -y install --no-install-recommends net-tools openssh-server curl wget sudo ca-certificates && \
    apt-get clean && \
    apt-get autoremove -y && \
    rm -rf /tmp/* /var/lib/apt/lists/* /var/tmp/*

# 修改sshd配置
# 参考：https://www.jianshu.com/p/e87bb207977c
# https://www.ssh.com/academy/ssh/config
RUN sed -i "s/UsePrivilegeSeparation.*/UsePrivilegeSeparation no/g" /etc/ssh/sshd_config && \
	sed -i "s/UsePAM.*/UsePAM no/g" /etc/ssh/sshd_config && \
	sed -i "s/#PermitRootLogin.*/PermitRootLogin yes/g" /etc/ssh/sshd_config && \
    sed -i "1i\export LESSCHARSET=utf-8" /etc/profile && \
    sed -i 's/#AllowTcpForwarding.*/AllowTcpForwarding yes/g' /etc/ssh/sshd_config && \
    # 运行网络链接端口
    sed -i 's/#GatewayPorts.*/GatewayPorts yes/g' /etc/ssh/sshd_config && \
    # 关闭GSS认证可以提高ssh连接速度
    sed -i 's/#GSSAPIAuthentication.*/GSSAPIAuthentication no/g' /etc/ssh/sshd_config && \
    # 禁用客户端dns验证
    sed -i 's/#UseDNS.*/UseDNS no/g' /etc/ssh/sshd_config && \
    # 禁用ssh方式登录
    sed -i 's/#PubkeyAuthentication.*/PubkeyAuthentication no/g' /etc/ssh/sshd_config && \
    # hosts: files dns就是配置的解析顺序,files指的是/etc/hosts文件,dns用的是/etc/resolv.conf
    sed -i 's/hosts:          files dns/hosts:          files/g' /etc/nsswitch.conf

# port
EXPOSE 22
EXPOSE 6800

# sh 脚本
RUN mkdir -p /idesh
COPY entrypoint_base.sh /idesh/entrypoint_base.sh
RUN chmod +x /idesh/entrypoint_base.sh

ENTRYPOINT ["/idesh/entrypoint_base.sh"]
