/*
SmartIDE - Dev Containers
Copyright (C) 2023 leansoftX.com

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package model

// 远程服务器上的根目录
const CONST_REMOTE_REPO_ROOT string = "project"

// 容器内部的web ide的默认端口
const CONST_Container_WebIDEPort int = 3000

// 容器内部的JetBrains IDE的默认端口
const CONST_Container_JetBrainsIDEPort int = 8887

// 容器内部的Opensumi IDE的默认端口
const CONST_Container_OpensumiIDEPort int = 8000

const (
	CONST_DevContainer_PortDesc_Opensumi string = "tools-webide-opensumi"
	CONST_DevContainer_PortDesc_Vscode   string = "tools-webide-vscode"
	CONST_DevContainer_PortDesc_JB       string = "tools-webide-jb"

	CONST_DevContainer_PortDesc_SSH string = "tools-ssh"
)

// 容器内部的ssh端口
const CONST_Container_SSHPort int = 22

// 临时文件夹的路径（相对于工作目录）
const CONST_GlobalTempDirPath string = "/.ide/.temp"

// k8s 文件夹
const CONST_GlobalK8sDirPath string = "/.ide/.k8s"

// .ide 文件的路径
const CONST_GlobalIDEDirPath string = "/.ide/"

// SSH 默认的本地绑定端口，默认是6822 可能会因为端口占用被改为其他的端口
const CONST_Local_Default_BindingPort_SSH int = 6822

// 主机绑定的webide端口
const CONST_Local_Default_BindingPort_WebIDE int = 6800

// 默认的配置文件相对路径
const CONST_Default_ConfigRelativeFilePath = ".ide/.ide.yaml"

//const CONST_Default_K8S_ConfigRelativeFilePath = ".ide/.k8s.ide.yaml"

// 环境变量名称，映射到容器里面，当前用户uid,windows默认1000
const CONST_LOCAL_USER_UID = "LOCAL_USER_UID"

// 环境变量名称，映射到容器里面，当前用户gid,windows默认1000
const CONST_LOCAL_USER_GID = "LOCAL_USER_GID"

// 环境变量名称，容器ssh账号密码，root,smartide密码一样
const CONST_ENV_NAME_LoalUserPassword = "LOCAL_USER_PASSWORD"

// 容器ssh,端口转发用户
const CONST_DEV_CONTAINER_ROOT = "root"

// 自定义的用户
const CONST_DEV_CONTAINER_CUSTOM_USER = "smartide"

// 容器ssh账号默认密码
const CONST_DEV_CONTAINER_USER_DEFAULT_PASSWORD = "smartide123.@IDE"

// 默认登录地址
const CONST_LOGIN_URL = "https://dev.smartide.cn"

const CONST_WS_URL = "ws://dev.smartide.cn/smartide/ws"

// 模板文件夹的名称
const TMEPLATE_DIR_NAME = "templates"
