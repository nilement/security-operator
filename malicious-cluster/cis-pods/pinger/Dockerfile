FROM golang:1.16-buster 

WORKDIR /src
COPY . .
RUN go build -o main
ENTRYPOINT [ "./main" ]