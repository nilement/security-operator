FROM golang:1.16-buster 

RUN mkdir -p /src/backups
WORKDIR /src

COPY . .
RUN go build -o main
COPY ./experiments.yaml .
ENTRYPOINT [ "./main" ]