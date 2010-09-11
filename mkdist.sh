#!/bin/bash

# You must have a working Go build environment and appropriate SDL libraries
# for both 32-bit and 64-bit Linux to run this script.

# Param $1: Architecture name, amd64 or 386 (the same scheme Go's $GOARCH uses)
function build {
    # Make a clean build for the given arch.
    make nuke
    make all GOARCH=$1

    NAME=teratogen-$1-`cmd/teratogen/teratogen -version`

    # Make a temporary dir where the package is assembled.
    DIR=`mktemp -d`
    mkdir $DIR/$NAME
    cp cmd/teratogen/teratogen $GOROOT/pkg/${GOOS}_$1/cgo_hyades_sdl.so $GOROOT/pkg/${GOOS}_$1/libcgo.so $DIR/$NAME

    # XXX: Go makes .so files executable for some reason. Fix it, it's ugly.
    chmod a-x $DIR/$NAME/*.so

    # Package the files
    pushd $DIR
    tar cjf $NAME.tar.bz2 $NAME
    popd

    # Copy the package over
    mkdir -p dist
    cp $DIR/$NAME.tar.bz2 dist

    # Delete the temp directory
    rm -rf $DIR
}

build amd64
build 386
