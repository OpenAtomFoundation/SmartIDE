---
title: "Server 安装说明"
linkTitle: "Server"
date: 2021-09-24
weight: 30
description: >
  SmartIDE Server为开发团队提供对远程容器化工作区的统一管理，支持开发者自助绑定linux主机或者k8s集群作为工作区运行资源，并通过输入Git代码库地址来一键启动远程容器化工作区。
  SmartIDE Server的基础功能是开源免费的，任何人都可以在自己的服务器上进行部署，本文档描述如何完成Server的部署过程。
---

## Server版部署架构

SmartIDE Server采用灵活可组合的部署方式，核心组件是 server-web 和 server-api，可以以采用容器的方式部署在任何支持容器的主机或者K8s集群上。Server的底层采用 Tekton流水线任务调用 SmartIDE CLI 完成对工作区的操作，包括：工作区创建、停止、删除容器、清理环境等动作，这部分的能力与 SmartIDE CLI 一致，因此底层使用的就是 CLI 本的能力。因此，SmartIDE Server 需要一个可以运行 Tekton流水线引擎的k8s环境，你可以采用最简单的 minikube 或者正式部署的k8s集群来运行Tekton流水线引擎，只要确保 server-api 可以正常与 Tekton流水线引擎的api节点进行通讯即可。

Server所管理的工作区与CLI一样，可以运行在任何支持容器的主机或者k8s集群上，这些主机和k8s集群不需要与server本身所运行的环境在同一个服务器或者集群中，可以灵活的进行组合和部署。

![](server-arch.png)

上图为SmartIDE Server版部署架构图，主要包括3大主要部分

1. SmartIDE Server：SmartIDE Server 为开发团队提供对开发环境的统一在线管理和访问能力，企业管理者可以通过SmartIDE Server为开发团队提供统一的一致的开发环境。
    - smartide-server-web：前端容器，主要为界面交互功能。
    - smartide-server-api：API层，主要负责接收Server前端请求以及cli请求，并实现与工作区管理流水线调度平台tekton的通讯，并将数据存储到数据库中。
    - tokton-pipeline：该组件运行在K8S环境之中，接收api层发送任务，并调度tekton-smartide-cli-task工作区管理执行器的编排与调度。
    - tekton-smartide-cli-task:工作区管理执行器，执行工作区的部署和管理的具体任务。
    - smartide-server-db：数据存储层。
2. 本地开发机CLI：通过SmartIDE CLI可以建立本地工作区环境并连接远程工作区，开始开发工作。同时，可以登录Server，获取Server管理工作区并建立连接，开始开发工作。
3. K8S/Host环境资源：作为开发环境的承载环境，同时支持主机以及K8S模式。

## 获取安装介质
### 2.1 安装介质列表
- Docker
- Docker-compse
- Git
- Kubectl
- MiniKube
- Tekton
- SmartIDE Server ？
- SmartIDE CLI ？

### 2.2 在线获取方式
在SmartIDE Server主机上执行如下命令，将一键为你安装好所有组件，并搭建好一个即刻可以使用的SmartIDE Server环境：
- **国内**
```bash
# 
curl -LO https://gitee.com/smartide/SmartIDE/raw/main/server/deployment/offline/deployment_cn.sh | bash
```
- **海外**
```bash
curl -LO https://raw.githubusercontent.com/SmartIDE/SmartIDE/main/server/deployment/offline/deployment.sh | bash
```

### 2.3 隔离网络内获取方式
- docker 20.10.14
  - 国内：https://smartidedl.blob.core.chinacloudapi.cn/docker/linux/docker-install.tar.gz
  - 海外：https://download.docker.com/linux/static/stable/x86_64/docker-20.10.14.tgz
- docker-compse 2.4.1
  - 国内：https://smartidedl.blob.core.chinacloudapi.cn/docker/compose/releases/download/1.29.2/docker-compose-Linux-x86_64
  - 海外：https://github.com/docker/compose/releases/download/1.29.2/docker-compose-linux-x86_64
- git 2.25.1
  - 国内：https://smartidedl.blob.core.chinacloudapi.cn/git/git-2.36.0.tar.gz
  - 海外：https://www.kernel.org/pub/software/scm/git/git-2.36.0.tar.gz
- kubectl 1.23.0
  - 国内：https://smartidedl.blob.core.chinacloudapi.cn/kubectl/v1.23.0/bin/linux/amd64/kubectl
  - 海外：https://storage.googleapis.com/kubernetes-release/release/v1.23.0/bin/linux/amd64/kubectl
- minikube 1.24.0
  - 国内：https://smartidedl.blob.core.chinacloudapi.cn/minikube/v1.24.0/minikube-linux-amd64
  - 海外：https://storage.googleapis.com/minikube/releases/v1.24.0/minikube-linux-amd64
- tekton
  - 国内：镜像包
  - 国外：镜像包
- SmartIDE Server
  - 国内：镜像包  镜像包需要重新设置config.docekr.yaml，并且设置superadmin初始化密码
  - 国外：镜像包
- SmartIDE CLI
  - 国内：见[SmartIDE CLI 安装手册]()
  - 国外：

## 3. 环境要求
### 3.1 单机版部署方式
单机版部署模式，建议使用2台虚拟机，即可完成基础环境的搭建，资源配置要求如下：
| 序号      | 虚拟机用途         | 推荐配置要求         |
| --------- | ----------- | ----------- |
| 1 | SmartIDE Server | CPU：8Core；内存：32G  磁盘：200G 操作系统：Ubuntu 20.04LTS |
| 2 | SmartIDE 开发资源主机 | CPU：8Core；内存：16G  磁盘：200G 操作系统：Ubuntu 20.04LTS |
### 3.2 K8S部署方式（TODO）
## 4. 单机版部署
### 4.1 环境准备
#### 4.1.1 虚拟机准备
虚拟机配置参考：3.1 单机版部署方式
#### 4.1.2 网络配置
2台虚拟机之间的网络配置
- Server -> 开发主机：SSH端口，默认为：22
- 开发主机 -> Server：Server网站端口，默认为8080
### 4.2 部署步骤
- 通过CLI客户端，使用superadmin
- config.docekr.yaml文件修改
  - 修改smartide.api-host地址
  - 修改api-host-xtoken 秘钥
### 4.3 环境初始化
- superAdmin账号初始化，admin账号创建

## 5. SmartIDE Server 版本

我们按照敏捷开发模式进行SmartIDE的开发，所有的版本都通过CI/CD流水线自动构建，打包，测试和发布。
为了同时满足外部用户对于稳定性的要求和开发团队以及早期使用者对新功能快速更新的要求，我们提供以下两个发布通道。

### 稳定版通道

稳定版的发布按照sprint进行，我们采用2周一个sprint的开发节奏，每个sprint结束后我们会发布一个稳定版到这个通道。这个版本经过开发团队相对完整的测试，确保所提供的功能稳定可用，同时我们会在每个sprint结束时同时发布“版本发布说明”并对当前版本的性能和改进进行说明。

流水线状态 
[![Build Status](https://dev.azure.com/leansoftx/smartide/_apis/build/status/smartide-server?branchName=release/release-11)](https://dev.azure.com/leansoftx/smartide/_build/latest?definitionId=74&branchName=release/release-11)

版本发布说明列表：

| 版本号      | 构建编号 | 发布日期      |   简要说明   |
| ----------- | ----------- | ----------- | ----------- |
| v     |  | 2022. | XXXXXX  |



