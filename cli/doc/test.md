
## smartide start
```
cd ~
[[ ! -d "projects/boat-house/boathouse-calculator" ]] && git clone https://github.com/idcf-boat-house/boathouse-calculator.git ~/projects/boat-house/boathouse-calculator || echo "git 库已存在"
cd ~/projects/boat-house/boathouse-calculator

smartide remove
smartide start

smartide restart

smartide stop
smartide restart

---

smartide start -f .ide/.ide.linkfile.yaml
smartide stop -f .ide/.ide.linkfile.yaml
smartide restart -f .ide/.ide.linkfile.yaml


```

## samrtide vm start
### 远程服务器使用ssh方式登录
### git库私有
### 链接docker-compose

```
## 在mac上执行ssh在运行命令可以，但是到windows上直接报错
ssh -T localadmin@experiment-002.southeastasia.cloudapp.azure.com<<'ENDSSH'
[[ -n $(docker ps -qa) ]] && docker rm -f $(docker ps -qa)
[[ -n $(docker images -q) ]] && docker rmi $(docker images -q) 
rm -rf ~/project 
echo "ssh"
exit
ENDSSH

smartide vm start --host experiment-002.southeastasia.cloudapp.azure.com --username localadmin --repourl git@ssh.dev.azure.com:v3/leansoftx/smartide/smartide-cli --branch releases/release-5 --filepath .ide/.ide.linkfile.yaml -d

smartide vm start --host experiment-002.southeastasia.cloudapp.azure.com --username localadmin --repourl git@ssh.dev.azure.com:v3/leansoftx/smartide/smartide-cli --branch feature/leixu/feature486-golive -d

smartide  start --host experiment-002.southeastasia.cloudapp.azure.com --username localadmin --repourl https://github.com/idcf-boat-house/boathouse-calculator.git -d
```
