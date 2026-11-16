#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")"

mkdir -p dist
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath \
  -ldflags "-s -w -H=windowsgui" \
  -o dist/FileChecker.exe .

echo "Built dist/FileChecker.exe"
