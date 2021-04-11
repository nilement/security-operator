#!/bin/bash
# Kubelet edit script
set -e
kubectl patch node minikube-m02 --type=json -p='[{"op": "remove", "path": "/spec/configSource"}]'