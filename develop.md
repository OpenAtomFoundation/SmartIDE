**编译windows可执行程序（exe）**

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
```

