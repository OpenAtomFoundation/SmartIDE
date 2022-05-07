# Copy Image to Aliyun


```shell

docker login --username=hi20172766@aliyun.com registry.cn-hangzhou.aliyuncs.com

# gcr.io/tekton-releases/github.com/tektoncd/triggers/cmd/controller:v0.18.0@sha256:c9bac56feb04c16a1b483a7fe50a723022c0f1dfe920d6704ca7566de8d473cf

docker pull gcr.io/tekton-releases/github.com/tektoncd/triggers/cmd/controller:v0.18.0@sha256:c9bac56feb04c16a1b483a7fe50a723022c0f1dfe920d6704ca7566de8d473cf
docker tag gcr.io/tekton-releases/github.com/tektoncd/triggers/cmd/controller:v0.18.0@sha256:c9bac56feb04c16a1b483a7fe50a723022c0f1dfe920d6704ca7566de8d473cf registry.cn-hangzhou.aliyuncs.com/smartide/tekton-releases-tektoncd-triggers-cmd-controller:v0.18.0
docker push registry.cn-hangzhou.aliyuncs.com/smartide/tekton-releases-tektoncd-triggers-cmd-controller:v0.18.0


# gcr.io/tekton-releases/github.com/tektoncd/triggers/cmd/eventlistenersink:v0.18.0@sha256:9453f8184a476433f9223172f75790efeedb0780172ba9bcaa564d6987d85c2b

docker pull gcr.io/tekton-releases/github.com/tektoncd/triggers/cmd/eventlistenersink:v0.18.0@sha256:9453f8184a476433f9223172f75790efeedb0780172ba9bcaa564d6987d85c2b
docker tag gcr.io/tekton-releases/github.com/tektoncd/triggers/cmd/eventlistenersink:v0.18.0@sha256:9453f8184a476433f9223172f75790efeedb0780172ba9bcaa564d6987d85c2b registry.cn-hangzhou.aliyuncs.com/smartide/tekton-releases-tektoncd-triggers-cmd-eventlistenersink:v0.18.0
docker push registry.cn-hangzhou.aliyuncs.com/smartide/tekton-releases-tektoncd-triggers-cmd-eventlistenersink:v0.18.0


# gcr.io/tekton-releases/github.com/tektoncd/triggers/cmd/webhook:v0.18.0@sha256:ccd1613eb4b64ff732e092619e9fb4594aa617d2b93dbd46ff091be394bfb0d7

docker pull gcr.io/tekton-releases/github.com/tektoncd/triggers/cmd/webhook:v0.18.0@sha256:ccd1613eb4b64ff732e092619e9fb4594aa617d2b93dbd46ff091be394bfb0d7
docker tag gcr.io/tekton-releases/github.com/tektoncd/triggers/cmd/webhook:v0.18.0@sha256:ccd1613eb4b64ff732e092619e9fb4594aa617d2b93dbd46ff091be394bfb0d7 registry.cn-hangzhou.aliyuncs.com/smartide/tekton-releases-tektoncd-triggers-cmd-webhook:v0.18.0
docker push registry.cn-hangzhou.aliyuncs.com/smartide/tekton-releases-tektoncd-triggers-cmd-webhook:v0.18.0


# gcr.io/tekton-releases/github.com/tektoncd/triggers/cmd/interceptors:v0.18.0@sha256:ca8025d2471deb7f51826227b89634413c465c66e785565c8e4db02b8f2c00e9

docker pull gcr.io/tekton-releases/github.com/tektoncd/triggers/cmd/interceptors:v0.18.0@sha256:ca8025d2471deb7f51826227b89634413c465c66e785565c8e4db02b8f2c00e9
docker tag gcr.io/tekton-releases/github.com/tektoncd/triggers/cmd/interceptors:v0.18.0@sha256:ca8025d2471deb7f51826227b89634413c465c66e785565c8e4db02b8f2c00e9 registry.cn-hangzhou.aliyuncs.com/smartide/tekton-releases-tektoncd-triggers-cmd-interceptors:v0.18.0
docker push registry.cn-hangzhou.aliyuncs.com/smartide/tekton-releases-tektoncd-triggers-cmd-interceptors:v0.18.0



```