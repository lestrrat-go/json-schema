#!/bin/bash

# Script to perform code generation. This exists to overcome
# the fact that go:generate doesn't really allow you to change directories

set -e

pushd internal/cmd/gennumeric
go build -o gennumeric main.go
popd

./internal/cmd/gennumeric/gennumeric 

rm internal/cmd/gennumeric/gennumeric
