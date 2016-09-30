FROM debian:jessie

# Packaged dependencies
RUN apt-get update && apt-get install -y \
	autoconf \
	automake \
	ca-certificates \
	curl \
	g++ \
	git \
	libtool \
	make \
	unzip \
	--no-install-recommends \
	&& rm -rf /var/lib/apt/lists/*

# Install Go
ENV GO_VERSION 1.6.3
RUN curl -fsSL "https://storage.googleapis.com/golang/go${GO_VERSION}.linux-amd64.tar.gz" \
	| tar -xzC /usr/local
ENV PATH /go/bin:/usr/local/go/bin:$PATH
ENV GOPATH /go

# Install google/protobuf
ENV PROTOBUF_VERSION v3.0.0-beta-2
RUN set -x \
	&& export PROTOBUF_PATH="$(mktemp -d)" \
	&& curl -fsSL "https://github.com/google/protobuf/archive/${PROTOBUF_VERSION}.tar.gz" \
		| tar -xzC "$PROTOBUF_PATH" --strip-components=1 \
	&& ( \
		cd "$PROTOBUF_PATH" \
		&& ./autogen.sh \
		&& ./configure --prefix=/usr/local \
		&& make \
		&& make install \
		&& ldconfig \
	) \
	&& rm -rf "$PROTOBUFPATH"

RUN go get github.com/golang/protobuf/proto \
	&& go get github.com/golang/protobuf/protoc-gen-go \
	&& go get github.com/golang/lint/golint

# Upload source
COPY . /go/src/github.com/jessfraz/hulk
WORKDIR /go/src/github.com/jessfraz/hulk
