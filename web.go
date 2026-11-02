package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

//go:embed ui.html
var uiHTML []byte

func serveUI(port int, browser bool, lang Lang, langExplicit bool) {
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		fmt.Println(lang.T("web.port", err))
		time.Sleep(10 * time.Second)
		return
	}
	url := "http://" + ln.Addr().String() + "/"
	prog := &Progress{}
	mux := http.NewServeMux()
	page := bytes.ReplaceAll(uiHTML, []byte("{{DEFAULT_LANG}}"), []byte(lang))
	exp := "false"
	if langExplicit {
		exp = "true"
	}
	page = bytes.ReplaceAll(page, []byte("{{LANG_EXPLICIT}}"), []byte(exp))
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(page)
	})
	writeJSON := func(w http.ResponseWriter, v any) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(v)
	}
	mux.HandleFunc("/api/defaults", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, defaultOptions()) })
	mux.HandleFunc("/api/info", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]any{"pick": folderPickerAvailable(), "os": runtime.GOOS})
	})
	mux.HandleFunc("/api/start", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST", 405)
			return
		}
		var o Options
		if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
			writeJSON(w, map[string]string{"error": err.Error()})
			return
		}
		L := NormalizeLang(o.Lang)
		o.Lang = string(L)
		prog.mu.Lock()
		running := prog.Running
		if !running {
			prog.Running = true
		}
		prog.mu.Unlock()
		if running {
			writeJSON(w, map[string]string{"error": L.T("web.running")})
			return
		}
		prog.cancel.Store(false)
		go Run(o, prog)
		writeJSON(w, map[string]string{"ok": "1"})
	})
	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, prog.snapshot()) })
	mux.HandleFunc("/api/cancel", func(w http.ResponseWriter, r *http.Request) {
		prog.cancel.Store(true)
		prog.set(func(p *Progress) {
			if p.Running {
				p.Phase = NormalizeLang(p.Lang).T("phase.stopping")
			}
		})
		writeJSON(w, "ok")
	})
	mux.HandleFunc("/api/pick", func(w http.ResponseWriter, r *http.Request) {
		L := NormalizeLang(r.URL.Query().Get("lang"))
		p, err := pickFolder(L.T("pick.title"))
		res := map[string]string{"path": p}
		if err != nil {
			if errors.Is(err, errNoPicker) {
				res["error"] = L.T("pick.err")
			} else {
				res["error"] = L.T("pick.errdetail", err)
			}
		}
		writeJSON(w, res)
	})
	serveFile := func(key string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			s := prog.snapshot()
			p, _ := s[key].(string)
			if p == "" {
				L := LangEN
				if ls, ok := s["lang"].(string); ok && ls != "" {
					L = NormalizeLang(ls)
				}
				http.Error(w, L.T("web.nofile"), 404)
				return
			}
			if filepath.Ext(p) == ".txt" {
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			}
			http.ServeFile(w, r, p)
		}
	}
	mux.HandleFunc("/report", serveFile("report"))
	mux.HandleFunc("/log", serveFile("log"))
	mux.HandleFunc("/api/open", func(w http.ResponseWriter, r *http.Request) {
		s := prog.snapshot()
		what := r.URL.Query().Get("what")
		var p string
		switch what {
		case "dest":
			p, _ = s["dest"].(string)
		case "reportdir":
			rp, _ := s["report"].(string)
			p = filepath.Dir(rp)
		}
		if p != "" {
			if _, err := os.Stat(p); err == nil {
				openFolder(p)
			}
		}
		writeJSON(w, "ok")
	})
	fmt.Println(lang.T("web.banner"))
	fmt.Println(lang.T("web.url", url))
	if browser {
		if err := openURL(url); err != nil {
			fmt.Println(lang.T("web.openfail"))
		}
	}
	fmt.Println(lang.T("web.keepopen"))
	http.Serve(ln, mux)
}
