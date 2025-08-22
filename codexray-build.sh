#!/bin/bash

set -e

targets=(
  "linux amd64"
  "linux arm64"
  "windows amd64"
  "windows arm64"
  "darwin amd64"
  "darwin arm64"
)

for target in "${targets[@]}"; do
  os=$(echo $target | cut -d' ' -f1)
  arch=$(echo $target | cut -d' ' -f2)
  ext=""
  if [ "$os" == "windows" ]; then
    ext=".exe"
  fi
  output="codexray-transformer-${os}-${arch}${ext}"
  echo "Building $output"
  GOOS=$os GOARCH=$arch go build -o $output
done
