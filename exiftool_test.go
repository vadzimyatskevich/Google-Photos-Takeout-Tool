package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestFindExifToolEmbedded(t *testing.T) {
	if len(embeddedExifTool) == 0 {
		t.Skip("no embedded ExifTool")
	}
	data := t.TempDir()
	t.Setenv("LOCALAPPDATA", data)
	t.Setenv("XDG_DATA_HOME", data)
	t.Setenv("HOME", data)

	p, err := findExifTool("")
	if err != nil {
		t.Fatal(err)
	}
	st, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	if st.IsDir() {
		t.Fatalf("expected file, got dir: %s", p)
	}
	if runtime.GOOS != "windows" {
		if filepath.Base(p) != "exiftool" {
			t.Fatalf("unix exe name: %s", p)
		}
		if !isPerlScript(p) {
			t.Fatalf("expected perl script: %s", p)
		}
		if _, err := os.Stat(filepath.Join(filepath.Dir(p), "lib", "Image", "ExifTool.pm")); err != nil {
			t.Fatal("missing ExifTool.pm:", err)
		}
		if err := ensurePerlIfNeeded(p); err != nil {
			t.Fatal(err)
		}
	}
	et, err := startExifTool(p)
	if err != nil {
		t.Fatal(err)
	}
	defer et.Close()
	so, se, err := et.Run("-ver")
	if err != nil {
		t.Fatalf("exiftool -ver: %v stderr=%s", err, se)
	}
	if !strings.Contains(so, exifToolVersion) {
		t.Fatalf("version %q not in %q", exifToolVersion, so)
	}
}
