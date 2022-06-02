#################################################
# SmartIDE Developer Container Image
# Licensed under GPL v3.0
# Copyright (C) leansoftX.com
#################################################

FROM registry.cn-hangzhou.aliyuncs.com/smartide/smartide-node-v2:latest

USER root

# install dotnet sdk
RUN wget https://packages.microsoft.com/config/ubuntu/16.04/packages-microsoft-prod.deb -O packages-microsoft-prod.deb
RUN dpkg -i packages-microsoft-prod.deb

RUN apt-get update && \ 
    apt-get install -y apt-transport-https && \
    # apt-get install -y dotnet-sdk-6.0 && \
    # apt-get install -y aspnetcore-runtime-6.0 && \
    apt-get autoremove -y && \
    rm -rf /tmp/* /var/lib/apt/lists/* /var/tmp/*

# 不工作
# RUN wget https://dot.net/v1/dotnet-install.sh && \ 
#     chmod +x dotnet-install.sh && \ 
#     ./dotnet-install.sh -c 3.1


# # RUN ./dotnet-install.sh -c 3.1
# # RUN ./dotnet-install.sh -c 5.0
# # RUN ./dotnet-install.sh -c 6.0