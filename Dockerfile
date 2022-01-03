FROM golang:latest

WORKDIR api_code
COPY *.go ./picapi/

RUN go mod init picapi
RUN go mod tidy

WORKDIR picapi
RUN go build

ENTRYPOINT ["./picapi"]