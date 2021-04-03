FROM ubuntu

RUN apt-get update && apt-get -y install curl xz-utils wget jq python3 netcat iproute2 && \
    curl -LO https://storage.googleapis.com/kubernetes-release/release/`curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt`/bin/linux/amd64/kubectl && \
    chmod +x ./kubectl && \
    mv ./kubectl /usr/local/bin/kubectl
RUN wget https://github.com/inguardians/peirates/releases/download/v1.0.30/peirates-linux-386.tar.xz