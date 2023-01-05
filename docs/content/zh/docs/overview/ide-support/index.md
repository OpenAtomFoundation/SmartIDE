---
title: "IDE支持"
linkTitle: "IDE支持"
weight: 40
description: >
  我们不会发明另外一个IDE，我们你使用自己喜欢的IDE，并更加高效。
---

作为工作区（远程/云端）与用户之间的交互方式（Interface），IDE工具是开发者的入口。为了能够让开发者使用自己喜欢的IDE工具，SmartIDE可以引入任何能够提供WebUI的工具作为IDE入口，从实现上来看，实际上就是一个WebApp而已。

当前，SmartIDE支持4种IDE接入方式，分别是

- WebIDE：任何可以提供WebUI的协助开发者进行软件开发的工具均被被认为是IDE的组件，比如
  - VSCode WebIDE：我们内置提供了基于OpenVSCode Server的WebIDE支持
  - JetBrains Projectors：Projectors是JetBrains提供的全系列IDE的WebUI版本，可以通过浏览器访问呢，界面和功能上和JetBrains的桌面IDE保持一致。当前我们内置提供了以下Projectors支持（根据不同的开发语言）。
    - IntelliJ IDEA 
    - WebStorm
    - Goland
    - PyCharm
    - Rider
    - PHPStorm
    - Clion
  - OpenSumi：阿里蚂蚁开源的WebIDE开发框架
  - Eclipse Theia: Eclipse基金会旗下的开源WebIDE开发框架（有限支持）
- SSH Remote 接入：开发者可以通过SSH协议，使用本地的终端程序，VSCode Remote SSH插件，JetBrains Gateway或者任何支持SSH远程接入的工具连接到SmartIDE工作区进行操作。
- Web Terminal 接入：SmartIDE工作区提供 webterminal addon，可以在现有工作区上提供浏览器内访问的终端程序，并允许用户进入工作区内的任何容器进行操作。使用Web Terminal不需要工作区对外暴露SSH端口。
- 其他Web应用：开发者在工作中还需要各种类型的第三方工具来管理工作区种的各类资源，比如：如果工作区内使用了MySQL作为后天数据库，那么开发会希望使用PHPMyAdmin来通过浏览器管理数据库，协助完成日常开发调试。类似的场景均可以通过SmartIDE的IDE配置文件实现。当前已经支持的其他Web应用工具包括：
  - PHPMyAdmin
  - Nacos
  - Juypter Notebook 