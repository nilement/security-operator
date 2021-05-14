FROM golang:1.16-buster 

RUN mkdir -p /src/backups
WORKDIR /src

RUN apt-get update \
    && apt-get install -y protobuf-compiler \
    && go get google.golang.org/protobuf/cmd/protoc-gen-go google.golang.org/grpc/cmd/protoc-gen-go-grpc

COPY . .
RUN go build -o main
COPY ./experiments.yaml .
ENTRYPOINT [ "./main" ]