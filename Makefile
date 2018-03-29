DOCKER_IMAGE := r.j3ss.co/paws

.PHONY: build
build:
	docker build --rm --force-rm -t $(DOCKER_IMAGE) .

sleeping-beauty: sleeping-beauty.asm
	nasm -o $@ $<
	chmod +x sleeping-beauty

.PHONY: clean
clean:
	rm -f sleeping-beauty
