---
title: "示例：安装手册"
linkTitle: "示例：安装手册"
date: 2021-09-24
description: >
  A short lead description about this content page. It can be **bold** or _italic_ and can be split over multiple paragraphs.
---

## 先决条件

smartIDE 运行环境依赖docker ，推荐安装最新版docker，安装docker参考[文档](https://yeasy.gitbooks.io/docker_practice/install/)

## 安装

## Mac

### 使用curl下载

> 您可以运行如下命令下载安装最新版

```bash
curl -sSL  https://smartidedl.blob.core.chinacloudapi.cn/releases/$(curl -L -s https://smartidedl.blob.core.chinacloudapi.cn/releases/stable.json | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')/smartide-osx-$(curl -L -s https://smartidedl.blob.core.chinacloudapi.cn/releases/stable.json | grep '"build_number"' | sed -E 's/.*"([^"]+)".*/\1/')-x64.zip | tar -xzC /usr/local/bin/

chmod +x /usr/local/bin/smartide
```

## windows

### 使用powershell 下载: Invoke-WebReques

> 您可以运行如下命令下载安装最新版

```powershell
Invoke-WebRequest -Uri ("https://smartidedl.blob.core.chinacloudapi.cn/releases/"+(Invoke-RestMethod https://smartidedl.blob.core.chinacloudapi.cn/releases/stable.json).tag_name+"/smartide-win-"+(Invoke-RestMethod https://smartidedl.blob.core.chinacloudapi.cn/releases/stable.json).build_number+"-x64.zip")  -OutFile "smartide.msi"

Expand-Archive -LiteralPath './smartide.zip' -DestinationPath .

 .\smartIDE.msi
```
