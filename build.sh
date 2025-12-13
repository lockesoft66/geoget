#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SOURCE_DIR="$SCRIPT_DIR/source"

build() {
  local os="$1"
  local arch="$2"
  local output="$3"
  local goarm="${4:-}"

  echo "Building ${output} (${os}/${arch}${goarm:+/v${goarm}})"
  if [[ -n "$goarm" ]]; then
    (cd "$SOURCE_DIR" && GOOS="$os" GOARCH="$arch" GOARM="$goarm" CGO_ENABLED=0 go build -o "$SCRIPT_DIR/$output" ./...)
  else
    (cd "$SOURCE_DIR" && GOOS="$os" GOARCH="$arch" CGO_ENABLED=0 go build -o "$SCRIPT_DIR/$output" ./...)
  fi
}

build linux amd64 geoget-linux
build linux arm geoget-linux-arm 7      # Raspberry Pi (ARMv7)
build windows amd64 geoget-win64.exe
build darwin arm64 geoget-mac64

echo "Done. Artifacts written to $SCRIPT_DIR"
