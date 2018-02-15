# md2pdf

[![Travis CI](https://travis-ci.org/jessfraz/md2pdf.svg?branch=master)](https://travis-ci.org/jessfraz/md2pdf)

Convert markdown files into nice looking pdfs with troff and ghostscript.

## Installation

You need `troff` and `dpost` from [heirloom-doctools](https://github.com/n-t-roff/heirloom-doctools)
and `ps2pdf` from [ghostscript](https://www.ghostscript.com/).

The easiest way to run this is with the docker image:

```console
$ docker run --rm -it -v $(pwd):/src -w /src --tmpfs /tmp \
    r.j3ss.co/md2pdf my-post.md
```

#### Binaries

- **darwin** [386](https://github.com/jessfraz/md2pdf/releases/download/v0.0.0/md2pdf-darwin-386) / [amd64](https://github.com/jessfraz/md2pdf/releases/download/v0.0.0/md2pdf-darwin-amd64)
- **freebsd** [386](https://github.com/jessfraz/md2pdf/releases/download/v0.0.0/md2pdf-freebsd-386) / [amd64](https://github.com/jessfraz/md2pdf/releases/download/v0.0.0/md2pdf-freebsd-amd64)
- **linux** [386](https://github.com/jessfraz/md2pdf/releases/download/v0.0.0/md2pdf-linux-386) / [amd64](https://github.com/jessfraz/md2pdf/releases/download/v0.0.0/md2pdf-linux-amd64) / [arm](https://github.com/jessfraz/md2pdf/releases/download/v0.0.0/md2pdf-linux-arm) / [arm64](https://github.com/jessfraz/md2pdf/releases/download/v0.0.0/md2pdf-linux-arm64)
- **solaris** [amd64](https://github.com/jessfraz/md2pdf/releases/download/v0.0.0/md2pdf-solaris-amd64)
- **windows** [386](https://github.com/jessfraz/md2pdf/releases/download/v0.0.0/md2pdf-windows-386) / [amd64](https://github.com/jessfraz/md2pdf/releases/download/v0.0.0/md2pdf-windows-amd64)

#### Via Go

```bash
$ go get github.com/jessfraz/md2pdf
```

## Usage

```console
$ md2pdf -h
               _ ____            _  __
 _ __ ___   __| |___ \ _ __   __| |/ _|
| '_ ` _ \ / _` | __) | '_ \ / _` | |_
| | | | | | (_| |/ __/| |_) | (_| |  _|
|_| |_| |_|\__,_|_____| .__/ \__,_|_|
                      |_|

 Convert markdown files into nice looking pdfs with troff and ghostscript.
 Version: v0.0.0
 Build: 129c85a-dirty

  -d    run in debug mode
  -v    print version and exit (shorthand)
  -version
        print version and exit
```

**The arguments passed are the paths to the files you wish to convert.**

```console
$ md2pdf thing.md thing2.md
thing-DRAFT.pdf
thing2-DRAFT.pdf
```
