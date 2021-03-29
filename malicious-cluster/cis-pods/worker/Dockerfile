FROM golang:1.16-buster 

RUN mkdir -p /src/backups
WORKDIR /src

RUN apt-get update \
    && apt-get install -y protobuf-compiler \
    && go get google.golang.org/protobuf/cmd/protoc-gen-go google.golang.org/grpc/cmd/protoc-gen-go-grpc

COPY ./main.go .
COPY ./restore.go .
RUN go build main.go
RUN go build restore.go
ENTRYPOINT [ "./main" ]