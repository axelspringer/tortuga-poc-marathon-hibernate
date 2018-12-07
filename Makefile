GO=go
RACE := $(shell test $$(go env GOARCH) != "amd64" || (echo "-race"))
GOFLAGS= 

BIN_HIBERTHON_TRIGGER=bin/hiberthon-trigger
BIN_HIBERTHON=bin/hiberthon

IMAGE_HIBERTHON="axelspringer/hiberthon"
IMAGE_HIBERTHON_TRIGGER="axelspringer/hiberthon-trigger"

VERSION := $(shell git rev-parse HEAD)
ROOT_DIR:=$(shell dirname $(realpath $(lastword $(MAKEFILE_LIST))))

all: build/local

help:
	@echo 'Available commands:'
	@echo
	@echo 'Usage:'
	@echo '    make deps     		          Install go deps.'
	@echo '    make build/docker              Create docker container with hiberthon and the trigger.'
	@echo '    make generate    	          Generate template assets.'
	@echo '    make build/hiberthon/static    Build static linked version of hiberthon'
	@echo '    make build/trigger/static      Build static linked version of hiberthon-trigger'

deps:
	go get ./...

build/docker:
	docker build -t $(IMAGE_HIBERTHON) .

push/docker:
	echo "$(DOCKER_PASSWORD)" | docker login -u "$(DOCKER_USERNAME)" --password-stdin
	docker push $(IMAGE)

generate:
	$(GO) generate src/net/bumper.go

build/hiberthon/static:
	CGO_ENABLED=0 $(GO) build -a -installsuffix cgo -ldflags "-w -s" -o $(BIN_HIBERTHON) src/cli/hiberthon/*

build/trigger/static:
	CGO_ENABLED=0 $(GO) build -a -installsuffix cgo -ldflags "-w -s" -o $(BIN_HIBERTHON_TRIGGER) src/cli/hiberthon-trigger/*

