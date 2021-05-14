import requests, json, yaml
import os, argparse
import subprocess
from subprocess import PIPE
import string
import random
import time
import logging
from string import Template

def get_random_string(length):
    # choose from all lowercase letter
    letters = string.ascii_lowercase
    return ''.join(random.choice(letters) for i in range(length))

def get_token():
    return open("/var/run/secrets/kubernetes.io/serviceaccount/token").read()

def ensure_ready_probe(created):
    # check configs
    attempts = 0
    active_config = 'null\n'

    while active_config != created:
        if attempts >= 5:
            return False
        time.sleep(5)
        get_name = subprocess.check_output("kubectl get no minikube-m02 -o json | jq '.status.config'", shell=True).decode()
        active_config = json.loads(get_name)
        if active_config is not None:
            if 'active' in active_config:
                active_config = active_config['active']['configMap']['name']
        attempts += 1
        
    # readiness probe
    f = open("/tmp/ready", "a")
    f.close()
    return True


params = {}
with open(r'./params.yaml') as file:
    params = yaml.load(file, Loader=yaml.FullLoader)

parser = argparse.ArgumentParser()
parser.add_argument('items', nargs='*')

args = parser.parse_args()

if len(args.items) == 0:
    print('No args provided!')
    os._exit(0)

token = get_token()
# token = ""
headers = {"Authorization": "Bearer " + token}

# https://kubernetes.io/docs/tasks/administer-cluster/reconfigure-kubelet/
r = requests.get('https://10.96.0.1:443/api/v1/nodes/minikube-m02/proxy/configz', verify=False, headers=headers)
# r = requests.get('http://localhost:8001/api/v1/nodes/minikube-m02/proxy/configz', verify=False, headers=headers)

config = r.json()['kubeletconfig']
config['kind'] = 'KubeletConfiguration'
config['apiVersion'] = 'kubelet.config.k8s.io/v1beta1'

for item in args.items:

    param = params[item]
    command = param['key'].split('.')
    lastarg = command[-1]

    z = config
    for a in range(len(command) - 1):
        arg = command[a]
        z = z[arg]
    z[lastarg] = param['value']


with open('edited.json', 'w') as outfile:
    json.dump(config, outfile)

generate_name = "cfg-" + get_random_string(4)

config_map_name = subprocess.check_output("kubectl -n kube-system create configmap {} --from-file=kubelet=edited.json --append-hash -o json | jq -r '.metadata|.name'".format(generate_name), universal_newlines=True, shell=True)
patch = Template('kubectl patch node minikube-m02 -p "{\\"spec\\":{\\"configSource\\":{\\"configMap\\":{\\"name\\":\\"$configname\\",\\"namespace\\":\\"kube-system\\",\\"kubeletConfigKey\\":\\"kubelet\\"}}}}"')
config_map_name=config_map_name.replace('\n', '')
output = patch.substitute(configname=config_map_name)
os.system(output)

ready = ensure_ready_probe(config_map_name)

if ready:
    logging.info("Misconfiguration applied to Kubelet")
    while True:
        logging.info("Misconfiguration still applied to Kubelet")
        time.sleep(30)
else:
    logging.error("Created configuration not applied to node")
