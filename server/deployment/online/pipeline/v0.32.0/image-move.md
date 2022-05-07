# Copy Image to Aliyun

TODO: create a github action 

```shell
docker tag gcr.io/tekton-releases/github.com/tektoncd/pipeline/cmd/controller:v0.32.0@sha256:0e4f92e95c9ae8140ddfc8751bb54cf54e1b00d27aa542c11d5ad8663c5067ae registry.cn-hangzhou.aliyuncs.com/smartide/tekton-releases-tektoncd-pipeline-cmd-controller:v0.32.0
docker push registry.cn-hangzhou.aliyuncs.com/smartide/tekton-releases-tektoncd-pipeline-cmd-controller:v0.32.0

docker login --username=hi20172766@aliyun.com registry.cn-hangzhou.aliyuncs.com

"-kubeconfig-writer-image", "gcr.io/tekton-releases/github.com/tektoncd/pipeline/cmd/kubeconfigwriter:v0.32.0@sha256:32fec74288f52ede279f091d8bac91d48ff6538fa3290251339b0075c59d0947", 
docker pull gcr.io/tekton-releases/github.com/tektoncd/pipeline/cmd/kubeconfigwriter:v0.32.0@sha256:32fec74288f52ede279f091d8bac91d48ff6538fa3290251339b0075c59d0947
docker tag gcr.io/tekton-releases/github.com/tektoncd/pipeline/cmd/kubeconfigwriter:v0.32.0@sha256:32fec74288f52ede279f091d8bac91d48ff6538fa3290251339b0075c59d0947 registry.cn-hangzhou.aliyuncs.com/smartide/tekton-releases-tektoncd-pipeline-cmd-kubeconfigwriter:v0.32.0
docker push registry.cn-hangzhou.aliyuncs.com/smartide/tekton-releases-tektoncd-pipeline-cmd-kubeconfigwriter:v0.32.0

"-git-image", "gcr.io/tekton-releases/github.com/tektoncd/pipeline/cmd/git-init:v0.32.0@sha256:fe3310b87b9fad4b5139ac93f0e570c25fb97dcb64a876a5b8eebbc877fc12e8", 
docker pull gcr.io/tekton-releases/github.com/tektoncd/pipeline/cmd/git-init:v0.32.0@sha256:fe3310b87b9fad4b5139ac93f0e570c25fb97dcb64a876a5b8eebbc877fc12e8 
docker tag gcr.io/tekton-releases/github.com/tektoncd/pipeline/cmd/git-init:v0.32.0@sha256:fe3310b87b9fad4b5139ac93f0e570c25fb97dcb64a876a5b8eebbc877fc12e8 registry.cn-hangzhou.aliyuncs.com/smartide/tekton-releases-tektoncd-pipeline-cmd-git-init:v0.32.0
docker push registry.cn-hangzhou.aliyuncs.com/smartide/tekton-releases-tektoncd-pipeline-cmd-git-init:v0.32.0


"-entrypoint-image", "gcr.io/tekton-releases/github.com/tektoncd/pipeline/cmd/entrypoint:v0.32.0@sha256:7f50901900925357460e6c6c985580f0b69c0d316ade75965228adb8b081614e", 
docker pull gcr.io/tekton-releases/github.com/tektoncd/pipeline/cmd/entrypoint:v0.32.0@sha256:7f50901900925357460e6c6c985580f0b69c0d316ade75965228adb8b081614e
docker tag gcr.io/tekton-releases/github.com/tektoncd/pipeline/cmd/entrypoint:v0.32.0@sha256:7f50901900925357460e6c6c985580f0b69c0d316ade75965228adb8b081614e registry.cn-hangzhou.aliyuncs.com/smartide/tekton-releases-tektoncd-pipeline-cmd-entrypoint:v0.32.0
docker push registry.cn-hangzhou.aliyuncs.com/smartide/tekton-releases-tektoncd-pipeline-cmd-entrypoint:v0.32.0


"-nop-image", "gcr.io/tekton-releases/github.com/tektoncd/pipeline/cmd/nop:v0.32.0@sha256:a8ffddd75b7a7078d5c07d09259d7c5db04614b4c5ba5c43e99b0632034f2479", 
docker pull gcr.io/tekton-releases/github.com/tektoncd/pipeline/cmd/nop:v0.32.0@sha256:a8ffddd75b7a7078d5c07d09259d7c5db04614b4c5ba5c43e99b0632034f2479
docker tag gcr.io/tekton-releases/github.com/tektoncd/pipeline/cmd/nop:v0.32.0@sha256:a8ffddd75b7a7078d5c07d09259d7c5db04614b4c5ba5c43e99b0632034f2479 registry.cn-hangzhou.aliyuncs.com/smartide/tekton-releases-tektoncd-pipeline-cmd-nop:v0.32.0
docker push registry.cn-hangzhou.aliyuncs.com/smartide/tekton-releases-tektoncd-pipeline-cmd-nop:v0.32.0

"-imagedigest-exporter-image", "gcr.io/tekton-releases/github.com/tektoncd/pipeline/cmd/imagedigestexporter:v0.32.0@sha256:2b39f19517523df8a00a366a0d3adb815ca2623fc9c51f05dd290773d5d308c7", 
docker pull gcr.io/tekton-releases/github.com/tektoncd/pipeline/cmd/imagedigestexporter:v0.32.0@sha256:2b39f19517523df8a00a366a0d3adb815ca2623fc9c51f05dd290773d5d308c7
docker tag gcr.io/tekton-releases/github.com/tektoncd/pipeline/cmd/imagedigestexporter:v0.32.0@sha256:2b39f19517523df8a00a366a0d3adb815ca2623fc9c51f05dd290773d5d308c7 registry.cn-hangzhou.aliyuncs.com/smartide/tekton-releases-tektoncd-pipeline-cmd-imagedigestexporter:v0.32.0
docker push registry.cn-hangzhou.aliyuncs.com/smartide/tekton-releases-tektoncd-pipeline-cmd-imagedigestexporter:v0.32.0

"-pr-image", "gcr.io/tekton-releases/github.com/tektoncd/pipeline/cmd/pullrequest-init:v0.32.0@sha256:632b5086dba4d7f30f5b1e77f9e5e551b06c9e897cf2afc93e100b26f9c32e39",
docker pull gcr.io/tekton-releases/github.com/tektoncd/pipeline/cmd/pullrequest-init:v0.32.0@sha256:632b5086dba4d7f30f5b1e77f9e5e551b06c9e897cf2afc93e100b26f9c32e39
docker tag gcr.io/tekton-releases/github.com/tektoncd/pipeline/cmd/pullrequest-init:v0.32.0@sha256:632b5086dba4d7f30f5b1e77f9e5e551b06c9e897cf2afc93e100b26f9c32e39 registry.cn-hangzhou.aliyuncs.com/smartide/tekton-releases-tektoncd-pipeline-cmd-pullrequest-init:v0.32.0
docker push registry.cn-hangzhou.aliyuncs.com/smartide/tekton-releases-tektoncd-pipeline-cmd-pullrequest-init:v0.32.0


"-gsutil-image", "gcr.io/google.com/cloudsdktool/cloud-sdk@sha256:27b2c22bf259d9bc1a291e99c63791ba0c27a04d2db0a43241ba0f1f20f4067f",
docker pull gcr.io/google.com/cloudsdktool/cloud-sdk@sha256:27b2c22bf259d9bc1a291e99c63791ba0c27a04d2db0a43241ba0f1f20f4067f
docker tag gcr.io/google.com/cloudsdktool/cloud-sdk@sha256:27b2c22bf259d9bc1a291e99c63791ba0c27a04d2db0a43241ba0f1f20f4067f registry.cn-hangzhou.aliyuncs.com/smartide/cloudsdktool-cloud-sdk
docker push  registry.cn-hangzhou.aliyuncs.com/smartide/cloudsdktool-cloud-sdk


# The shell image must be root in order to create directories and copy files to PVCs.
# gcr.io/distroless/base:debug as of October 21, 2021
# image shall not contains tag, so it will be supported on a runtime like cri-o
"-shell-image", "gcr.io/distroless/base@sha256:cfdc553400d41b47fd231b028403469811fcdbc0e69d66ea8030c5a0b5fbac2b",
docker pull gcr.io/distroless/base@sha256:cfdc553400d41b47fd231b028403469811fcdbc0e69d66ea8030c5a0b5fbac2b
docker tag gcr.io/distroless/base@sha256:cfdc553400d41b47fd231b028403469811fcdbc0e69d66ea8030c5a0b5fbac2b registry.cn-hangzhou.aliyuncs.com/smartide/distroless-base
docker push registry.cn-hangzhou.aliyuncs.com/smartide/distroless-base


# for script mode to work with windows we need a powershell image
# pinning to nanoserver tag as of July 15 2021
"-shell-image-win", "mcr.microsoft.com/powershell:nanoserver@sha256:b6d5ff841b78bdf2dfed7550000fd4f3437385b8fa686ec0f010be24777654d6"
docker pull mcr.microsoft.com/powershell:nanoserver@sha256:b6d5ff841b78bdf2dfed7550000fd4f3437385b8fa686ec0f010be24777654d6
docker tag mcr.microsoft.com/powershell:nanoserver@sha256:b6d5ff841b78bdf2dfed7550000fd4f3437385b8fa686ec0f010be24777654d6 registry.cn-hangzhou.aliyuncs.com/smartide/powershell:nanoserver
docker push registry.cn-hangzhou.aliyuncs.com/smartide/powershell:nanoserver



gcr.io/tekton-releases/github.com/tektoncd/pipeline/cmd/webhook:v0.32.0@sha256:f0e31a5b1218bef6ad6323c05b4ed54412555accf542ac8a9dd0221629f33189
docker pull gcr.io/tekton-releases/github.com/tektoncd/pipeline/cmd/webhook:v0.32.0@sha256:f0e31a5b1218bef6ad6323c05b4ed54412555accf542ac8a9dd0221629f33189
docker tag gcr.io/tekton-releases/github.com/tektoncd/pipeline/cmd/webhook:v0.32.0@sha256:f0e31a5b1218bef6ad6323c05b4ed54412555accf542ac8a9dd0221629f33189 registry.cn-hangzhou.aliyuncs.com/smartide/tekton-releases-tektoncd-pipeline-cmd-webhook:v0.32.0
docker push registry.cn-hangzhou.aliyuncs.com/smartide/tekton-releases-tektoncd-pipeline-cmd-webhook:v0.32.0

gcr.io/tekton-releases/github.com/tektoncd/dashboard/cmd/dashboard:v0.23.0@sha256:4f70cd5f10bb6c8594b7810cf1fd8a8950d535ef0bb95e2c5f214a620646d720
docker pull gcr.io/tekton-releases/github.com/tektoncd/dashboard/cmd/dashboard:v0.23.0@sha256:4f70cd5f10bb6c8594b7810cf1fd8a8950d535ef0bb95e2c5f214a620646d720
docker tag gcr.io/tekton-releases/github.com/tektoncd/dashboard/cmd/dashboard:v0.23.0@sha256:4f70cd5f10bb6c8594b7810cf1fd8a8950d535ef0bb95e2c5f214a620646d720 registry.cn-hangzhou.aliyuncs.com/smartide/tekton-releases-tektoncd-dashboard-cmd-dashboard:v0.23.0
docker push registry.cn-hangzhou.aliyuncs.com/smartide/tekton-releases-tektoncd-dashboard-cmd-dashboard:v0.23.0
```