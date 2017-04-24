DOCKER_IMAGE=r.j3ss.co/bespoke

public: build
	docker run --rm -it -v $(CURDIR)/public:/usr/src/app/dist $(DOCKER_IMAGE) gulp

build:
	docker build --rm --force-rm -t $(DOCKER_IMAGE) .

serve: build
	docker run --rm -it --net host -v $(CURDIR)/src:/usr/src/app/src $(DOCKER_IMAGE) gulp serve

GCLOUD_BUCKET:=misc.j3ss.co
publish:
	gsutil -m cp -r -a public-read public/* gs://$(GCLOUD_BUCKET)/decks/$(shell basename $(CURDIR))/
