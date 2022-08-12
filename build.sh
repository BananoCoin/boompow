#!/bin/bash

echo "Build for macOS & Linux" 

cd ./apps/client
go get

BINARY="boompow-client"
rm -f ./target/${BINARY}

mkdir -p ./target
go build -o ./target/${BINARY} -tags cl -race -ldflags "-w -s -X main.WSUrl=wss://boompow.banano.cc/ws/worker -X main.GraphQLURL=https://boompow.banano.cc/graphql -X main.Version=`git tag --sort=-version:refname | head -n 1`" .
echo "Output ./apps/client/target/${BINARY}"