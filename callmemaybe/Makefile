.PHONY: docker

all: callmemaybe

callmemaybe: *.go
	CGO_ENABLED=0 go build -ldflags "-s" -a -installsuffix cgo -ldflags "-w" -o callmemaybe .

docker: callmemaybe
	docker build --rm --force-rm -t jess/callmemaybe .

clean:
	rm -f callmemaybe
