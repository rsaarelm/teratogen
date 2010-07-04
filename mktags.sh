#!/bin/bash

make clean
tago `find . -name "*.go"` `find $GOROOT/src/pkg -name "*.go"`