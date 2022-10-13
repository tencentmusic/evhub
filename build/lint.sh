#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

ROOT=$(dirname "${BASH_SOURCE[0]}")/..

golangci-lint run -c ${ROOT}/.golangci.yml

