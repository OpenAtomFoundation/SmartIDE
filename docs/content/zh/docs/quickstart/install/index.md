---
title: "安装手册"
linkTitle: "安装手册"
date: 2021-09-24
weight: 1
description: >
  本文档介绍如何完成SmartIDE的安装。当前SmartIDE支持MacOS和Windows两种操作系统，我们提供了一键安装脚方便开发人员快速完成安装。
---

## 先决条件

使用SmartIDE需要首先安装 Docker Desktop 工具，请从以下网盘地址下载最新版的 Docker Desktop 安装包 

> 链接: https://pan.baidu.com/s/1xQcif-oeVNNzonawywK7Lw 提取码: vu8c 

下载完成后请参考以下文档完成 Docker Desktop 的安装：

- [在 Windows 上安装 Docker Desktop](docker-install-windows)
- [在 MacOS 上安装 Docker Desktop](docker-install-osx)

## SmartIDE 安装手册

我们按照敏捷开发模式进行SmartIDE的开发，所有的版本都通过CI/CD流水线自动构建，打包，测试和发布。为了同时满足外部用户对于稳定性的要求和开发团队以及早期使用者对新功能快速更新的要求，我们提供以下两个发布通道。

### 稳定版通道

稳定版的发布按照sprint进行，我们采用2周一个sprint的开发节奏，每个sprint结束后我们会发布一个稳定版到这个通道。这个版本经过开发团队相对完整的测试，确保所提供的功能稳定可用，同时我们会在每个sprint结束时同时发布“版本发布说明”并对当前版本的性能和改进进行说明。

版本发布说明列表：

| 版本号      | 构建编号 | 发布日期      |   简要说明   |
| ----------- | ----------- | ----------- | ----------- |
| [v0.1.5](/zh/blog/2021-1105-vm-start/)          | 819 | 2021.11.05 | 完善本地模式和远程主机模式下对Git的支持       |
| [v0.1.2](/zh/blog/2021-1024-first-release/)     | 933 | 2021.10.24 | 第一个公开发行版，本地模式       |

{{< tabs name="stable_install" >}}
{{% tab name="MacOS" %}}
```bash
# SmartIDE 稳定版通道安装脚本
# 打开终端窗口，复制粘贴以下脚本即可安装稳定版SmartIDE CLI应用
# 再次执行此命令即可更新版本
curl -OL  "https://smartidedl.blob.core.chinacloudapi.cn/releases/$(curl -L -s https://smartidedl.blob.core.chinacloudapi.cn/releases/stable.txt)/smartide" \
&& mv -f smartide /usr/local/bin/smartide \
&& chmod +x /usr/local/bin/smartide
```
{{% /tab %}}
{{% tab name="Windows" %}}
```powershell
# SmartIDE 稳定版通道安装脚本
# 打开PowerShell终端窗口，复制粘贴以下脚本即可自动下载稳定版SmartIDE MSI安装包，并启动安装程序
# 再次执行此命令即可更新版本
Invoke-WebRequest -Uri ("https://smartidedl.blob.core.chinacloudapi.cn/releases/"+(Invoke-RestMethod https://smartidedl.blob.core.chinacloudapi.cn/releases/stable.txt)+"/SetupSmartIDE.msi")  -OutFile "smartide.msi"
 .\smartIDE.msi
```
{{% /tab %}}
{{< /tabs >}}

### 每日构建版通道

SmartIDE CI/CD 流水线每晚8点（GMT+8 Beijing) 会自动将当天提交到sprint分支上的代码进行自动化构建和测试，并发布到每日构建版通道；产品组每天早上会对这个版本进行冒烟测试，确保基本功能可以正常运行，如果在冒烟测试过程中发现任何问题，我们会优先进行修复并立即通过这个通道发布修复后的版本。

每日构建版本通道的目标用户是产品组成员和早期试用社区用户，大家可以通过我们的SmartIDE GitHub首页提交issue或者通过微信公众号，B站等社区渠道提供反馈给我们。我们非常希望得到社区的反馈，并会尽力为社区用户提供最快速度的响应。

{{< tabs name="daily_install" >}}
{{% tab name="MacOS" %}}
```bash
# SmartIDE 每日构建版通道安装脚本
# 打开终端窗口，复制粘贴以下脚本即可安装每日构建版SmartIDE CLI应用
# 再次执行此命令即可更新版本
curl -OL  "https://smartidedl.blob.core.chinacloudapi.cn/builds/$(curl -L -s https://smartidedl.blob.core.chinacloudapi.cn/builds/stable.txt)/smartide" \
&& mv -f smartide /usr/local/bin/smartide \
&& chmod +x /usr/local/bin/smartide
```
{{% /tab %}}

{{% tab name="Windows" %}}
```powershell
# SmartIDE 每日构建版通道安装脚本
# 打开PowerShell终端窗口，复制粘贴以下脚本即可自动下载每日构建版SmartIDE MSI安装包，并启动安装程序
# 再次执行此命令即可更新版本
Invoke-WebRequest -Uri ("https://smartidedl.blob.core.chinacloudapi.cn/builds/"+(Invoke-RestMethod https://smartidedl.blob.core.chinacloudapi.cn/builds/stable.txt)+"/SetupSmartIDE.msi")  -OutFile "smartide.msi"
 .\smartIDE.msi
```
{{% /tab %}}

{{< /tabs >}}

## 获取并理解版本号

安装好SmartIDE后，您可以通过以下命令获取当前版本号

```shell
# 键入以下命令获取当前版本
smartide version
# 输出如下
v0.1.6.992
Version number: v0.1.6.992
Build number: 20211113.2_992_release-6_Schedule
Commit record: 783718398ed96d9d07714715fede2709fc405485
Company: leansoftX.com
```

说明：

- Version number: 代表当前版本号，格式：[主版本].[小版本].[Sprint编号].[构建号]
- Build number: CI/CD流水线的完整构建编码，格式：[日期].[序号]_构建号_分支名_构建类型
- Commit record: 当前版本所对应的git commit hash
- Company: 发行商名称



