//go:build windows

package main

import _ "embed"

//go:embed assets/exiftool_win64.zip
var embeddedExifTool []byte

const exifToolVersion = "13.60"
