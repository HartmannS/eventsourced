FROM golang:1.12-alpine3.9

RUN sed -i "s/dl-cdn.alpinelinux.org/dl-2.alpinelinux.org/g" /etc/apk/repositories
RUN apk add -U --no-cache git make upx
