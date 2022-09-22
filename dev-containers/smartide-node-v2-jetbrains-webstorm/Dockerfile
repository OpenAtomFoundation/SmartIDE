#################################################
# SmartIDE Developer Container Image
# Licensed under GPL v3.0
# Copyright (C) leansoftX.com
#################################################

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

FROM registry.cn-hangzhou.aliyuncs.com/smartide/smartide-node-v2:latest

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
    && apt-get install --no-install-recommends git wget curl bash-completion net-tools sudo ca-certificates -y \
# clean apt to reduce image size:
    && rm -rf /var/lib/apt/lists/* \
    && rm -rf /var/cache/apt

ENV DEBIAN_FRONTEND=noninteractive
ENV TZ=Asia/Shanghai
# copy the Projector dir:
ENV PROJECTOR_DIR /projector
COPY --from=projectorStaticFiles $PROJECTOR_DIR $PROJECTOR_DIR


ENV PROJECTOR_USER_NAME smartide

RUN true \
# Any command which returns non-zero exit code will cause this shell script to exit immediately:
    && set -e \
# Activate debugging to show execution details: all commands will be printed before execution
    && set -x \
# move run scipt: ##will remove
    && mv $PROJECTOR_DIR/run.sh run.sh \ 
# change user to non-root (http://pjdietz.com/2016/08/28/nginx-in-docker-without-root.html):
    # && mv $PROJECTOR_DIR/$PROJECTOR_USER_NAME /home \
    && chmod g+rw /home && mkdir -p /home/smartide && mkdir -p /home/project \
    #&& useradd -d /home/$PROJECTOR_USER_NAME -s /bin/bash -G sudo $PROJECTOR_USER_NAME \
    && echo '%sudo ALL=(ALL) NOPASSWD:ALL' >> /etc/sudoers \
    && chown -R $PROJECTOR_USER_NAME.$PROJECTOR_USER_NAME /home/$PROJECTOR_USER_NAME \
    && chown -R $PROJECTOR_USER_NAME.$PROJECTOR_USER_NAME /home/project \
    && chown -R $PROJECTOR_USER_NAME.$PROJECTOR_USER_NAME $PROJECTOR_DIR/ide/bin \
    ##will remove && chmod +x run.sh 
    && chmod +x run.sh && chmod +x /projector/ide/bin/ide-projector-launcher.sh

USER $PROJECTOR_USER_NAME
ENV HOME /home/$PROJECTOR_USER_NAME

# use sudo so that user does not get sudo usage info on (the first) login
RUN sudo echo "Running 'sudo' for smartide: success" && \
    # create .bashrc.d folder and source it in the bashrc
    mkdir -p /home/smartide/.bashrc.d && \
    (echo; echo "for i in \$(ls -A \$HOME/.bashrc.d/); do source \$HOME/.bashrc.d/\$i; done"; echo) >> /home/smartide/.bashrc


# WORKDIR $HOME
# ENV NODE_VERSION=14.17.6
# ENV NVM_DIR /home/$PROJECTOR_USER_NAME/.nvm
# RUN curl -fsSL https://raw.githubusercontent.com/nvm-sh/nvm/v0.38.0/install.sh | sh \
#     && . $NVM_DIR/nvm.sh  \
#     && nvm install $NODE_VERSION \
#     && nvm alias default $NODE_VERSION \
#     && npm install -g yarn node-gyp \
#     && echo ". ~/.nvm/nvm-lazy.sh"  >> /home/$PROJECTOR_USER_NAME/.bashrc.d/50-node
# # above, we are adding the lazy nvm init to .bashrc, because one is executed on interactive shells, the other for non-interactive shells (e.g. plugin-host)
# COPY --chown=$PROJECTOR_USER_NAME:$PROJECTOR_USER_NAME nvm-lazy.sh /home/$PROJECTOR_USER_NAME/.nvm/nvm-lazy.sh
# ENV PATH $NVM_DIR/versions/node/v$NODE_VERSION/bin:$PATH


# USER root

# RUN echo "export NVM_DIR_ROOT=\"/home/$PROJECTOR_USER_NAME/.nvm\"" >> /etc/profile && \
#     echo "[ -s \"\$NVM_DIR_ROOT/nvm.sh\" ] && \. \"\$NVM_DIR_ROOT/nvm.sh\""  >> /etc/profile  && \
#     echo "[ -s \"\$NVM_DIR_ROOT/bash_completion\" ] && \. \"\$NVM_DIR_ROOT/bash_completion\""  >> /etc/profile


USER root

EXPOSE 8887

#CMD ["bash", "-c", "/run.sh"]

COPY gosu_entrypoint_node.sh /idesh/gosu_entrypoint_node.sh
RUN chmod +x /idesh/gosu_entrypoint_node.sh

ENTRYPOINT ["/idesh/gosu_entrypoint_node.sh"]

