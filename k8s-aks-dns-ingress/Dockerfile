FROM golang:alpine as builder
MAINTAINER Jessica Frazelle <jess@linux.com>

ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go

RUN	apk add --no-cache \
	ca-certificates

COPY . /go/src/github.com/jessfraz/k8s-aks-dns-ingress

RUN set -x \
	&& apk add --no-cache --virtual .build-deps \
		git \
		gcc \
		libc-dev \
		libgcc \
		make \
	&& cd /go/src/github.com/jessfraz/k8s-aks-dns-ingress \
	&& make static \
	&& mv k8s-aks-dns-ingress /usr/bin/k8s-aks-dns-ingress \
	&& apk del .build-deps \
	&& rm -rf /go \
	&& echo "Build complete."

FROM scratch

COPY --from=builder /usr/bin/k8s-aks-dns-ingress /usr/bin/k8s-aks-dns-ingress
COPY --from=builder /etc/ssl/certs/ /etc/ssl/certs

ENTRYPOINT [ "k8s-aks-dns-ingress" ]
CMD [ "--help" ]
