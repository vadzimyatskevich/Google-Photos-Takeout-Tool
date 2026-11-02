//go:build !windows

package main

import _ "embed"

//go:embed assets/exiftool.tar.gz
var embeddedExifTool []byte

const exifToolVersion = "13.60"
