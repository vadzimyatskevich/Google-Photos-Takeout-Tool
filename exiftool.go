package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

type ExifTool struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Reader
	stderr *bufio.Reader
	seq    int
	mu     sync.Mutex
}

// findExifTool returns the exiftool executable: explicit path, embedded copy (Windows), next to exe, or PATH.
func findExifTool(explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}
	if len(embeddedExifTool) > 0 {
		return extractEmbedded()
	}
	if exe, err := os.Executable(); err == nil {
		d := filepath.Dir(exe)
		for _, c := range []string{"exiftool.exe", "exiftool/exiftool.exe", "exiftool(-k).exe", "exiftool"} {
			p := filepath.Join(d, c)
			if st, err := os.Stat(p); err == nil && !st.IsDir() {
				return p, nil
			}
		}
	}
	return exec.LookPath("exiftool")
}

func extractEmbedded() (string, error) {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		base = os.TempDir()
	}
	dir := filepath.Join(base, "TakeoutRestore", "exiftool-"+exifToolVersion)
	exe := filepath.Join(dir, "exiftool.exe")
	if _, err := os.Stat(filepath.Join(dir, ".complete")); err == nil {
		return exe, nil
	}
	os.RemoveAll(dir)
	zr, err := zip.NewReader(bytes.NewReader(embeddedExifTool), int64(len(embeddedExifTool)))
	if err != nil {
		return "", err
	}
	for _, f := range zr.File {
		name := f.Name
		if i := strings.Index(name, "/"); i >= 0 { // strip top folder
			name = name[i+1:]
		}
		if name == "" || f.FileInfo().IsDir() {
			continue
		}
		if strings.HasPrefix(strings.ToLower(name), "exiftool(-k)") && strings.HasSuffix(strings.ToLower(name), ".exe") {
			name = "exiftool.exe"
		}
		dst := filepath.Join(dir, filepath.FromSlash(name))
		os.MkdirAll(filepath.Dir(dst), 0o755)
		rc, err := f.Open()
		if err != nil {
			return "", err
		}
		out, err := os.Create(dst)
		if err != nil {
			rc.Close()
			return "", err
		}
		_, err = io.Copy(out, rc)
		rc.Close()
		out.Close()
		if err != nil {
			return "", err
		}
	}
	os.WriteFile(filepath.Join(dir, ".complete"), []byte("ok"), 0o644)
	return exe, nil
}

func startExifTool(path string) (*ExifTool, error) {
	cmd := exec.Command(path, "-stay_open", "True", "-@", "-")
	hideWindow(cmd)
	in, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	out, _ := cmd.StdoutPipe()
	er, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("не удалось запустить exiftool (%s): %v", path, err)
	}
	return &ExifTool{cmd: cmd, stdin: in, stdout: bufio.NewReaderSize(out, 1<<16), stderr: bufio.NewReaderSize(er, 1<<16)}, nil
}

func readUntil(r *bufio.Reader, marker string) (string, error) {
	var sb strings.Builder
	for {
		line, err := r.ReadString('\n')
		if strings.TrimRight(line, "\r\n") == marker {
			return sb.String(), nil
		}
		sb.WriteString(line)
		if err != nil {
			return sb.String(), err
		}
	}
}

// Run executes one command; args must not contain newlines.
func (et *ExifTool) Run(args ...string) (string, string, error) {
	et.mu.Lock()
	defer et.mu.Unlock()
	et.seq++
	mark := fmt.Sprintf("{ready%d}", et.seq)
	var b strings.Builder
	b.WriteString("-charset\nfilename=utf8\n")
	for _, a := range args {
		b.WriteString(strings.NewReplacer("\r", " ", "\n", " ").Replace(a))
		b.WriteByte('\n')
	}
	fmt.Fprintf(&b, "-echo4\n%s\n-execute%d\n", mark, et.seq)
	if _, err := io.WriteString(et.stdin, b.String()); err != nil {
		return "", "", err
	}
	so, err := readUntil(et.stdout, mark)
	if err != nil {
		return so, "", err
	}
	se, err := readUntil(et.stderr, mark)
	return so, se, err
}

func (et *ExifTool) Close() {
	io.WriteString(et.stdin, "-stay_open\nFalse\n")
	et.stdin.Close()
	et.cmd.Wait()
}

// ReadTags returns existing tag values (group0:tag -> string).
func (et *ExifTool) ReadTags(path string) map[string]string {
	so, _, err := et.Run("-j", "-G0", "-api", "QuickTimeUTC=1", "-api", "LargeFileSupport=1",
		"-EXIF:DateTimeOriginal", "-EXIF:CreateDate", "-EXIF:ModifyDate", "-EXIF:OffsetTimeOriginal",
		"-XMP:DateTimeOriginal", "-XMP:CreateDate", "-XMP:DateCreated",
		"-QuickTime:CreateDate", "-QuickTime:MediaCreateDate", "-Keys:CreationDate",
		"-GPSLatitude#", "-GPSLongitude#", "-EXIF:ImageDescription", "-XMP:Description", "-IPTC:Caption-Abstract",
		"-Keys:Description", "-XMP:Rating", path)
	res := map[string]string{}
	if err != nil {
		return res
	}
	var arr []map[string]any
	if json.Unmarshal([]byte(so), &arr) == nil && len(arr) > 0 {
		for k, v := range arr[0] {
			s := strings.TrimSpace(fmt.Sprint(v))
			if s == "" || strings.HasPrefix(s, "0000:00:00") || s == "    :  :     :  :  " {
				continue
			}
			res[k] = s
		}
	}
	return res
}

func defaultWorkers() int {
	n := runtime.NumCPU()
	if n > 4 {
		n = 4
	}
	if n < 1 {
		n = 1
	}
	return n
}
