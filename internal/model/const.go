/*
 * @Author: kenan
 * @Date: 2021-12-30 19:48:33
 * @LastEditors: kenan
 * @LastEditTime: 2022-02-18 16:13:41
 * @FilePath: /smartide-cli/internal/model/const.go
 * @Description:
 *
 * Copyright (c) 2022 by kenanlu@leansoftx.com, All Rights Reserved.
 */
package model

// 容器内部的web ide的默认端口
const CONST_Container_WebIDEPort int = 3000

// 容器内部的JetBrains IDE的默认端口
const CONST_Container_JetBrainsIDEPort int = 8887

// 容器内部的ssh端口
const CONST_Container_SSHPort int = 22

// 临时文件夹的路径（相对于工作目录）
const CONST_TempDirPath string = "/.ide/.temp"

// .ide 文件的路径
const CONST_IDEDirPath string = "/.ide/"

// SSH 默认的本地绑定端口，默认是6822 可能会因为端口占用被改为其他的端口
const CONST_Local_Default_BindingPort_SSH int = 6822

// 主机绑定的webide端口
const CONST_Local_Default_BindingPort_WebIDE int = 6800

// 默认的配置文件相对路径
const CONST_Default_ConfigRelativeFilePath = ".ide/.ide.yaml"

//环境变量名称，映射到容器里面，当前用户uid,windows默认1000
const CONST_LOCAL_USER_UID = "LOCAL_USER_UID"

//环境变量名称，映射到容器里面，当前用户gid,windows默认1000
const CONST_LOCAL_USER_GID = "LOCAL_USER_GID"

//环境变量名称，容器ssh账号密码，root,smartide密码一样
const CONST_ENV_NAME_LoalUserPassword = "LOCAL_USER_PASSWORD"

// 容器ssh,端口转发用户
const CONST_DEV_CONTAINER_ROOT = "root"

// 自定义的用户
const CONST_DEV_CONTAINER_CUSTOM_USER = "smartide"

//容器ssh账号默认密码
const CONST_DEV_CONTAINER_USER_DEFAULT_PASSWORD = "smartide123.@IDE"

//默认登录地址
const CONST_LOGIN_URL = "https://dev.smartide.cn"
