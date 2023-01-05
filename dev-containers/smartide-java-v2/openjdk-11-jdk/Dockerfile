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

FROM --platform=$TARGETPLATFORM registry.cn-hangzhou.aliyuncs.com/smartide/smartide-node-v2:latest
ARG TARGETPLATFORM 
USER root
RUN apt-get update && \
    apt-get -y install --no-install-recommends tzdata openjdk-8-jdk openjdk-11-jdk openjdk-17-jdk maven && \
    apt-get clean && \
    apt-get autoremove -y && \
    rm -rf /tmp/* /var/lib/apt/lists/* /var/tmp/*

# RUN curl -fsSL https://smartidedl.blob.core.chinacloudapi.cn/docker/linux/docker-install.sh | sh && \
#     curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"  &&\
#     install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

RUN wget https://dlcdn.apache.org/maven/maven-3/3.8.6/binaries/apache-maven-3.8.6-bin.tar.gz
RUN tar -xvzf apache-maven-3.8.6-bin.tar.gz
RUN cp -r apache-maven-3.8.6 /opt/maven

RUN echo 'JAVA_HOME=/usr/lib/jvm/java-1.17.0-openjdk-amd64' >> /root/.bashrc
RUN echo 'M2_HOME=/opt/maven' >> /root/.bashrc
RUN echo 'MAVEN_HOME=/opt/maven' >> /root/.bashrc
RUN echo 'PATH=${M2_HOME}/bin:${PATH}' >> /root/.bashrc

RUN echo 'JAVA_HOME=/usr/lib/jvm/java-1.17.0-openjdk-amd64' >> .bashrc
RUN echo 'M2_HOME=/opt/maven' >> .bashrc
RUN echo 'MAVEN_HOME=/opt/maven' >> .bashrc
RUN echo 'PATH=${M2_HOME}/bin:${PATH}' >> .bashrc

RUN rm -rf apache-maven-3.8.6-bin.tar.gz apache-maven-3.8.6 