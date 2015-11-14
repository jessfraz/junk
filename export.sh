#!/bin/bash
set -e

(
compression=$1
tmpdir=$(pwd)/tmp
docker export $(docker create jess/floppercon) > floppercon.tar
mkdir -p ${tmpdir}
tar xvf floppercon.tar -C ${tmpdir}
cd ${tmpdir}

if [[ "$compression" == "xz" ]]; then
	XZ_OPT=-9 tar cvJf ../floppercon.tar.xz .
else
	GZIP=-9 tar cvzf ../floppercon.tar.gz .
fi

rm -rf ${tmpdir}
rm -rf floppercon.tar
)
