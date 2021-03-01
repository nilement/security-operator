# Security Operator Project

<p>This is a project for Kubernetes security experiments.</p>


<p>
Test cluster to run experiments on 

1. `sh kind/cluster-with-registry.sh`
2. `kubectl apply -f nginx/deployment.yaml nginx/lb.yaml`
3. `kubectl port-forward service/example-service 13337:13337`
</p>

Current experiments: 

- Changing Kubelet service file permission on a Node. To be detected by kube-hunter.
- Uploading a different image to the image registry after deployment, before scaling up. Difference detected in logs.
