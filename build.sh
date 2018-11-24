#!/bin/bash

set -e
pushd ../kw1281
go test
popd
pushd ../skytraq
go test
popd
go test ./...
export GOOS=linux
export GOARCH=arm
export GOARM=7
unset GOBIN
go install ./...
scp ~/go/bin/linux_arm/juicer $SSH_TARGET
