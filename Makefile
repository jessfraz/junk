# Set an output prefix, which is the local directory if not specified
PREFIX?=$(shell pwd)

.PHONY: clean all fmt vet lint build test install
.DEFAULT: default

all: clean build fmt lint test vet install

build:
	@echo "+ $@"
	@go build ./...

fmt:
	@echo "+ $@"
	@gofmt -s -l .

lint:
	@echo "+ $@"
	@golint ./...

test: fmt lint vet
	@echo "+ $@"
	@go test -v ./...

vet:
	@echo "+ $@"
	@go vet ./...

clean:
	@echo "+ $@"
	@rm -rf cliaoke

install:
	@echo "+ $@"
	@go install -v .
