

## 目录结构 
```
- cmd             运行命令
- doc             文档
- internal        内部使用的包
- pkg             可共用的包
```


## 多语言
1. 引用
- 根据语言加载对应的语言文件
go get github.com/leansoftX/i18n
- 变量命名
 {类型}\_{模块}\_[{子模块}_]\_{文本内容描述}

   - 类型， error、warn、info
   - 模块， 比如start、list、stop、remove、get... , 还有可能是公共的
   - 子模块， 比如workspace、help、help_flag


2. 示例代码
```
var i18nInstance = i18n.GetInstance()
i18nInstance.Remove.Info_project_dir_removing
```

## 交叉编译

1. mac
> CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o /build/smartide.exe main.go

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


## 清理环境（在服务器上运行）
```
docker rm -f $(docker ps -qa)
docker rmi $(docker images -q)
rm -rf ~/project 
```

## 设置当前开发目录下的文件为环境变量
macos
``` macos
go build -o /usr/local/bin/smartide -ldflags="-X 'main.BuildTime=$(date "+%Y-%m-%d %H:%M:%S")' -w -s"
chmod +x /usr/local/bin/smartide

```

windows - cmd
``` cmd
go build -o "C:\Program Files (x86)\SmartIDE\SmartIDE.exe" -ldflags="-X 'main.BuildTime=%date:~0,4%-%date:~5,2%-%date:~8,2% %time%'"

```

windows - powershell
``` powershell
go build -o "C:\Program Files (x86)\SmartIDE\SmartIDE.exe" -ldflags="-X 'main.BuildTime=$(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')'"
```

linux 
```
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /usr/local/bin/smartide -ldflags="-X 'main.BuildTime=$(date "+%Y-%m-%d %H:%M:%S")' -w -s"
chmod +x /usr/local/bin/smartide
```

## 压缩
```
## 示例，upx -9 -o <压缩后文件路径> <原始文件路径>
upx -9 -o /usr/local/bin/se /usr/local/bin/smartide
chmod +x /usr/local/bin/se
```

## linux 版本安装脚本
```
curl -OL  "https://smartidedl.blob.core.chinacloudapi.cn/builds/$(curl -L -s https://smartidedl.blob.core.chinacloudapi.cn/builds/stable.txt)/smartide-linux" \
&& sudo mv -f smartide-linux /usr/local/bin/smartide \
&& sudo ln -s -f /usr/local/bin/smartide /usr/local/bin/se \
&& sudo chmod +x /usr/local/bin/smartide 
```

## 维护命令

**发布静态文件**

golang 在1.16后，引入了embed，可以灵活的导入静态文件，不需要把静态文件改为go后缀名这种麻烦的办法。

***删除容器***
```
docker rm -f ide_product-service-db_1
docker rm -f ide_product-service-dev_1
```
***merge release分支***
```
git checkout releases/release-21 && git pull && git checkout - && git merge releases/release-21
```
***重置某个分支***
```
git branch -D releases/release-6
git fetch
git checkout releases/release-6
git pull
git checkout -
```

***发布smmartide cli的镜像，并部署***
1. ***上传镜像*** 运行smartide cli流水线（codesign），上传最新的镜像到阿里镜像仓库、docker hub 镜像仓库
2. ***更新tekton中镜像的版本*** 更新 https://github.com/SmartIDE/smartide-tekton-install 中的镜像版本为流水线的build number
3. ***部署tekton流水线*** 运行 smartide server 对应的流水线，重新部署tekton流水线，流水线会自动调用github上的tekton yaml文件


## 开发技巧
``` bash
## 运行单元测试
go test ./...

## 格式化当前项目下的所有go文件中的代码
go fmt

## 清除没有使用的引用
go mod tidy -v

## 统计代码行数
export GOPATH="$HOME/go" 
export PATH="$PATH:$GOPATH/bin"
go get -u github.com/hhatto/gocloc/cmd/gocloc
gocloc .
```


