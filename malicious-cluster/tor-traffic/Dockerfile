FROM ubuntu

WORKDIR /usr/src/app

RUN apt-get update && \
    apt-get install --no-install-recommends -y \
    python3.8 python3-pip python3.8-dev

RUN apt install tor -y

RUN apt-get -y install python3-pip

COPY ./requirements.txt .

RUN pip3 install -r requirements.txt

COPY . .

CMD ["bash", "./init.sh"]