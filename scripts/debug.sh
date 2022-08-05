#!/bin/bash

# Simple helper to start debugging easier
# Run this, then run the "Launch Debugger" task in vscode

usage() { echo "Usage: $0 [-t <server|client>]" 1>&2; exit 1; }

unset -v type

while getopts t: opt; do
  case $opt in
    t) type=$OPTARG ;;
    *) usage ;;
  esac
done

shift "$(( OPTIND - 1 ))"

if [ -z "$type" ]; then
  usage
  exit 1
fi

if [ "$type" != "server" ] && [ "$type" != "client" ]; then
  usage
  exit 1
fi

if [ "$type" == "server" ]; then
  dlv debug --headless --listen=:2345 --api-version=2 --log github.com/bananocoin/boompow-next/services/server -- server
elif [ "$type" == "client" ]; then
  dlv debug --headless --listen=:2345 --api-version=2 --log github.com/bananocoin/boompow-next/services/client
fi
