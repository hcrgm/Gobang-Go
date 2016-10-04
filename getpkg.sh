#!/bin/bash
export GOPATH=`pwd`
go get github.com/labstack/echo
go get github.com/auth0/go-jwt-middleware
go get github.com/gorilla/websocket
go get github.com/bitly/go-simplejson
go get github.com/kataras/go-sessions