FROM golang:latest
MAINTAINER Jess Frazelle <jess@docker.com>

RUN go get github.com/coreos/go-etcd/etcd && \
    go get github.com/Sirupsen/logrus && \
    go get code.google.com/p/go-uuid/uuid

ADD . /go/src/github.com/jfrazelle/nginxd
RUN cd /go/src/github.com/jfrazelle/nginxd && go install . ./...
ENV PATH $PATH:/go/bin

ENTRYPOINT ["nginxd"]
