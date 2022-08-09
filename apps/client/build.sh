#!/bin/bash

export GOPATH=$HOME/go

BINARY="boompow-client"

if [ $? == 0 ]; then
  if [ "${GOOS}" == "windows" ]; then
    if [ "${GITHUB_ACTIONS}" == "true" ]; then
      go build -v -ldflags="-X main.gitver=$(git describe --always --long --dirty)" -o ${BINARY}.exe -tags cl . 1>> debug.out
      echo "${BINARY}.exe"
    else
      go build -v -ldflags="-X main.gitver=$(git describe --always --long --dirty)" -o ${BINARY}.exe -tags cl .
    fi
  else
    if [ "${GITHUB_ACTIONS}" == "true" ]; then
      go build -v -ldflags="-X main.gitver=$(git describe --always --long --dirty)" -o ${BINARY} -tags cl . 1>> debug.out
      echo "${BINARY}"
    else
      go build -v -ldflags="-X main.gitver=$(git describe --always --long --dirty)" -o ${BINARY} -tags cl .
    fi
  fi
fi