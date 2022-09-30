#################################################
# SmartIDE Developer Container Image
# Licensed under GPL v3.0
# Copyright (C) leansoftX.com
#################################################

FROM registry.cn-hangzhou.aliyuncs.com/smartide/smartide-dotnet-v2:latest

USER root 

WORKDIR /home
#复制IDE文件
COPY openvscode-images opvscode

#fix 1686
RUN ln -sf /home/opvscode/bin/remote-cli/openvscode-server /home/opvscode/bin/remote-cli/code
RUN dir /home/opvscode/bin/remote-cli

ENV LANG=C.UTF-8 \
    LC_ALL=C.UTF-8 \
    EDITOR=code \
    VISUAL=code \
    GIT_EDITOR="code --wait" \
    OPENVSCODE_SERVER_ROOT=/home/opvscode

COPY gosu_entrypoint_node.sh /idesh/gosu_entrypoint_node.sh
RUN chmod +x /idesh/gosu_entrypoint_node.sh

#替换Smartide插件市场
WORKDIR opvscode
RUN find ./  -name "*.js" | xargs perl -pi -e 's|https://open-vsx.org|https://marketplace.smartide.cn|g' 
RUN find ./  -name "*.json" | xargs perl -pi -e 's|https://open-vsx.org|https://marketplace.smartide.cn|g'

ENTRYPOINT ["/idesh/gosu_entrypoint_node.sh"]
