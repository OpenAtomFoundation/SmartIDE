---
title: "SmartIDE - 开发从未如此简单"
linkTitle: "快速开始"
weight: 30
description: >
  使用SmartIDE，你可以在5分钟之内启动任何一个项目的编码调试，无需安装任何SDK，无需配置任何工具。
---

作为开发者，你无需了解什么是云，什么是容器，也无需学习复杂的docker命令，你所需要的就是学会一个简单的命令（smartide start），即可真正实现“一键启动”开发环境.

你也无需在本地安装IDE软件，只需要浏览器就够了。SmartIDE内置了Web版的vscode，你只需要打开浏览器就可以进行编码，使用智能提示，设置断点并且进行交互式单步调试，就如同使用一个全功能的IDE软件一样的体验。

{{% pageinfo %}}
*为了让你快速体验SmartIDE的快速开发体验，我们准备了一个示例应用 [Boathouse计算器](/zh/docs/examples/sample-calculator/)，无论你是否熟悉这个应用的代码，或者它所使用的 Node.Js 技术栈，你都可以在5分钟之内完成这个应用的开发和调试。*
{{% /pageinfo %}}

> 第一步：安装smartide cli工具

详细安装说明请参考 [安装手册](/zh/docs/install/)

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

> 第二步：克隆代码并运行 smartide start

```shell
git clone https://gitee.com/idcf-boat-house/boathouse-calculator.git
cd boathouse-calculator
smartide start
```
运行后的效果如下：
![smartide start](smartide-start.png)

> 第三步：设置断点、启动调试

SmartIDE会自动启动内置的WebIDE，你会看到一个类似vscode的IDE窗口在你的默认浏览器中出现。

现在你可以在WebIDE内置的Terminal中运行以下命令完成依赖项的安装

**说明**: npm是node.js的包管理器，npm install的作用是根据当前代码库的配置获取应用所需要的依赖包。一般来说，为了能够正确运行node.js应用，你首先需要安装node.js的sdk环境，但是SmartIDE已经为你完成了这个动作，作为开发者的你不再需要关心开发环境搭建的问题。

```shell
npm install
```
![npm install](npm-install.png)

完成以上操作后，你可以直接点击WebIDE左侧的调试工具栏，启动调试。

你也可以像我一样在 /api/controllers/arithmeticController.js 文件的第47行设置一个端点，并启动另外一个浏览器打开 http://localhost:3001 即可进入交互式调试体验。

![smartide debugging](smartide-debugging.png)

到这里，你已经完成了 [Boathouse计算器示例应用](/zh/docs/examples/sample-calculator/) 的开发调试过程，一切的操作就像在本地使用vscode一样的顺畅。

> 下一步

现在你已经完成了基本的SmartIDE操作，下一步你可以点击 **[安装手册](/zh/docs/install/)** 了解如何获取最新版的SmartIDE工具，以及如何进行更新。作为一款面向开发人员的工具，我们的更新速度非常快，基本上每天都会发布新版本。通过 **[安装手册](/zh/docs/install/)** 你可以详细了解我们的更新策略和不同更新通道的获取方式。

以上所展示的仅仅是 **[Boathouse计算器示例应用](/zh/docs/examples/sample-calculator/)** 的最简单开发场景，SmartIDE的更多使用场景包括：使用本地VSCode连接到SmartIDE工作区，使用远程主机作为工作区，GitConfig和SSH密钥同步等等，你可以点击这个链接了解完整的 **[Boathouse计算器示例应用](/zh/docs/examples/sample-calculator/)** 使用体验。

> 更多SmartIDE 开发示例

另外，我们还对很多常用的开发框架进行了SmartIDE适配，你可以通过 [示例应用](/zh/docs/examples/) 部分的文档获取更多示例应用的体验文档，这些示例包括：

- Vue Element Admin
- Element UI
- iTop

*更多示例会持续更新* 
