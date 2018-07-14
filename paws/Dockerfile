FROM golang:1.10-alpine AS go-base
RUN apk add --no-cache \
	bash \
	build-base \
	gcc \
	git \
	libpcap-dev \
	make
COPY . /go/src/github.com/jessfraz/paws
WORKDIR /go/src/github.com/jessfraz/paws
ENV CGO_ENABLED=1
RUN go build -installsuffix netgo -ldflags '-w -extldflags -static' -tags 'netgo cgo static_build' -o /usr/bin/moarpackets ./moarpackets/

FROM alpine:latest AS assembly-base
RUN apk add --no-cache \
	make \
	nasm
COPY . /usr/src
WORKDIR /usr/src
RUN make sleeping-beauty

FROM alpine:latest
COPY --from=go-base /usr/bin/moarpackets /usr/bin/
COPY --from=assembly-base /usr/src/sleeping-beauty /usr/bin/
ENTRYPOINT ["moarpackets"]
