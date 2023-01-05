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

FROM ubuntu:20.04 AS ideDownloader
# -------------下载IDE文件
# prepare tools:
RUN apt-get update && apt-get -y install --no-install-recommends wget ca-certificates tar
# download IDE to the /ide dir:
WORKDIR /download

# https://download.jetbrains.com/idea/ideaIC-2021.2.3.tar.gz
ARG downloadUrl
RUN wget -q $downloadUrl -O - | tar -xz
RUN find . -maxdepth 1 -type d -name * -execdir mv {} /ide \;

# -------------构建projector-server库，运行依赖
FROM smartide/projector-server:latest as projectorGradleBuilder


# -------------处理IDE运行程序，和构建好的rojector-server库
FROM ubuntu:20.04 AS projectorStaticFiles

# prepare tools:
RUN apt-get update && apt-get -y install --no-install-recommends unzip
# create the Projector dir:
ENV PROJECTOR_DIR /projector
RUN mkdir -p $PROJECTOR_DIR
# copy IDE:
COPY --from=ideDownloader /ide $PROJECTOR_DIR/ide
# copy projector files to the container:
COPY static $PROJECTOR_DIR
# copy projector:
COPY --from=projectorGradleBuilder $PROJECTOR_DIR/projector-server/projector-server/build/distributions/projector-server.zip $PROJECTOR_DIR
# prepare IDE - apply projector-server:
RUN unzip $PROJECTOR_DIR/projector-server.zip
RUN rm $PROJECTOR_DIR/projector-server.zip
RUN find . -maxdepth 1 -type d -name projector-server-* -exec mv {} projector-server \;
RUN mv projector-server $PROJECTOR_DIR/ide/projector-server
RUN mv $PROJECTOR_DIR/ide-projector-launcher.sh $PROJECTOR_DIR/ide/bin
RUN chmod 644 $PROJECTOR_DIR/ide/projector-server/lib/*



# -------------配置运行环境
FROM registry.cn-hangzhou.aliyuncs.com/smartide/smartide-cpp-v2:latest

RUN true \
# Any command which returns non-zero exit code will cause this shell script to exit immediately:
   && set -e \
# Activate debugging to show execution details: all commands will be printed before execution
   && set -x \
# install packages:
    && apt-get update \
# packages for awt:
    && apt-get install --no-install-recommends libxext6 libxrender1 libxtst6 libxi6 libfreetype6 -y \
# packages for user convenience:
    && apt-get install --no-install-recommends git wget curl bash-completion net-tools sudo ca-certificates procps -y \
# clean apt to reduce image size:
    && rm -rf /var/lib/apt/lists/* \
    && rm -rf /var/cache/apt

# # "https://download.jetbrains.com/idea/ideaIC-2021.2.3.tar.gz"
# ARG downloadUrl

# RUN true \
# # 返回非零退出代码的任何命令都将导致此shell脚本立即退出：
#     && set -e \
# # 激活调试以显示执行详细信息：在执行之前将打印所有命令
#     && set -x \
# # 为IDE安装特定包：
#     && apt-get update \
#     && if [ "${downloadUrl#*CLion}" != "$downloadUrl" ]; then apt-get install build-essential clang -y; else echo "-------------------------Not CLion"; fi \
#     && if [ "${downloadUrl#*pycharm}" != "$downloadUrl" ]; then apt-get install python2 python3 python3-distutils python3-pip python3-setuptools -y; else echo "-------------------------Not pycharm"; fi \
#     && if [ "${downloadUrl#*rider}" != "$downloadUrl" ]; then apt install apt-transport-https dirmngr gnupg ca-certificates -y \
#     && apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys 3FA7E0328081BFF6A14DA29AA6A19B38D3D831EF && echo "deb https://download.mono-project.com/repo/ubuntu stable-focal main" | tee /etc/apt/sources.list.d/mono-official-stable.list \
#     && apt-get update && apt install mono-devel -y --no-install-recommends; else echo "-------------------------Not rider"; fi \
#     # && wget https://packages.microsoft.com/config/ubuntu/20.04/packages-microsoft-prod.deb -O packages-microsoft-prod.deb && dpkg -i packages-microsoft-prod.deb && rm packages-microsoft-prod.deb \
#     # && apt-get update && apt-get install -y --no-install-recommends apt-transport-https dotnet-sdk-6.0 dotnet-runtime-6.0; else echo "-------------------------Not rider"; fi \
# # clean apt to reduce image size:
#     && rm -rf /var/lib/apt/lists/* \
#     && rm -rf /var/cache/apt

# copy the Projector dir:
ENV PROJECTOR_DIR /projector
COPY --from=projectorStaticFiles $PROJECTOR_DIR $PROJECTOR_DIR

ENV PROJECTOR_USER_NAME smartide

RUN true \
# Any command which returns non-zero exit code will cause this shell script to exit immediately:
    && set -e \
# Activate debugging to show execution details: all commands will be printed before execution
    && set -x \
# move run scipt:
    && mv $PROJECTOR_DIR/run.sh run.sh \
# change user to non-root (http://pjdietz.com/2016/08/28/nginx-in-docker-without-root.html):
    # && mv $PROJECTOR_DIR/$PROJECTOR_USER_NAME /home \
    && chmod g+rw /home && mkdir -p /home/smartide && mkdir -p /home/project \
    # && useradd -d /home/$PROJECTOR_USER_NAME -s /bin/bash -G sudo $PROJECTOR_USER_NAME \
    && echo '%sudo ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers \
    && chown -R $PROJECTOR_USER_NAME.$PROJECTOR_USER_NAME /home/$PROJECTOR_USER_NAME \
    && chown -R $PROJECTOR_USER_NAME.$PROJECTOR_USER_NAME /home/project \
    && chown -R $PROJECTOR_USER_NAME.$PROJECTOR_USER_NAME $PROJECTOR_DIR/ide/bin \
    && chmod +x run.sh && chmod +x /projector/ide/bin/ide-projector-launcher.sh

USER $PROJECTOR_USER_NAME
ENV HOME /home/$PROJECTOR_USER_NAME

# use sudo so that user does not get sudo usage info on (the first) login
RUN sudo echo "Running 'sudo' for smartide: success" && \
    # create .bashrc.d folder and source it in the bashrc
    mkdir -p /home/smartide/.bashrc.d && \
    (echo; echo "for i in \$(ls -A \$HOME/.bashrc.d/); do source \$HOME/.bashrc.d/\$i; done"; echo) >> /home/smartide/.bashrc


EXPOSE 8887
USER root

COPY gosu_entrypoint.sh /idesh/gosu_entrypoint.sh
RUN chmod +x /idesh/gosu_entrypoint.sh

ENTRYPOINT ["/idesh/gosu_entrypoint.sh"]
