---
title: "安装手册"
linkTitle: "安装手册"
date: 2021-09-24
weight: 1
description: >
  目前smartIDE支持Windows和macOS两种环境安装，如下文档分别介绍如何在这两种环境中安装.
---

## 先决条件

使用SmartIDE需要首先安装 Docker Desktop 工具，请从以下网盘地址下载最新版的 Docker Desktop 安装包 

> 链接: https://pan.baidu.com/s/1xQcif-oeVNNzonawywK7Lw 提取码: vu8c 

下载完成后请参考以下文档完成 Docker Desktop 的安装：

- [在 Windows 上安装 Docker Desktop](docker-install-windows)
- [在 Linux 上安装 Docker Desktop](docker-install-osx)

## SmartIDE 安装手册

### Mac

**使用curl下载并安装**

> 您可以运行如下命令下载安装最新版

```bash
curl -sSL  https://smartidedl.blob.core.chinacloudapi.cn/releases/$(curl -L -s https://smartidedl.blob.core.chinacloudapi.cn/releases/stable.json | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')/smartide-osx-$(curl -L -s https://smartidedl.blob.core.chinacloudapi.cn/releases/stable.json | grep '"build_number"' | sed -E 's/.*"([^"]+)".*/\1/')-x64.zip | tar -xzC /usr/local/bin/

chmod +x /usr/local/bin/smartide
```

### windows

**使用powershell 下载并安装**

> 您可以运行如下命令下载安装最新版

```powershell
Invoke-WebRequest -Uri ("https://smartidedl.blob.core.chinacloudapi.cn/releases/"+(Invoke-RestMethod https://smartidedl.blob.core.chinacloudapi.cn/releases/stable.json).tag_name+"/smartide-win-"+(Invoke-RestMethod https://smartidedl.blob.core.chinacloudapi.cn/releases/stable.json).build_number+"-x64.zip")  -OutFile "smartide.msi"

Expand-Archive -LiteralPath './smartide.zip' -DestinationPath .

 .\smartIDE.msi
```
