#!/bin/bash

export GOPATH=$HOME/go

BINARY="boompow-client"

if [ $? == 0 ]; then
  if [ "${GOOS}" == "windows" ]; then
    if [ "${GITHUB_ACTIONS}" == "true" ]; then
      go build -v -ldflags="-X main.gitver=$(git describe --always --long --dirty)" -o ${BINARY}.exe -tags cl ./apps/client 1>> debug.out
      echo "${BINARY}.exe"
    else
      go build -v -ldflags="-X main.gitver=$(git describe --always --long --dirty)" -o ${BINARY}.exe -tags cl ./apps/client
    fi
  else
    if [ "${GITHUB_ACTIONS}" == "true" ]; then
      go build -v -ldflags="-X main.gitver=$(git describe --always --long --dirty)" -o ${BINARY} -tags cl ./apps/client 1>> debug.out
      echo "${BINARY}"
    else
      go build -v -ldflags="-X main.gitver=$(git describe --always --long --dirty)" -o ${BINARY} -tags cl ./apps/client
    fi
  fi
fi