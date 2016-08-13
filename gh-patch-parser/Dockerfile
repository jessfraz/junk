FROM alpine
MAINTAINER Jessica Frazelle <jess@docker.com>

ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go

RUN	apk update && apk add \
	ca-certificates \
	curl \
	git \
	&& rm -rf /var/cache/apk/*

# make git happy
RUN git config --global user.name gh-patch-parser && \
    git config --global user.email gh-patch-parser@dockerproject.com

ENV GOFMT_VERSION 1.5.1
RUN curl -sSL https://storage.googleapis.com/golang/go${GOFMT_VERSION}.linux-amd64.tar.gz | tar -C /tmp/ -xz \
	&& mv /tmp/go/bin/gofmt /usr/local/bin \
	&& rm -rf /tmp/go

COPY . /go/src/github.com/jfrazelle/gh-patch-parser

RUN buildDeps=' \
		go \
		gcc \
		libc-dev \
		libgcc \
	' \
	set -x \
	&& apk update \
	&& apk add $buildDeps \
	&& cd /go/src/github.com/jfrazelle/gh-patch-parser \
	&& go get -d -v github.com/jfrazelle/gh-patch-parser \
	&& go build -o /usr/bin/gh-patch-parser . \
	&& apk del $buildDeps \
	&& rm -rf /var/cache/apk/* \
	&& rm -rf /go \
	&& echo "Build complete."


ENTRYPOINT [ "gh-patch-parser" ]
