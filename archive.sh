#!/bin/sh

set -e  # abort on the first error
set -v  # display script source locations in error messages
set -x  # print script steps

rm -rf archives
mkdir archives

gobin=$GOPATH/bin

xcroot=$gobin/goon-xc/snapshot
goonroot=$GOPATH/src/goon

$gobin/goxc xc

cd $xcroot
for dir in *; do
    cd $dir
    tar -czf $goonroot/archives/goon_${dir}.tar.gz goon*
    cd ..
done
