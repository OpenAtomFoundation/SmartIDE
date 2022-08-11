#################################################
# SmartIDE Developer Container Image
# Licensed under GPL v3.0
# Copyright (C) leansoftX.com
#################################################

FROM registry.cn-hangzhou.aliyuncs.com/smartide/smartide-java-v2:latest

USER root

WORKDIR /home
#复制IDE文件
COPY openvscode-images opvscode

ENV LANG=C.UTF-8 \
    LC_ALL=C.UTF-8 \
    EDITOR=code \
    VISUAL=code \
    GIT_EDITOR="code --wait" \
    OPENVSCODE_SERVER_ROOT=/home/opvscode

COPY gosu_entrypoint_node.sh /idesh/gosu_entrypoint_node.sh
RUN chmod +x /idesh/gosu_entrypoint_node.sh

WORKDIR /opvscode
RUN contents="$(jq '.extensionsGallery.serviceUrl = "https://marketplace.smartide.cn/vscode/gallery" | .extensionsGallery.itemUrl ="https://marketplace.smartide.cn/vscode/item" | .linkProtectionTrustedDomains=["https://marketplace.smartide.cn"]' product.json )" && \
RUN echo -E "${contents}" > product.json 

ENTRYPOINT ["/idesh/gosu_entrypoint_node.sh"]
