DOCKER_REGISTRY := r.j3ss.co
DOCKER_IMAGE_PAWS := $(DOCKER_REGISTRY)/paws
DOCKER_IMAGE_TOTESSAFE := $(DOCKER_REGISTRY)/totessafe

.PHONY: build
build: paws totessafe

.PHONY: paws
paws:
	docker build --rm --force-rm -t $(DOCKER_IMAGE_PAWS) .

.PHONY: totessafe
totessafe:
	docker build --rm --force-rm -f totessafe/Dockerfile -t $(DOCKER_IMAGE_TOTESSAFE) .

sleeping-beauty: sleeping-beauty.asm
	nasm -o $@ $<
	chmod +x sleeping-beauty

.PHONY: clean
clean:
	rm -f sleeping-beauty
