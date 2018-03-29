FROM golang:1.10-alpine AS go-base
RUN apk add --no-cache \
	bash \
	build-base \
	gcc \
	git \
	libpcap-dev \
	make
COPY moar-packets /go/src/github.com/jessfraz/paws/moar-packets
WORKDIR /go/src/github.com/jessfraz/paws/moar-packets
ENV CGO_ENABLED=1
RUN go get -u github.com/google/gopacket
RUN go build -installsuffix netgo -ldflags '-w -extldflags -static' -tags 'netgo cgo static_build' -o /usr/bin/moar-packets .

FROM alpine:latest AS assembly-base
RUN apk add --no-cache \
	make \
	nasm
COPY . /usr/src
WORKDIR /usr/src
RUN make sleeping-beauty

FROM alpine:latest
COPY --from=go-base /usr/bin/moar-packets /usr/bin/
COPY --from=assembly-base /usr/src/sleeping-beauty /usr/bin/
CMD ["moar-packets"]
