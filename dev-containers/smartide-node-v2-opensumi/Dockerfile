#################################################
# SmartIDE Developer Container Image
# Licensed under GPL v3.0
# Copyright (C) leansoftX.com
#################################################

FROM registry.cn-hangzhou.aliyuncs.com/smartide/smartide-node-v2:2782

USER root

WORKDIR /home

RUN mkdir -p /release \
    && mkdir -p /home/smartide/.sumi/extensions

#复制IDE文件
COPY opensumi-release /release

#复制Extension文件
COPY opensumi-extension /home/smartide/.sumi/extensions

ENV WORKSPACE_DIR /home/project
ENV EXTENSION_DIR /home/smartide/.sumi/extensions
ENV EXT_MODE js
ENV NODE_ENV production

EXPOSE 8000

COPY gosu_entrypoint.sh /idesh/gosu_entrypoint.sh
RUN chmod +x /idesh/gosu_entrypoint.sh

ENTRYPOINT ["/idesh/gosu_entrypoint.sh"]