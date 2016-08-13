.PHONY: certs docker gzip xz

all: floppercon

certs:
	./certs/update.sh

floppercon: *.go
	CGO_ENABLED=0 go build -ldflags "-s" -a -installsuffix cgo -ldflags "-w" -o floppercon .

docker: floppercon
	docker build --rm --force-rm -t jess/floppercon .

gzip: docker
	./export.sh gzip

xz: docker
	./export.sh xz

clean:
	rm -f floppercon
	rm -rf tmp
	rm -rf floppercon.tar*
