FROM golang:1.16.3-alpine as builder

RUN apk update \
    && apk upgrade \
    && apk add --no-cache make

WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o tp -ldflags="-w -s" .

FROM alpine:3.11

WORKDIR /app

ENV PORT="8080"


COPY --from=builder /app/tp /usr/bin/tp

CMD ["tp"]
