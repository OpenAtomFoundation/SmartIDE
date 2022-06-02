---
title: "镜像和模版"
linkTitle: "镜像和模版"
weight: 60
description: >
  SmartIDE开发容器镜像和模版库说明
---

## 开发者镜像

SmartIDE 提供了 `主流开发语言的SDK` 的容器镜像，包括 node/javascript、java、golang、python、dotnet/C#、php、C/C++、，集成了VSCode和JetBrains两大主流IDE体系，并在近期完成了对OpenSumi国产IDE的node SDK支持。开发者可以直接使用这些作为自己的通用开发环境，或者以这些镜像为基础来构建自己专属的开发镜像。

SmartIDE所提供的开发者容器镜像中已经针对开发调试和测试场景进行了一系列的优化，相对于自己从头构建容器镜像来说，使用SmartIDE的开发者镜像会有一系列的好处：

- 非root用户运行：docker默认采用root方式运行容器，这带来了很多方便的同时也造成一些不安全的因素，比如用户可能在容器内越权操作主机系统，在容器内创建的文件如果被映射到主机将会作为root用户所有等等。这些问题对于使用容器单纯运行应用不易构成太大的问题，但是对于直接使用容器进行应开发来说就会造成巨大的安全隐患以及不方便。SmartIDE所提供的所有开发者容器镜像均采用普通用户权限运行，并且允许用户在启动容器时指定容器内用户的密码，方便开发者控制容器内环境同时也避免容器内操作越权。
- 内置SSH服务支持：对于开发者来说，对容器内环境需要非常高的操作便利性，提供SSH访问能力可以极大的方便开发者对自己的开发环境进行各类操作和控制。这一点上与运维用途的容器也非常不同，一般来说运维用途的容器会尽量避免用户直接进入，而尽量采用 `不可变` 原则来进行管理，而开发者容器在利用 `不可变` 原则提供环境一致性的同时还要允许开发者进行定制，因此提供SSH直接访问就非常的重要。

### 国内国外双托管

为了同时兼顾国内和国外开发者使用，所有SmartIDE的镜像同时托管至阿里云和Docker Hub，方便开发者根据自己的地理位置选择最快速的拉取地址

- 国内阿里云地址：`registry.cn-hangzhou.aliyuncs.com/smartide/<镜像名称:TAG>`
- 国外 Docker Hub 地址：`registry.hub.docker.com/smartide/<镜像名称:TAG>`

> 对于以下列表中所列出的所有镜像，均可以通过替换地址前缀的方式更换拉取源。


## 模版库

为了方便开发者使用我们所提供的开发者镜像，我们在 SmartIDE CLI 中内置了环境模版功能，开发者可以使用 `smartide new` 指令获取所有可用的模版，这些模版与以上的开发者镜像一一对应，可以用来一键创建标准化的开发环境。

SmartIDE模版库本身是开源的，地址为

- 国内Gitee: https://gitee.com/smartide/smartide-templates
- 国外GitHub: https://github.com/smartide/smartide-templates

完整的指令列表如下
```shell
## 完整技术栈和IDE匹配列表
smartide new node|java|golang|dotnet|python|php|cpp [-t vscode|(webstorm|idea|rider|goland|pycharm｜phpstorm|clion)|opensumi]
```

各个技术栈相关的模版启动指令如下：

### Node/JavaScript/前端

```shell
#########################################
# Node/JavaScript 前端技术栈
#########################################

## 创建带有node全版本sdk的开发容器，无IDE，可通过VSCode SSH Remote或者JetBrains Gateway接入
smartide new node
## 创建带有node全版本sdk的开发容器，使用VSCode WebIDE
smartide new node -t vscode
## 创建带有node全版本sdk的开发容器，使用JetBrains WebStorm WebIDE
smartide new node -t webstorm
## 创建带有node全版本sdk的开发容器，使用Opensumi WebIDE
smartide new node -t opensumi
```

### Java语言

```shell
#########################################
# Java语言
#########################################

## 创建带有JDK的开发容器，无IDE，可通过VSCode SSH Remote或者JetBrains Gateway接入
smartide new java
## 创建带有JDK开发容器，使用VSCode WebIDE
smartide new java -t vscode
## 创建带有JDK开发容器，使用JetBrains IntelliJ IDEA WebIDE
smartide new java -t idea
```

### Go语言

```shell
#########################################
# Go语言
#########################################

## 创建带有Go的开发容器，无IDE，可通过VSCode SSH Remote或者JetBrains Gateway接入
smartide new golang
## 创建带有Go开发容器，使用VSCode WebIDE
smartide new golang -t vscode
## 创建带有Go开发容器，使用JetBrains Goland WebIDE
smartide new golang -t goland
```

### DotNet (跨平台版本)

```shell
#########################################
# DotNet (跨平台版本)
#########################################

## 创建带有.Net的开发容器，无IDE，可通过VSCode SSH Remote或者JetBrains Gateway接入
smartide new dotnet
## 创建带有.Net开发容器，使用VSCode WebIDE
smartide new dotnet -t vscode
## 创建带有.Net开发容器，使用JetBrains Rider WebIDE
smartide new dotnet -t rider
```

### Python

```shell
#########################################
# Python
#########################################

## 创建带有Python的开发容器，无IDE，可通过VSCode SSH Remote或者JetBrains Gateway接入
smartide new python
## 创建带有Python开发容器，使用VSCode WebIDE
smartide new python -t vscode
## 创建带有Python开发容器，使用JetBrains PyCharm WebIDE
smartide new python -t pycharm
```

### PHP

```shell
#########################################
# PHP
#########################################

## 创建带有PHP的开发容器，无IDE，可通过VSCode SSH Remote或者JetBrains Gateway接入
smartide new php
## 创建带有PHP开发容器，使用VSCode WebIDE
smartide new php -t vscode
## 创建带有PHP开发容器，使用JetBrains PhpStorm WebIDE
smartide new php -t phpstorm
```

### C/C++

```shell
#########################################
# C/C++
#########################################

## 创建带有C/C++的开发容器，无IDE，可通过VSCode SSH Remote或者JetBrains Gateway接入
smartide new cpp
## 创建带有C/C++开发容器，使用VSCode WebIDE
smartide new cpp -t vscode
## 创建带有C/C++开发容器，使用JetBrains Clion WebIDE
smartide new cpp -t clion
```

## 镜像和模版指令列表

为了方便开发者使用我们的镜像，当前所有公开镜像的地址公布如下并会持续更新。有关以下镜像体系的详细说明，请参考 [SmartIDE Sprint 9 (v0.1.9）发布说明](/zh/blog/2022-0104-sprint9/#smartide-image-v2)

SmartIDE开发者镜像分成3层提供，分别提供不同的能力。

### L0 - 基础镜像

提供基础能力，比如：认证，权限，SSH；使用Ubuntu满足日常开发基础组件的需求


| **开发语言** | **镜像类型** | **tag**| **Pull命令**| **new指令**| **备注**|
|----------|----------|---------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------|----------------------------|-------------------------------------------------------------|
| base     | 基础       | latest                                       | `docker pull registry.cn-hangzhou.aliyuncs.com/smartide/smartide-base-v2:latest`                                        | `se new base`              | 基于ubuntu:20.04，集成git、ssh server等基础库                         |
| base     | 基础       | latest                                       | `docker pull registry.cn-hangzhou.aliyuncs.com/smartide/smartide-base-v2-vscode:latest`                                        | `se new base -t vscode`              | 基于ubuntu:20.04，集成git、ssh server等基础库，内置VSCode WebIDE                         |

### L1 - SDK镜像

SDK镜像提供开发语言环境支持能力，同时提供SDK Only的使用方式，允许本地IDE将SDK镜像作为开发环境直接使用，不嵌入WebIDE

| **开发语言** | **镜像类型** | **tag**| **Pull命令**| **new指令**| **备注**|
|----------|----------|---------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------|----------------------------|-------------------------------------------------------------|
| node     | SDK      | all-version,latest                           | `docker pull registry.cn-hangzhou.aliyuncs.com/smartide/smartide-node-v2:latest`                                        | `se new node`              | 在base 镜像的基础上，集成了Node V14.17.6(默认)、V12.22.7 V16.7.0 SDK及nvm  |
| Java     | SDK      | openjdk-11-jdk,2801,latest                        | `docker pull registry.cn-hangzhou.aliyuncs.com/smartide/smartide-java-v2:2801`                                        | `se new java`              | 在Node SDK 镜像的基础上，集成Java Open JDK 11及maven                   |
| golang   | SDK      | 1.17.5,2800,latest; 1.16.12,1745                  | `docker pull registry.cn-hangzhou.aliyuncs.com/smartide/smartide-golang-v2:2800`                                      | `se new golang`            | 在Node SDK 镜像的基础上，集成Go SDK，分为1.17.5、1.16.12两个版本              |
| python   | SDK      | all-version,2848,latest                           | `docker pull registry.cn-hangzhou.aliyuncs.com/smartide/smartide-python-v2:2848`                                      | `se new python`            | 在Node SDK 镜像的基础上，集成python2和python3                          |
| dotnet   | SDK      | 6.0,2798,latest                                   | `docker pull registry.cn-hangzhou.aliyuncs.com/smartide/smartide-dotnet-v2:2798`                                      | `se new dotnet`            | 在Node SDK 镜像的基础上，集成Net6.0 SDK 和asp.net core                 |
| php      | SDK      | php7.4,2802,latest                                | `docker pull registry.cn-hangzhou.aliyuncs.com/smartide/smartide-php-v2:2802`                                             | `se new php`               | 在Node SDK 镜像的基础上，集成php7.4和Apache2                           |
| C++ | SDK    | clang,2797,latest          | `docker pull registry.cn-hangzhou.aliyuncs.com/smartide/smartide-cpp-v2:2797`                          | `se new cpp`           | 在Node SDK 镜像的基础上,集成cmake、clang       |

### L2 - WebIDE镜像

WebIDE镜像，在SDK镜像的基础上嵌入常用的IDE

VSCode WebIDE

| **开发语言** | **镜像类型** | **tag**| **Pull命令**| **new指令**| **备注**|
|----------|----------|---------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------|----------------------------|-------------------------------------------------------------|
| node     | VSCode   | all-version,2898,latest                           | `docker pull registry.cn-hangzhou.aliyuncs.com/smartide/smartide-node-v2-vscode:2898`                                 | `se new node -t vscode`    | 在SDK镜像的基础上集成VSCode WebIDE                                   |
| Java     | VSCode   | openjdk-11-jdk,2902,latest                        | `docker pull registry.cn-hangzhou.aliyuncs.com/smartide/smartide-java-v2-vscode:2902`                                 | `se new java -t vscode`    | 在SDK镜像的基础上集成VSCode WebIDE                                   |
| golang   | VSCode   | 1.17.5,2903,latest;1.16.12,1747                   | `docker pull registry.cn-hangzhou.aliyuncs.com/smartide/smartide-golang-v2-vscode:2903`                               | `se new golang -t vscode`  | 在SDK镜像的基础上集成VSCode WebIDE                                   |
| python   | VSCode   | all-version,2901,latest                           | `docker pull registry.cn-hangzhou.aliyuncs.com/smartide/smartide-python-v2-vscode:2901`                               | `se new python -t vscode`  | 在SDK镜像的基础上集成VSCode WebIDE                                   |
| dotnet   | VSCode   | 6.0,2904,latest                                   | `docker pull registry.cn-hangzhou.aliyuncs.com/smartide/smartide-dotnet-v2-vscode:2904`                               | `se new dotnet -t vscode`  | 在SDK镜像的基础上集成VSCode WebIDE                                   |
| php      | VSCode   | php7.4,2900                                       | `docker pull registry.cn-hangzhou.aliyuncs.com/smartide/smartide-php-v2-vscode:2900`                                  | `se new php -t vscode`     | 在SDK镜像的基础上集成VSCode WebIDE                                   |
| C++ | VSCode | clang,2899,latest          | `docker pull registry.cn-hangzhou.aliyuncs.com/smartide/smartide-cpp-v2-vscode:2899`                   | `se new cpp -t vscode` | 在SDK镜像的基础上集成VSCode WebIDE            |

JetBrains Projector WebIDE

| **开发语言** | **镜像类型** | **tag**| **Pull命令**| **new指令**| **备注**|
|----------|----------|---------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------|----------------------------|-------------------------------------------------------------|
| node     | WebStorm | 2021.3.2-2834,2021.3.2-latest,latest              | `docker pull registry.cn-hangzhou.aliyuncs.com/smartide/smartide-node-v2-jetbrains-webstorm:2021.3.2-2834`            | `se new node -t webstorm`  | 在SDK镜像的基础上集成WebStorm V2021.3.2 WebIDE                       |
| Java     | IDEA     | 2021.2.3-openjdk-11-jdk-2832,2021.2.3-2832,2021.2.3-latest,latest | `docker pull registry.cn-hangzhou.aliyuncs.com/smartide/smartide-java-v2-jetbrains-idea:2021.2.3-openjdk-11-jdk-2832` | `se new java -t idea`      | 在SDK镜像的基础上集成IDEA社区版 V2021.2.3 WebIDE                        |
| golang   | Goland   | 2021.3.3-2830,2021.3.3-latest,latest              | `docker pull registry.cn-hangzhou.aliyuncs.com/smartide/smartide-golang-v2-jetbrains-goland:2021.3.3-2830`            | `se new golang -t goland`  | 在SDK镜像的基础上集成Goland  V2021.3.3 WebIDE                        |
| python   | Pycharm  | all-version,2021.2.3-2850,latest                  | `docker pull registry.cn-hangzhou.aliyuncs.com/smartide/smartide-python-v2-jetbrains-pycharm:,2021.2.3-2850`          | `se new python -t pycharm` | 在SDK镜像的基础上集成 Pycharm  V2021.2.5 WebIDE                      |
| dotnet   | Rider    | 6.0,2021.3.3-2828,latest                          | `docker pull registry.cn-hangzhou.aliyuncs.com/smartide/smartide-dotnet-v2-jetbrains-rider:2021.3.3-2828`             | `se new dotnet -t rider`   | 在SDK镜像的基础上集成 Rider V2021.3.3 WebIDE                         |
| php      | PhpStorm | 2021.3.2-php7.4-2837,2021.3.2-2837,2021.3.2-latest,latest       | `docker pull registry.cn-hangzhou.aliyuncs.com/smartide/smartide-php-v2-jetbrains-phpstorm:2021.3.2-php7.4-2837`      | `se new php -t phpstorm`   | 在SDK镜像的基础上集成PhpStorm社区版 V2021.2.7 WebIDE                    |
| C++ | Clion  | clang,2021.3.3-2827,latest | `docker pull registry.cn-hangzhou.aliyuncs.com/smartide/smartide-cpp-v2-jetbrains-clion:2021.3.3-2827` | `se new cpp -t clion`  | 在SDK镜像的基础上集成 Clion V2021.3.3 WebIDE  |

OpenSumi WebIDE

| **开发语言** | **镜像类型** | **tag**| **Pull命令**| **new指令**| **备注**|
|----------|----------|---------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------|----------------------------|-------------------------------------------------------------|
| node     | OpenSumi | 2887,all-version,latest              | `docker pull registry.cn-hangzhou.aliyuncs.com/smartide/smartide-node-v2-opensumi:2887`            | `se new node -t opensumi`  | 在SDK镜像的基础上集成OpenSumi WebIDE                       |