#!/usr/bin/env bash

workdir=$(cd `dirname $0`; pwd)

docker-compose -f ${workdir}/docker-compose.yml stop

