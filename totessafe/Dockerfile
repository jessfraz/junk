FROM golang:1.10-alpine AS go-base
RUN apk add --no-cache \
	bash \
	build-base \
	gcc \
	git \
	make
COPY . /go/src/github.com/jessfraz/paws
WORKDIR /go/src/github.com/jessfraz/paws
ENV CGO_ENABLED=1
RUN go build -installsuffix netgo -ldflags '-w -extldflags -static' -tags 'netgo cgo static_build' -o /usr/bin/totessafe ./totessafe/

FROM alpine:latest
COPY --from=go-base /usr/bin/totessafe /usr/bin/
ENTRYPOINT ["totessafe"]
