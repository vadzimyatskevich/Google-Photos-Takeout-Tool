#!/bin/sh
# Builds TakeoutRestore.exe (Windows x64) with ExifTool embedded.
set -e
cd "$(dirname "$0")"
ET_VER="${ET_VER:-13.60}"
if [ ! -f assets/exiftool_win64.zip ]; then
  echo "Downloading ExifTool $ET_VER for Windows..."
  mkdir -p assets
  curl -fL -o assets/exiftool_win64.zip "https://sourceforge.net/projects/exiftool/files/exiftool-${ET_VER}_64.zip/download"
fi
gofmt -l . ; go vet ./...
mkdir -p dist
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o dist/TakeoutRestore.exe .
cd dist && rm -f TakeoutRestore.zip && mkdir -p pkg/TakeoutRestore && cp TakeoutRestore.exe pkg/TakeoutRestore/ && cp ../docs/README_ru.txt "pkg/TakeoutRestore/ПРОЧТИ.txt" && (cd pkg && zip -r ../TakeoutRestore.zip TakeoutRestore) && rm -rf pkg
ls -la
