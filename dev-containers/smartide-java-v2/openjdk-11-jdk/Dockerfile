#################################################
# SmartIDE Developer Container Image
# Licensed under GPL v3.0
# Copyright (C) leansoftX.com
#################################################

FROM registry.cn-hangzhou.aliyuncs.com/smartide/smartide-node-v2:latest

USER root
RUN apt-get update && \
    apt-get -y install --no-install-recommends tzdata openjdk-8-jdk openjdk-11-jdk openjdk-17-jdk maven libsecret-1-dev zip apt-utils && \
    apt-get clean && \
    apt-get autoremove -y && \
    rm -rf /tmp/* /var/lib/apt/lists/* /var/tmp/*

# RUN curl -fsSL https://smartidedl.blob.core.chinacloudapi.cn/docker/linux/docker-install.sh | sh && \
#     curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"  &&\
#     install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

RUN wget https://dlcdn.apache.org/maven/maven-3/3.8.6/binaries/apache-maven-3.8.6-bin.tar.gz
RUN tar -xvzf apache-maven-3.8.6-bin.tar.gz
RUN cp -r apache-maven-3.8.6 /opt/maven

RUN export JAVA_HOME=/usr/lib/jvm/java-1.17.0-openjdk-amd64
RUN export M2_HOME=/opt/maven
RUN export MAVEN_HOME=/opt/maven
RUN export PATH=${M2_HOME}/bin:${PATH}

USER root

