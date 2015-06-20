FROM debian:jessie
MAINTAINER Jessica Frazelle <jess@docker.com>

RUN apt-get update && apt-get install -y \
	make \
	python \
	--no-install-recommends

ADD https://jesss.s3.amazonaws.com/binaries/nsqexec /usr/local/bin/nsqexec

RUN chmod +x /usr/local/bin/nsqexec

ENTRYPOINT [ "/usr/local/bin/nsqexec" ]
