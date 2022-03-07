---
title: "工作区"
linkTitle: "工作区"
weight: 50
draft: false
description: >
  工作区(Workspace)是SmartIDE中的最重要的概念，SmartIDE的所有功能都是围绕工作区展开的。SmartIDE支持3种工作区运行模式，本地模式、远程模式和k8s模式。
---

工作区(Workspace)是SmartIDE中的核心概念，工作区为开发人员提供某个应用开发所需要的完整组件，一般来说会包括：代码库、代码编辑器、开发语言SDK、开发调试工具、测试工具、依赖环境/中间件以及管理工具。SmartIDE使用一个叫做 .ide.yaml 的文件对这些组件进行描述，并通过容器的方式标准化这些组件的获取过程，达到简化开发环境管理的目标。

SmartIDE的工作区可以通过以下4种方式启动，覆盖开发者所需要的所有使用场景：

- 本地工作区：工作区运行在开发者自己的开发机上，开发机可以是Windows, MacOS或者是Linux均可。
- 远程主机工作区：工作区运行在远程的linux主机上，这台linux主机可以运行在任何位置（本地虚拟机、局域网中或者云端），只要开发者可以通过SSH访问这台主机即可。
- k8s工作区：工作区运行在k8s集群中，一般来说一个工作区就是一个pod。
- server工作区：开发者在SmartIDE Server的网页上启动的工作区，可以以远程模式或者k8s模式运行。严格来说，Server模式并不是一个独立的模式，而是远程模式和k8s模式的另外一种操作方式。

## localhost访问

SmartIDE工作区集成了 **SSH隧道** 技术，无论你采用那种以下哪种方式启动工作区，SmartIDE CLI都会自动帮助你建立起本地开发机和工作区资源之间的隧道，并将工作区内部资源的端口转发到 `localhost` 上，这样你就可以完全通过 `localhost:<port>` 来访问这些资源，而不用关心这些资源具体运行在本地、远程主机或者k8s集群中。

这样做的另外一个好处是，如果你已经习惯了使用一些开发环境管理工具，比如：MySQL Workbench来连接你的MySQL服务器，那么对于SmartIDE工作区种的MySQL服务器就可以直接通过 `localhost:3306` 来访问；这种体验和MySQL服务器就运行在你本地开发机上完全一致，但是又不会占用你的本地资源。

使用 **SSH隧道** 技术的另外一项优势是：避免了在远程服务器环境（linux主机或者k8s集群）上对外暴露端口，因为所有的转发都通过 **SSH隧道** 完成，因此远程服务器资源只需要提供最基本的访问方式，而不必另外打通防火墙或者进行额外的网络配置。

一般来说SmartIDE工作区都提供WebIDE和SSH终端方式，那么你就可以通过

- `http://localhost:6800` 访问WebIDE （6800是SmartIDE内集成的WebIDE的惯用端口号）
- `ssh smartide@localhost -p 6822` 访问工作区的终端 (6822是SmartIDE工作区内置的SSH服务的惯用端口号)

**注意：** 为了避免因为端口占用造成端口转发失败，SmartIDE CLI会自动检测当前端口被占用的情况，并自动将端口值增加100以便避免冲突。

## 四类工作区

### 本地工作区

本地工作区完全运行在开发者自己的开发机上面，但是采用容器的方式运行。开发者可以使用Windows，MacOS或者Linux三种操作系统中的任何一种运行本地工作区，唯一的前提条件是已经安装了 **Docker桌面版** 工具，具体的安装方式请参考我们的 [安装](/zh/docs/install) 文档。

启动本地工作区的cli指令非常简单，只需要运行以下一个指令即可：

```shell
smartide start <代码库Url>
```

### 远程主机工作区

远程主机工作区运行在linux主机上，同样采用容器的方式运行。这台linux主机可以运行在任意位置，包括：本地虚拟机(VirutalBox/VMWare/HyperV)，本地网络上的服务器或者云服务器；唯一的要求是开发者可以通过SSH直接登录这台主机。

启动远程主机工作区并不需要开发者在远程主机上安装 SmartIDE CLI 工具，开发者只需要按照我们的 [安装](/zh/docs/install) 文档对linux主机进行初始化以后，即可在本地开发机上使用 SmartIDE CLI 工具远程操作这台linux主机。在这种场景下，开发者本地的开发机也不需要安装 **Docker桌面版** 。

启动远程主机工作区的cli指令也非常简单：

```shell
## 添加远程主机到SmartIDE主机资源列表
smartide host add <远程主机的IP地址> --username <SSH登录用户名> --password <SSH密码/如果使用key的方式认证则不需要输入>
## 获取SmartIDE主机资源列表
smartide host list
## 通过<主机ID>在远程主机上启动项目
smartide start --host <主机ID> <代码库Url>
```

### k8s工作区

k8s工作区运行在k8s集群中，采用容器的方式运行。开发者需要首先获取k8s集群的操作密钥并且在本地安装好Kubectl工具，接可以通过 SmartIDE CLI 工具在k8s其中启动工作区。在这种情况下，开发者本地的开发机也不需要安装 **Docker桌面版** 。

```shell
## 在K8s集群上启动项目
smartide start --k8s <实例名称> --namespace <命令空间名称> <代码库Url>
```

### Server工作区

SmartIDE Server为用户管理工作区提供可视化的网页操作，但是Server工作区本质上仍然是运行在远程linux主机上的远程主机工作区或者运行在k8s集群中的k8s工作区。SmartIDE Server允许用户讲自己的linux主机或者k8s集群注册到资源列表，并通过 工作区管理 页面创建server工作区。

在用户使用Server工作区的过程中，需要使用SmartIDE CLI执行 `smartide connect` 指令允许cli监听正在运行的Server工作区列表并建立 **SSH隧道** 以便用户可以继续使用 `localhost:port` 的方式访问工作区资源。

