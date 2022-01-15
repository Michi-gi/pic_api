FROM golang:latest

WORKDIR api_code
COPY *.go ./picapi/

WORKDIR picapi
RUN go mod init picapi
RUN go mod tidy

RUN go build

ENTRYPOINT ["./picapi"]
