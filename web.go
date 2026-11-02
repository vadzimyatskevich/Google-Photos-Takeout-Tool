package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

//go:embed ui.html
var uiHTML []byte

func serveUI(port int, browser bool) {
	ln, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	if err != nil {
		fmt.Println("Не удалось открыть порт:", err)
		time.Sleep(10 * time.Second)
		return
	}
	url := "http://" + ln.Addr().String() + "/"
	prog := &Progress{}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(uiHTML)
	})
	writeJSON := func(w http.ResponseWriter, v any) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(v)
	}
	mux.HandleFunc("/api/defaults", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, defaultOptions()) })
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
		prog.mu.Lock()
		running := prog.Running
		if !running {
			prog.Running = true
		}
		prog.mu.Unlock()
		if running {
			writeJSON(w, map[string]string{"error": "уже выполняется"})
			return
		}
		go Run(o, prog)
		writeJSON(w, map[string]string{"ok": "1"})
	})
	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, prog.snapshot()) })
	mux.HandleFunc("/api/cancel", func(w http.ResponseWriter, r *http.Request) { prog.cancel.Store(true); writeJSON(w, "ok") })
	mux.HandleFunc("/api/pick", func(w http.ResponseWriter, r *http.Request) {
		p, err := pickFolder()
		res := map[string]string{"path": p}
		if err != nil {
			res["error"] = err.Error()
		}
		writeJSON(w, res)
	})
	serveFile := func(key string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			s := prog.snapshot()
			p, _ := s[key].(string)
			if p == "" {
				http.Error(w, "нет файла", 404)
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
	fmt.Println("TakeoutRestore — восстановление Google Фото из Takeout")
	fmt.Println("Интерфейс открыт в браузере:", url)
	fmt.Println("Не закрывайте это окно, пока идёт работа. Закрыть программу — закрыть это окно.")
	if browser {
		openURL(url)
	}
	http.Serve(ln, mux)
}
