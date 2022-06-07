#!/bin/bash
echo ">>>>> SmartIDE Server Deployment Start..."
mkdir ~/smartide
cd ~/smartide
echo ">>>>> SmartIDE Server Installing : 1.Basic Component"
# 1.Basic Component
echo ">>>>> SmartIDE Server Installing : 1.1 docker & docker-compose"
# 1.1 docker & docker-compose
curl -o- https://smartidedl.blob.core.chinacloudapi.cn/docker/linux/docker-install.sh | bash
echo ">>>>> SmartIDE Server Installing : 1.2 Git"
# 1.2 Git
sudo apt-get update && sudo apt-get install git -y
echo ">>>>> SmartIDE Server Installing : 1.3 Kubectl"
# 1.3 Kubectl
curl -LO https://smartidedl.blob.core.chinacloudapi.cn/kubectl/v1.23.0/bin/linux/amd64/kubectl
sudo install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

echo ">>>>> SmartIDE Server Installing : 2.MiniKube"
# 2.MiniKube Install And Configrate
echo ">>>>> SmartIDE Server Installing : 2.1 Minikube"
# 2.1 Minikube
curl -LO https://smartidedl.blob.core.chinacloudapi.cn/minikube/v1.24.0/minikube-linux-amd64
sudo install minikube-linux-amd64 /usr/local/bin/minikube
echo ">>>>> SmartIDE Server Installing : 2.2 Build Minikube Env"
# 2.2 Build Minikube Env
minikube start --driver=docker --cpus=2 --memory=4096mb
minikube addons enable ingress

echo ">>>>> SmartIDE Server Installing : 3.Tekton Pipeline"
# 3.Tekton Pipeline
echo ">>>>> SmartIDE Server Installing : 3.1 Tekton Pipeline And DashBoard"
# 3.1 Tekton Pipeline And DashBoard
kubectl apply -f https://raw.githubusercontent.com/SmartIDE/SmartIDE/main/server/deployment/pipeline/v0.32.0/tekton-release.yaml
kubectl apply -f https://raw.githubusercontent.com/SmartIDE/SmartIDE/main/server/deployment/dashboard/v0.32.0/tekton-dashboard-release.yaml
echo ">>>>> SmartIDE Server Installing : Tekton Trigger"
# 3.2 Tekton Trigger
kubectl apply -f https://raw.githubusercontent.com/SmartIDE/SmartIDE/main/server/deployment/trigger/v0.18.0/release.yaml
kubectl apply -f https://raw.githubusercontent.com/SmartIDE/SmartIDE/main/server/deployment/trigger/v0.18.0/interceptors.yaml
echo ">>>>> SmartIDE Server Installing : 3.3 Tekton SmartIDE Pipeline Configrate"
# 3.3 Tekton SmartIDE Pipeline Configrate
kubectl apply -f https://raw.githubusercontent.com/SmartIDE/SmartIDE/main/server/deployment/smartide-pipeline/dockerhub/trigger.yaml
kubectl apply -f https://raw.githubusercontent.com/SmartIDE/SmartIDE/main/server/deployment/smartide-pipeline/dockerhub/trigger-template.yaml
kubectl apply -f https://raw.githubusercontent.com/SmartIDE/SmartIDE/main/server/deployment/smartide-pipeline/dockerhub/trigger-binding.yaml
kubectl apply -f https://raw.githubusercontent.com/SmartIDE/SmartIDE/main/server/deployment/smartide-pipeline/dockerhub/trigger-event-listener.yaml

kubectl apply -f https://raw.githubusercontent.com/SmartIDE/SmartIDE/main/server/deployment/smartide-pipeline/dockerhub/task-smartide-cli-release.yaml
kubectl apply -f https://raw.githubusercontent.com/SmartIDE/SmartIDE/main/server/deployment/smartide-pipeline/dockerhub/pipeline-smartide-cli.yaml
sleep 20
kubectl apply -f https://raw.githubusercontent.com/SmartIDE/SmartIDE/main/server/deployment/smartide-pipeline/ingress-el-trigger-listener-smartide-cli.yaml

echo ">>>>> SmartIDE Server Installing : 4.SmartIDE Server"
# 4.SmartIDE Server
curl -LO https://raw.githubusercontent.com/SmartIDE/SmartIDE/main/server/deployment/docker-compose.yaml
curl -LO https://raw.githubusercontent.com/SmartIDE/SmartIDE/main/server/deployment/docker-compose.env
docker-compose -f docker-compose.yaml --env-file docker-compose.env up -d

echo ">>>>> SmartIDE Server Installing : 5.Build SmartIDE Server Network Connect With Minikube "
# 5.Build SmartIDE Server Network Connect With Minikube 
docker network connect smartide_smartide-server-network minikube

echo ">>>>> SmartIDE Server Deployment Finish : SmartIDE Server Success！"