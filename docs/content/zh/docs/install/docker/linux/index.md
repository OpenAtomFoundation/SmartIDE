---
title: "Docker & Docker-Compose 安装手册 (Linux服务器)"
linkTitle: "Docker 安装手册 (Linux)"
weight: 40
date: 2021-11-16
description: >
  本文档描述如何在 Linux服务器 上正确安装 Docker 和 Docker-Compose
---

SmartIDE可以使用任意安装了docker和docker-compose工具的linux主机作为开发环境远程主机，你的主机可以运行在公有云、私有云、企业数据中心甚至你自己家里，只要这台机器可以和互联网连通，就可以作为SmartIDE的远程主机。本文档描述正确安装docker和docker-compose工具以便确保SmartIDE可以正常管理这台远程主机。

## 环境要求列表

使用SmartIDE的远程主机需要满足以下要求

- 操作系统：
  - Ubuntu 20.04LTS, 18.04LTS, 16.04LTS
  - CentOS 7.2 以上
- 软件需求：
  - Git
  - Docker
  - Docker-Compose

> 特别说明：你无需在linux主机上安装SmartIDE CLI命令行工具，因为所有的操作都可以在本地开发机（Windows/MacOS）上完成，SmartIDE本身内置使用SSH远程操作linux主机。

## 配置非root用户登陆服务器

使用SmartIDE远程控制的linux主机需要使用非root用户登陆，这是为了确保更加安全的开发环境以及主机和容器内文件系统权限的适配。请按照以下操作创建用户并赋予免密码sudo权限。

**备注：以下操作需要使用root账号或者sudo方式运行。**

以下脚本将创建一个叫做 smartide 的用户并赋予面密码的sudo权限。

```shell
## 创建用户及用户文件系统
useradd -m smartide
## 为用户设置密码，请在命令执行后输入你需要设置的密码，确保你将这个密码记录下来
passwd smartide
## 为用户添加sudo权限
usermod -aG sudo smartide
## 编辑 /etc/sudoers
vim /etc/sudoers
```

**如果使用vim编辑器，按 "i" 进入插入编辑模式，编辑完成后按ESC退出编辑模式，输入 ":wq!" 保存退出。**

在 /etc/sudoers 文件中添加如下内容

```shell
smartide   ALL=(ALL) NOPASSWD: ALL
```

同时修改开头为 %sudo 的行为以下内容

```shell
%sudo   ALL=(ALL) NOPASSWD: ALL
```

![](images/sudoer_nopwd.png)

## 配置用户使用bash作为默认的shell

SmartIDE 使用一些特定的bash指令完成工作区调度，因此请确保你所使用的用户或者以上所创建的smartide用户的默认shell为bash。

```shell
vim /etc/passwd
```

更新 

```shell
smartide:x:1000:1000::/home/smartide:/bin/sh 
```

改为

```shell
smartide:x:1000:1000::/home/smartide:/bin/bash
```


## 一键安装docker和docker-compose工具

> 使用以上创建的smartide用户或者其他符合要求的非root用户登陆服务器。

使用以下命令在Linux主机上安装docker和docker-compose工具，**运行完成之后请从当前终端登出并从新登入以便脚本完成生效**。

```bash
# 通过ssh连接到你的Linux主机，复制并粘贴此命令到终端
# 注意不要使用sudo方式运行此脚本
curl -o- https://smartidedl.blob.core.chinacloudapi.cn/docker/linux/docker-install.sh | bash
# 退出当前登录
exit
```

完成以上操作后，请运行以下命令测试 docker 和 docker-compose 正常安装。

```shell
docker run hello-world
docker-compose -version
```

如何可以出现类似以下结果，则表示docker和docker-compose均已经正常安装。

![验证docker和docker-compose安装正确](images/docker-install-linux001.png)

完成以上2项设置之后，你就可以继续按照 [快速启动](/zh/docs/quickstart/) 中 **远程模式** 部分的说明继续体验SmartIDE的功能了。

## 配置Sysbox
> SmartIDE 的部分开发者镜像内使用了 VM-Like-Container 的能力，比如需要在容器内运行完整的docker，也就是嵌套容器（Docker in Docker - DinD）。类似的能力需要我们在主机上针对容器运行时（container runtime）进行一些改进。为了满足类似的能力，SmartIDE 采用了开源项目 sysbox  提供类似的容器运行时。

以下文档描述了如何在主机上安装 Sysbox容器运行时，在执行此操作之前，请确保已经按照本文档的上半部分正确安装了 docker环境。

本文档针对 sysbox 社区版的安装过程进行说明，如果需要安装企业版，请自行参考 [sysbox](https://github.com/nestybox/sysbox)  官方文档。

```shell
## 国内安装地址
wget https://smartidedl.blob.core.chinacloudapi.cn/hybrid/sysbox/sysbox-ce_0.5.2-0.linux_amd64.deb

## 国际安装地址
wget https://downloads.nestybox.com/sysbox/releases/v0.5.2/sysbox-ce_0.5.2-0.linux_amd64.deb
```

安装前需要通过执行以下命令移除当前正在运行的容器

```shell
docker rm $(docker ps -a -q) -f
```

执行以下指令完成安装

```shell
sudo apt-get install ./sysbox-ce_0.5.2-0.linux_amd64.deb
```

安装成功后通过执行以下命令来验证Sysbox是否安装成功并已启动服务，注意查看 active (running) 的提示信息

```shell
sudo systemctl status sysbox -n20
```
输出的信息如下图：
![输入图片说明](images/Sysbox.png)