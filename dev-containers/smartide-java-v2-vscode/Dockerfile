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

FROM --platform=$TARGETPLATFORM registry.cn-hangzhou.aliyuncs.com/smartide/smartide-java-v2:latest
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

#替换Smartide插件市场
WORKDIR opvscode
RUN find ./  -name "*.js" | xargs perl -pi -e 's|https://open-vsx.org|https://marketplace.smartide.cn|g' 
RUN find ./  -name "*.json" | xargs perl -pi -e 's|https://open-vsx.org|https://marketplace.smartide.cn|g'

ENTRYPOINT ["/idesh/gosu_entrypoint_node.sh"]
