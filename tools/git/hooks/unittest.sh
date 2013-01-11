#!/bin/sh

scratch=$(mktemp -d)
function finish {
  rm -rf "$scratch"
}
trap finish EXIT

git checkout-index --prefix=$scratch/ -a
pushd $scratch/
make test
exit $?
