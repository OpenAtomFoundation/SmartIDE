---
title: "应用插件市场安装手册"
linkTitle: "应用插件市场安装手册"
weight: 24
date: 2022-04-27
description: >
    描述如何部署安装SmartIDE Marketplace。
---

1. 简介
https://marketplace.smartide.cn 是基于Eclipse OpenVSX 开源项目搭建的类VSCode插件市场，此文档旨在描述 SmartIDE Marketplace 的详细部署过程，内容分为简要介绍、组件介绍、部署细节三部分。

SmartIDE Marketplace服务搭建参考以下文档细节：
  1. Deploying Open VSX · eclipse/OpenVSX Wiki (github.com)
  2. eclipse/OpenVSX: An open-source registry for VS Code extensions (github.com)
  3. Dashboard — Gitpod-OpenVSX
SmartIDE Marketplace服务部署均使用容器化方式进行，各模块整体架构如下图所示：

  1. 主体为OpenVSX-Server，spring boot框架的java服务，我们在部署时需要自行添加application.yml server 配置文件，并将其放置对应位置，以便Server启动后加载。
  2. 后台数据库使用PostgreSql，并使用Pgadmin进行数据库的管理和查询等操作，数据库的创建和表结构的初始化过程 server进程启动时会自动执行。
  3. 前端界面为NodeJS架构的Web UI，我们在部署时会将此代码库构建的静态网站结果，放入Server服务的对应文件夹，使其二者变为一个进程即Server进程加入前端界面。这也是Sprint Boot框架的灵活性功能，即使用者可以基于Web UI代码库自定制前端界面，并将自定制的前端页面嵌入Server服务（Deploying Open VSX · eclipse/openvsx Wiki (github.com)）
  4. 用户登陆验证，目前只支持OAuth Provider，官方文档中声明目前只支持Github AuthApp和 Eclipse OAuth，我们在部署时使用Github AuthApp。（问题：私有化部署时还需要研究如何在企业内部登陆）
  5. 插件存储可以使用数据库（默认），Google Storage或 Azure Blob Storage三种模式，推荐添加Google Storage或 Azure Blob Storage以避免数据库过大的情况出现。
  6. 插件搜索服务支持数据库搜索和附加Elastic Search服务两种模式，推荐有条件的情况下添加Elastic Search搜索服务提高搜索效率，降低数据库压力。
  7. 除以上架构图中展现内容外，Marketplace网站需要配置https 证书，这样才服务器的扩展才能够正确被IDE加载，我们使用Nginx进行服务器的部署端口转发。
