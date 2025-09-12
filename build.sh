#!/bin/bash
##############################
#####Setting Environments#####
echo "Setting Environments"
set -e
export PATH=$PATH:$GOPATH/bin:$HOME/bin:$GOROOT/bin
##############################
######Install Dependence######
echo "Installing Dependence"
#########Running Clear#########
#########Running Test#########
echo "Running Test"
pkgs="\
  github.com/wfunc/go/baseapi\
  github.com/wfunc/go/basedb\
  github.com/wfunc/go/session\
  github.com/wfunc/go/sms\
  github.com/wfunc/go/upgrade\
  github.com/wfunc/go/util\
"
echo "mode: set" >a.out
for p in $pkgs; do
  go build $p
  go test -v --coverprofile=c.out $p
  cat c.out | grep -v "mode" >>a.out
  go install $p
done
gocov convert a.out >coverage.json

##############################
#####Create Coverage Report###
echo "Create Coverage Report"
cat coverage.json | gocov-xml -source $GOPATH/src > coverage.xml
cat coverage.json | gocov-html coverage.json >coverage.html
