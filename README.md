# junk

[![Build Status](https://travis-ci.org/jessfraz/junk.svg?branch=master)](https://travis-ci.org/jessfraz/junk)

A place for everything without a home.

**IF YOU ARE LOOKING AT THIS AND YOU ARE NOT JESS, YOU WILL FIND NOTHING COOL
HERE.** It's all half finished or shitty crap.

## Using the Makefile

```console
$ make help
all                            Runs a clean, build, fmt, lint, test, and vet.
build                          Builds dynamic executables and/or packages.
clean                          Cleanup any build binaries or packages.
fmt                            Verifies all files have been `gofmt`ed.
lint                           Verifies `golint` passes.
move-repo                      Moves a local repository into this repo as a sub-directory (ex. REPO=~/dumb-shit).
test                           Runs the go tests.
vet                            Verifies `go vet` passes.
```
