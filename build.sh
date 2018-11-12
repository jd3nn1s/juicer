#!/bin/bash

export GOOS=linux
export GOARCH=arm
export GOARM=7
go install -v -x
scp ~/go/bin/linux_arm/juicer $SSH_TARGET
