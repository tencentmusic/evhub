#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# The root of the build/dist directory
evhub_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd -P)"

# Generates *.pb.go from the protobuf file under $1 and formats it correctly
function evhub::protoc::generate_proto() {
  evhub::protoc::check_protoc
  evhub::protoc::check_protoc_pkg_go
  evhub::protoc::check_protoc_pkg_go_grpc

  local package=${1}
  evhub::protoc::protoc "${package}"

  for i in `find ${package} -name "*.pb.go"`
  do
    evhub::protoc::format "$i"
  done
}

# Checks that the current protoc version is at least version 3.0.0-beta1
# exit 1 if it's not the case
function evhub::protoc::check_protoc() {
  if [[ -z "$(which protoc)" || "$(protoc --version)" != "libprotoc 3."* ]]; then
    echo "Generating protobuf requires protoc 3.0.0-beta1 or newer. Please download and"
    echo "install the platform appropriate Protobuf package for your OS: "
    echo
    echo "  https://github.com/google/protobuf/releases"
    echo
    echo "WARNING: Protobuf changes are not being validated"
    exit 1
  fi
}

# Checks that the current protoc-gen-gofast
function evhub::protoc::check_protoc_pkg_gofast() {
  if [[ -z "$(which protoc-gen-gofast)" ]]; then
    echo "Generating protobuf requires protoc-gen-gofast. Please download and"
    echo "install the platform appropriate package for your OS: "
    echo
    echo "  https://github.com/gogo/protobuf"
    echo
    exit 1
  fi
}

function evhub::protoc::check_protoc_pkg_go() {
  if [[ -z "$(which protoc-gen-go)" ]]; then
    echo "Generating protobuf requires protoc-gen-go. Please download and"
    echo "install the platform appropriate package for your OS: "
    echo
    echo "  go get google.golang.org/protobuf/cmd/protoc-gen-go"
    echo
  fi
}

function evhub::protoc::check_protoc_pkg_go_grpc() {
  if [[ -z "$(which protoc-gen-go-grpc)" ]]; then
    echo "Generating grpc server requires protoc-gen-go-grpc. Please download and"
    echo "install the platform appropriate package for your OS: "
    echo
    echo "  go get google.golang.org/grpc/cmd/protoc-gen-go-grpc"
    echo
  fi
}

# Generates *.pb.go from the protobuf file under $1
function evhub::protoc::protoc() {
  local package=${1}

  for file in $(find ${package} -name "*.proto"); do
    protoc \
      --proto_path="${package}" \
      --go_opt=paths=source_relative \
      --go_out="${package}" ${file}
  done
}

# Formats $1
function evhub::protoc::format() {
  local package=${1}

  # Run gofmt and goimports to clean up the generated code.
  gofmt -l -s -w "${package}" > /dev/null
  goimports -w "${package}" > /dev/null
}
