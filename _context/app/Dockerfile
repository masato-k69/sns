FROM golang:1.23.0-alpine

ENV GO111MODULE on

RUN apk update && \
    apk add --no-cache curl

WORKDIR /app

CMD [ "go", "run", "main.go" ]