#!/bin/sh
# Builds TakeoutRestore GUI + CLI for Windows/Linux/macOS.
# Both variants are produced for every platform that this host can compile.
# Desktop GUI: -tags desktop,production,gui (Wails). CLI+web: no gui tag (CGO-free).
# Linux desktop GUI needs CGO + GTK/WebKit (Ubuntu: libgtk-3-dev libwebkit2gtk-4.1-dev).
# macOS desktop GUI needs a macOS SDK; linux/arm64 GUI needs an aarch64 GTK sysroot.
set -e
cd "$(dirname "$0")"

ET_VER="${ET_VER:-13.60}"
GUI_TAGS=desktop,production,gui
HOST_OS="$(go env GOOS)"
HOST_ARCH="$(go env GOARCH)"
LINUX_GUI_TAGS="$GUI_TAGS"

if pkg-config --exists webkit2gtk-4.1 >/dev/null 2>&1; then
  LINUX_GUI_TAGS="$GUI_TAGS,webkit2_41"
fi

gtk_ok=0
if pkg-config --exists gtk+-3.0 >/dev/null 2>&1; then
  if pkg-config --exists webkit2gtk-4.1 >/dev/null 2>&1 || pkg-config --exists webkit2gtk-4.0 >/dev/null 2>&1; then
    gtk_ok=1
  fi
fi

if [ ! -f assets/exiftool_win64.zip ]; then
  echo "Downloading ExifTool $ET_VER for Windows..."
  mkdir -p assets
  curl -fL -o assets/exiftool_win64.zip "https://sourceforge.net/projects/exiftool/files/exiftool-${ET_VER}_64.zip/download"
fi

if [ ! -f assets/exiftool.tar.gz ]; then
  echo "Downloading ExifTool $ET_VER for Linux/macOS..."
  mkdir -p assets
  curl -fL -o assets/exiftool.tar.gz "https://sourceforge.net/projects/exiftool/files/Image-ExifTool-${ET_VER}.tar.gz/download"
fi

if [ ! -f frontend/dist/index.html ] || [ ! -f frontend/dist/app.js ]; then
  echo "error: frontend/dist is incomplete (need index.html and app.js for GUI builds)" >&2
  exit 1
fi

gofmt -l .
CGO_ENABLED=0 go vet ./...
if [ "$gtk_ok" = 1 ] && [ "$HOST_OS" = linux ]; then
  go vet -tags="$LINUX_GUI_TAGS" ./...
else
  echo "Skipping GUI vet: Linux desktop GUI needs GTK/WebKit headers (pkg-config gtk+-3.0 + webkit2gtk)."
fi
go mod tidy
mkdir -p dist

ensure_windows_icon() {
  if [ ! -f build/windows/icon256.png ]; then
    echo "Missing build/windows/icon256.png (app icon); continuing without refresh."
    return 0
  fi
  if command -v go-winres >/dev/null 2>&1; then
    echo "Refreshing Windows icon resource..."
    go-winres simply \
      --icon build/windows/icon256.png \
      --arch amd64 \
      --manifest gui \
      --product-name TakeoutRestore \
      --file-description "Restore Google Takeout photos" \
      --original-filename TakeoutRestore.exe \
      --product-version 1.0.0 \
      --file-version 1.0.0 \
      --copyright MIT \
      --out rsrc
  elif [ ! -f rsrc_windows_amd64.syso ]; then
    echo "Warning: rsrc_windows_amd64.syso missing and go-winres not installed; exe will have no custom icon."
  fi
}

copy_docs() {
  dest_dir="$1"
  cp docs/README.txt "$dest_dir/README.txt"
  cp docs/README_ru.txt "$dest_dir/ПРОЧТИ.txt"
}

pack_windows_gui() {
  echo "Building Windows amd64 (desktop GUI)..."
  ensure_windows_icon
  gui_out=dist/TakeoutRestore-windows-amd64-gui.exe
  GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -tags "$GUI_TAGS" -trimpath -ldflags "-s -w -H windowsgui" -o "$gui_out" .
  rm -rf dist/pkg
  mkdir -p dist/pkg/TakeoutRestore
  # Zip keeps a short name for end users who double-click the app.
  cp "$gui_out" dist/pkg/TakeoutRestore/TakeoutRestore.exe
  copy_docs dist/pkg/TakeoutRestore
  ( cd dist/pkg && zip -r ../TakeoutRestore.zip TakeoutRestore )
  rm -rf dist/pkg
}

pack_unix_gui() {
  os="$1"
  arch="$2"
  echo "Building ${os} ${arch} (desktop GUI)..."
  bin="TakeoutRestore-${os}-${arch}"
  if [ "$os" != "$HOST_OS" ] || [ "$arch" != "$HOST_ARCH" ]; then
    echo "Skipping ${os}/${arch} desktop GUI: Wails needs CGO and native WebKit/GTK (or a macOS SDK)."
    return 0
  fi
  if [ "$os" = linux ] && [ "$gtk_ok" != 1 ]; then
    echo "Skipping linux/${arch} desktop GUI: install libgtk-3-dev and libwebkit2gtk-4.1-dev (or 4.0)."
    return 0
  fi
  tags="$GUI_TAGS"
  cgo=0
  if [ "$os" = linux ]; then
    tags="$LINUX_GUI_TAGS"
    cgo=1
  elif [ "$os" = darwin ]; then
    cgo=1
  fi
  GOOS="$os" GOARCH="$arch" CGO_ENABLED="$cgo" go build -tags "$tags" -trimpath -ldflags "-s -w" -o "dist/${bin}" .
  stage=$(mktemp -d)
  mkdir -p "$stage/TakeoutRestore"
  cp "dist/${bin}" "$stage/TakeoutRestore/TakeoutRestore"
  chmod +x "$stage/TakeoutRestore/TakeoutRestore"
  copy_docs "$stage/TakeoutRestore"
  tar -C "$stage" -czf "dist/${bin}.tar.gz" TakeoutRestore
  rm -rf "$stage"
}

pack_cli() {
  os="$1"
  arch="$2"
  echo "Building ${os} ${arch} (CLI + web, no gui tag)..."
  out="dist/TakeoutRestore-${os}-${arch}-cli"
  if [ "$os" = windows ]; then
    out="${out}.exe"
  fi
  GOOS="$os" GOARCH="$arch" CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o "$out" .
}

# GUI
pack_windows_gui
pack_unix_gui linux amd64
pack_unix_gui linux arm64
pack_unix_gui darwin amd64
pack_unix_gui darwin arm64

# CLI (same platform matrix)
pack_cli windows amd64
pack_cli linux amd64
pack_cli linux arm64
pack_cli darwin amd64
pack_cli darwin arm64

echo
echo "Artifacts in dist/:"
ls -la dist
