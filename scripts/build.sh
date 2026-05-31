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

  echo "[1/5] Building $NAME"

  mkdir -p package

  CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH \
    go build -ldflags="-s -w -X main.version=${VERSION}" \
    -o package/burro ./cmd/burro

  echo "[2/5] Copying configs"

  cp -R runtime package/
  find package/runtime -type f \( -name ".gitignore" -o -name ".gitkeep" \) -delete

  echo "[3/5] Creating archive"

  tar -czf "dist/${NAME}.tar.gz" -C package .

  echo "[4/5] Cleanup package"
  rm -rf package

done

echo "[5/5] Generating checksums"
sha256sum dist/*.tar.gz > dist/checksums.txt

echo "Done → dist/"
