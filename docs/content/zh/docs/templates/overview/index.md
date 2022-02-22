---
title: "概述"
linkTitle: "概述"
weight: 10
description: >
  镜像和模版概述
---

## 开发镜像

smartide 提供了 各主流语言的SDK 镜像，包括 node、java、golang、python、dotnet、php、C++、，以及集成VSCode和JetBrains 各语言对应的WebIDE。开发者使用这些镜像可以一键创建自己需要的开发环境。镜像托管至阿里云的容器镜像服务中地址为：`registry.cn-hangzhou.aliyuncs.com/smartide/<镜像名称:TAG>`


## 模版库

使用smartide 提供的模板和镜像，开发人员可以快速的创建对应语言的开发环境，可以是新项目，也可以用于适配已有项目，使用方式参考下表new指令一列。
模板地址托管至gitee中，地址为：https://gitee.com/smartide/smartide-templates


## 镜像和模版指令列表



| **开发语言** | **镜像类型** | **tag**| **Pull命令**| **new指令**| **备注**|
|----------|----------|---------------------------------------------------|-----------------------------------------------------------------------------------------------------------------------|----------------------------|-------------------------------------------------------------|
| base     | 基础       | 1729,latest                                       | `docker pull registry.cn-hangzhou.aliyuncs.com/smartide/smartide-base-v2:1729`                                        | `se new base`              | 基于ubuntu:20.04，集成git、ssh server等基础库                         |
| node     | SDK      | all-version,1725,latest                           | `docker pull registry.cn-hangzhou.aliyuncs.com/smartide/smartide-node-v2:1725`                                        | `se new node`              | 在base 镜像的基础上，集成了Node V14.17.6(默认)、V12.22.7 V16.7.0 SDK及nvm  |
| node     | VSCode   | all-version,1748,latest                           | `docker pull registry.cn-hangzhou.aliyuncs.com/smartide/smartide-node-v2-vscode:1748`                                 | `se new node -t vscode`    | 在SDK镜像的基础上集成VSCode WebIDE                                   |
| node     | WebStorm | 2021.3.2-2188,2021.3.2-latest,latest              | `docker pull registry.cn-hangzhou.aliyuncs.com/smartide/smartide-node-v2-jetbrains-webstorm:2021.3.2-2188`            | `se new node -t webstorm`  | 在SDK镜像的基础上集成WebStorm V2021.3.2 WebIDE                       |
| Java     | SDK      | openjdk-11-jdk,2098,latest                        | `docker pull registry.cn-hangzhou.aliyuncs.com/smartide/smartide-java-v2:2098`                                        | `se new java`              | 在Node SDK 镜像的基础上，集成Java Open JDK 11及maven                   |
| Java     | VSCode   | openjdk-11-jdk,latest,2192                        | `docker pull registry.cn-hangzhou.aliyuncs.com/smartide/smartide-java-v2-vscode:2192`                                 | `se new java -t vscode`    | 在SDK镜像的基础上集成VSCode WebIDE                                   |
| Java     | IDEA     | 2021.2.3-openjdk-11-jdk-2081,2021.2.3-2081,latest | `docker pull registry.cn-hangzhou.aliyuncs.com/smartide/smartide-java-v2-jetbrains-idea:2021.2.3-openjdk-11-jdk-2081` | `se new java -t idea`      | 在SDK镜像的基础上集成IDEA社区版 V2021.2.3 WebIDE                        |
| golang   | SDK      | 1.17.5,1746,latest; 1.16.12,1745                  | `docker pull registry.cn-hangzhou.aliyuncs.com/smartide/smartide-golang-v2:1746`                                      | `se new golang`            | 在Node SDK 镜像的基础上，集成Go SDK，分为1.17.5、1.16.12两个版本              |
| golang   | VSCode   | 1.17.5,1749,latest;1.16.12,1747                   | `docker pull registry.cn-hangzhou.aliyuncs.com/smartide/smartide-golang-v2-vscode:1749`                               | `se new golang -t vscode`  | 在SDK镜像的基础上集成VSCode WebIDE                                   |
| golang   | Goland   | 2021.3.3-2191,2021.3.3-latest,latest              | `docker pull registry.cn-hangzhou.aliyuncs.com/smartide/smartide-golang-v2-jetbrains-goland:2021.3.3-2191`            | `se new golang -t goland`  | 在SDK镜像的基础上集成Goland  V2021.3.3 WebIDE                        |
| python   | SDK      | all-version,2197,latest                           | `docker pull registry.cn-hangzhou.aliyuncs.com/smartide/smartide-python-v2:2197`                                      | `se new python`            | 在Node SDK 镜像的基础上，集成python2和python3                          |
| python   | VSCode   | all-version,2198,latest                           | `docker pull registry.cn-hangzhou.aliyuncs.com/smartide/smartide-python-v2-vscode:2198`                               | `se new python -t vscode`  | 在SDK镜像的基础上集成VSCode WebIDE                                   |
| python   | Pycharm  | all-version,2021.2.3-2202,latest                  | `docker pull registry.cn-hangzhou.aliyuncs.com/smartide/smartide-python-v2-jetbrains-pycharm:,2021.2.3-2201`          | `se new python -t pycharm` | 在SDK镜像的基础上集成 Pycharm  V2021.2.5 WebIDE                      |
| dotnet   | SDK      | 6.0,2141,latest                                   | `docker pull registry.cn-hangzhou.aliyuncs.com/smartide/smartide-dotnet-v2:2141`                                      | `se new dotnet`            | 在Node SDK 镜像的基础上，集成Net6.0 SDK 和asp.net core                 |
| dotnet   | VSCode   | 6.0,2143,latest                                   | `docker pull registry.cn-hangzhou.aliyuncs.com/smartide/smartide-dotnet-v2-vscode:2143`                               | `se new dotnet -t vscode`  | 在SDK镜像的基础上集成VSCode WebIDE                                   |
| dotnet   | Rider    | 6.0,2021.3.3-2142,latest                          | `docker pull registry.cn-hangzhou.aliyuncs.com/smartide/smartide-dotnet-v2-jetbrains-rider:2021.3.3-2142`             | `se new dotnet -t rider`   | 在SDK镜像的基础上集成 Rider V2021.3.3 WebIDE                         |
| php      | SDK      | php7.4,2107,latest                                | `docker pull registry.cn-hangzhou.aliyuncs.com/smartide/smartide-php-v2:`                                             | `se new php`               | 在Node SDK 镜像的基础上，集成php7.4和Apache2                           |
| php      | VSCode   | php7.4,2108                                       | `docker pull registry.cn-hangzhou.aliyuncs.com/smartide/smartide-php-v2-vscode:2108`                                  | `se new php -t vscode`     | 在SDK镜像的基础上集成VSCode WebIDE                                   |
| php      | PhpStorm | 2021.3.2-php7.4-2109,2021.3.2-latest,latest       | `docker pull registry.cn-hangzhou.aliyuncs.com/smartide/smartide-php-v2-jetbrains-phpstorm:2021.3.2-php7.4-2109`      | `se new php -t phpstorm`   | 在SDK镜像的基础上集成PhpStorm社区版 V2021.2.7 WebIDE                    |
| C++ | SDK    | clang,2156,latest          | `docker pull registry.cn-hangzhou.aliyuncs.com/smartide/smartide-cpp-v2:2156`                          | `se new cpp`           | 在Node SDK 镜像的基础上,集成cmake、clang       |
| C++ | VSCode | clang,2159,latest          | `docker pull registry.cn-hangzhou.aliyuncs.com/smartide/smartide-cpp-v2-vscode:2159`                   | `se new cpp -t vscode` | 在SDK镜像的基础上集成VSCode WebIDE            |
| C++ | Clion  | clang,2021.3.3-2160,latest | `docker pull registry.cn-hangzhou.aliyuncs.com/smartide/smartide-cpp-v2-jetbrains-clion:2021.3.3-2160` | `se new cpp -t clion`  | 在SDK镜像的基础上集成 Clion V2021.3.3 WebIDE  |


