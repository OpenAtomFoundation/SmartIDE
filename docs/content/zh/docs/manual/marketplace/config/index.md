---
title: "配置连接手册"
linkTitle: "配置连接手册"
weight: 41
description: >
    本文档描述如何更新Visual Studio Code以及兼容IDE的配置文件连接到SmartIDE Marketplace，包括：VSCode, Codium, Code Server, OpenVSCode Server和OpenSumi。
---

## 1. 原理
参考自[Using-Open-VSX-in-VS-Code](https://github.com/eclipse/openvsx/wiki/Using-Open-VSX-in-VS-Code)，通过修改IDE product.json 文件中 extensionsGallery.serviceUrl & extensionsGallery.itemUrl & linkProtectionTrustedDomains 节点的值，让IDE链接应用市场时改变API指向达到链接SmartIDE Marketplace的目的。

    "extensionsGallery": {
        "serviceUrl": "https://marketplace.smartide.cn/vscode/gallery",
        "itemUrl": "https://marketplace.smartide.cn/vscode/item"
    }

    "linkProtectionTrustedDomains": [
        "https://marketplace.smartide.cn"
    ]

## 2. 配置
以 Visual Studio Code 为例展示配置过程：
- 打开 VSCode 安装目录按照如下路径找到 product.json
![](./images/marketplace-config-01.jpg)
- 关闭VSCode 正在运行的进程，用其他编辑器打开 product.json，并参照第1章节修改对应内容
![](./images/marketplace-config-02.jpg)
![](./images/marketplace-config-03.jpg)
- 打开 VSCode 进入扩展页面查看插件市场
![](./images/marketplace-config-04.jpg)