package main

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
)

// RestoreApp is the shared control plane used by the web UI and the desktop GUI.
type RestoreApp struct {
	prog Progress
}

func NewRestoreApp() *RestoreApp { return &RestoreApp{} }

func (a *RestoreApp) Defaults() Options { return defaultOptions() }

func (a *RestoreApp) Info() map[string]any {
	return map[string]any{"pick": folderPickerAvailable(), "os": runtime.GOOS}
}

// Start begins a restore job in a background goroutine.
// Returns an error message string (empty on success) for easy JS binding.
func (a *RestoreApp) Start(o Options) map[string]string {
	L := NormalizeLang(o.Lang)
	o.Lang = string(L)
	a.prog.mu.Lock()
	running := a.prog.Running
	if !running {
		a.prog.Running = true
	}
	a.prog.mu.Unlock()
	if running {
		return map[string]string{"error": L.T("web.running")}
	}
	a.prog.cancel.Store(false)
	a.prog.pause.Store(false)
	go Run(o, &a.prog)
	return map[string]string{"ok": "1"}
}

func (a *RestoreApp) Status() map[string]any { return a.prog.snapshot() }

func (a *RestoreApp) Cancel() {
	a.prog.pause.Store(false) // wake workers blocked in Pause
	a.prog.cancel.Store(true)
	a.prog.set(func(p *Progress) {
		p.Paused = false
		if p.Running {
			p.Phase = NormalizeLang(p.Lang).T("phase.stopping")
		}
	})
}

// Pause freezes the current job in place (no re-scan). Resume continues it.
func (a *RestoreApp) Pause() {
	a.prog.mu.Lock()
	running := a.prog.Running
	cancelled := a.prog.cancel.Load()
	a.prog.mu.Unlock()
	if !running || cancelled {
		return
	}
	a.prog.pause.Store(true)
	a.prog.set(func(p *Progress) {
		if !p.Running || p.cancel.Load() {
			return
		}
		p.Paused = true
		if p.phaseWork == "" {
			p.phaseWork = p.Phase
		}
		p.Phase = NormalizeLang(p.Lang).T("phase.paused")
	})
}

func (a *RestoreApp) Resume() {
	a.prog.pause.Store(false)
	a.prog.set(func(p *Progress) {
		p.Paused = false
		if p.Running && !p.cancel.Load() {
			if p.phaseWork != "" {
				p.Phase = p.phaseWork
			}
		}
	})
}

// IsRunning reports whether a job is currently in progress (for window-close confirm).
func (a *RestoreApp) IsRunning() bool {
	a.prog.mu.Lock()
	defer a.prog.mu.Unlock()
	return a.prog.Running
}

func (a *RestoreApp) PickFolder(lang string) map[string]string {
	L := NormalizeLang(lang)
	p, err := pickFolder(L.T("pick.title"))
	res := map[string]string{"path": p}
	if err != nil {
		if errors.Is(err, errNoPicker) {
			res["error"] = L.T("pick.err")
		} else if p == "" {
			res["error"] = L.T("pick.errdetail", err)
		}
	}
	return res
}

func (a *RestoreApp) OpenPath(what string) map[string]string {
	s := a.prog.snapshot()
	var p string
	switch what {
	case "dest":
		p, _ = s["dest"].(string)
	case "report":
		p, _ = s["report"].(string)
	case "log":
		p, _ = s["log"].(string)
	case "reportdir":
		rp, _ := s["report"].(string)
		if rp != "" {
			p = filepath.Dir(rp)
		}
	}
	if p == "" {
		return map[string]string{"error": "missing"}
	}
	if _, err := os.Stat(p); err != nil {
		return map[string]string{"error": err.Error()}
	}
	switch what {
	case "report", "log":
		_ = openURL(p)
	default:
		_ = openFolder(p)
	}
	return map[string]string{"ok": "1"}
}

func (a *RestoreApp) Progress() *Progress { return &a.prog }
