FROM debian:jessie
MAINTAINER Jessica Frazelle <jess@docker.com>

ADD https://jesss.s3.amazonaws.com/binaries/gh-patch-parser /usr/local/bin/gh-patch-parser

RUN apt-get update && apt-get install -y \
    ca-certificates \
    --no-install-recommends \
    && chmod +x /usr/local/bin/gh-patch-parser

ENTRYPOINT [ "/usr/local/bin/gh-patch-parser" ]
