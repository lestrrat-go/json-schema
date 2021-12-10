#!/bin/bash

# Script to perform code generation. This exists to overcome
# the fact that go:generate doesn't really allow you to change directories

set -e

pushd internal/cmd/genobjects
go mod tidy
go build -o genobjects main.go
popd

./internal/cmd/genobjects/genobjects -objects=internal/cmd/genobjects/objects.yml

rm internal/cmd/genobjects/genobjects
