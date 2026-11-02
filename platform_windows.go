//go:build windows

package main

import (
	"os/exec"
	"strings"
	"syscall"
	"time"
	"unsafe"
)

var (
	kernel32        = syscall.NewLazyDLL("kernel32.dll")
	procGetDiskFree = kernel32.NewProc("GetDiskFreeSpaceExW")
)

func hideWindow(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
}

func freeSpace(path string) (uint64, error) {
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return 0, err
	}
	var avail, total, free uint64
	r, _, e := procGetDiskFree.Call(uintptr(unsafe.Pointer(p)), uintptr(unsafe.Pointer(&avail)), uintptr(unsafe.Pointer(&total)), uintptr(unsafe.Pointer(&free)))
	if r == 0 {
		return 0, e
	}
	return avail, nil
}

// setFileTimes sets creation, access and modification time.
func setFileTimes(path string, t time.Time) error {
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return err
	}
	h, err := syscall.CreateFile(p, 0x100 /*FILE_WRITE_ATTRIBUTES*/, syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE|syscall.FILE_SHARE_DELETE, nil, syscall.OPEN_EXISTING, syscall.FILE_FLAG_BACKUP_SEMANTICS, 0)
	if err != nil {
		return err
	}
	defer syscall.CloseHandle(h)
	ft := syscall.NsecToFiletime(t.UnixNano())
	return syscall.SetFileTime(h, &ft, &ft, &ft)
}

func openURL(u string) error {
	return exec.Command("rundll32", "url.dll,FileProtocolHandler", u).Start()
}

func openFolder(p string) error {
	return exec.Command("explorer", p).Start()
}

func pickFolder() (string, error) {
	script := `Add-Type -AssemblyName System.Windows.Forms;` +
		`$o=New-Object System.Windows.Forms.Form -Property @{TopMost=$true;ShowInTaskbar=$false};` +
		`$f=New-Object System.Windows.Forms.FolderBrowserDialog;$f.Description='Выберите папку';$f.ShowNewFolderButton=$true;` +
		`[Console]::OutputEncoding=[Text.Encoding]::UTF8;` +
		`if($f.ShowDialog($o) -eq 'OK'){[Console]::Out.Write($f.SelectedPath)}`
	cmd := exec.Command("powershell", "-NoProfile", "-STA", "-ExecutionPolicy", "Bypass", "-Command", script)
	hideWindow(cmd)
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}
