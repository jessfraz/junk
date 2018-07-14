# Set an output prefix, which is the local directory if not specified.
PREFIX?=$(shell pwd)

SHELL = /bin/bash
.SHELLFLAGS = -O extglob -c

BUILDTAGS=

.PHONY: all
all: clean build fmt lint test vet ## Runs a clean, build, fmt, lint, test, and vet.

.PHONY: build
build: ## Builds dynamic executables and/or packages.
	@echo "+ $@"
	@go build -tags "$(BUILDTAGS) cgo" ./...

.PHONY: fmt
fmt: ## Verifies all files have been `gofmt`ed.
	@echo "+ $@"
	@gofmt -s -l . | grep -v '.pb.go:' | grep -v vendor | tee /dev/stderr

.PHONY: lint
lint: ## Verifies `golint` passes.
	@echo "+ $@"
	@golint ./... | grep -v '.pb.go:' | grep -v vendor | grep -v paws/totessafe/reflector | tee /dev/stderr

.PHONY: test
test: ## Runs the go tests.
	@echo "+ $@"
	@go test -v -tags "$(BUILDTAGS) cgo" $(shell go list ./... | grep -v vendor)

.PHONY: vet
vet: ## Verifies `go vet` passes.
	@echo "+ $@"
	@go vet $(shell go list ./... | grep -v vendor) | grep -v '.pb.go:' | tee /dev/stderr

# if this session isn't interactive, then we don't want to allocate a
# TTY, which would fail, but if it is interactive, we do want to attach
# so that the user can send e.g. ^C through.
DOCKER_FLAGS = --rm -i
INTERACTIVE := $(shell [ -t 0 ] && echo 1 || echo 0)
ifeq ($(INTERACTIVE), 1)
	DOCKER_FLAGS += -t
endif

DOCKER_IMAGE := r.j3ss.co/junk:dev
.PHONY: ci
ci: ## Run the makefile in a docker container.
	@echo "+ $@"
	docker build --rm --force-rm -f Dockerfile.dev -t $(DOCKER_IMAGE) .
	docker run \
		--disable-content-trust=true \
		$(DOCKER_FLAGS) \
		$(DOCKER_IMAGE) make

check_defined = \
				$(strip $(foreach 1,$1, \
				$(call __check_defined,$1,$(strip $(value 2)))))
__check_defined = \
				  $(if $(value $1),, \
				  $(error Undefined $1$(if $2, ($2))$(if $(value @), \
				  required by target `$@')))

check_exists = \
			   $(strip $(foreach 1,$1, \
			   $(call __check_exists,$1)))
__check_exists = \
				 $(if $(wildcard $(value $1)/*),, \
				 $(error $(value $1) does not exist, \
				 required by target `$@'))

TMP_REMOTE := tmp
.PHONY: move-repo
move-repo: ## Moves a local repository into this repo as a sub-directory (ex. REPO=~/dumb-shit).
	@:$(call check_defined, REPO, path to the repository)
	@:$(call check_exists, REPO)
	cd "$(REPO)" && \
		git checkout . && \
		$(RM) -r "$(notdir $(REPO))" && \
		mkdir -p "$(notdir $(REPO))" && \
		mv !($(notdir $(REPO))) "$(notdir $(REPO))" && \
		mv .!(|.|git) "$(notdir $(REPO))" && \
		git add "$(notdir $(REPO))" && \
		git commit -a -m "Preparing $(notdir $(REPO)) for move"
	git remote add "$(TMP_REMOTE)" "$(REPO)"
	git fetch "$(TMP_REMOTE)"
	git merge --allow-unrelated-histories "$(TMP_REMOTE)/master"
	git remote rm "$(TMP_REMOTE)"

.PHONY: clean
clean: ## Cleanup any build binaries or packages.
	@echo "+ $@"

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-\/]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
