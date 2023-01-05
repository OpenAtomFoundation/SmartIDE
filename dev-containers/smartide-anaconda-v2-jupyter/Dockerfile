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
#复制IDE文件
COPY openvscode-images opvscode
COPY openvscode-images/bin/remote-cli/openvscode-server opvscode/bin/remote-cli/code

ENV LANG=C.UTF-8 \
    LC_ALL=C.UTF-8 \
    EDITOR=code \
    VISUAL=code \
    GIT_EDITOR="code --wait" \
    OPENVSCODE_SERVER_ROOT=/home/opvscode

# hadolint ignore=DL3008
RUN set -x && \
    sed -i s@/archive.ubuntu.com/@/mirrors.aliyun.com/@g /etc/apt/sources.list && \
    apt-get clean && \
    apt-get update --fix-missing && \
    apt-get install -y --no-install-recommends \
        bzip2 \
        ca-certificates \
        git \
        libglib2.0-0 \
        libsm6 \
        libxcomposite1 \
        libxcursor1 \
        libxdamage1 \
        libxext6 \
        libxfixes3 \
        libxi6 \
        libxinerama1 \
        libxrandr2 \
        libxrender1 \
        mercurial \
        openssh-client \
        procps \
        subversion \
        wget \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/* && \
    UNAME_M="$(uname -m)" && \
    if [ "${UNAME_M}" = "x86_64" ]; then \
        ANACONDA_URL="https://repo.anaconda.com/archive/Anaconda3-2021.11-Linux-x86_64.sh"; \
        SHA256SUM="fedf9e340039557f7b5e8a8a86affa9d299f5e9820144bd7b92ae9f7ee08ac60"; \
    elif [ "${UNAME_M}" = "s390x" ]; then \
        ANACONDA_URL="https://repo.anaconda.com/archive/Anaconda3-2021.11-Linux-s390x.sh"; \
        SHA256SUM="1504e9259816c5804eff1304fe7e339517b9fc1a08bfd991bc525a7efb6568f1"; \
    elif [ "${UNAME_M}" = "aarch64" ]; then \
        ANACONDA_URL="https://repo.anaconda.com/archive/Anaconda3-2021.11-Linux-aarch64.sh"; \
        SHA256SUM="4daacb88fbd3a6c14e28cd3b37004ed4c2643e2b187302e927eb81a074e837bc"; \
    elif [ "${UNAME_M}" = "ppc64le" ]; then \
        ANACONDA_URL="https://repo.anaconda.com/archive/Anaconda3-2021.11-Linux-ppc64le.sh"; \
        SHA256SUM="7eb6a95925ee756240818599f8dcbba7a155adfb05ef6cd5336aa3c083de65f3"; \
    fi && \
    wget "${ANACONDA_URL}" -O anaconda.sh -q && \
    echo "${SHA256SUM} anaconda.sh" > shasum && \
    sha256sum --check --status shasum && \
    /bin/bash anaconda.sh -b -p /opt/conda && \
    rm anaconda.sh shasum && \
    ln -s /opt/conda/etc/profile.d/conda.sh /etc/profile.d/conda.sh && \
    echo ". /opt/conda/etc/profile.d/conda.sh" >> ~/.bashrc && \
    echo "conda activate base" >> ~/.bashrc && \
    find /opt/conda/ -follow -type f -name '*.a' -delete && \
    find /opt/conda/ -follow -type f -name '*.js.map' -delete && \
    /opt/conda/bin/conda clean -afy

WORKDIR /opt/conda/lib/python3.9/site-packages/matplotlib/mpl-data/fonts/ttf
COPY simhei.ttf .
WORKDIR /home

VOLUME /home/project
EXPOSE 8888

COPY gosu_entrypoint_node.sh /idesh/gosu_entrypoint_node.sh
RUN chmod +x /idesh/gosu_entrypoint_node.sh
ENTRYPOINT ["/idesh/gosu_entrypoint_node.sh"]