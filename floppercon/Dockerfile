FROM scratch
MAINTAINER Jess Frazelle <jess@docker.com>

COPY floppercon /floppercon
COPY certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

ENTRYPOINT ["./floppercon"]
