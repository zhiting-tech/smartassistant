# zhitingtech/builder

FROM golang:1.16-alpine
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk add --no-cache build-base
ARG VERSION=latest
