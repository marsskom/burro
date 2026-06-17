#!/usr/bin/env bash
set -euo pipefail

VERSION=${VERSION:-dev}

echo "[0/5] Cleaning workspace"
rm -rf dist package
mkdir -p dist

TARGETS=(
  "linux amd64"
  "linux arm64"
  "darwin amd64"
  "darwin arm64"
)

for t in "${TARGETS[@]}"; do
  GOOS=$(echo "$t" | awk '{print $1}')
  GOARCH=$(echo "$t" | awk '{print $2}')

  NAME="burro-${VERSION}-${GOOS}-${GOARCH}"

  echo "[1/4] Building $NAME"

  mkdir -p package

  CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH \
    go build -ldflags="-s -w -X main.version=${VERSION}" \
    -o package/burro ./cmd/burro

  echo "[2/4] Creating archive"

  tar -czf "dist/${NAME}.tar.gz" -C package .

  echo "[3/4] Cleanup package"
  rm -rf package

done

echo "[4/4] Generating checksums"
sha256sum dist/*.tar.gz > dist/checksums.txt

echo "Done → dist/"
