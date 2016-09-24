#!/bin/bash
echo "Getting packages"
GOPATH=`pwd` go get -u github.com/labstack/echo
echo "Installing"
GOPATH=`pwd` go install gobang
echo "Get 'gobang' binary file at bin directory"
