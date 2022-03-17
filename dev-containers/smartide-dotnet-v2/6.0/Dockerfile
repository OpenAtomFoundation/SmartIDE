#################################################
# SmartIDE Developer Container Image
# Licensed under GPL v3.0
# Copyright (C) leansoftX.com
#################################################

#FROM #{SMARTIDEBASEVERSION}#
FROM registry.cn-hangzhou.aliyuncs.com/smartide/smartide-node-v2:1725

USER root

# install dotnet sdk
RUN wget https://packages.microsoft.com/config/ubuntu/16.04/packages-microsoft-prod.deb -O packages-microsoft-prod.deb
RUN dpkg -i packages-microsoft-prod.deb

RUN apt-get update && \ 
    apt-get install -y apt-transport-https && \
    apt-get install -y dotnet-sdk-6.0 && \
    apt-get install -y aspnetcore-runtime-6.0 && \
    apt-get autoremove -y && \
    rm -rf /tmp/* /var/lib/apt/lists/* /var/tmp/*

ENV DOTNET_ROOT=$HOME/dotnet
ENV PATH=$HOME/dotnet:$PATH
ENV PATH=$HOME/.dotnet/tools:$PATH
ENV DOTNET_CLI_TELEMETRY_OPTOUT=false

# RUN wget https://dot.net/v1/dotnet-install.sh
# RUN ./dotnet-install.sh -c Current
# RUN ./dotnet-install.sh -c 3.1
# RUN ./dotnet-install.sh -c 5.0
# RUN ./dotnet-install.sh -c 6.0