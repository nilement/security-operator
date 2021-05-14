FROM debian:buster


RUN apt update && \
      apt install -y curl jq python3 python3-pip && \
      curl -LO https://storage.googleapis.com/kubernetes-release/release/`curl -s https://storage.googleapis.com/kubernetes-release/release/stable.txt`/bin/linux/amd64/kubectl && \
      chmod +x ./kubectl && \
      mv ./kubectl /usr/local/bin/kubectl

WORKDIR /src

COPY ./requirements.txt /src
RUN pip3 install -r requirements.txt

COPY ./main.py ./params.yaml ./recovery.py /src/

ENV PYTHONUNBUFFERED=1

ENTRYPOINT [ "python3",  "main.py" ]