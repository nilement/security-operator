FROM ubuntu:20.04

ARG XMRIG_VERSION='v6.12.1'
RUN  apt-get update
RUN  DEBIAN_FRONTEND="noninteractive" apt-get install git build-essential cmake libuv1-dev libssl-dev libhwloc-dev -y

RUN apt install tor -y

RUN git clone https://github.com/xmrig/xmrig && \
    mkdir xmrig/build && \
    cd xmrig && git checkout ${XMRIG_VERSION}

RUN cd xmrig/build && \
    cmake .. -DCMAKE_BUILD_TYPE=Release && \
    make -j$(nproc)

RUN apt-get install curl wget python3 -y

RUN wget http://ftp.barfooze.de/pub/sabotage/tarballs/proxychains-ng-4.14.tar.xz \
    && tar Jxf proxychains-ng-4.14.tar.xz \
    && cd proxychains-ng-4.14 \
    && ./configure \
    && make \
    && make install

COPY ./proxychains.conf /usr/local/etc/proxychains.conf

COPY ./main.py /home/main.py

CMD [ "python3", "/home/main.py" ]