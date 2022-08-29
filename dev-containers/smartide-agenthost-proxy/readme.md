<!--
 * @Date: 2022-08-26 09:59:05
 * @LastEditors: Jason Chen
 * @LastEditTime: 2022-08-26 10:10:13
 * @FilePath: /smartide/dev-containers/smartide-agenthost-proxy/readme.md
-->


```
cd \dev-containers\smartide-agenthost-proxy
docker build -f ./Dockerfile -t "smartide-agenthost-proxy:latest" .
docker tag $(docker images smartide-agenthost-proxy:latest -q) registry.cn-hangzhou.aliyuncs.com/smartide/smartide-agenthost-proxy:latest
```