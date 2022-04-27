---
title: "应用插件市场"
linkTitle: "应用插件市场"
weight: 60
draft: false
description: >
    SmartIDE Marketplace是基于 Eclipse OpenVSX server 开源项目的一个fork，我们针对中国开发者的使用习惯和网络状况对这个开源项目进行了本地化，包括：界面的中文翻译处理和将插件同步到中国大陆的地址上供开发者下载使用。同时，对于无法直接使用微软官方marketplace的类VSCode IDE来说，比如：Codium, Code Server 和 OpenVSCode Server这些VSCode fork，可以使用SmartIDE Marketplace作为自己的插件市场，方便这些工具的使用者获取与VSCode类似的插件安装体验。
---

现今非常多的企业开发者在使用VSCode作为自己的主力IDE工具，但是由于很多企业对开发者连接互联网有严格的限制，大多数企业开发者都在采用离线安装的方式来获取VSCode的插件，这样做不仅操作繁琐，而且没有办法及时获取插件的更新，同时也会对企业的研发管理带来直接的安全隐患。

SmartIDE Marketplace 的目标并不是替代微软的Marketplace或者 Eclipse 的 open-vsx.org 而是希望为国内的开发者以及企业内部的开发者提供一种安全可靠，而且高效的插件管理机制。

SmartIDE Marketplace 与 Eclipse OpenVSX 一样是开源项目，并且我们提供了国内Gitee的开源库地址与Github保持同步，开源库地址如下：
- [Github](https://github.com/SmartIDE/eclipse-openvsx)
- [Gitee](https://gitee.com/SmartIDE/eclipse-openvsx)

相关文档和常见问题如下：
- [部署手册](https://smartide.cn/zh/docs/install/marketplace/)
- [操作手册](https://smartide.cn/zh/docs/manual/marketplace/)
  - [配置连接手册](https://smartide.cn/zh/docs/manual/marketplace/config/) ：如何更新Visual Studio Code以及兼容IDE的配置文件连接到SmartIDE Marketplace，包括：VSCode, Codium, Code Server, OpenVSCode Server和OpenSumi 
  - [插件安装手册](https://smartide.cn/zh/docs/manual/marketplace/usage/)：如何使用SmartIDE Marketplace安装插件 
  - [插件同步机制](https://smartide.cn/zh/docs/manual/marketplace/extension-sync/)：SmartIDE Marketplace 插件初始化同步机制
  - [插件发布手册](https://smartide.cn/zh/docs/manual/marketplace/publish-extension/)：如何发布插件到 SmartIDE Marketplace
- [私有化部署服务](https://smartide.cn/zh/services/)