#################################################
# SmartIDE Developer Container Image
# Licensed under GPL v3.0
# Copyright (C) leansoftX.com
#################################################

#FROM #{SMARTIDEBASEVERSION}#
FROM registry.cn-hangzhou.aliyuncs.com/smartide/smartide-node-v2:1725

USER root

# install cpp sdk
RUN apt-get update && \
    apt-get install -y build-essential cmake clang libssl-dev

RUN apt-get update && \
    apt-get install -y libcurl4-openssl-dev libssl-dev uuid-dev zlib1g-dev libpulse-dev

