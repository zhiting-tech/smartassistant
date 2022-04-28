.PHONY: start build lint test format
VERSION=latest
DOCKER_REGISTRY=scregistry.zhitingtech.com/
#DOCKER_REGISTRY="docker.yctc.tech/"

start:
	@go run cmd/smartassistant/main.go -c ./app.yaml

# make build-all VERSION=1.0.0
build-all: build

build:
	docker build -f build/docker/Dockerfile --build-arg GIT_COMMIT=$(shell git log -1 --format=%h) -t smartassistant:$(VERSION) .

push:
	docker image tag smartassistant:$(VERSION) $(DOCKER_REGISTRY)zhitingtech/smartassistant:$(VERSION)
	docker push $(DOCKER_REGISTRY)zhitingtech/smartassistant:$(VERSION)

run:
	docker-compose -f build/docker/docker-compose.yaml up

lint:
	@golangci-lint run ./...

test:
	go test -cover -v ./modules/...
	go test -cover -v ./pkg/...

format:
	@find -type f -name '*.go' | $(XARGS) gofmt -s -w

.PHONY: install.golangci-lint
install.golangci-lint:
	@go get -u github.com/golangci/golangci-lint/cmd/golangci-lint

.PHONY: install.goimports
install.goimports:
	@go get -u golang.org/x/tools/cmd/goimports

build-plugin-demo:
	cd examples/plugin-demo; docker build -f Dockerfile -t demo-plugin  .

