import json, yaml
import logging

backup = open('./backup/kube-apiserver.yaml', 'r')
backed_up_config = yaml.load(backup, Loader=yaml.FullLoader)
backup.close()

api_pod = None
with open('/etc/kubernetes/manifests/kube-apiserver.yaml', 'r') as file:
    api_pod = yaml.load(file, Loader=yaml.FullLoader)

with open('/etc/kubernetes/manifests/kube-apiserver.yaml', 'w') as file:
    api_pod['spec']['containers'][0]['command'] = backed_up_config
    yaml.dump(api_pod, file)

logging.info("Old configuration restored")