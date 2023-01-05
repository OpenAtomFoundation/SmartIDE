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