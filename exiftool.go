package main

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"bytes"
	"compress/gzip"
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
	cmd       *exec.Cmd
	stdin     io.WriteCloser
	stdout    *bufio.Reader
	stderr    *bufio.Reader
	seq       int
	mu        sync.Mutex
	closeOnce sync.Once
}

// findExifTool returns the exiftool executable: explicit path, embedded copy, next to exe, or PATH.
func findExifTool(explicit string) (string, error) {
	if explicit != "" {
		st, err := os.Stat(explicit)
		if err != nil {
			return "", err
		}
		if st.IsDir() {
			return "", fmt.Errorf("%s", explicit)
		}
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

func appDataDir() string {
	switch runtime.GOOS {
	case "windows":
		if base := os.Getenv("LOCALAPPDATA"); base != "" {
			return base
		}
	case "darwin":
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, "Library", "Application Support")
		}
	default:
		if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
			return xdg
		}
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, ".local", "share")
		}
	}
	return os.TempDir()
}

func embeddedExifPaths() (dir, exe string) {
	dir = filepath.Join(appDataDir(), "TakeoutRestore", "exiftool-"+exifToolVersion)
	if runtime.GOOS == "windows" {
		return dir, filepath.Join(dir, "exiftool.exe")
	}
	return dir, filepath.Join(dir, "exiftool")
}

func extractEmbedded() (string, error) {
	dir, exe := embeddedExifPaths()
	if _, err := os.Stat(filepath.Join(dir, ".complete")); err == nil {
		if st, err := os.Stat(exe); err == nil && !st.IsDir() {
			return exe, nil
		}
	}
	os.RemoveAll(dir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	var err error
	if len(embeddedExifTool) >= 2 && embeddedExifTool[0] == 0x1f && embeddedExifTool[1] == 0x8b {
		err = extractTarGz(embeddedExifTool, dir)
	} else {
		err = extractZip(embeddedExifTool, dir)
	}
	if err != nil {
		os.RemoveAll(dir)
		return "", err
	}
	if runtime.GOOS != "windows" {
		_ = os.Chmod(exe, 0o755)
	}
	if err := os.WriteFile(filepath.Join(dir, ".complete"), []byte("ok"), 0o644); err != nil {
		return "", err
	}
	return exe, nil
}

func stripArchiveRoot(name string) string {
	name = strings.ReplaceAll(name, "\\", "/")
	name = strings.TrimPrefix(name, "./")
	if i := strings.Index(name, "/"); i >= 0 {
		return name[i+1:]
	}
	return name
}

func safeArchiveDest(dest, name string) (string, bool) {
	name = filepath.Clean(filepath.FromSlash(name))
	if name == "." || name == "" {
		return "", false
	}
	if filepath.IsAbs(name) {
		return "", false
	}
	dst := filepath.Join(dest, name)
	rel, err := filepath.Rel(dest, dst)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	return dst, true
}

func copyFileFrom(r io.Reader, dst string, mode os.FileMode) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return err
	}
	_, err = io.Copy(out, r)
	cerr := out.Close()
	if err != nil {
		return err
	}
	return cerr
}

func extractZip(data []byte, dir string) error {
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return err
	}
	for _, f := range zr.File {
		name := stripArchiveRoot(f.Name)
		if name == "" || f.FileInfo().IsDir() {
			continue
		}
		if strings.HasPrefix(strings.ToLower(name), "exiftool(-k)") && strings.HasSuffix(strings.ToLower(name), ".exe") {
			name = "exiftool.exe"
		}
		dst, ok := safeArchiveDest(dir, name)
		if !ok {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		err = copyFileFrom(rc, dst, 0o644)
		rc.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func extractTarGz(data []byte, dir string) error {
	gr, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return err
	}
	defer gr.Close()
	tr := tar.NewReader(gr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if hdr.Typeflag != tar.TypeReg {
			continue
		}
		name := stripArchiveRoot(hdr.Name)
		if name != "exiftool" && !strings.HasPrefix(name, "lib/") {
			continue
		}
		dst, ok := safeArchiveDest(dir, name)
		if !ok {
			continue
		}
		mode := os.FileMode(hdr.Mode) & 0o777
		if mode == 0 {
			mode = 0o644
		}
		if name == "exiftool" {
			mode |= 0o111
		}
		if err := copyFileFrom(tr, dst, mode); err != nil {
			return err
		}
	}
	return nil
}

func isPerlScript(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()
	var hdr [2]byte
	n, _ := f.Read(hdr[:])
	return n >= 2 && hdr[0] == '#' && hdr[1] == '!'
}

func ensurePerlIfNeeded(path string) error {
	if runtime.GOOS == "windows" || !isPerlScript(path) {
		return nil
	}
	_, err := exec.LookPath("perl")
	return err
}

func startExifTool(path string) (*ExifTool, error) {
	var cmd *exec.Cmd
	if runtime.GOOS != "windows" && isPerlScript(path) {
		perl, err := exec.LookPath("perl")
		if err != nil {
			return nil, err
		}
		cmd = exec.Command(perl, path, "-stay_open", "True", "-@", "-")
	} else {
		cmd = exec.Command(path, "-stay_open", "True", "-@", "-")
	}
	hideWindow(cmd)
	in, err := cmd.StdinPipe()
	if err != nil {
		return nil, err
	}
	out, _ := cmd.StdoutPipe()
	er, _ := cmd.StderrPipe()
	if err := cmd.Start(); err != nil {
		return nil, err
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

// Kill unblocks a command stuck in -stay_open. Close still waits for the process.
func (et *ExifTool) Kill() {
	if et != nil && et.cmd != nil && et.cmd.Process != nil {
		et.cmd.Process.Kill()
	}
}

func (et *ExifTool) Close() {
	et.closeOnce.Do(func() {
		if et.stdin != nil {
			io.WriteString(et.stdin, "-stay_open\nFalse\n")
			et.stdin.Close()
		}
		if et.cmd != nil {
			et.cmd.Wait()
		}
	})
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
			// 0000:00:00 and the 1904-01-01 QuickTime epoch are "no date".
			// With QuickTimeUTC=1, ExifTool 12.x reports the epoch as 1904:01:01.
			if s == "" || strings.HasPrefix(s, "0000:00:00") || strings.HasPrefix(s, "1904:01:01 00:00:00") || s == "    :  :     :  :  " {
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
