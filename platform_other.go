//go:build !windows

package main

import (
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"
	"time"
)

func hideWindow(cmd *exec.Cmd) {}

func ensureConsole() {}

func freeSpace(path string) (uint64, error) {
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return 0, err
	}
	return st.Bavail * uint64(st.Bsize), nil
}

func setFileTimes(path string, t time.Time) error { return os.Chtimes(path, t, t) }

func openURL(target string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", target).Start()
	default:
		if _, err := exec.LookPath("xdg-open"); err != nil {
			return err
		}
		return exec.Command("xdg-open", target).Start()
	}
}

func openFolder(p string) error { return openURL(p) }

func folderPickerAvailable() bool {
	switch runtime.GOOS {
	case "darwin":
		_, err := exec.LookPath("osascript")
		return err == nil
	default:
		if _, err := exec.LookPath("zenity"); err == nil {
			return true
		}
		_, err := exec.LookPath("kdialog")
		return err == nil
	}
}

func pickFolder(title string) (string, error) {
	switch runtime.GOOS {
	case "darwin":
		if _, err := exec.LookPath("osascript"); err != nil {
			return "", errNoPicker
		}
		script := "POSIX path of (choose folder with prompt " + quoteApple(title) + ")"
		return pickFrom(exec.Command("osascript", "-e", script))
	default:
		if p, err := exec.LookPath("zenity"); err == nil {
			return pickFrom(exec.Command(p, "--file-selection", "--directory", "--title", title))
		}
		if p, err := exec.LookPath("kdialog"); err == nil {
			return pickFrom(exec.Command(p, "--getexistingdirectory", ".", title))
		}
		return "", errNoPicker
	}
}

func quoteApple(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, ``) + `"`
}

func pickFrom(cmd *exec.Cmd) (string, error) {
	out, err := cmd.Output()
	s := strings.TrimSpace(string(out))
	if err != nil {
		if s == "" {
			return "", nil
		}
		return "", err
	}
	return s, nil
}
