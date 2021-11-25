---
title: "Node.Js 快速启动教程"
linkTitle: "node.js"
weight: 30
description: >
  本文档描述如何使用SmartIDE完成一个Node Express应用的完整开发，调试和代码提交过程。
---

SmartIDE内置了node.js环境模版，你可以通过一个简单的指令创建内置了WebIDE的开发环境，并立即开始编码和调试。

如果你还没有完成SmartIDE安装，请参考以下文档安装SmartIDE命令行工具。

- [SmartIDE 安装手册](/zh/docs/install)

> 说明：SmartIDE的命令行工具可以在Windows和MacOS操作系统上运行，对大多数命令来说，操作是完全一致的。为了方便不同平台的开发者更容易使用本文档，我们提供了两种平台的操作指导和截图。

运行以下命令创建node开发环境：

```shell
mkdir smartide-node-quickstart
cd smartide-node-quickstart
smartide new node
```

运行后的效果如下

![node quickstart](images/quickstart-node001.png)

```shell
npm config set registry https://registry.npmmirror.com
```