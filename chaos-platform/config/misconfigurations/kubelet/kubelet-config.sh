#!/bin/bash
# Kubelet edit script
set -e
KUBE_TOKEN=$(cat /var/run/secrets/kubernetes.io/serviceaccount/token) 
curl -sSLk  -H "Authorization: Bearer $KUBE_TOKEN" "https://192.168.99.125:10250/configz" | jq '.kubeletconfig|.kind="KubeletConfiguration"|.apiVersion="kubelet.config.k8s.io/v1beta1"' > kubelet_configz_current
cat kubelet_configz_current | jq '.enableDebuggingHandlers = false' > edited.json
configmapname=$(kubectl -n kube-system create configmap test-config-5 --from-file=kubelet=edited.json --append-hash -o json | jq -r '.metadata|.name')
kubectl patch node minikube-m02 -p "{\"spec\":{\"configSource\":{\"configMap\":{\"name\":\"$configmapname\",\"namespace\":\"kube-system\",\"kubeletConfigKey\":\"kubelet\"}}}}"