//go:build gui

package main

import (
	"context"
	"embed"
	"fmt"
	"net/http"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/linux"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

//go:embed all:frontend/dist
var guiAssets embed.FS

//go:embed build/appicon.png
var appIcon []byte

const guiEnabled = true

// DesktopApp is the Wails-bound controller for the desktop window.
type DesktopApp struct {
	*RestoreApp
	ctx          context.Context
	defaultLang  Lang
	langExplicit bool
	progressStop chan struct{}
}

func (g *DesktopApp) startup(ctx context.Context) {
	g.ctx = ctx
	g.progressStop = make(chan struct{})
	go g.emitProgress()
}

func (g *DesktopApp) shutdown(ctx context.Context) {
	if g.progressStop != nil {
		select {
		case <-g.progressStop:
		default:
			close(g.progressStop)
		}
	}
	g.Cancel()
}

func (g *DesktopApp) beforeClose(ctx context.Context) (prevent bool) {
	if !g.IsRunning() {
		return false
	}
	g.Cancel()
	deadline := time.Now().Add(800 * time.Millisecond)
	for g.IsRunning() && time.Now().Before(deadline) {
		time.Sleep(50 * time.Millisecond)
	}
	return false
}

func (g *DesktopApp) emitProgress() {
	t := time.NewTicker(250 * time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-g.progressStop:
			return
		case <-t.C:
			if g.ctx == nil {
				continue
			}
			runtime.EventsEmit(g.ctx, "progress", g.Status())
		}
	}
}

// Config returns initial UI settings for the frontend.
func (g *DesktopApp) Config() map[string]any {
	return map[string]any{
		"defaultLang":  string(g.defaultLang),
		"langExplicit": g.langExplicit,
		"os":           goruntime.GOOS,
		"pick":         true,
	}
}

// PickFolderDialog uses the native Wails directory dialog.
func (g *DesktopApp) PickFolderDialog(lang string) map[string]string {
	L := NormalizeLang(lang)
	p, err := runtime.OpenDirectoryDialog(g.ctx, runtime.OpenDialogOptions{
		Title: L.T("pick.title"),
	})
	if err != nil {
		return map[string]string{"error": err.Error()}
	}
	return map[string]string{"path": p}
}

// PickZipDialog opens a multi-file dialog for zip archives.
func (g *DesktopApp) PickZipDialog(lang string) map[string]any {
	L := NormalizeLang(lang)
	paths, err := runtime.OpenMultipleFilesDialog(g.ctx, runtime.OpenDialogOptions{
		Title: L.T("pick.title"),
		Filters: []runtime.FileFilter{
			{DisplayName: "ZIP", Pattern: "*.zip"},
		},
	})
	if err != nil {
		return map[string]any{"error": err.Error()}
	}
	return map[string]any{"paths": paths}
}

func (g *DesktopApp) OpenPath(what string) map[string]string {
	s := g.Status()
	var p string
	switch what {
	case "dest":
		p, _ = s["dest"].(string)
	case "report":
		p, _ = s["report"].(string)
	case "log":
		p, _ = s["log"].(string)
	case "reportdir":
		return g.RestoreApp.OpenPath("reportdir")
	}
	if p == "" {
		return map[string]string{"error": "missing"}
	}
	if g.ctx != nil {
		runtime.BrowserOpenURL(g.ctx, fileURL(p))
		return map[string]string{"ok": "1"}
	}
	return g.RestoreApp.OpenPath(what)
}

func fileURL(p string) string {
	p = strings.ReplaceAll(p, "\\", "/")
	if !strings.HasPrefix(p, "/") && !strings.Contains(p, ":/") {
		p = "/" + p
	}
	if len(p) >= 2 && p[1] == ':' {
		// Windows drive: C:/...
		return "file:///" + p
	}
	return "file://" + p
}

func (g *DesktopApp) previewHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/preview" {
			http.NotFound(w, r)
			return
		}
		p := r.URL.Query().Get("p")
		if p == "" || !filepath.IsAbs(p) || !previewableImage(p) {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		dest, _ := g.Status()["dest"].(string)
		if dest == "" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		rel, err := filepath.Rel(filepath.Clean(dest), filepath.Clean(p))
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		http.ServeFile(w, r, p)
	})
}

func launchGUI(lang Lang, langExplicit bool) {
	app := &DesktopApp{
		RestoreApp:   NewRestoreApp(),
		defaultLang:  lang,
		langExplicit: langExplicit,
	}
	err := wails.Run(&options.App{
		Title:            "TakeoutRestore",
		Width:            920,
		Height:           740,
		MinWidth:         720,
		MinHeight:        560,
		Frameless:        true,
		CSSDragProperty:  "--wails-draggable",
		CSSDragValue:     "drag",
		BackgroundColour: &options.RGBA{R: 238, G: 242, B: 246, A: 255},
		AssetServer: &assetserver.Options{
			Assets:  guiAssets,
			Handler: app.previewHandler(),
		},
		OnStartup:     app.startup,
		OnShutdown:    app.shutdown,
		OnBeforeClose: app.beforeClose,
		Bind:          []interface{}{app},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
		},
		Linux: &linux.Options{
			Icon: appIcon,
		},
	})
	if err != nil {
		fmt.Println("GUI error:", err)
		time.Sleep(5 * time.Second)
	}
}
