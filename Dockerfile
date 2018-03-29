FROM golang:1.10-alpine AS go-base
RUN apk add --no-cache \
	bash \
	build-base \
	gcc \
	git \
	libpcap-dev \
	make
COPY moarpackets /go/src/github.com/jessfraz/paws/moarpackets
WORKDIR /go/src/github.com/jessfraz/paws/moarpackets
ENV CGO_ENABLED=1
RUN go get -u github.com/google/gopacket
RUN go build -installsuffix netgo -ldflags '-w -extldflags -static' -tags 'netgo cgo static_build' -o /usr/bin/moarpackets .

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
CMD ["moarpackets"]
