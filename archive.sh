#!/bin/sh

mkdir -p archives

root=../../bin/goon-xc/snapshot
for dir in $(ls $root); do
    tar -czf archives/goon_${dir}.tar.gz $root/$dir/goon*
done
