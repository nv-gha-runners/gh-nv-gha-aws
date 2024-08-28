#!/bin/bash

set -euo pipefail

TAG=$1

PLATFORMS=(
  darwin-amd64
  darwin-arm64
  linux-amd64
  linux-arm64
  windows-amd64
  windows-arm64
)

mkdir -p dist

for p in "${PLATFORMS[@]}"; do
  GOOS="${p%-*}"
  GOARCH="${p#*-}"
  EXT=""
  if [ "$GOOS" = "windows" ]; then
    EXT=".exe"
  fi
  make GOOS="${GOOS}" GOARCH="${GOARCH}" BIN_NAME="dist/gh-nv-gha-aws_${TAG}_${GOOS}-${GOARCH}${EXT}"
done

ls dist
