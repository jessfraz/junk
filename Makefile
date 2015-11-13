.PHONY: certs docker

all: floppercon

certs:
	./certs/update.sh

floppercon: *.go
	CGO_ENABLED=0 go build -ldflags "-s" -a -installsuffix cgo -ldflags "-w" -o floppercon .

docker: floppercon
	docker build --rm --force-rm -t jess/floppercon .

clean:
	rm -f floppercon
