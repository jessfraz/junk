FROM progrium/busybox
MAINTAINER Jessica Frazelle <jess@docker.com>

ADD https://jesss.s3.amazonaws.com/binaries/gh-patch-parser /usr/local/bin/gh-patch-parser

RUN chmod +x /usr/local/bin/gh-patch-parser

ENTRYPOINT [ "/usr/local/bin/gh-patch-parser" ]
