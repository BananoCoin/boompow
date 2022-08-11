#!/bin/bash

go get

BINARY="boompow-client"

if [ $? == 0 ]; then
  if [ "${GOOS}" == "windows" ]; then
    go build -o ${BINARY}.exe -tags cl -race -ldflags "-w -s -X main.WSUrl=wss://boompow.banano.cc/ws/worker -X main.GraphQLURL=https://boompow.banano.cc/graphql -X main.Version=`git tag --sort=-version:refname | head -n 1`" .
  else
    go build -o ${BINARY} -tags cl -race -ldflags "-w -s -X main.WSUrl=wss://boompow.banano.cc/ws/worker -X main.GraphQLURL=https://boompow.banano.cc/graphql -X main.Version=`git tag --sort=-version:refname | head -n 1`" .
  fi
fi