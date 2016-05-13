# GH Patch Parser

[![Travis CI](https://travis-ci.org/jfrazelle/gh-patch-parser.svg?branch=master)](https://travis-ci.org/jfrazelle/gh-patch-parser)

GH Patch Parser to auto label incoming patches, triggered from nsq messages.

```console
$ gh-patch-parser -h
Usage of gh-patch-parser:
  -channel="patch-parser": nsq channel
  -d=false: run in debug mode
  -gh-token="": github access token
  -lookupd-addr="nsqlookupd:4161": nsq lookupd address
  -topic="hooks-docker": nsq topic
  -v=false: print version and exit (shorthand)
  -version=false: print version and exit

```

Example docker run command:

```bash
$ docker run -d --restart always \
    --link nsqlookupd1:nsqlookupd \
    --privileged \
    --name patch-parser \
    jess/gh-patch-parser -d \
    -gh-token="YOUR_AUTH_TOKEN" \
    -topic hooks-docker -channel patch-parser \
    -lookupd-addr nsqlookupd:4161
```
