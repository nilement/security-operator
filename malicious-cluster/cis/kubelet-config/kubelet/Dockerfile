FROM debian:buster
RUN apt update && \
      apt install -y curl && \
      curl -LO https://storage.googleapis.com/kubernetes-release/release/`curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt`/bin/linux/amd64/kubectl && \
      chmod +x ./kubectl && \
      mv ./kubectl /usr/local/bin/kubectl

RUN apt install -y jq
COPY ./kubelet-config.sh /home/kubelet-config.sh
COPY ./kubelet-recovery.sh /home/kubelet-recovery.sh

ENTRYPOINT [ "bash", "/home/kubelet-config.sh" ] 