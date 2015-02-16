FROM debian:jessie
MAINTAINER Jessica Frazelle <jess@docker.com>

ADD https://jesss.s3.amazonaws.com/binaries/gh-patch-parser /usr/local/bin/gh-patch-parser

RUN apt-get update && apt-get install -y \
    ca-certificates \
    curl \
    git \
    --no-install-recommends \
    && chmod +x /usr/local/bin/gh-patch-parser

# We still support compiling with older Go, so need to grab older "gofmt"
ENV GOFMT_VERSION 1.3.3
RUN curl -sSL https://storage.googleapis.com/golang/go${GOFMT_VERSION}.linux-amd64.tar.gz | tar -C /usr/local/bin -xz --strip-components=2 go/bin/gofmt

# make git happy
RUN git config --global user.name gh-patch-parser && \
    git config --global user.email gh-patch-parser@dockerproject.com

ENTRYPOINT [ "/usr/local/bin/gh-patch-parser" ]
