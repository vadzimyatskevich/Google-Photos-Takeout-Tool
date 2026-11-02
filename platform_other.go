//go:build !windows

package main

import (
	"errors"
	"os"
	"os/exec"
	"syscall"
	"time"
)

func hideWindow(cmd *exec.Cmd) {}

func freeSpace(path string) (uint64, error) {
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return 0, err
	}
	return st.Bavail * uint64(st.Bsize), nil
}

func setFileTimes(path string, t time.Time) error { return os.Chtimes(path, t, t) }

func openURL(u string) error      { return exec.Command("xdg-open", u).Start() }
func openFolder(p string) error   { return exec.Command("xdg-open", p).Start() }
func pickFolder() (string, error) { return "", errors.New("folder dialog not supported on this OS") }
