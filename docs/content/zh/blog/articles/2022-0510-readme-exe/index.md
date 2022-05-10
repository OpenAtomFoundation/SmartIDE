---
title: "README.exe"
linkTitle: "README.exe"
date: 2022-05-10
description: >
  SmartIDE让你的README变成可执行文档，再也不用编写无用的文档，再也不必操心环境问题。
---

作为开发者，拿到一个新的代码库的时候一般都会先去看README文件，通过这个文件可以知道这套代码所需要安装的环境，工具和操作方式。这件事情本来应该是一件很愉悦的事情，因为每一套新代码其实都是开发者的新玩具，拿到新玩具的心情那肯定是不错的。但是，当你阅读玩具说明书之后，发现这份说明书完全不配套的时候，那心里一定是一万匹草泥马在奔腾。当然，这也很容易理解，开发者不爱写文档，特别是那些没有用的文档。至少，README对写的人来说其实没啥用，因为写的人都已经清楚了文档中的内容，至于看的人感受如何，那就呵呵吧。

这个问题的根源在于README只能看，不能运行！如果我们能够让README活起来，从 README.md 变成 README.exe，那是不是就可以解决这个问题了呢？答案是肯定的！因为写的人自己也可以用这份文档来启动项目。这样，写文档的人有了动力，看（运行）文档的人也会很爽。

> **这就是SmartIDE的核心功能: IDE as Code能力。**

## 神奇的IDE配置文件

SmartIDE最初始的设计灵感就是如何能够让README活起来？为了做到这一点，我们设计了一个 **IDE配置文件** (默认文件名 .ide.yaml）文件格式。这个文件中完整描述了运行当前代码所需要的环境配置，包括 **基础环境SDK，应用服务器，应用端口，配置文件，网络依赖以及所使用的IDE工具**。有了这个文件，开发者就可以真正实现一键启动开发调试，而不会再听到：“在我这里工作呀，是你的环境问题！” 这种骇人听闻的抱怨。

有了这个文件，任何人都可以使用一个简单的指令来一键搭建一模一样的开发环境，指令如下：

```shell
smartide start https://github.com/idcf-boat-house/boathouse-calculator
```

这个指令会自动识别代码库中的 **IDE配置文件**，根据文件中对开发环境的描述完成获取开发者镜像、启动开发容器、克隆代码、运行初始化脚本等一系列动作，最终一个 **完整可运行的环境** 呈现在开发者的面前，下面这段视频展示了整个运行过程：

{{< bilibili 336989627 >}}

## IDE配置文件详解

以上示例中所使用的 `.ide.yaml` 文件如下

```yaml
version: smartide/v0.3
orchestrator:
  type: docker-compose
  version: 3
workspace:
  dev-container: # 开发者容器设置
    service-name: boathouse-calculator
    
    ports: # 申明端口
      tools-webide-vscode: 6800
      tools-ssh: 6822
      apps-application: 3001
    
    ide-type: vscode  #sdk-only | vscode | opensumi | jb-projector
    volumes: # 本地配置映射，支持映射git-config和ssh密钥信息进入容器
      git-config: true
      ssh-key: true
    command: # 环境启动脚本
      - npm config set registry https://registry.npmmirror.com
      - npm install
  services:
    boathouse-calculator:
      image: registry.cn-hangzhou.aliyuncs.com/smartide/smartide-node-v2-vscode:all-version
      restart: always
      environment:
        ROOT_PASSWORD: root123
        LOCAL_USER_PASSWORD: root123       
      volumes:
      - .:/home/project
      ports:
        - 3001:3001
        - 6822:22
        - 6800:3000
      networks:
        - smartide-network

  networks:
    smartide-network:
      external: true
```

这个文件内容非常通俗易懂，是个程序员应该都能看明白，不过我还是简单说明一下：

- `orchestrator` - 环境调度工具设置，用来制定调度容器环境的底层工具，这里我们使用的是 docker-compose
- `workspace` - 工作区配置，[工作区](/zh/docs/overview/workspace/) 是SmartIDE中最重要的概念，包含开发者用来进行开发调试的所有环境信息
  - `dev-container` - 开发者容器设置
    - `service-name` - 开发者容器所对应的 docker-compose 服务名称
    - `ports` - 开发者容器对外暴露的端口
    - `ide-type` - 开发者容器所所用的IDE类型，支持：vscode, sdk-only, jb-projector (Jetbrains系列全家桶）和 opensumi
    - `volumes` - 配置映射，支持将开发者的git-config和ssh密钥导入容器
    - `commands` - 开发环境启动脚本，对环境进行初始化；比如以上脚本中就完成了2个关键操作：1）设置npm国内镜像源 2）获取npm依赖。这部分脚本开发者可以根据自己代码库的情况设置，SmartIDE会在环境启动后自动运行这些脚本，让程序处于开发者所希望的状态上。
- `services` - 开发环境内的服务列表
  - 这里其实就是一段标准的 docker-compose 配置文件

## IDE as Code 重新定义IDE

这种做法称为 `IDE as Code` 也就是 “集成开发环境即代码”，将你的开发环境配置变成一个 **IDE配置文件** 的配置文件放置在代码库中，然后根据这个配置文件生成对应的自动化脚本，完成“集成开发环境” 的创建。因此，SmartIDE中的IDE不仅仅是一个开发工具，而是 **包含了环境的IDE**。

`IDE as Code` 的做法源自DevOps的核心实践 `Infrastructure as Code` ，也就是 “基础设施即代码” 简称 `IaC`。其核心思想是将环境配置代码化，常见的k8s的yaml文件其实就是典型的基础设施即代码实现。在运维领域常用的工具比如chef/puppet/ansible，还有 HashiCorp 的 Terraform 也都是 `IaC` 的经典实现。`IaC` 的主要价值和目标就是将环境搭建过程标准化，让任何人在任何环境中都可以获得 **一致、稳定、可靠** 的环境搭建体验。SmartIDE所创造的 **IDE配置文件** 延续IaC了的使用场景，并将其基本思路应用到了开发测试环境中，这其实就是 SmartIDE 的产品核心能力。

基于 `IDE as Code` 的实现，SmartIDE在产品实现过程中一直秉承一个原则：能够让用户通过配置文件实现的东西，就不要通过代码实现。这个核心原则给予了SmartIDE非常强的灵活性，比如以下这段视频中所演示的 [若依管理系统](/zh/docs/examples/ruoyi/) 项目

{{< bilibili 851717139 >}}

这个项目包括比较复杂的环境配置，包括：

- JDK 基础环境
- MAVEN 包管理工具
- MySQL 服务器
- Redis 服务器
- 可选：数据库管理工具 PHPMyAdmin 

若依项目的官方文档用了整整2页的篇幅描述开发环境的搭建 (参考链接：[环境部署 | RuoYi](http://doc.ruoyi.vip/ruoyi/document/hjbs.html)） 。使用了 SmartIDE 以后，开发者可以一个统一的指令 `smartide start` 来启动这个复杂的项目，不再需要去阅读这个文档。无论项目的代码是简单亦或复杂，`smartide start` 指令都可以进行适配，因为其背后的复杂配置已经全部通过 **IDE配置文件** 和代码保存在一起了。使用 `IDE as Code` 的另外一个好处是，由于配置和代码保存在一起，进行代码变更的开发者可以同步更新环境配置，并且一起提交进行评审，也就是说：你的环境也和代码一样是可评审，可审计的。

> IDE文件配置就是README活文档。

当然，SmartIDE能做的绝不仅仅如此，我们已经发布了 [SmartIDE Server](/zh/docs/quickstart/server/) 版，允许开发者在网页上即可完成开发环境的搭建和完整的开发调试过程。Server与CLI一样使用这个 **IDE配置文件** 来识别和搭建环境，与cli保持一致的体验。这一切的目的都是让开发者的日常工作变得简单，让开发者体验 #开发从未如此简单 的快乐。

如果你对云原生开发环境感兴趣，请扫描以下二维码加入我们的 **SmartIDE社区早鸟计划**

<img src="/images/smartide-s-qrcode.png" style="width:120px;height:auto;padding: 1px;"/>

谢谢您对SmartIDE的关注，让我们一起成为云原生时代的 *Smart开发者*, 享受 *开发从未如此简单* 的快乐。

2022年5月10日
