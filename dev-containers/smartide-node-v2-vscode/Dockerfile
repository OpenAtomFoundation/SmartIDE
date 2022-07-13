#################################################
# SmartIDE Developer Container Image
# Licensed under GPL v3.0
# Copyright (C) leansoftX.com
#################################################

FROM --platform=$TARGETPLATFORM registry.cn-hangzhou.aliyuncs.com/smartide/smartide-node-v2:latest
ARG TARGETPLATFORM 
USER root

WORKDIR /home
#复制IDE文件
COPY openvscode-images-arm64 opvscode-arm64
COPY openvscode-images-amd64 opvscode-amd64



#复制IDE文件
SHELL ["/bin/bash", "-c"]
RUN if [ "$TARGETPLATFORM" = "linux/arm64" ];then mv opvscode-arm64/* opvscode;mv opvscode/bin/remote-cli/openvscode-server opvscode/bin/remote-cli/code;else mv opvscode-amd64/* opvscode;mv opvscode/bin/remote-cli/openvscode-server opvscode/bin/remote-cli/code;fi

ENV LANG=C.UTF-8 \
    LC_ALL=C.UTF-8 \
    EDITOR=code \
    VISUAL=code \
    GIT_EDITOR="code --wait" \
    OPENVSCODE_SERVER_ROOT=/home/opvscode

COPY gosu_entrypoint_node.sh /idesh/gosu_entrypoint_node.sh
RUN chmod +x /idesh/gosu_entrypoint_node.sh

ENTRYPOINT ["/idesh/gosu_entrypoint_node.sh"]
