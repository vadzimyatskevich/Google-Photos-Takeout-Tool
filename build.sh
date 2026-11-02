#!/bin/sh
# Builds TakeoutRestore.exe (Windows x64, ExifTool embedded) and Unix archives.
set -e
cd "$(dirname "$0")"
ET_VER="${ET_VER:-13.60}"
if [ ! -f assets/exiftool_win64.zip ]; then
  echo "Downloading ExifTool $ET_VER for Windows..."
  mkdir -p assets
  curl -fL -o assets/exiftool_win64.zip "https://sourceforge.net/projects/exiftool/files/exiftool-${ET_VER}_64.zip/download"
fi
gofmt -l .
go vet ./...
mkdir -p dist

echo "Building Windows amd64..."
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o dist/TakeoutRestore.exe .
rm -rf dist/pkg
mkdir -p dist/pkg/TakeoutRestore
cp dist/TakeoutRestore.exe dist/pkg/TakeoutRestore/
cp docs/README.txt "dist/pkg/TakeoutRestore/README.txt"
cp docs/README_ru.txt "dist/pkg/TakeoutRestore/ПРОЧТИ.txt"
( cd dist/pkg && zip -r ../TakeoutRestore.zip TakeoutRestore )
rm -rf dist/pkg

pack_unix() {
  os="$1"
  arch="$2"
  echo "Building ${os} ${arch}..."
  bin="TakeoutRestore-${os}-${arch}"
  GOOS="$os" GOARCH="$arch" CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o "dist/${bin}" .
  stage=$(mktemp -d)
  mkdir -p "$stage/TakeoutRestore"
  cp "dist/${bin}" "$stage/TakeoutRestore/TakeoutRestore"
  chmod +x "$stage/TakeoutRestore/TakeoutRestore"
  cp docs/README.txt "$stage/TakeoutRestore/README.txt"
  cp docs/README_ru.txt "$stage/TakeoutRestore/ПРОЧТИ.txt"
  tar -C "$stage" -czf "dist/${bin}.tar.gz" TakeoutRestore
  rm -rf "$stage"
}

pack_unix linux amd64
pack_unix linux arm64
pack_unix darwin amd64
pack_unix darwin arm64
ls -la dist
