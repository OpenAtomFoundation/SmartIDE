# Author: SmartIDE
# Github: https://github.com/SmartIDE/SmartIDE

# K8S Ingress服务及Service配置
kubectl apply -f https://gitee.com/smartide/SmartIDE/raw/main/deployment/k8s/ingress-controller/ingress-controller.yaml
kubectl apply -f https://gitee.com/smartide/SmartIDE/raw/main/deployment/k8s/ingress-controller/ingress-controller-service.yaml
# K8S Service证书配置
kubectl apply -f https://gitee.com/smartide/SmartIDE/raw/main/deployment/k8s/cert-manager/cert-manager.yaml
kubectl apply -f https://gitee.com/smartide/SmartIDE/raw/main/deployment/k8s/cert-manager/cluster-issuer.yaml
# K8S StroageClass配置
kubectl apply -f https://gitee.com/smartide/SmartIDE/raw/main/deployment/k8s/file-storageclass/smartide-file-storageclass.yaml

