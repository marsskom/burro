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
    -o package/burro ./cmd/proxy

  CGO_ENABLED=0 GOOS=$GOOS GOARCH=$GOARCH \
    go build -ldflags="-s -w" \
    -o package/certgen ./cmd/certgen

  echo "[2/5] Copying configs"

  cp config.yml package/
  cp .env.example package/.env
  cp -R data package/

  find plugins -type f -name '*.yml' | while read -r file; do
    mkdir -p "package/$(dirname "$file")"
    cp "$file" "package/$file"
  done

  echo "[3/5] Creating archive"

  tar -czf "dist/${NAME}.tar.gz" -C package .

  echo "[4/5] Cleanup package"
  rm -rf package

done

echo "[5/5] Generating checksums"
sha256sum dist/*.tar.gz > dist/checksums.txt

echo "Done → dist/"
