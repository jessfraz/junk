FROM golang:1.11-alpine

ENV PATH /go/bin:$PATH
ENV GOPATH /go

RUN apk add --no-cache \
	bash \
	gcc \
	git \
	libpcap-dev \
	make \
	musl-dev

# Install golint
RUN go get golang.org/x/lint/golint

COPY . /go/src/github.com/jessfraz/junk
WORKDIR /go/src/github.com/jessfraz/junk
