---
title: "开发语言和中间件支持"
linkTitle: "开发语言支持"
weight: 30
description: >
  全语言技术栈支持，为不同开发语言技术栈提供一致的开发环境启动方式。
---

SmartIDE中的开发语言支持只可扩展，可插拔的。通过IDE配置文件，开发者镜像和模板的配合，开发者可以将任何可以在容器中运行的开发语言SDK与SmartIDE进行适配，整个适配过程无需修改SmartIDE本身的代码逻辑。当前我们已经提供的开发语言支持包括：

- JavaScript/Node
- Java
- C# (跨平台版本，包括：.net core 和 .net 6以上，暂时不支持.net framework)
- Python
- Golang
- PHP
- C/C++
- Anacoda (Juypter Notebook)

对于各种类型的中间件的支持方式也是一样的，任何剋有在容器中运行的中间件均可以适配到SmartIDE工作区中使用。当前我们已经提供支持的中间件以及配套工具包括

- MySQL + PHPMyAdmin
- Redis
- Nacos
