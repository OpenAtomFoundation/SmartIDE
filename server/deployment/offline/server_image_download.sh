#!/bin/bash
echo ">>>>> 1. Download Tekton Pipelin Images"
IMAGE_REGISTRY=registry.cn-hangzhou.aliyuncs.com
IMAGE_NAMESPAGE=smartide

IMAGE=tekton-releases-tektoncd-pipeline-cmd-controller:v0.32.0
docker pull $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE
docker save -o $IMAGE.tar $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE

IMAGE=tekton-releases-tektoncd-pipeline-cmd-kubeconfigwriter:v0.32.0
docker pull $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE
docker save -o $IMAGE.tar $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE

IMAGE=tekton-releases-tektoncd-pipeline-cmd-git-init:v0.32.0
docker pull $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE
docker save -o $IMAGE.tar $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE

IMAGE=tekton-releases-tektoncd-pipeline-cmd-entrypoint:v0.32.0
docker pull $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE
docker save -o $IMAGE.tar $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE

IMAGE=tekton-releases-tektoncd-pipeline-cmd-nop:v0.32.0
docker pull $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE
docker save -o $IMAGE.tar $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE

IMAGE=tekton-releases-tektoncd-pipeline-cmd-imagedigestexporter:v0.32.0
docker pull $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE
docker save -o $IMAGE.tar $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE

IMAGE=tekton-releases-tektoncd-pipeline-cmd-pullrequest-init:v0.32.0
docker pull $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE
docker save -o $IMAGE.tar $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE

IMAGE=cloudsdktool-cloud-sdk
docker pull $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE
docker save -o $IMAGE.tar $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE

IMAGE=smartide/distroless-base
docker pull $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE
docker save -o $IMAGE.tar $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE

IMAGE=tekton-releases-tektoncd-pipeline-cmd-webhook:v0.32.0
docker pull $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE
docker save -o $IMAGE.tar $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE

echo ">>>>> 2. Download Tekton DashBoard Images"
IMAGE=tekton-releases-tektoncd-dashboard-cmd-dashboard:v0.23.0
docker pull $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE
docker save -o $IMAGE.tar $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE

echo ">>>>> 3. Download Tekton Trigger Images"
IMAGE=tekton-releases-tektoncd-triggers-cmd-controller:v0.18.0
docker pull $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE
docker save -o $IMAGE.tar $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE

IMAGE=tekton-releases-tektoncd-triggers-cmd-webhook:v0.18.0
docker pull $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE
docker save -o $IMAGE.tar $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE

IMAGE=tekton-releases-tektoncd-triggers-cmd-eventlistenersink:v0.18.0
docker pull $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE
docker save -o $IMAGE.tar $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE

echo ">>>>> 4. Download SmartIDE Tekton CLI  Image"
IMAGE=smartide-cli:3175
docker pull $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE
docker save -o $IMAGE.tar $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE

echo ">>>>> 5. Download Dev Images"
IMAGE=smartide-java-v2-jetbrains-idea:2021.2.3-openjdk-11-jdk-2081
docker pull $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE
docker save -o $IMAGE.tar $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE

echo ">>>>> 6. Download Minikube Images"
IMAGE_REGISTRY=registry.cn-hangzhou.aliyuncs.com
IMAGE_NAMESPAGE=smartide
IMAGE=storage-provisioner:v5
docker pull $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE
docker save -o $IMAGE.tar $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE

echo ">>>>> 7. Download SmartIDE Server Images"
IMAGE_REGISTRY=acrsmartide.azurecr.io

IMAGE=smartide-web:3294
docker pull $IMAGE_REGISTRY/$IMAGE
docker save -o $IMAGE.tar $IMAGE_REGISTRY/$IMAGE

IMAGE=smartide-api:3294
docker pull $IMAGE_REGISTRY/$IMAGE
docker save -o $IMAGE.tar $IMAGE_REGISTRY/$IMAGE

IMAGE=mysql:8.0.21
docker pull $IMAGE_REGISTRY/$IMAGE
docker save -o $IMAGE.tar $IMAGE_REGISTRY/$IMAGE

IMAGE=redis:6.0.6
docker pull $IMAGE_REGISTRY/$IMAGE
docker save -o $IMAGE.tar $IMAGE_REGISTRY/$IMAGE

IMAGE=portainer:1.24.2
docker pull $IMAGE_REGISTRY/$IMAGE
docker save -o $IMAGE.tar $IMAGE_REGISTRY/$IMAGE

IMAGE=phpmyadmin:5.1.1
docker pull $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE
docker save -o $IMAGE.tar $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE




