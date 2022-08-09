#!/bin/bash

go get

BINARY="boompow-client"

if [ $? == 0 ]; then
  if [ "${GOOS}" == "windows" ]; then
    go build -o ${BINARY}.exe -tags cl -ldflags "-X main.WSUrl=wss://boompow.banano.cc/ws/worker -X main.GraphQLURL=https://boompow.banano.cc/graphql" .
  else
    go build -o ${BINARY} -tags cl -ldflags "-X main.WSUrl=wss://boompow.banano.cc/ws/worker -X main.GraphQLURL=https://boompow.banano.cc/graphql" .
  fi
fi