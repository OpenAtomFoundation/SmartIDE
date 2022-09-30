#################################################
# SmartIDE Developer Container Image
# Licensed under GPL v3.0
# Copyright (C) leansoftX.com
#################################################

FROM registry.cn-hangzhou.aliyuncs.com/smartide/smartide-dotnet-v2:6.0

USER root

RUN mkdir $HOME/dotnet_install && cd $HOME/dotnet_install && \
    curl -L https://aka.ms/install-dotnet-preview -o install-dotnet-preview.sh && \
    sudo bash install-dotnet-preview.sh && \
    dotnet --list-sdks

# install dotnet sdk
# RUN wget https://packages.microsoft.com/config/ubuntu/20.04/packages-microsoft-prod.deb -O packages-microsoft-prod.deb
# RUN dpkg -i packages-microsoft-prod.deb

# RUN apt-get update && \ 
#     apt-get install -y apt-transport-https && \
#     apt-get install -y dotnet-sdk-7.0 && \
#     apt-get install -y aspnetcore-runtime-7.0 && \
#     apt-get autoremove -y && \
#     rm -rf /tmp/* /var/lib/apt/lists/* /var/tmp/*

# ENV DOTNET_ROOT=$HOME/dotnet
# ENV PATH=$HOME/dotnet:$PATH
# ENV PATH=$HOME/.dotnet/tools:$PATH
# ENV DOTNET_CLI_TELEMETRY_OPTOUT=false

# # RUN wget https://dot.net/v1/dotnet-install.sh
# # RUN ./dotnet-install.sh -c Current
# # RUN ./dotnet-install.sh -c 3.1
# # RUN ./dotnet-install.sh -c 5.0
# # RUN ./dotnet-install.sh -c 6.0