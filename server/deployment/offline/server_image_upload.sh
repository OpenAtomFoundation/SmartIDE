#!/bin/bash
echo ">>>>> 1. Upload Tekton Pipelin Images"
IMAGE_REGISTRY=registry.cn-hangzhou.aliyuncs.com
IMAGE_NAMESPAGE=smartide

IMAGE=tekton-releases-tektoncd-pipeline-cmd-controller:v0.32.0
docker load < $IMAGE.tar
docker push $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE

IMAGE=tekton-releases-tektoncd-pipeline-cmd-kubeconfigwriter:v0.32.0
docker load < $IMAGE.tar
docker push $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE

IMAGE=tekton-releases-tektoncd-pipeline-cmd-git-init:v0.32.0
docker load < $IMAGE.tar
docker push $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE

IMAGE=tekton-releases-tektoncd-pipeline-cmd-entrypoint:v0.32.0
docker load < $IMAGE.tar
docker push $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE

IMAGE=tekton-releases-tektoncd-pipeline-cmd-nop:v0.32.0
docker load < $IMAGE.tar
docker push $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE

IMAGE=tekton-releases-tektoncd-pipeline-cmd-imagedigestexporter:v0.32.0
docker load < $IMAGE.tar
docker push $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE

IMAGE=tekton-releases-tektoncd-pipeline-cmd-pullrequest-init:v0.32.0
docker load < $IMAGE.tar
docker push $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE

IMAGE=cloudsdktool-cloud-sdk
docker load < $IMAGE.tar
docker push $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE

IMAGE=smartide/distroless-base
docker load < $IMAGE.tar
docker push $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE

IMAGE=tekton-releases-tektoncd-pipeline-cmd-webhook:v0.32.0
docker load < $IMAGE.tar
docker push $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE

echo ">>>>> 2. Upload Tekton DashBoard Images"
IMAGE=tekton-releases-tektoncd-dashboard-cmd-dashboard:v0.23.0
docker load < $IMAGE.tar
docker push $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE

echo ">>>>> 3. Upload Tekton Trigger Images"
IMAGE=tekton-releases-tektoncd-triggers-cmd-controller:v0.18.0
docker load < $IMAGE.tar
docker push $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE

IMAGE=tekton-releases-tektoncd-triggers-cmd-webhook:v0.18.0
docker load < $IMAGE.tar
docker push $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE

IMAGE=tekton-releases-tektoncd-triggers-cmd-eventlistenersink:v0.18.0
docker load < $IMAGE.tar
docker push $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE

echo ">>>>> 4. Upload SmartIDE Tekton CLI  Image"
IMAGE=smartide-cli:3175
docker load < $IMAGE.tar
docker push $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE

echo ">>>>> 5. Upload SmartIDE Server Images"
IMAGE=mysql:8.0.21
docker load < $IMAGE.tar
docker push $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE

IMAGE=redis:6.0.6
docker load < $IMAGE.tar
docker push $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE

IMAGE=portainer:1.24.2
docker load < $IMAGE.tar
docker push $IMAGE_REGISTRY/$IMAGE_NAMESPAGE/$IMAGE

echo ">>>>> 6. Upload Minikube Images"
IMAGE=google_containers
docker load < $IMAGE.tar
docker push $IMAGE_REGISTRY/$IMAGE





