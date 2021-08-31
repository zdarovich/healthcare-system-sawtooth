FROM golang:1.16.3-alpine as builder

RUN apk update \
    && apk upgrade \
    && apk add --no-cache make \
    && apk add --no-cache zeromq-dev musl-dev pkgconfig alpine-sdk libsodium-dev openssl libressl-dev

WORKDIR /app
COPY . .
RUN go test -v test/*_test.go
