BINARY_NAME ?= haggar
DOCKER_NAMESPACE ?= gorsuch
DOCKER_TAG ?= latest
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
GOBUILD=CGO_ENABLED=0 go build -trimpath

.PHONY: build
build:
	$(GOBUILD) -o bin/${BINARY_NAME}-$(GOOS)-$(GOARCH) .

.PHONY: build-linux
build-linux:
	GOOS=linux GOARCH=amd64 $(MAKE) build

.PHONY: build-docker
build-docker: build-linux
	docker build --no-cache -t $(DOCKER_NAMESPACE)/$(BINARY_NAME):$(DOCKER_TAG) -f Dockerfile .