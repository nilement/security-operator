import requests, json, yaml
import os, argparse
import subprocess
from subprocess import PIPE
import string
import random
from string import Template

def get_random_string(length):
    # choose from all lowercase letter
    letters = string.ascii_lowercase
    return ''.join(random.choice(letters) for i in range(length))

params = {}
with open(r'./params.yaml') as file:
    params = yaml.load(file, Loader=yaml.FullLoader)

parser = argparse.ArgumentParser()
parser.add_argument('items', nargs='*')

args = parser.parse_args()

if len(args.items) == 0:
    print('No args provided!')
    os._exit(0)

r = requests.get('http://localhost:8001/api/v1/nodes/minikube-m02/proxy/configz')

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

generate_name = "config"

config_map_name = subprocess.check_output("kubectl -n kube-system create configmap {} --from-file=kubelet=edited.json --append-hash -o json | jq -r '.metadata|.name'".format(generate_name), universal_newlines=True, shell=True)
patch = Template('kubectl patch node minikube-m02 -p "{\\"spec\\":{\\"configSource\\":{\\"configMap\\":{\\"name\\":\\"$configname\\",\\"namespace\\":\\"kube-system\\",\\"kubeletConfigKey\\":\\"kubelet\\"}}}}"')
config_map_name=config_map_name.replace('\n', '')
output = patch.substitute(configname=config_map_name)
# print(config_map_name)
print(output)
os.system(output)