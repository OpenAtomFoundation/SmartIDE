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
- [在 MacOS 上安装 Docker Desktop](docker-install-osx)

## SmartIDE 安装手册

### Mac

**使用curl下载并安装**

> 最新稳定版安装命令

```bash
curl -OL  "https://smartidedl.blob.core.chinacloudapi.cn/releases/$(curl -L -s https://smartidedl.blob.core.chinacloudapi.cn/releases/stable.txt)/smartide" \
&& mv -f smartide /usr/local/bin/smartide \
&& chmod +x /usr/local/bin/smartide
```

> 每日构建版本安装命令

```bash
curl -OL  "https://smartidedl.blob.core.chinacloudapi.cn/builds/$(curl -L -s https://smartidedl.blob.core.chinacloudapi.cn/builds/stable.txt)/smartide" \
&& mv -f smartide /usr/local/bin/smartide \
&& chmod +x /usr/local/bin/smartide
```


### windows

**使用powershell 下载并安装**

> 最新稳定版安装命令

```powershell
Invoke-WebRequest -Uri ("https://smartidedl.blob.core.chinacloudapi.cn/releases/"+(Invoke-RestMethod https://smartidedl.blob.core.chinacloudapi.cn/releases/stable.txt)+"/SetupSmartIDE.msi")  -OutFile "smartide.msi"

 .\smartIDE.msi
```

> 每日构建版本安装命令

```powershell
Invoke-WebRequest -Uri ("https://smartidedl.blob.core.chinacloudapi.cn/builds/"+(Invoke-RestMethod https://smartidedl.blob.core.chinacloudapi.cn/builds/stable.txt)+"/SetupSmartIDE.msi")  -OutFile "smartide.msi"

 .\smartIDE.msi
```

