<!--
 * @Author: kenan
 * @Date: 2021-09-29 16:41:13
 * @LastEditors: kenan
 * @LastEditTime: 2021-10-13 20:00:58
 * @Description: file content
-->

## 资源文件

1.  

## **编译windows可执行程序（exe）**

1. mac
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o /build/smartide.exe main.go

- CGO_ENABLED
- GOOS：目标可执行程序运行操作系统，支持 darwin，freebsd，linux，windows
- GOARCH：目标可执行程序操作系统构架，包括 386，amd64，arm
- -o，编译后的文件保存路径和名称

2. windows 

``` 
SET CGO_ENABLED=0 
SET GOOS=windows 
SET GOARCH=amd64 
go build -o ./build/smartide.exe main.go
GO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -o smartide
```

**发布静态文件**
golang 在1.16后，引入了embed，可以灵活的导入静态文件，不需要把静态文件改为go后缀名这种麻烦的办法。


## 设置当前开发目录下的文件为环境变量
go build -o /usr/local/bin/smartide
chmod +x /usr/local/bin/smartide

***删除容器***
docker rm -f ide_product-service-db_1
docker rm -f ide_product-service-dev_1