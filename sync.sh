#!/bin/bash
set -xe

cd ~/go/src/github.com/wfunc/util
util_sha=`git rev-parse HEAD`

cd ~/go/src/github.com/wfunc/web
web_sha=`git rev-parse HEAD`


cd ~/go/src/github.com/wfunc/crud
crud_sha=`git rev-parse HEAD`

cd ~/go/src/github.com/wfunc/go
go get github.com/wfunc/util@$util_sha
go get github.com/wfunc/web@$web_sha
go get github.com/wfunc/crud@$crud_sha
go mod tidy

