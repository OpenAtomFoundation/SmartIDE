---
title: "CLI 安装说明"
linkTitle: "CLI"
date: 2021-09-24
weight: 20
description: >
  本文档介绍如何完成SmartIDE CLI的安装。当前SmartIDE CLI支持MacOS和Windows两种操作系统，我们提供了一键安装脚本方便开发人员快速完成安装。
---

## 先决条件

SmartIDE CLI 通过调用docker和docker-compose来实现容器化环境的管理，因此你需要确保在自己的环境中已经正确安装了这两个软件。如果你已经在使用这两个工具，那么可以直接进入以下 [SmartIDE 安装手册](#smartide-安装手册) 部分。

> **说明：** 如果你只需要使用smartide远程主机模式，那么本地开发机上并不需要安装 Docker Desktop 工具，因为所有的环境调度都在远程主机上完成，但是你仍然需要确保远程主机上已经正确安装了docker和docker-compose工具。

如果您需要docker和docker-compose的安装指导，可以参考以下链接：

- [在 Windows 上安装 Docker 桌面版 (包含Docker-Compose)](/zh/docs/install/docker/windows/)
- [在 MacOS 上安装 Docker 桌面版 (包含Docker-Compose)](/zh/docs/install/docker/osx/)
- [在 Linux 服务器上安装 Docker 和 Docker-Compose](/zh/docs/install/docker/linux/)

> **说明：** 我们在国内提供了安装包镜像地址以及一键安装脚本以便提升大家的安装体验，具体请参考以上文档中的说明。

## SmartIDE 安装手册

我们按照敏捷开发模式进行SmartIDE的开发，所有的版本都通过CI/CD流水线自动构建，打包，测试和发布。为了同时满足外部用户对于稳定性的要求和开发团队以及早期使用者对新功能快速更新的要求，我们提供以下两个发布通道。

### 稳定版通道

稳定版的发布按照sprint进行，我们采用2周一个sprint的开发节奏，每个sprint结束后我们会发布一个稳定版到这个通道。这个版本经过开发团队相对完整的测试，确保所提供的功能稳定可用，同时我们会在每个sprint结束时同时发布“版本发布说明”并对当前版本的性能和改进进行说明。

流水线状态 
[![Build Status](https://dev.azure.com/leansoftx/smartide/_apis/build/status/smartide-codesign-ci?branchName=main)](https://dev.azure.com/leansoftx/smartide/_build/latest?definitionId=32&branchName=main)

版本发布说明列表：

| 版本号      | 构建编号 | 发布日期      |   简要说明   |
| ----------- | ----------- | ----------- | ----------- |
| [v0.1.18](/zh/blog/2022-0519-sprint18/)     | 3538 | 2022.05.19 | 增加CLi K8s模式，并支持主机资源非22端口连接。插件市场历史版本同步，阿里 OpenSumi使用SmartIDE插件市场作为官方市场。修复了若干Bug。  |
| [v0.1.17](/zh/blog/2022-0507-sprint17/)     | 3332 | 2022.05.09 | 模版支持远程主机并集成到Server，插件市场公测  |
| [v0.1.16](/zh/blog/2022-0419-sprint16/)     | 3137 | 2022.04.19 | 修复若干bug，增加工作区日志支持，K8S模式完善，优化基础镜像启动等。  |
| v0.1.15     | 2980 | 2022.04.07 | 增加了对阿里 & 蚂蚁自研开源 IDE 研发框架 OpenSumi 的支持、客户端K8S启动模式，并且完善了工作区全生命周期的状态控制，以及修复了若干Bug。       |
| [v0.1.14](/zh/blog/2022-0309-sprint14/)     | 2505 | 2022.03.21 | 修复若干bug，Linux 版 SmartIDE CLI 工具发布，完成Server版开发集成       |
| v0.1.11     | 2035 | 2022.02.08 | 修复若干bug，启动Server版开发集成       |
| v0.1.10     | 1829 | 2022.01.18 | 简化版指令集，k8s start可用，开源项目gin-vue-admin适配       |
| [v0.1.9](/zh/blog/2022-0104-sprint9/)          | 1753 | 2022.01.01 | 稳定性和体验性加强，k8s Start, Image V2 体系开发，JetBrain体系WebIDE支持       |
| [v0.1.7](/zh/blog/2021-1203-state-management/)          | 1456 | 2021.12.02 | 公开发布，状态管理，远程模式稳定性，项目模板       |
| [v0.1.5](/zh/blog/2021-1105-vm-start/)          | 819 | 2021.11.05 | 完善本地模式和远程主机模式下对Git的支持       |
| [v0.1.2](/zh/blog/2021-1024-first-release/)     | 933 | 2021.10.24 | 第一个公开发行版，本地模式       |

{{< tabs name="stable_install" >}}
{{% tab name="MacOS" %}}
```bash
# SmartIDE 稳定版通道安装脚本
# 打开终端窗口，复制粘贴以下脚本即可安装稳定版SmartIDE CLI应用
# 再次执行此命令即可更新版本
curl -OL  "https://smartidedl.blob.core.chinacloudapi.cn/releases/$(curl -L -s https://smartidedl.blob.core.chinacloudapi.cn/releases/stable.txt)/smartide-osx" \
&& mv -f smartide-osx /usr/local/bin/smartide \
&& ln -s -f /usr/local/bin/smartide /usr/local/bin/se \
&& chmod +x /usr/local/bin/smartide
```
{{% /tab %}}
{{% tab name="Linux" %}}
```bash
# SmartIDE 每日构建版通道安装脚本
# 打开终端窗口，复制粘贴以下脚本即可安装每日构建版SmartIDE CLI应用
# 再次执行此命令即可更新版本
curl -OL  "https://smartidedl.blob.core.chinacloudapi.cn/releases/$(curl -L -s https://smartidedl.blob.core.chinacloudapi.cn/releases/stable.txt)/smartide-linux" \
&& sudo mv -f smartide-linux /usr/local/bin/smartide \
&& sudo ln -s -f /usr/local/bin/smartide /usr/local/bin/se \
&& sudo chmod +x /usr/local/bin/smartide
```
{{% /tab %}}
{{% tab name="Windows" %}}
```powershell
# SmartIDE 稳定版通道安装脚本
# 打开PowerShell终端窗口，复制粘贴以下脚本即可自动下载稳定版SmartIDE MSI安装包，并启动安装程序
# 再次执行此命令即可更新版本
# 如果是第一次在Windows上安装，请重新打开PowerShell窗口以便PATH设置生效
Invoke-WebRequest -Uri ("https://smartidedl.blob.core.chinacloudapi.cn/releases/"+(Invoke-RestMethod https://smartidedl.blob.core.chinacloudapi.cn/releases/stable.txt)+"/SetupSmartIDE.msi")  -OutFile "smartide.msi"
 .\smartIDE.msi
```
{{% /tab %}}
{{< /tabs >}}

### 每日构建版通道

SmartIDE CI/CD 流水线每晚8点（GMT+8 Beijing) 会自动将当天提交到sprint分支上的代码进行自动化构建和测试，并发布到每日构建版通道；产品组每天早上会对这个版本进行冒烟测试，确保基本功能可以正常运行，如果在冒烟测试过程中发现任何问题，我们会优先进行修复并立即通过这个通道发布修复后的版本。

每日构建版本通道的目标用户是产品组成员和早期试用社区用户，大家可以通过我们的SmartIDE GitHub首页提交issue或者通过微信公众号，B站等社区渠道提供反馈给我们。我们非常希望得到社区的反馈，并会尽力为社区用户提供最快速度的响应。

流水线状态 
[![Build Status](https://dev.azure.com/leansoftx/smartide/_apis/build/status/smartide-codesign-ci?branchName=releases/release-8)](https://dev.azure.com/leansoftx/smartide/_build/latest?definitionId=32&branchName=releases/release-8)

{{< tabs name="daily_install" >}}
{{% tab name="MacOS" %}}
```bash
# SmartIDE 每日构建版通道安装脚本
# 打开终端窗口，复制粘贴以下脚本即可安装每日构建版SmartIDE CLI应用
# 再次执行此命令即可更新版本
curl -OL  "https://smartidedl.blob.core.chinacloudapi.cn/builds/$(curl -L -s https://smartidedl.blob.core.chinacloudapi.cn/builds/stable.txt)/smartide-osx" \
&& mv -f smartide-osx /usr/local/bin/smartide \
&& ln -s -f /usr/local/bin/smartide /usr/local/bin/se \
&& chmod +x /usr/local/bin/smartide
```
{{% /tab %}}

{{% tab name="Linux" %}}
```bash
# SmartIDE 每日构建版通道安装脚本
# 打开终端窗口，复制粘贴以下脚本即可安装每日构建版SmartIDE CLI应用
# 再次执行此命令即可更新版本
curl -OL  "https://smartidedl.blob.core.chinacloudapi.cn/builds/$(curl -L -s https://smartidedl.blob.core.chinacloudapi.cn/builds/stable.txt)/smartide-linux" \
&& sudo mv -f smartide-linux /usr/local/bin/smartide \
&& sudo ln -s -f /usr/local/bin/smartide /usr/local/bin/se \
&& sudo chmod +x /usr/local/bin/smartide
```
{{% /tab %}}

{{% tab name="Windows" %}}
```powershell
# SmartIDE 每日构建版通道安装脚本
# 打开PowerShell终端窗口，复制粘贴以下脚本即可自动下载每日构建版SmartIDE MSI安装包，并启动安装程序
# 再次执行此命令即可更新版本
# 如果是第一次在Windows上安装，请重新打开PowerShell窗口以便PATH设置生效
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
v0.1.11.2035
版本号: v0.1.11.2035
构建号: 20220208.2_2035_main_IndividualCI
编译时间: 2022-02-08 01:15:49
提交记录: 2030b4ab1492c7afcc13f918a782b06c36f11a6d
```

说明：

- Version number: 代表当前版本号，格式：[主版本].[小版本].[Sprint编号].[构建号]
- Build number: CI/CD流水线的完整构建编码，格式：[日期].[序号]_构建号_分支名_构建类型
- Commit record: 当前版本所对应的git commit hash
- Company: 发行商名称



