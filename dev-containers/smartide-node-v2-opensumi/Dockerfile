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

FROM registry.cn-hangzhou.aliyuncs.com/smartide/smartide-node-v2:latest

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