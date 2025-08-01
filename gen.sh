#!/bin/bash

# Script to perform code generation. This exists to overcome
# the fact that go:generate doesn't really allow you to change directories

set -e

dir=$(cd $(dirname $0); pwd -P)

pushd $dir/internal/cmd/genobjects
go mod tidy
go build -o genobjects main.go
popd

pushd $dir
$dir/internal/cmd/genobjects/genobjects -objects=$dir/internal/cmd/genobjects/objects.yml
popd

rm $dir/internal/cmd/genobjects/genobjects
