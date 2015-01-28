FROM progrium/busybox
MAINTAINER Jessica Frazelle <jess@docker.com>

ADD https://jesss.s3.amazonaws.com/binaries/nsqexec /usr/local/bin/nsqexec

RUN chmod +x /usr/local/bin/nsqexec

ENTRYPOINT [ "/usr/local/bin/nsqexec" ]
