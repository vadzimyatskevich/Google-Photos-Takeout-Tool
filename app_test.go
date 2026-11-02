package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRestoreAppStartCancel(t *testing.T) {
	src := filepath.Join("test", "manual_fixture")
	if _, err := os.Stat(src); err != nil {
		t.Skip("manual fixture missing:", err)
	}
	dest := filepath.Join("test", "manual_out_cancel")
	_ = os.RemoveAll(dest)

	app := NewRestoreApp()
	o := defaultOptions()
	o.Sources = []string{src}
	o.Dest = dest
	o.DryRun = true
	o.TZ = "Europe/Minsk"

	if r := app.Start(o); r["error"] != "" {
		t.Fatalf("start: %v", r)
	}
	if r := app.Start(o); r["error"] == "" {
		t.Fatal("expected second start to fail while running")
	}
	app.Cancel()

	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		s := app.Status()
		running, _ := s["running"].(bool)
		if !running {
			finished, _ := s["finished"].(bool)
			if !finished {
				t.Fatalf("not running but not finished: %#v", s)
			}
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("cancel did not finish in time")
}

func TestRestoreAppPauseResume(t *testing.T) {
	src := filepath.Join("test", "manual_fixture")
	if _, err := os.Stat(src); err != nil {
		t.Skip("manual fixture missing:", err)
	}
	dest := filepath.Join("test", "manual_out_pause")
	_ = os.RemoveAll(dest)

	app := NewRestoreApp()
	o := defaultOptions()
	o.Sources = []string{src}
	o.Dest = dest
	o.DryRun = true
	o.TZ = "Europe/Minsk"
	o.Workers = 1

	if r := app.Start(o); r["error"] != "" {
		t.Fatalf("start: %v", r)
	}
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		s := app.Status()
		if running, _ := s["running"].(bool); running {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	app.Pause()
	pausedSeen := false
	for i := 0; i < 50; i++ {
		s := app.Status()
		if paused, _ := s["paused"].(bool); paused {
			pausedSeen = true
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if !pausedSeen {
		app.Cancel()
		t.Fatal("expected paused status")
	}
	time.Sleep(150 * time.Millisecond)
	s := app.Status()
	if running, _ := s["running"].(bool); !running {
		t.Fatal("paused job should still be running")
	}
	app.Resume()
	for i := 0; i < 50; i++ {
		s = app.Status()
		if paused, _ := s["paused"].(bool); !paused {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if paused, _ := s["paused"].(bool); paused {
		app.Cancel()
		t.Fatal("expected resume to clear paused")
	}
	app.Cancel()
	deadline = time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		s = app.Status()
		if running, _ := s["running"].(bool); !running {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	t.Fatal("cancel after pause did not finish in time")
}
