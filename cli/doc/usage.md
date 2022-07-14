
# 概述
使用**Smart IDE CLI**是一个跨平台的命令行工具，目前支持windows、mac、linux（未测试），**Smart IDE CLI**可以启动、停止、删除 Web IDE。

会根据不同的语言环境自动切换语言，目前简体中文、繁体中文操作系统会显示简体中文的说明文字，除此均显示为英文。

目前支持的命令有start、stop、remove、init

- start，本地拉取和运行镜像
- stop，停止容器的运行
- remove，停止和删除镜像
- init（开发中），初始化配置文件


# 安装及使用说明

环境要求：docker

1. git clone {repos} 
  例如：git clone https://github.com/idcf-boat-house/boathouse-calculator.git
2. 从github、gitee、ads上下载各个平台对应的可执行文件（golang生成的二进制包），使用blew的方式请联系 @周文洋
3. 把可执行文件拷贝到本地的工作目录，或者设置系统环境变量
4. 在工作目录，创建和配置 **.ide.yaml**
   > 目前的yaml文件格式为暂定，后续还会调整
    ```
    # 版本
    version: smartide/v0.1
    workspace:
      # 基础镜像的地址
      image: registry.cn-hangzhou.aliyuncs.com/smartide/smartide-node:latest
      # 项目的名称，会成为容器的名称
      name: smartide
      # 绑定的对外暴露端口，例如本地可以使用 http://localhost:3030 访问webide
      idePort: 3030
      # 容器内部应用的调试端口
      appDebugPort: 8080
      # 调试时对外暴露的端口，例如本地访问 http://localhost:8080 可以触发断点
      appHostPort: 8080
    ```

5. 在工作目录运行 ./smartide start，回去自动拉去和运行镜像，完成后会自动打开浏览器

## MAC

## Windows

## Linux


# 使用说明
