FROM alpine
MAINTAINER Jessica Frazelle <jess@docker.com>

ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go

RUN	apk update && apk add \
	ca-certificates \
	git \
	make \
	python \
	&& rm -rf /var/cache/apk/*

COPY . /go/src/github.com/jfrazelle/nsqexec

RUN buildDeps=' \
		go \
		gcc \
		libc-dev \
		libgcc \
	' \
	set -x \
	&& apk update \
	&& apk add $buildDeps \
	&& cd /go/src/github.com/jfrazelle/nsqexec \
	&& go get -d -v github.com/jfrazelle/nsqexec \
	&& go build -o /usr/bin/nsqexec . \
	&& apk del $buildDeps \
	&& rm -rf /var/cache/apk/* \
	&& rm -rf /go \
	&& echo "Build complete."


ENTRYPOINT [ "nsqexec" ]
