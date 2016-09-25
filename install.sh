#!/bin/bash
set -e
export GOPATH=`pwd`
echo "Getting packages"
go get github.com/labstack/echo
echo "Installing"
go build gobang
echo "Build success"
