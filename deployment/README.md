# 在线安装
## 1. 环境准备
- 虚拟机准备：
    - 保证安装目录/home及/var目录至少有15G的空间
    - 新建非root账号（这里是因为minikube start driver=docker时，不允许使用root账号。）
### 1.1 新增用户smartide,并设置sudo权限
```
sudo useradd smartide
sudo passwd smartide
sudo vim /etc/sudoers
```
在## Allow root to run any commands anywhere 的root下方设置sudo免密权限：
```
smartide   ALL=(ALL) NOPASSWD: ALL
```
### 1.12 网络联通性要求：
- SmartIDE Server与开发资源主机之间的网络访问应全部开放，这里使用到的端口主要为：
- Server -> 开发主机：SSH端口，默认为：22
- 开发主机 -> Server：Server网站端口，默认为8080
## 2 环境安装
### 2.1 安装步骤介绍：
安装主要分为以下步骤：
- 1.基础环境安装，包括:Docker、Docker-Compose、Git、Kubectl、Minikube。
- 2.Tekton安装。这是建立和管理SmartIDE工作区所使用的流水线组件。
- 3.SmartIDE Server安装。这里包括SmartIDE Server所依赖的数据库环境以及应用程序前后端，以及维护管理工具。
### 2.3 执行一键安装脚本：
这里已为你准备好一键安装脚本，执行以下命令，即可完成全部安装过程。
- 国内：
```language=bash
curl -o- https://gitee.com/chileeb/SmartIDE/raw/main/deployment/online/deployment_cn.sh | bash
```
- 国外：
```language=bash
curl -o- https://gitee.com/chileeb/SmartIDE/raw/main/deployment/online/deployment.sh | bash
```
安装完成后，我们可以看到命令行中显示：
SmartIDE Server Deployment Success！
此时，SmartIDE已安装完毕。
- 
## 3. 访问地址
- SmartIDE Server：http://{deploment host ip}:8080
- MySQL DB ：http://{deploment host ip}:8090
- Container Manager : http://{deploment host ip}:9000
- Tekton DashBoard : http://{deploment host ip}:9097/

- SmartIDE Server：用户名：superadmin 默认密码：SmartIDE@123

注：若需使用Tekton DashBoard查看流水线执行情况，需执行以下命令：
```
kubectl --namespace tekton-pipelines port-forward svc/tekton-dashboard 9097:9097 --address 0.0.0.0 &
```
### 4 配置参数说明
安装完成后，修改配置文件：
```
vim smartide-server/config.docker.yaml
```
修改以下配置内容：
```
smartide:
  api-host: http://gva-web:8080
  api-host-xtoken: XXXX
  tasks:
  - name: templateinit
    spec: '@hourly'
    start: true
  tekton-trigger-host: http://minikube
  template-git-url: https://gitee.com/smartide/smartide-templates.git
```
- api-host：无须修改。这里的地址是流水线CLI程序调用SmartIDE Server API程序的入口地址。
- api-host-xtoken：无须修改。这里默认为superadmin账号的token密码。如修改管理员密码后，需要通过客户端，使用superadmin登录，将获取到的token填写于此处。
- tekton-trigger-host：这里是SmartIDE Server调用Tekton流水线服务的地址，这里需要修改IP地址为部署SmartIDE Server的主机IP地址。
- template-git-url：这里的地址为SmartIDE模板库，用作使用模板创建工作区。

## 5.客户端安装:见[CLI 安装说明]
