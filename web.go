package main

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"path/filepath"
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
	app := NewRestoreApp()
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
	mux.HandleFunc("/api/defaults", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, app.Defaults()) })
	mux.HandleFunc("/api/info", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, app.Info()) })
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
		writeJSON(w, app.Start(o))
	})
	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, app.Status()) })
	mux.HandleFunc("/api/cancel", func(w http.ResponseWriter, r *http.Request) {
		app.Cancel()
		writeJSON(w, "ok")
	})
	mux.HandleFunc("/api/pause", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST", 405)
			return
		}
		app.Pause()
		writeJSON(w, "ok")
	})
	mux.HandleFunc("/api/resume", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST", 405)
			return
		}
		app.Resume()
		writeJSON(w, "ok")
	})
	mux.HandleFunc("/api/pick", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, app.PickFolder(r.URL.Query().Get("lang")))
	})
	serveFile := func(key string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			s := app.Status()
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
		writeJSON(w, app.OpenPath(r.URL.Query().Get("what")))
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
