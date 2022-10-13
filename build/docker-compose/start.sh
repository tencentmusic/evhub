#!/usr/bin/env bash

workdir=$(cd `dirname $0`; pwd)

if !(command -v grpcui > /dev/null 2>&1); then
    go install github.com/fullstorydev/grpcui/cmd/grpcui@latest
fi

if [ $(docker images | grep -c "evhub/producer") -lt 1 ]; then
  make image TARGET=producer VERSION=latest;
fi

if [ $(docker images | grep -c "evhub/processor") -lt 1 ]; then
  make image TARGET=processor VERSION=latest;
fi

if [ $(docker images | grep -c "evhub/admin") -lt 1 ]; then
  make image TARGET=admin VERSION=latest;
fi

docker-compose -f ${workdir}/docker-compose.yml up -d

sleep 10
grpcui -plaintext localhost:9000