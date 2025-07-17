#!/bin/bash

# Script to perform code generation. This exists to overcome
# the fact that go:generate doesn't really allow you to change directories

set -e

dir=$(cd $(dirname $0); pwd -P)
pushd $dir/internal/cmd/gennumeric
go build -o gennumeric main.go
popd

$dir/internal/cmd/gennumeric/gennumeric -output $dir

rm $dir/internal/cmd/gennumeric/gennumeric
