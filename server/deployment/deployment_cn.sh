#!/bin/bash
# Author: SmartIDE
# Github: https://github.com/SmartIDE/SmartIDE

#######color code########
RED="31m"      
GREEN="32m"  
YELLOW="33m" 
BLUE="36m"
FUCHSIA="35m"

colorEcho(){
    COLOR=$1
    echo -e "\033[${COLOR}${@:2}\033[0m"
}

echo -e "$(colorEcho $YELLOW SmartIDE Server Deployment Start...)"

# 0.初始化安装SmartIDE Server机器的IP地址
echo -n -e "请输入本机对外服务的IP地址："
read serverIp

# 0.创建安装目录
mkdir -p ~/smartide-install
cd ~/smartide-install

# 1.Basic Component
echo -e "$(colorEcho $GREEN SmartIDE Server Deployment : 1.Basic Component)"
echo -e "$(colorEcho $BLUE SmartIDE Server Deployment : 1.1 docker and docker-compose)"
# 1.1 docker & docker-compose
curl -o- https://smartidedl.blob.core.chinacloudapi.cn/docker/linux/docker-install.sh | bash
sudo groupadd docker
sudo usermod -aG docker ${USER}
newgrp docker <<EONG
echo "Docker Will Restart..."
EONG
sudo chown $USER /var/run/docker.sock
sudo systemctl restart docker
docker ps
# 1.2 Git
echo -e "$(colorEcho $BLUE SmartIDE Server Deployment : 1.2 Git)"
if [ ! -e "/usr/bin/git" ] 
then
  sudo apt-get update && sudo apt-get install git -y
  git version
else
  echo "Git Already Installed."
fi
# 1.3 Kubectl
echo -e "$(colorEcho $BLUE SmartIDE Server Deployment : 1.3 Kubectl)"
curl -LO https://smartidedl.blob.core.chinacloudapi.cn/kubectl/v1.23.0/bin/linux/amd64/kubectl
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl
echo -e "$(colorEcho $GREEN SmartIDE Server Deployment : 1.Basic Component Installed Successfully.)"

echo -e "$(colorEcho $GREEN SmartIDE Server Deployment : 2.MiniKube)"
# 2.MiniKube Install And Configrate
# 2.1 Minikube
echo -e "$(colorEcho $BLUE SmartIDE Server Deployment : 2.1 Minikube Install)"
curl -LO https://smartidedl.blob.core.chinacloudapi.cn/minikube/v1.24.0/minikube-linux-amd64
sudo install minikube-linux-amd64 /usr/local/bin/minikube
echo -e "$(colorEcho $BLUE SmartIDE Server Deployment : 2.2 Build Minikube Env)"
# 2.2 Build Minikube Env
minikube delete
minikube start --image-mirror-country=cn --driver=docker --cpus=2 --memory=4096mb
minikube addons enable ingress

# 3.Tekton Pipeline
echo -e "$(colorEcho $GREEN SmartIDE Server Deployment : 3.Tekton Pipeline)"
echo -e "$(colorEcho $BLUE SmartIDE Server Deployment : 3.1 Kubectl Apply Tekton Pipeline And DashBoard)"
kubectl apply -f https://gitee.com/smartide/SmartIDE/raw/main/server/deployment/pipeline/v0.32.0/smartide-tekton-release.yaml
kubectl apply -f https://gitee.com/smartide/SmartIDE/raw/main/server/deployment/dashboard/v0.32.0/smartide-tekton-dashboard-release.yaml
# 3.2 Tekton Trigger
echo -e "$(colorEcho $BLUE SmartIDE Server Deployment : 3.2 Kubectl Apply Tekton Trigger)"
kubectl apply -f https://gitee.com/smartide/SmartIDE/raw/main/server/deployment/trigger/v0.18.0/smartide-release.yaml
kubectl apply -f https://gitee.com/smartide/SmartIDE/raw/main/server/deployment/trigger/v0.18.0/smartide-interceptor.yaml
# 3.3 Tekton SmartIDE Pipeline Configrate
echo -e "$(colorEcho $BLUE SmartIDE Server Deployment : 3.3 Kubectl Apply Tekton SmartIDE Pipeline Configrate)"
sleep 15
kubectl apply -f https://gitee.com/smartide/SmartIDE/raw/main/server/deployment/smartide-pipeline/aliyun/trigger.yaml
kubectl apply -f https://gitee.com/smartide/SmartIDE/raw/main/server/deployment/smartide-pipeline/aliyun/trigger-template.yaml
kubectl apply -f https://gitee.com/smartide/SmartIDE/raw/main/server/deployment/smartide-pipeline/aliyun/trigger-binding.yaml
kubectl apply -f https://gitee.com/smartide/SmartIDE/raw/main/server/deployment/smartide-pipeline/aliyun/trigger-event-listener.yaml
sleep 20
kubectl apply -f https://gitee.com/smartide/SmartIDE/raw/main/server/deployment/smartide-pipeline/aliyun/pipeline-smartide-cli.yaml
sleep 20
kubectl apply -f https://gitee.com/smartide/SmartIDE/raw/main/server/deployment/smartide-pipeline/aliyun/task-smartide-cli-release.yaml
kubectl apply -f https://gitee.com/smartide/SmartIDE/raw/main/server/deployment/smartide-pipeline/ingress-el-trigger-listener-smartide-cli.yaml
echo -e "$(colorEcho $GREEN SmartIDE Server Deployment : 3.Tekton Pipeline Installed Successfully.)"

# 4.SmartIDE Server
echo -e "$(colorEcho $GREEN SmartIDE Server Deployment : 4.SmartIDE Server)"
curl -LO https://gitee.com/smartide/SmartIDE/raw/main/server/deployment/docker-compose.yaml
curl -LO https://gitee.com/smartide/SmartIDE/raw/main/server/deployment/docker-compose_cn.env
curl -LO https://gitee.com/smartide/SmartIDE/raw/main/server/deployment/config.docker.yaml
sed -i 's/gva-web/'"$serverIp"'/g' config.docker.yaml

docker network create smartide-server-network
docker-compose -f docker-compose.yaml --env-file docker-compose_cn.env down
docker-compose -f docker-compose.yaml --env-file docker-compose_cn.env up -d
echo -e "$(colorEcho $GREEN SmartIDE Server Deployment : 4.SmartIDE Server Installed Successfully.)"

# 5.Build SmartIDE Server Network Connection With Minikube 
echo -e "$(colorEcho $GREEN SmartIDE Server Deployment : 5.Build SmartIDE Server Network Connection With Minikube.)"
docker network connect smartide-server-network minikube
echo -e "$(colorEcho $GREEN SmartIDE Server Deployment : 5.Build SmartIDE Server Network Connection With Minikube Successfully.)"
echo -e "$(colorEcho $YELLOW SmartIDE Server 服务地址：http://$serverIp:8080)"
echo -e "$(colorEcho $YELLOW SmartIDE Server Deployment Successfully！)"