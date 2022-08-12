#!/bin/bash

echo "Build linux binary in docker..."

docker run -it --rm -v $(pwd)/apps/client/target:/src/apps/client/target $(docker build -q -f ./apps/client/Dockerfile.build .)