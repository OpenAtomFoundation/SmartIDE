![](/docs/content/zh/docs/overview/images/smartide-logo-small.png)
<p align="center"> 
  中文 ｜ <a href="README-EN.md">English</a>
</p>
<h3 align="center">Be a Smart Developer, 开发从未如此简单</h3>
<p align="center">
  中国主页: <a href="https://smartide.cn/zh/" target="_blank">https://smartide.cn</a> 
  海外主页: <a href="https://smartide.dev/en/" target="_blank">https://smartide.dev</a> 
</p>
<p align="center">
  <img src="https://dev.azure.com/leansoftx/smartide/_apis/build/status/smartide-codesign-ci?branchName=main" />
  <img src="https://github.com/smartide/smartide/actions/workflows/smartide-docs-publish.yml/badge.svg" />
  <br/>
  <img src="https://dev.azure.com/leansoftx/945b0f40-4baa-4f8a-be2c-4997b4c0af6a/23e62cbd-3bd8-42a3-a414-df7a1957a69b/_apis/work/boardbadge/37313877-3fed-46f3-87c0-dd28ce47e264" />
</p>
<hr />

SmartIDE可以帮助你完成开发环境的一键搭建，你只需要学会一个命令 (smartide start) 就可以在自己所需要的环境中，使用自己喜欢的开发工具进行编码和开发调试了，不再需要安装任何工具，SDK，调试器，编译器，环境变量等繁琐的操作。如果我们把Vscode和JetBrain这些IDE称为传统IDE的话，这些传统IDE最大的问题是：他们虽然在 I (Integration) 和 D (Development) 上面都做的非常不错，但是都没有解决 E (Environment) 的问题。

**SmartIDE的重点就是要解决 E 的问题。**

## 产品安装方式

我们按照敏捷开发模式进行SmartIDE的开发，所有的版本都通过CI/CD流水线自动构建，打包，测试和发布。为了同时满足外部用户对于稳定性的要求和开发团队以及早期使用者对新功能快速更新的要求，我们提供以下两个发布通道。

- [稳定版](https://smartide.cn/zh/docs/install/#%E7%A8%B3%E5%AE%9A%E7%89%88%E9%80%9A%E9%81%93)
- [每日构建版](https://smartide.cn/zh/docs/install/#%E6%AF%8F%E6%97%A5%E6%9E%84%E5%BB%BA%E7%89%88%E9%80%9A%E9%81%93)

## 快速启动

请参考以下资料开始使用SmartIDE

- 文档 [5分钟快速启动](https://smartide.cn/zh/docs/quickstart/) 
- 文档 [Node.Js 快速启动教程](https://smartide.cn/zh/docs/quickstart/new-node/)
- 视频 [SmartIDE 开发从未如此简单](https://www.bilibili.com/video/BV1pR4y147wn)
- 视频 [开发从未如此简单 - 谁说vue.js和大型微服务Java Spring项目不能在WebIDE中运行，谁说不能一键部署WebIDE到远程主机](https://www.bilibili.com/video/BV1zb4y1i7B1)
- 视频 [SmartIDE 产品发布会](https://www.bilibili.com/video/BV1xR4y1s7sx)

## SmartIDE 三种形态

![](/docs/content/zh/blog/releases/2021-1203-state-management/images/smartide-3modes.png)

- **本地模式**: 本地模式通过一个简单的 smartide start 命令，根据嵌入在代码库中的环境说明文(.ide.yaml)完成环境的启动，让开发者可以无需搭建任何开发环境即可通过容器的方式开始编码调试以及基本的源代码管理(Git)操作。
- **远程主机模式**: 远程主机模式允许用户在 smartide start 命令中增加 --host 参数直接调度一台远程Linux完成开发环境的启动。相对于本地模式，远程主机模式更加能够体现SmartIDE的能力，开发者可以利用远程主机更为强大的算力，更庞大的存储以及更高速的网络获取更好的开发体验，还可以完成一些本地模式下无法完成的开发操作，比如：AI和大数据开发，大型微服务项目的开发等等。SmartIDE对于开发者使用的远程主机没有任何限制，只需要开发者可以通过SSH方式访问主机即可，远程主机可以位于任何位置，包括：公有云，私有云，企业数据中心甚至开发者自己家里。
- **k8s模式**: 将远程主机模式命令中的 --host 替换成 --k8s，开发者即可将开发环境一键部署到 Kubernetes (k8s) 集群中。与远程主机模式一样，SmartIDE对于开发者所使用的k8s集群没有任何限制，无论是公有云托管式集群，还有自行部署的集群均可。只要开发者对于某个namespace具备部署权限，借款通过SmartIDE完成开发环境的一键部署。k8s模式将为使用SmartIDE的开发者开辟一个全新的天地，借助k8s所提供的高度灵活高效的环境调度能力，我们可以为开发者提供更加丰富的使用场景和更为强大的开发环境。

## 路线图

![](/docs/content/zh/blog/releases/2021-1203-state-management/images/smartide-roadmap.png)

从当前我们所提供的 smartide-cli 应用将贯穿未来的整个路线图，作为开发者与开发工作区(Workspace)之间进行交互的主要桥梁，在此基础上我们也将为开发者提供更加易于使用的GUI工具，包括本地GUI工具和Web端管理能力，协助开发者完成更为复杂的环境调度和团队协作场景。SmartIDE针对独立开发者和中小团队的功能将采用开源免费的方式提供，而针对企业的版本则会提供企业授权和更为完善的产品技术支持。

## 社区推广计划

欢迎大家通过以下渠道与SmartIDE产品开发团队保持联系: 

- Smart Meetup: 我们将通过【冬哥有话说栏目】每周推介一款好用的开源代码库给到大家，整个推介过程控制在15分钟内，全程通过演示的方式使用SmartIDE来启动开源代码库的编码调试，让开发者在了解了开源项目本身的价值的同时了解SmartIDE带来的快速便捷开发体验。
- Smart早鸟计划: 我们将持续的在社区中招募希望提前体验SmartIDE的开发者，加入我们的微信群。作为一款由开发者为开发者打造的开发工具，我们希望听取真正使用者的意见，持续改进我们的产品，和开发者一起将这个产品做下去。


## Copyright 

&copy;[leansoftX.com](https://leansoftx.com) 2021
