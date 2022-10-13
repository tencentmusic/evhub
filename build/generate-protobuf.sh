#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

ROOT=$(dirname "${BASH_SOURCE[0]}")/..
source "${ROOT}/build/lib/protoc.sh"
source "${ROOT}/build/lib/util.sh"

proto_path="${ROOT}/api"
protoc="${ROOT}/build/lib/proto/prototool"



if [ "$(evhub::util::host_os)" = 'darwin' ]; then
  	 prototool generate ./api/proto
else
    ${protoc} generate ./api/proto
fi





