#!/bin/bash
set -e
export GOPATH=`pwd`
echo "Getting packages"
./getpkg.sh
echo "Installing"
go build src/gobang.go
echo "Build success"
