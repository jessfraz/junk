# Set an output prefix, which is the local directory if not specified
PREFIX?=$(shell pwd)

# Setup name variables for the package/tool
NAME := paws
PKG := github.com/jessfraz/$(NAME)

# Set any default go build tags
BUILDTAGS :=

# Set the build dir, where built client binaries will be output
BUILDDIR := ${PREFIX}/bin

DOCKER_REGISTRY := r.j3ss.co
DOCKER_IMAGE_PAWS := $(DOCKER_REGISTRY)/paws
DOCKER_IMAGE_TOTESSAFE := $(DOCKER_REGISTRY)/totessafe

.PHONY: totessafectl
totessafectl: ## Builds the totessafectl executable.
	mkdir -p $(BUILDDIR)
	go build -o $(BUILDDIR)/$@ ./$@/

sleeping-beauty: sleeping-beauty.asm
	nasm -o $@ $<
	chmod +x sleeping-beauty

.PHONY: images
images: paws totessafe ## Builds dockerfiles for paws and totessafe.

.PHONY: paws
paws: ## Builds the dockerfile for paws.
	docker build --rm --force-rm -t $(DOCKER_IMAGE_PAWS) .

.PHONY: totessafe
totessafe: ## Builds the dockerfile for totessafe.
	docker build --rm --force-rm -f totessafe/Dockerfile -t $(DOCKER_IMAGE_TOTESSAFE) .

all: clean totessafectl fmt lint test staticcheck vet ## Runs a clean, totessafectl, fmt, lint, test, staticcheck, and vet

.PHONY: fmt
fmt: ## Verifies all files have men `gofmt`ed
	@echo "+ $@"
	@gofmt -s -l . | grep -v '.pb.go:' | grep -v vendor | tee /dev/stderr

.PHONY: lint
lint: ## Verifies `golint` passes
	@echo "+ $@"
	@golint ./... | grep -v '.pb.go:' | grep -v vendor | tee /dev/stderr

.PHONY: test
test: ## Runs the go tests
	@echo "+ $@"
	@go test -v -tags "$(BUILDTAGS) cgo" $(shell go list ./... | grep -v vendor)

.PHONY: vet
vet: ## Verifies `go vet` passes
	@echo "+ $@"
	@go vet $(shell go list ./... | grep -v vendor) | grep -v '.pb.go:' | tee /dev/stderr

.PHONY: staticcheck
staticcheck: ## Verifies `staticcheck` passes
	@echo "+ $@"
	@staticcheck $(shell go list ./... | grep -v vendor) | grep -v '.pb.go:' | tee /dev/stderr

.PHONY: cover
cover: ## Runs go test with coverage
	@echo "" > coverage.txt
	@for d in $(shell go list ./... | grep -v vendor); do \
		go test -race -coverprofile=profile.out -covermode=atomic "$$d"; \
		if [ -f profile.out ]; then \
			cat profile.out >> coverage.txt; \
			rm profile.out; \
		fi; \
	done;

.PHONY: AUTHORS
AUTHORS:
	@$(file >$@,# This file lists all individuals having contributed content to the repository.)
	@$(file >>$@,# For how it is generated, see `make AUTHORS`.)
	@echo "$(shell git log --format='\n%aN <%aE>' | LC_ALL=C.UTF-8 sort -uf)" >> $@

.PHONY: clean
clean: ## Cleanup any build binaries or packages
	@echo "+ $@"
	$(RM) $(NAME)
	$(RM) sleeping-beauty
	$(RM) -r $(BUILDDIR)

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
