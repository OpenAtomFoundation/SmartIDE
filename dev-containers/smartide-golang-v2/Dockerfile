#################################################
# SmartIDE Developer Container Image
# Licensed under GPL v3.0
# Copyright (C) leansoftX.com
#################################################

FROM #{SMARTIDEBASEVERSION}#

USER root

RUN set -eux; \
    apt-get update; \
    apt-get install -y --no-install-recommends \
    g++ gcc libc6-dev make pkg-config; \
    apt-get clean && \
    apt-get autoremove -y && \
    rm -rf /tmp/* /var/lib/apt/lists/* /var/tmp/*

ARG BUILD_GOLANG_VERSION
ENV GOLANG_VERSION=$BUILD_GOLANG_VERSION
ENV GOPATH=/home/smartide/go
ENV GOROOT=/usr/local/go
# go get -v  安装包需要，路径不能使用变量配置
ENV PATH=/usr/local/go/bin:$PATH
ENV PATH=/home/smartide/go/bin:$PATH

RUN mkdir -p /home/smartide && mkdir -p /home/smartide/go && \
    mkdir -p "$GOPATH/src" "$GOPATH/bin" "$GOPATH/pkg/mod" && chmod -R 777 "$GOPATH" && \
    echo "export PATH=/usr/local/go/bin:/home/smartide/go/bin:$PATH" >> /etc/profile

RUN wget -O go.tgz https://go.dev/dl/go$GOLANG_VERSION.linux-amd64.tar.gz --progress=dot:giga \
    && tar -C /usr/local -xzf go.tgz \
    && rm -rf go.tgz
