# zhitingtech/runtime
FROM alpine
RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories
RUN apk add --no-cache tzdata
ARG VERSION=latest
ARG GIT_COMMIT=unspecified
ENV GIT_COMMIT=$GIT_COMMIT
LABEL org.opencontainers.image.version=$VERSION
LABEL org.opencontainers.image.revision=$GIT_COMMI