# Security Operator Project

chaos-platform - Kubernetes operator for running experiments.
malicious-cluster - Experiments and validators. 

## Recreating the test environment

Test environment on Google Cloud can be recreated with the following commands.

`gcloud compute networks create vulnerable-k8s --subnet-mode custom`

`gcloud compute networks subnets create k8s-nodes \
  --network vulnerable-k8s \
  --range 10.240.0.0/24`
  
`gcloud compute firewall-rules create vulnerable-k8s-allow-internal \
  --allow tcp,udp,icmp,ipip \
  --network vulnerable-k8s \
  --source-ranges 10.240.0.0/24`
  
  
`gcloud compute firewall-rules create vulnerable-k8s-allow-external \
  --allow tcp:22,tcp:6443,icmp \
  --network vulnerable-k8s \
  --source-ranges 0.0.0.0/0`
  
  
`gcloud compute firewall-rules create vulnerable-k8s-allow-matas-external \
  --allow tcp \
  --network vulnerable-k8s \
  --destination-ranges 213.32.242.0/24`
