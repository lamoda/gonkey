export GO111MODULE=on

REPO=lamoda
NAME=gonkey
VERSION=$(shell git describe --tags 2> /dev/null || git rev-parse --short HEAD)

DOCKER_TAG ?= latest

.PHONY: @dockerbuild @push @stub test

build: @build

@dockerbuild:
	docker build --force-rm --pull -t $(REPO)/$(NAME):$(DOCKER_TAG) .

@build:
	go build -a -o gonkey

@push:
	docker push $(REPO)/$(NAME):$(DOCKER_TAG)

@stub:
	echo "VERSION=$(VERSION)" > .artifact

test:
	go test ./...
