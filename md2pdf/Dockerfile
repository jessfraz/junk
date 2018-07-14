FROM golang:alpine as builder
MAINTAINER Jessica Frazelle <jess@linux.com>

ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go

RUN	apk add --no-cache \
	ca-certificates \
	ghostscript \
	heirloom-doctools

COPY . /go/src/github.com/jessfraz/md2pdf

RUN set -x \
	&& apk add --no-cache --virtual .build-deps \
		git \
		gcc \
		libc-dev \
		libgcc \
		make \
	&& cd /go/src/github.com/jessfraz/md2pdf \
	&& make static \
	&& mv md2pdf /usr/bin/md2pdf \
	&& apk del .build-deps \
	&& rm -rf /go \
	&& echo "Build complete."

ENTRYPOINT [ "md2pdf" ]
CMD [ "--help" ]
