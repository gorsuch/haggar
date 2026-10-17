BINARY_NAME ?= haggar
DOCKER_NAMESPACE ?= gorsuch
DOCKER_TAG ?= latest
DOCKER_IMAGE ?= $(DOCKER_NAMESPACE)/$(BINARY_NAME):$(DOCKER_TAG)
DOCKER_PLATFORMS ?= linux/amd64,linux/arm64
DOCKER_BUILDER ?= haggar-builder
GOOS ?= $(shell go env GOOS)
GOARCH ?= $(shell go env GOARCH)
GOBUILD=CGO_ENABLED=0 go build -trimpath

.PHONY: build
build:
	$(GOBUILD) -o bin/${BINARY_NAME}-$(GOOS)-$(GOARCH) .

.PHONY: build-linux
build-linux:
	GOOS=linux GOARCH=amd64 $(MAKE) build

# Builds for the local platform and loads the image into Docker.
.PHONY: build-docker
build-docker:
	docker build -t $(DOCKER_IMAGE) .

# Builds for every platform in DOCKER_PLATFORMS and pushes to the registry.
# Multi-platform builds need a docker-container buildx builder, which is
# created on first use.
.PHONY: push-docker
push-docker: docker-builder
	docker buildx build --builder $(DOCKER_BUILDER) --platform $(DOCKER_PLATFORMS) -t $(DOCKER_IMAGE) --push .

.PHONY: docker-builder
docker-builder:
	docker buildx inspect $(DOCKER_BUILDER) >/dev/null 2>&1 || docker buildx create --name $(DOCKER_BUILDER) --driver docker-container
