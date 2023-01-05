---
title: "Java 快速启动教程"
linkTitle: "Java"
weight: 31
description: >
  本文档描述如何使用SmartIDE完成一个Java SpringBoot Web 应用的完整开发、调试和代码提交过程。
---

SmartIDE内置了Java开发环境模板，你可以通过一个简单的指令创建带有WebIDE的开发环境，并立即开始编码和调试。   

如果你还没有完成SmartIDE安装，请参考 [SmartIDE 安装手册](/zh/docs/install) 安装SmartIDE命令行工具。

> 说明：SmartIDE的命令行工具可以在Windows和MacOS操作系统上运行，对大多数命令来说，操作是完全一致的。本文档中虽然使用的是MacOS上的截图，但是Windows环境的所有日志和工作状态完全一致。对于脚本格式有区别的地方，我们会同时提供2套脚本。


## VSCode

### 完整操作视频

为了便于大家更直观地了解和使用SmartIDE创建Java环境，并开始Spring Web项目的开发和调试，我们在B站提上提供了视频供大家参考，视频如下：
{{< bilibili 852887843 >}}

跳转到B站：<a href="https://www.bilibili.com/video/av852887843" target="_blank"> ` https://www.bilibili.com/video/av852887843 `</a>

###  1. 创建开发环境

运行以下命令创建Java开发环境：

```shell
# 在 MacOS/Windows 上打开 终端（Terminal）或者 PowerShell 应用
# 执行以下命令
mkdir sample-java-vscode 
cd sample-java-vscode
smartide new java -T vscode
```

运行后的效果如下，你可以通过命令窗口中的日志详细了解 SmartIDE 的启动过程，当 SmartIDE 启动完毕之后，会自动打开浏览器窗口并导航到 WebIDE 界面。

![Java Quickstart DevEnv](images/quickstart-java-vscode01-01.png)

###  2. 安装Java及Spring常用插件包

Spring Boot应用，本质上是Java应用，所以我们需要Java和Spring两方面的扩展支持。VS Code为了方便安装，提供一种Extension Pack的方式，把相关技术的多个常用Extension打包成一个Pack，可以实现一键安装，所以我们可以直接安装以下两方面的扩展包即可：

1. Java支持：Java Extension Pack(vscjava.vscode-java-pack）

    功能包括：流行的Java开发扩展，提供Java自动补全、调试、测试、Maven/Gradle支持、项目依赖管理等等。

2. Spring支持：Spring Boot Extension Pack(pivotal.vscode-boot-dev-pack)

    功能包括：用于开发Spring Boot应用程序的扩展集合。

SmartIDE已经通过初始化脚本的方式，为开发环境自动安装了这两个扩展包。在WebIDE启动后请稍等片刻，SmartIDE会自动启动Terminal并执行以上两个插件的安装脚本，安装完毕后，如下图所示：

![Install Extension](images/quickstart-java-vscode02-01.png)

对应初始化插件包安装脚本如下：

> 脚本供参考，你无需手工执行这些脚本

```shell
# 安装Java扩展包集合（Extension Pack for Java）
code --install-extension vscjava.vscode-java-pack
# 安装Spring扩展包集合（Spring Boot Extension Pack）
code --install-extension pivotal.vscode-boot-dev-pack
```

###  3. 创建并配置项目
**使用快捷键运行Spring向导，快速创建一个Spring Boot应用。**

使用快捷键Ctrl + Shift + P，然后输入Spring Initializr，选择 **Spring Initializr:Create a Maven Project...** 创建项目，进入创建向导：

![Spring Initializr](images/quickstart-java-vscode03-01.png)

在这里指定 Spring Boot 版本，**2.6.4**，如下图：

![Spring Boot Version](images/quickstart-java-vscode03-02.png)

指定语言，**Java**，如下图：

![Spring Boot Java Language](images/quickstart-java-vscode03-03.png)

输入maven项目Group Id: **cn.smartide**，并回车，如下图：

![Spring Boot Group Id](images/quickstart-java-vscode03-04.png)

输入maven项目Artifact Id: **smartide-demo**，并回车，如下图：

![Spring Boot Artifact Id](images/quickstart-java-vscode03-05.png)

指定打包方式: **Jar**，如下图：

![Spring Boot Packaging Type](images/quickstart-java-vscode03-06.png)

指定 Java (JDK) 版本: **11**，如下图：

![Spring Boot Java Version](images/quickstart-java-vscode03-07.png)

选择项目依赖: **Spring Web Web**，如下图：

![Spring Boot Dependencies](images/quickstart-java-vscode03-08.png)

可对依赖进行多选，这里我们直接回车确认不再添加其他依赖包，如下图：

![Spring Boot Dependencies OK](images/quickstart-java-vscode03-09.png)

代码目录：**/home/project**，如下图：

![Spring Boot Code Folder](images/quickstart-java-vscode03-10.png)

确认后，扩展将为我们创建Spring Boot Web项目。

创建成功后，IDE的右下方会给出提示，点击：**Add to Workspace**，添加项目到工作区，如下图：

![Spring Boot Add Workspace](images/quickstart-java-vscode03-11.png)


执行完毕后的效果如下，左侧文件管理器里面已经出现了 smartide-demo 文件夹，并在其中创建了 spring web 应用的基础代码结构。如下图：

![Spring Web Generator](images/quickstart-java-vscode03-12.png)

此时，IDE会自动打开Java项目并进行初始化构建操作，如下图所示：

![Building](images/quickstart-java-vscode03-13.png)

点击查看详细信息，查看项目的初始化构建情况，如下图所示：

![Java Build Status](images/quickstart-java-vscode03-14.png)

初始化完成后，如下图所示，所有操作都显示为Done：

![Build Done](images/quickstart-java-vscode03-15.png)

工作区添加完毕后，会提示项目导入成功，如下图：

![Imported WorkSpace](images/quickstart-java-vscode03-16.png)

默认，构建时会通过maven官方源下载依赖，我们可以修改 pom.xml，将依赖拉取地址改为国内阿里源。在 pom.xml 中 **第5行** ，添加：

```xml
<repositories>
    <repository>
        <id>alimaven</id>
        <name>aliyun maven</name>
        <url>http://maven.aliyun.com/nexus/content/groups/public/</url>
        <releases>
            <enabled>true</enabled>
        </releases>
        <snapshots>
            <enabled>false</enabled>
        </snapshots>
    </repository>
</repositories>
<pluginRepositories>
    <pluginRepository>
        <id>alimaven</id>
        <name>aliyun maven</name>
        <url>http://maven.aliyun.com/nexus/content/groups/public/</url>
        <releases>
            <enabled>true</enabled>
        </releases>
        <snapshots>
            <enabled>false</enabled>
        </snapshots>
    </pluginRepository>
</pluginRepositories>
```

这样可以更快地拉取依赖，进行maven构建。新建终端，执行命令：

```shell
# 进入spring web项目目录
cd /home/project/smartide-demo
# 运行mvn构建安装
mvn install
```

![Mvn Install](images/quickstart-java-vscode03-17.png)


###  4. 开发调试

完成以上配置之后，你的代码已经完全准备好，可以开始进行编码调试了。

首先，我们建立一个controller包，路径为cn.smartide.smartidedemo.controller，并且新建一个类：SmartIDEController.java，代码如下

```java
/*SmartIDEController.java*/

package cn.smartide.smartidedemo.controller;

import org.springframework.boot.autoconfigure.EnableAutoConfiguration;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

@RestController
@EnableAutoConfiguration
public class SmartIDEController {
    @RequestMapping("/home")
    String home(@RequestParam String language) {
        String hello = "Hello, SmartIDE Users!";
        return hello + " The dev language is:" + language + ".";
    }
}
```
新建完成后，在启动类 **SmartideDemoApplication.java** 点击Debug，通过Debug模式启动项目。启动调试后，请注意 smartide 客户端的日志输出，SmartIDE 会在后台持续监控容器内的进程情况，并将所有端口转发到 localhost 上
![Start Debug](images/quickstart-java-vscode04-01.png)

在**SmartIDEController.java**文件的 **第13行** 代码处 **单击设置断点** 

![设置断点](images/quickstart-java-vscode04-02.png)

在浏览器中输入地址： **http://localhost:8080/home?language=Java** 触发我们之前所设置的断点，

![调试状态](images/quickstart-java-vscode04-03.png)

现在，进入到调试状态，注意上图中的几个关键点

1. 通过打开 http://localhost:8080/home?language=Java 这个地址触发我们预设的断点
2. 将鼠标移动到特定的变量上以后，IDE 会自动加载当前变量的结构体以及赋值状态（实时），方便开发者观察运行时状态
3. Variables (变量) 窗口实时显示当前运行时内的变量状态
4. Call Stack (调用堆栈) 窗口实时显示当前运行时堆栈状态

点击单步运行，并跳出断点，页面将返回输出结果，如下图所示：

![结束调试](images/quickstart-java-vscode04-04.png)


**至此，我们已经使用 SmartIDE 完成了一个 Spring Web 应用程序的创建，配置和编码调试过程。**

###  5. 提交并分享

SmartIDE 环境中已经内置了 Git 的支持，你可以点击 **菜单栏左侧 ｜ 源代码管理 ｜ 点击 Initialize Repository 按钮** 将当前工作区初始化成一个 Git代码库。

![初始化Git库](images/quickstart-java-vscode05-01.png)

在 **提交注释** 中填写 **使用SmartIDE创建**，然后点击 **提交按钮** 

![Commit](images/quickstart-java-vscode05-02.png)

点击 **Remote | Add Remote** 按钮，添加一个远端 Git库 地址。SmartIDE 支持任何Git服务，包括：GitHub, Gitlab, Azure DevOps, Gitee 等等。

![Add Remote](images/quickstart-java-vscode05-03.png)

> 这时，我们可以将创建的这份代码推送到了类似Gitee的代码仓库上，代码库地址类似如下
> https://gitee.com/smartide/sample-java-vscode

至此，我们已经使用 SmartIDE 完成了一个 Spring Boot 应用从环境搭建，创建基础代码结构，配置调试环境，完成编码开发到提交代码的全过程。

**现在可以将你的代码库发送给其他的小伙伴，让他通过以下指令一键启动你的应用。**

```shell
smartide start https://gitee.com/smartide/sample-java-vscode
```

是不是很爽！

## JetBrains IntelliJ IDEA

###  1. 新建开发环境

运行以下命令创建Spring Web项目开发环境：

```
mkdir sample-java-idea
cd sample-java-idea
smartide new java -T idea
```

运行后的效果如下，你可以通过命令窗口中的日志详细了解 SmartIDE 的启动过程，当 SmartIDE 启动完毕之后，会自动打开浏览器窗口并导航到 WebIDE 界面。

![node quickstart](images/quickstart-java-idea01-01.png)

###  2. 创建并配置项目

SmartIDE启动的JetBrains IDEA环境为 **社区版**，在新建Spring Web项目时，建议安装Spring初始化插件，这样新建项目会更加方便。
这里我们安装第三方插件：**Spring Initializr and Assistant**。

点击插件，在搜索框输入：**Spring Initializr and Assistant**，并安装插件。

![Spring Initializr](images/quickstart-java-idea02-01.png)

接受第三方插件安装:

![Third-Party Plugins](images/quickstart-java-idea02-02.png)

安装完毕后，新建项目：

![New Project](images/quickstart-java-idea02-03.png)

使用 Spring Initializr向导 新建项目：

![Spring Initializr New Project](images/quickstart-java-idea02-04.png)

输入项目信息：

![Project Properties](images/quickstart-java-idea02-05.png)

选择项目依赖项目信息，这里我们选择了两个依赖，分别是：
- 1）Developer Tools -> Lombok
- 2）Web -> Spring Web

![Project Dpendencied](images/quickstart-java-idea02-06.png)

修改项目名称，以及默认 **保存路径**：

![Project Name And Path](images/quickstart-java-idea02-07.png)

至此，Spring Web项目已创建完成。

下面，打开**File**—**Project Structure**—**Modules** ( Win快捷键：`Ctrl+Alt+Shift+S`，Mac快捷键 `Command+;` ) ，设置Maven项目的目录结构：
- 1）`src/main/java`：Sources文件夹
- 2）`src/main/resources`：Resources文件夹
- 3）`src/test/java`：Tests文件夹

![Maven ](images/quickstart-java-idea02-09.png)

接着，我们设置依赖拉取地址为国内阿里源，在pom.xml中**第5行**，添加：

```xml
<repositories>
    <repository>
        <id>alimaven</id>
        <name>aliyun maven</name>
        <url>http://maven.aliyun.com/nexus/content/groups/public/</url>
        <releases>
            <enabled>true</enabled>
        </releases>
        <snapshots>
            <enabled>false</enabled>
        </snapshots>
    </repository>
</repositories>
<pluginRepositories>
    <pluginRepository>
        <id>alimaven</id>
        <name>aliyun maven</name>
        <url>http://maven.aliyun.com/nexus/content/groups/public/</url>
        <releases>
            <enabled>true</enabled>
        </releases>
        <snapshots>
            <enabled>false</enabled>
        </snapshots>
    </pluginRepository>
</pluginRepositories>
```
设置完毕后，在pom.xml文件中，指定**spring-boot-maven-plugin**插件版本：

![Maven Plugin Version](images/quickstart-java-idea02-10.png)

点击**Reload All Maven Projects**，重新加载Maven项目。Maven项目加载完毕后，点击mvn install，触发依赖下载与安装：
![Maven Reload](images/quickstart-java-idea02-11.png)

至此，我们设置完毕项目所有的Maven依赖，这时可以启动Spring Web项目了。

###  3. 开发调试

完成以上配置之后，基础代码已经完全准备好，就可以开始进行编码调试了。

首先，我们建立一个controller包，路径为cn.smartide.demo.controller，并且新建一个类：SmartIDEController.java，代码如下:
```java
/*SmartIDEController.java*/

package cn.smartide.demo.controller;

import org.springframework.boot.autoconfigure.EnableAutoConfiguration;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RequestParam;
import org.springframework.web.bind.annotation.RestController;

@RestController
@EnableAutoConfiguration
public class SmartIDEController {
    @RequestMapping("/home")
    String home(@RequestParam String language) {
        String hello = "Hello, SmartIDE Users!";
        return hello + " The dev language is:" + language + ".";
    }
}
```

新建完成后，在启动类 **SmartideDemoApplication.java** 上右键点击Debug，通过Debug模式启动项目。启动调试后，请注意 smartide 客户端的日志输出，SmartIDE 会在后台持续监控容器内的进程情况，并将所有端口转发到 localhost 上

![Start Debug](images/quickstart-java-idea02-12.png)

在**SmartIDEController.java**文件的 **第15行** 代码处 **单击设置断点** 

![设置断点](images/quickstart-java-idea02-13.png)

在浏览器中输入地址：**http://localhost:8080/home?language=Java** 触发我们之前所设置的断点，

![调试状态](images/quickstart-java-idea02-14.png)

现在，进入到调试状态，注意上图中的几个关键点

1. 通过打开 http://localhost:8080/home?language=Java 这个地址触发我们预设的断点
2. 将鼠标移动到特定的变量上以后，IDE 会自动加载当前变量的结构体以及赋值状态（实时），方便开发者观察运行时状态
3. Variables (变量) 窗口实时显示当前运行时内的变量状态
4. Debugger (调试) 窗口实时显示当前运行时堆栈状态

点击单步运行，并跳出断点，页面将返回输出结果，如下图所示：

![结束调试](images/quickstart-java-idea02-15.png)


**至此，我们已经使用 SmartIDE 完成了一个 Spring Web 应用程序的创建，配置和编码调试过程。**


###  4. 提交并分享

环境中已内置了git命令，这里我们通过命令行的方式，将代码提交到远端 Git库。
SmartIDE 支持任何Git服务，包括：GitHub, Gitlab, Azure DevOps, Gitee 等等。

![终端](images/quickstart-java-idea02-16.png)

下面我们执行git提交命令：
```shell
# 初始化git仓库
git init
# 添加项目文件
git add .
# 提交
git commit -m "使用SmartIDE创建smartide-java-demo"
# 添加远端地址
git remote add origin https://gitee.com/smartide/sample-java-idea
# 推送到远端
git push -u origin "master"
```

> 这时，我们可以将创建的这份代码推送到了类似Gitee的代码仓库上，代码库地址类似如下 https://gitee.com/smartide/sample-java-idea

至此，我们已经使用 SmartIDE 完成了一个Spring Web 应用从环境搭建，创建基础代码结构，完成编码开发调试到提交代码的全过程。

**现在可以将你的代码库发送给其他的小伙伴，让他通过以下指令一键启动你的应用。**

```shell
smartide start https://gitee.com/smartide/sample-java-idea
```

是不是很爽！

###  5. 快速创建Spring Boot项目
SmartIDE模板中提供了创建Spring Boot的示例模板，可以更快速地创建一个Spring Web项目。

命令如下：
```shell
smartide new java -T springboot-idea
```

详情请参考：https://gitee.com/smartide/smartide-springboot-template

基于Spring Boot并包含前端、数据库等的Java项目，可参考示例应用-**[若依项目](../../examples/ruoyi/)**。

## 远程开发

上面我们已经使用SmartIDE的本地工作区模式完成了一个应用的创建和开发过程，这个过程和你所熟悉的开发模式有2个区别，1）我们使用了WebIDE；2）开发环境全部通过容器获取并运行。

在这个过程中你的项目代码也已经具备了远程开发的能力，你可以按照以下文档中的说明使用任意一种远程工作区来开发调试你的应用

- [远程主机工作区](/zh/docs/overview/remote-workspace/#远程主机工作区)
- [k8s工作区](/zh/docs/overview/remote-workspace/#k8s工作区)
- [Server工作区](/zh/docs/overview/remote-workspace/#server工作区)

另外，你也可以通过VSCode或者JetBrains内置的远程开发模式进行Hybird模式的远程开发，具体请参考

- [IDE远程开发操作手册](/zh/docs/manual/ide-remote/)

---
**感谢您对SmartIDE的支持：Be a Smart Developer，开发从未如此简单。**
