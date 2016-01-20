# Set an output prefix, which is the local directory if not specified
PREFIX?=$(shell pwd)
BUILDTAGS=
DOCKER_IMAGE=hulk-dev

.PHONY: clean all fmt vet lint build test install static protoc dbuild dtestbuild drun shell
.DEFAULT: default

all: build static fmt lint test vet

build:
	@echo "+ $@"
	@go build -tags "$(BUILDTAGS) cgo" .

dbuild:
	@echo "+ $@"
	@docker build --rm --force-rm -t jess/hulk -f Dockerfile .

drun: dbuild
	@echo "+ $@"
	docker run --rm -it -v $(CURDIR)/bundles/state:/run/hulk -v $(CURDIR)/bundles/artifacts:/var/lib/hulk jess/hulk -d server

dtestbuild:
	@echo "+ $@"
	@docker build --rm --force-rm -t hulk-dev -f Dockerfile.test .

shell: dtestbuild
	@echo "+ $@"
	@docker run --rm -it -v $(CURDIR):/go/src/github.com/jfrazelle/hulk hulk-dev bash

static:
	@echo "+ $@"
	CGO_ENABLED=1 go build -tags "$(BUILDTAGS) cgo static_build" -ldflags "-w -extldflags -static" -o hulk .

protoc:
	protoc -I ./api/grpc/types ./api/grpc/types/api.proto --go_out=plugins=grpc:api/grpc/types

fmt:
	@echo "+ $@"
	@gofmt -s -l . | grep -v '.pb.go:' | tee /dev/stderr

lint:
	@echo "+ $@"
	@golint ./... | grep -v '.pb.go:' | tee /dev/stderr

test: fmt lint vet
	@echo "+ $@"
	@go test -v -tags "$(BUILDTAGS) cgo" ./...

vet:
	@echo "+ $@"
	@go vet ./... | grep -v '.pb.go:' | tee /dev/stderr

clean:
	@echo "+ $@"
	@rm -rf hulk bundles

install:
	@echo "+ $@"
	@go install -v .
