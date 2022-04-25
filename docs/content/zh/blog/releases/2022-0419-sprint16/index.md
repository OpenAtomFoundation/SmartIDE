---
title: "SmartIDE v0.1.16 已经发布 - 支持自阿里&蚂蚁开源的国产 IDE OpenSumi"
linkTitle: "v0.1.16 OpenSumi"
date: 2022-04-19
description: >
  支持自阿里&蚂蚁开源的国产 IDE OpenSumi
---

[SmartIDE v0.1.16 (Build 3137)](/zh/docs/install/) 已经在**2022年4月19日**发布到稳定版通道，我们在这个版本中增加了阿里和蚂蚁发布的国产IDE OpenSumi的支持，以及其他一些改进。SmartIDE 从 Sprint 11 (v0.1.11) 开始已经将重心转向 [Server版](/zh/docs/quickstart/server/) 的开发，并且已经针对社区开放了server的内测。但是对于 [CLI](/zh/docs/quickstart/cli/) 的改进和增强一直没有停止，因为 CLI 是 SmartIDE 的核心，实际上我们的 Server 版对于 [工作区](/zh/docs/overview/workspace/) 的管理也是通过云原生开源流水线框架 Tekton 调度 CLI 实现的。

我们将在近期发布更加完善的 Server 版安装部署手册和文档，同时 Server 版 和 CLI 核心代码也将在近期开源。SmartIDE 的核心代码将采用GPL协议开源，允许任何组织和个人免费使用我们的代码搭建自己的云原生IDE环境。

## 来自阿里&蚂蚁的国产IDE - OpenSumi 简介

严格来说，阿里的 [OpenSumi](https://opensumi.com/) 并不是一个IDE产品，而是一个IDE二次开发框架。这个定位与 [Eclipse Cheia](https://theia-ide.org/) 的定位相同。SmartIDE 的早期版本也支持 Eclipse Theia，但是由于其操作体验与VSCode还是存在一定的差距，后续我们将重心转向类VSCode的IDE支持，比如对 [OpenVSCode Server](https://github.com/gitpod-io/openvscode-server) 的支持，以及 [JetBrains](https://www.jetbrains.com/) 系列IDE全家桶的支持。阿里&蚂蚁的开发团队在2022年3月3日发布了OpenSumi以后，SmartIDE团队对这款IDE进行了研究，认为可以替代Eclipse Theia 作为未来提供 “定制化IDE” 解决方案的基座，因此将重心转向了对 OpenSumi的支持，按照阿里&蚂蚁相关文章的说明：

> *“OpenSumi 是一款面向垂直领域，低门槛、高性能、高定制性的双端（Web 及 Electron）IDE 研发框架，基于 TypeScript+React 进行编码，实现了包含资源管理器、编辑器、调试、Git 面板、搜索⾯板等核新功能模块。开发者只要基于起步项目进行简单配置，就可以快速搭建属于自己的本地或云端 IDE 产品。”*  -- [原文链接](https://mp.weixin.qq.com/s/wVXCOO8WloKs-LWERA2_vA)

![OpenSumi官网](images/opensumi000.png)

## 如何使用 SmartIDE 启动OpenSumi WebIDE 

OpenSumi的定位非常符合SmartIDE对IDE定制化解决方案的需求，因此我们针对OpenSumi进行了适配和集成，开发者可以使用一个非常简单的指令即可在浏览器中启动一个基于OpenSumi WebIDE 的 node.js 开发环境，具体请参考 [Node快速启动](/zh/docs/quickstart/node/#opensumi) 文档

```shell
## 使用OpenSumi WebIDE开启Node开发环境
smartide new node -t opensumi
```

以下是处于单步调试状态的 OpenSumi WebIDE

![OpenSumi调试状态](images/opensumi001.png)

或者也可以通过我们的 计算器 示例应用体验使用OpenSumi开发调试Node.js应用的过程：

```shell
## 使用OpenSumi调试计算器示例
smartide start https://gitee.com/idcf-boat-house/boathouse-calculator.git --filepath .ide/opensumi.ide.yaml
```
以下是正在单步调试 计算器示例应用 的OpenSumi WebIDE，[B站视频](https://www.bilibili.com/video/bv14Y4y187hC)

{{< bilibili 641037405 >}}

感谢你对SmartIDE的关注，欢迎从SmartIDE官网下载体验我们的产品，或者加入我们的早鸟群，及时了解SmartIDE的开发进展。




