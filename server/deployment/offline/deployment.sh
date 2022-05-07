#!/bin/bash
export DOCKER_HUB=registry.cn-hangzhou.aliyuncs.com
export DOCKER_HUB_NAMESPACE=smartide
export DOCKER_HUB_URI=${DOCKER_HUB}/${DOCKER_HUB_NAMESPACE}

echo ">>>>> SmartIDE Server Deployment Start..."
echo ">>>>> SmartIDE Server Installing : 1.Basic Component"
# 1.Basic Component
echo ">>>>> SmartIDE Server Installing : 1.1 docker & docker-compose"
# 1.1 docker
tar -zxvf docker-install.tar.gz
cd docker-install
chmod +x install.sh
sudo ./install.sh -f docker-20.10.14.tgz
sudo groupadd docker
sudo usermod -aG docker ${USER}
sudo systemctl restart docker
docker ps
# 1.2 docker-compose
sudo cp docker-compose-Linux-x86_64 /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
docker-compose version
echo ">>>>> SmartIDE Server Installing : 1.2 Git"
# 1.2 Git(默认系统自带git，若需要单独进行离线安装)
sudo apt-get update && sudo apt-get install git -y
echo ">>>>> SmartIDE Server Installing : 1.3 Kubectl"
# 1.3 Kubectl
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

echo ">>>>> SmartIDE Server Installing : 2.MiniKube"
# 2.MiniKube Install And Configrate
echo ">>>>> SmartIDE Server Installing : 2.1 Minikube"
# 2.1 Minikube
sudo install minikube-linux-amd64 /usr/local/bin/minikube
echo ">>>>> SmartIDE Server Installing : 2.2 Build Minikube Env"
# 2.2 Build Minikube Env
minikube start --driver=docker --cpus=2 --memory=4096mb --cache-images=true
minikube addons enable ingress

echo ">>>>> SmartIDE Server Installing : 3.Tekton Pipeline"
# 3.Tekton Pipeline
echo ">>>>> SmartIDE Server Installing : 3.1 Tekton Pipeline And DashBoard"
# 3.1 Tekton Pipeline And DashBoard
envsubst < app.yaml | kubectl apply -f -
kubectl apply -f ./pipeline/v0.32.0/smartide-tekton-release.yaml
kubectl apply -f ./dashboard/v0.32.0/smartide-tekton-dashboard-release.yaml
echo ">>>>> SmartIDE Server Installing : Tekton Trigger"
# 3.2 Tekton Trigger
kubectl apply -f https://gitee.com/smartide/SmartIDE/raw/main/server/deployment/trigger/v0.18.0/smartide-release.yaml
kubectl apply -f https://gitee.com/smartide/SmartIDE/raw/main/server/deployment/trigger/v0.18.0/smartide-interceptor.yaml
echo ">>>>> SmartIDE Server Installing : 3.3 Tekton SmartIDE Pipeline Configrate"
# 3.3 Tekton SmartIDE Pipeline Configrate
kubectl apply -f ./smartide-pipeline/aliyun/trigger.yaml
kubectl apply -f ./smartide-pipeline/aliyun/trigger-template.yaml
kubectl apply -f ./smartide-pipeline/aliyun/trigger-binding.yaml
kubectl apply -f ./smartide-pipeline/aliyun/trigger-event-listener.yaml
kubectl apply -f ./smartide-pipelinetask-smartide-cli-release.yaml
kubectl apply -f ./smartide-pipeline/pipeline-smartide-cli.yaml
kubectl apply -f ./smartide-pipeline/ingress-el-trigger-listener-smartide-cli.yaml

echo ">>>>> SmartIDE Server Installing : 4.SmartIDE Server"
# 4.SmartIDE Server
docker network create martide-server-network
docker-compose -f docker-compose.yaml --env-file docker-compose_cn.env up -d

echo ">>>>> SmartIDE Server Installing : 5.Build SmartIDE Server Network Connect With Minikube "
# 5.Build SmartIDE Server Network Connect With Minikube
docker network connect smartide-server-network minikube

echo ">>>>> SmartIDE Server Deployment Finish : SmartIDE Server Success！"