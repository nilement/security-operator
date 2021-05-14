FROM debian:buster


RUN apt update && \
      apt install -y curl jq python3 python3-pip

RUN mkdir -p /src/backup
WORKDIR /src

COPY ./requirements.txt /src
RUN pip3 install -r requirements.txt

COPY ./main.py ./params.yaml ./recovery.py /src/

ENV PYTHONUNBUFFERED=1

ENTRYPOINT [ "python3",  "main.py" ]