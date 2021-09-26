---
title: "示例：安装手册"
linkTitle: "示例：安装手册"
date: 2021-09-24
description: >
  目前smartIDE支持Windows和macOS两种环境安装，如下文档分别介绍如何在这两种环境中安装.
---

## 先决条件

smartIDE 运行环境依赖docker ，推荐安装最新版docker，安装docker参考[文档](https://yeasy.gitbooks.io/docker_practice/install/)

## 安装

## Mac

### 使用curl下载

> 您可以运行如下命令下载安装最新版

```bash
curl -sSL  https://smartidedl.blob.core.chinacloudapi.cn/releases/$(curl -L -s https://smartidedl.blob.core.chinacloudapi.cn/releases/stable.txt)/smartide-osx-x64.zip | tar -xzC /usr/local/bin/

chmod +x /usr/local/bin/smartide
```

## windows

### 使用powershell 下载: Invoke-WebReques

> 您可以运行如下命令下载安装最新版

```powershell
Invoke-WebRequest -Uri ("https://smartidedl.blob.core.chinacloudapi.cn/releases/"+(Invoke-RestMethod https://smartidedl.blob.core.chinacloudapi.cn/releases/stable.txt)+"/SetupSmartIDE.msi")  -OutFile "smartide.msi"

 .\smartIDE.msi
 
```
