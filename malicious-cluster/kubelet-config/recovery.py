import os

cmd = 'kubectl patch node minikube-m02 --type=json -p=\'[{"op": "remove", "path": "/spec/configSource"}]\''
os.system(cmd)