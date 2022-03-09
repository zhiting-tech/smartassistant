# 开始前先 go mod vendor
FROM golang:1.16-alpine as builder
RUN apk add build-base
COPY . /app
WORKDIR /app
ARG VERSION=latest
RUN go env -w GOPROXY="goproxy.cn,direct"
# --tags json1 for sqlite json support
RUN go build --tags json1,release -ldflags="-w -s -X 'github.com/zhiting-tech/smartassistant/modules/types.Version=$VERSION'" -o smartassistant cmd/smartassistant/main.go

FROM alpine
RUN apk add --no-cache tzdata
WORKDIR /app
ARG GIT_COMMIT=unspecified
ENV GIT_COMMIT=$GIT_COMMIT
LABEL org.opencontainers.image.revision=$GIT_COMMIT
COPY --from=builder /app/smartassistant /app/smartassistant
RUN mkdir -p /mnt/data/zt-smartassistant/config
RUN mkdir -p /mnt/data/zt-smartassistant/data/smartassistant
## 创建插件上传目录
RUN mkdir /app/plugins
COPY ./app.yaml.example /mnt/data/zt-smartassistant/config/smartassistant.yaml
COPY ./static ./static
ENTRYPOINT ["/app/smartassistant"]
