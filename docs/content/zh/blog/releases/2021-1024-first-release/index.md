---
title: "SmartIDE v0.1.2 发布说明"
linkTitle: "版本 0.1.2 发布说明"
date: 2021-10-24
description: >
  SmartIDE的第一次版本发布。
---

各位开发者，大家好。在2021年10月24日程序员节这一天，SmartIDE的第一个公开发行版 v0.1.2 终于对外发布，我们并没有特意选择这个日子，但是冥冥之中我们的代码就把我们推送到了这个日子和大家见面，难道SmartIDE的代码是有生命的？


既然SmartIDE的代码是如此的懂得开发者，那么就让我们来认识一下这位新朋友。到底SmartIDE是谁？她能做些什么？

“人如其名”，SmartIDE是一个Smart的IDE（智能化集成开发环境）。你可能在想：好吧，又是一个新的IDE，我已经有了vscode/jetbrain全家桶，为什么我还需要另外一个IDE？

## 我们为什么要开发SmartIDE？

> 在当今这个软件吞噬世界的时代，每一家公司都是一家软件公司 ... ...

如今，软件确实已经深入我们生活的方方面面，没有软件你甚至无法点餐，无法购物，无法交水电费；我们生活的每一个环节都已经被软件包裹。在这些软件的背后是云计算，大数据和人工智能等等各种高新科技；这些现代IT基础设施在过去的5年左右获得了非常显著的发展，我们每个人都是这些高科技成果的受益者 ... ... 但是，为你提供这些高科技成果的开发者们自己所使用的软件（开发工具）却仍然像 “刀耕火种” 一般落后。

你可能会觉得这是危言耸听，那么让我来举一个简单的例子：大家一定都通过微信给自己的朋友发送过图片，这个过程非常简单，拿出手机，拍照，打开微信点击发送图片，完成。收到图片的朋友直接打开就可以看到你拍摄的照片了。但是对于开发者来说，如果要将一份代码快照发送给另外一位开发者，那么对方可能要用上几天的时间才能看到这份代码运行的样子。作为普通人，你恐怕无法理解这是为什么，如果你是一名开发者，你一定知道我在说什么！当然，我们也不指望普通人能够理解我们，对么？

![Dilbert的漫画](dilbert.png)

这样的场景是不是很熟悉？开发环境的搭建对于开发者来说理所当然的是要占用大量时间和精力的，但是对于 “产品经理/领导/老板/老婆/老妈/朋友” 来说，开始写代码就应该像打开Word写个文档一样简单，只有开发者自己知道这其实很不简单。

但是开发者已经有了非常好用的IDE了，Visual Studio Code, JetBrain全家桶都已经非常成熟，并不需要另外一个IDE了。确实，SmartIDE也并不是另外一个IDE，我们不会重新造轮子，我们只是希望你的轮子可以转的更快、更高效、更简单。

如果我们可以

```shell
git clone https://github.com/idcf-boat-house/boathouse-calculator.git
cd boathouse-calculator
smartide start
```

然后就可以进行开发和调试，是不是很爽？

![](smartide-sample-calcualtor.png)

图中重点：

- 通过右下角的的终端，你可以看到仅用一个简单的命令（smartide start）就完成了开发环境的搭建
- 在右上角的浏览器中运行着一个大家熟悉的Visual Studio Code，并且已经进入了单步调试状态，可以通过鼠标悬停在变量上就获取变量当前的赋值，vscode左侧的调用堆栈，变量监视器等的都在实时跟踪应用运行状态
- 左侧的浏览器中是我们正在调试的程序，这是一个用node.js编写的计算器应用并处于调试终端状态
- 以上全部的操作都通过浏览器的方式运行，无需提前安装任何开发环境，SDK或者IDE。你所需要的只有代码库和SmartIDE。
- 以上环境可以运行在你本地电脑或者云端服务器，但开发者全部都可以通过localhost访问，无需在服务器上另外开启任何端口。

SmartIDE可以帮助你完成开发环境的一键搭建，你只需要学会一个命令 (smartide start) 就可以在自己所需要的环境中，使用自己喜欢的开发工具进行编码和开发调试了，不再需要安装任何工具，SDK，调试器，编译器，环境变量等繁琐的操作。如果我们把Vscode和JetBrain这些IDE称为传统IDE的话，这些传统IDE最大的问题是：他们虽然在 I (Integration) 和 D (Development) 上面都做的非常不错，但是都没有解决 E (Environment) 的问题。SmartIDE的重点就是要解决 E 的问题。

## SmartIDE v0.1.2 发布说明

这是SmartIDE的第一个公开发行版，我们已经实现了如下功能：

- **本地模式** 本地模式允许开发者使用一个简单命令smartide start，一键在本地启动基于docker容器的开发环境。全部的开发工具，SDK全部根据一个存放在代码库中的.ide.yaml文件通过docker容器镜像获取
- **远程主机模式** 远程主机模式与本地开发模式提供一致的体验，开发者是需要使用 smartide vm start 命令就可以将开发环境在一台可以通过SSH访问的远程Linux主机上一键启动。 
- **开发容器镜像** 当前提供了 node.js, java, go 三类环境的开发容器镜像
- **WebIDE支持** 支持使用 vscode 和 Eclipse Theia 两种 WebIDE 进行编码开发和单步调试
- **Web Terminal支持** 通过内置于vscode/Elcipse Theia内部的terminal直接操作开发容器环境
- **传统IDE支持** 支持使用 vscode 本地 IDE 通过 SSH 方式连接到开发容器进行开发
- **远程开发本地体验** 通过 ssh tunnel 实现容器内端口的本地访问，允许开发者使用习惯的 http://localhost:port 方式访问开发容器所提供的各项能力，包括：WebIDE，传统IDE支持，Terminal支持等。
- **多语言** 当前提供中文版和英文版两种语言版本，通过自动识别开发者操作系统环境自动切换显示语言
- **跨平台支持** 当前smartide可以在 Windows 和 MacOS 两种操作系统上运行，并可以远程操作 Linux 主机，包括从 Windows远程操作 Linux 和 Mac 远程操作 Linux.
- **一键安装和升级**：开发者可以从 [安装手册](/zh/docs/getting-started/install/) 获取一键安装命令，并支持在现有版本上的一键升级。

欢迎大家通过以下资源体验SmartIDE的快捷开发调试功能，并通过 [GitHub](https://github.com/SmartIDE/SmartIDE/issues) 给我们提供反馈

- [安装手册](/zh/docs/getting-started/install/)：提供Windows和Mac两种环境的一键安装和升级脚本，并提供2个更新通道。
  - 稳定版通道：提供经过完整测试的稳定版本，每2周更新
  - 每日构建版通道：包括最新的功能但可能会有不稳定的情况，每天更新
- [示例操作手册](/zh/docs/getting-started/sample-calculator/)
  - Boathouse-Calcuator 是我们通过 [IDCF Boathouse 开源共创项目](https://idcf.org.cn) 提供给社区的示例应用之一，我们已经在这个应用中适配了SmartIDE的配置文件([.ide.yaml](https://github.com/idcf-boat-house/boathouse-calculator/tree/master/.ide))，您可以按照以上操作手册的内容体验SmartIDE的功能

## 祝各位开发者1024节快乐

SmartIDE选择了在这样一个特殊的日子跟大家说 Hello World。希望在后续的日子里面一直都有SmartIDE的陪伴，让每一名开发者的编码人生更加高效，快乐！

> 赋能每一名开发者，赋能每一家企业
> <br/> -- SmartIDE 团队愿景

2021.10.24 徐磊 @ 北京






