---
title: "Docker & Docker-Compose 安装手册 (Linux服务器)"
linkTitle: "Docker 安装手册 (Linux)"
weight: 23
date: 2021-11-16
description: >
  本文档描述如何在 Linux服务器 上正确安装 Docker 和 Docker-Compose
---

SmartIDE可以使用任意安装了docker和docker-compose工具的linux主机作为开发环境远程主机，你的主机可以运行在公有云、私有云、企业数据中心甚至你自己家里，只要这台机器可以和互联网连通，就可以作为SmartIDE的远程主机。本文档描述正确安装docker和docker-compose工具以便确保SmartIDE可以正常管理这台远程主机。

## 一键安装docker和docker-compose工具

使用以下命令在Linux主机上安装docker和docker-compose工具，运行完成之后请从当前终端登出并从新登入以便脚本完成生效。

```bash
# 通过ssh连接到你的Linux主机，复制并粘贴此命令到终端
# 注意*不要*使用sudo方式运行此脚本
curl -o- https://smartidedl.blob.core.chinacloudapi.cn/docker/linux/docker-install.sh | bash
```

完成以上操作后，请运行以下命令测试 docker 和 docker-compose 正常安装。

```shell
docker run hello-world
docker-compose -version
```

如何可以出现类似以下结果，则表示docker和docker-compose均已经正常安装。

![验证docker和docker-compose安装正确](images/docker-install-linux001.png)

## 可选项 - 配置docker使用user namespace启动

当你完成以上安装脚本之后，你的主机已经可以作为SmartIDE的远程主机使用了。但是由于docker本身会默认使用root作为容器内的运行用户，而SmartIDE为了保证你的代码变更可以被临时保存（未提交的代码），会将容器内的代码映射到主机上。这会造成主机上的代码文件的所有者变成root。这会造成你无法直接登录主机修改这些文件，在某些情况下会不太方便。

如果你希望解决这个问题，可以使用以下脚本配置docker将当前用户映射到容器内的root用户，以便避免以上问题。

> 注意：我们建议你仅在可以独享当前主机的情况下进行以下配置，因为使用如上所述的用户映射方式会对其他用户造成不可预知的问题。

```bash
# 通过ssh连接到你的Linux主机，复制并粘贴此命令到终端
# 注意*不要*使用sudo方式运行此脚本
curl -o- https://smartidedl.blob.core.chinacloudapi.cn/docker/linux/docker-user-namespace.sh | bash
```

完成以上2项设置之后，你就可以继续按照 [快速启动](/zh/docs/quickstart/) 中 **远程主机** 部分的说明继续体验SmartIDE的功能了。