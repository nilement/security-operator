import requests, json, yaml
import os, argparse
import subprocess
import time
import psutil
import logging

def backup_config(config):
    w = open("./backup/kube-apiserver.yaml", "w")
    yaml.dump(config, w)

def set_param(config, param, value):
    for c in range(len(config)):
        s = config[c].split('=')
        if s[0] == param:
            newc = s[0] + '=' + value
            config[c] = newc
            return
    
    config.append(param + '=' + str(value))

def push_param(config, param, value):
    for c in range(len(config)):
        s = config[c].split('=')
        if s[0] == param:
            newc = s[0] + '=' + value + ',' + s[1]
            config[c] = newc
            return

def ensure_ready_probe(misconfigs):
    ensured = False
    attempts = 0

    while not ensured:
        if attempts > 5:
            return False
        time.sleep(5)
        apiserver_proc = None
        for proc in psutil.process_iter():
            if proc.name() == 'kube-apiserver':
                apiserver_proc = proc
        
        if not apiserver_proc:
            print("Can't find kube-apiserver process!")
            attempts += 1
            continue

        params = apiserver_proc.cmdline()
        # params = ['kube-apiserver', '--advertise-address=192.168.99.135', '--allow-privileged=true', '--authorization-mode=Node,RBAC', '--client-ca-file=/var/lib/minikube/certs/ca.crt', '--enable-admission-plugins=NamespaceLifecycle,LimitRanger,ServiceAccount,DefaultStorageClass,DefaultTolerationSeconds,NodeRestriction,MutatingAdmissionWebhook,ValidatingAdmissionWebhook,ResourceQuota', '--enable-bootstrap-token-auth=true', '--etcd-cafile=/var/lib/minikube/certs/etcd/ca.crt', '--etcd-certfile=/var/lib/minikube/certs/apiserver-etcd-client.crt', '--etcd-keyfile=/var/lib/minikube/certs/apiserver-etcd-client.key', '--etcd-servers=https://127.0.0.1:2379', '--insecure-port=0', '--kubelet-client-certificate=/var/lib/minikube/certs/apiserver-kubelet-client.crt', '--kubelet-client-key=/var/lib/minikube/certs/apiserver-kubelet-client.key', '--kubelet-preferred-address-types=InternalIP,ExternalIP,Hostname', '--proxy-client-cert-file=/var/lib/minikube/certs/front-proxy-client.crt', '--proxy-client-key-file=/var/lib/minikube/certs/front-proxy-client.key', '--requestheader-allowed-names=front-proxy-client', '--requestheader-client-ca-file=/var/lib/minikube/certs/front-proxy-ca.crt', '--requestheader-extra-headers-prefix=X-Remote-Extra-', '--requestheader-group-headers=X-Remote-Group', '--requestheader-username-headers=X-Remote-User', '--secure-port=8443', '--service-account-issuer=https://kubernetes.default.svc.cluster.local', '--service-account-key-file=/var/lib/minikube/certs/sa.pub', '--service-account-signing-key-file=/var/lib/minikube/certs/sa.key', '--service-cluster-ip-range=10.96.0.0/12', '--tls-cert-file=/var/lib/minikube/certs/apiserver.crt', '--tls-private-key-file=/var/lib/minikube/certs/apiserver.key', '--anonymous-auth=true']

        applied = []
        failed = []

        for misconfig in misconfigs:
            found = False
            for param in params:
                n = param.split('=')
                if n[0] == misconfig['parameter']:
                    if misconfig['action'] == 'set':
                        if n[1] == misconfig['value']:
                            found = True
                    elif misconfig['action'] == 'push':
                        if misconfig['value'] in n[1]:
                            found = True
                    elif misconfig['action'] == 'remove':
                        if misconfig['value'] not in n[1]:
                            found = True
                    break
            
            if found:
                applied.append(misconfig['key'])
            else:
                failed.append(misconfig['key'])

        print("Applied: ", applied)
        print("Failed: ", failed)
        if len(applied) == len(misconfigs):
            ensured = True
            # readiness probe
            f = open("/tmp/ready", "a")
            f.close()
            return True

        attempts += 1

logging.info("Applying Api-Server")
params = {}
with open(r'./params.yaml') as file:
    params = yaml.load(file, Loader=yaml.FullLoader)

parser = argparse.ArgumentParser()
parser.add_argument('items', nargs='*')

args = parser.parse_args()

if len(args.items) == 0:
    print('No args provided!')
    os._exit(1)

api_config = {}
with open(r'/etc/kubernetes/manifests/kube-apiserver.yaml') as file:
    api_pod = yaml.load(file, Loader=yaml.FullLoader)
    api_config =  api_pod['spec']['containers'][0]['command']
    backup_config(api_config)

misconfigs = []

for item in args.items:
    misconfig = list(filter(lambda x: x['key'] == item, params))
    if misconfig:
        misconfigs.append(misconfig[0])

for misconfig in misconfigs:
    param = misconfig['parameter']
    if misconfig['action'] == 'set':
        set_param(api_config, param, misconfig['value'])
    if misconfig['action'] == 'push':
        push_param(api_config, param, misconfig['value'])
    # if misconfig['action'] == 'remove':
    #     api_config[param].remove(misconfig[''])

with open('/etc/kubernetes/manifests/kube-apiserver.yaml', 'w') as file:
    yaml.dump(api_pod, file)

ready = ensure_ready_probe(misconfigs)

if ready:
    logging.info("Misconfiguration applied to Kubelet")
    while True:
        logging.info("Misconfiguration still applied to Kubelet")
        time.sleep(30)
else:
    logging.error("Created configuration not applied to node")
    while True:
        logging.info("Misconfiguration not completely applied to Kubelet")
        time.sleep(30)
