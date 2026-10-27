// A small HTTP service in Go with private files on Object Storage: what you
// send is stored in the bucket bound to this service, and comes back through
// links that expire in 15 minutes.
package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"time"
)

//go:embed page.html
var pages embed.FS

var page = template.Must(template.New("page.html").Funcs(template.FuncMap{
	"size": func(n int64) string {
		if n >= 1<<20 {
			return fmt.Sprintf("%.1f MB", float64(n)/(1<<20))
		}
		return fmt.Sprintf("%d KB", max(1, (n+512)/1024))
	},
}).ParseFS(pages, "page.html"))

const maxUpload = 10 << 20

func main() {
	store, err := newStorage()
	if err != nil {
		log.Printf("storage: %v", err)
	}

	mux := http.NewServeMux()

	// The platform's health check: the process is up.
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	})

	mux.HandleFunc("GET /status.json", func(w http.ResponseWriter, r *http.Request) {
		c := store.check(r.Context())
		status := http.StatusOK
		if !c.OK {
			status = http.StatusServiceUnavailable
		}
		writeJSON(w, status, map[string]any{"ok": c.OK, "checks": map[string]check{"storage": c}})
	})

	mux.HandleFunc("GET /files", func(w http.ResponseWriter, r *http.Request) {
		files, err := store.list(r.Context())
		if err != nil {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"files": files})
	})

	// A file to the bucket. JSON for API calls, back to the page for the form.
	mux.HandleFunc("POST /files", func(w http.ResponseWriter, r *http.Request) {
		fromForm := strings.Contains(r.Header.Get("Accept"), "text/html")
		r.Body = http.MaxBytesReader(w, r.Body, maxUpload+1<<20)
		f, h, err := r.FormFile("file")
		if err != nil || h.Size == 0 {
			if fromForm {
				http.Redirect(w, r, "/?error=empty", http.StatusSeeOther)
				return
			}
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": `Send the file in the "file" field.`})
			return
		}
		defer f.Close()
		stored, err := store.put(r.Context(), h.Filename, h.Header.Get("Content-Type"), f, h.Size)
		if err != nil {
			log.Printf("upload failed: %v", err)
			if fromForm {
				http.Redirect(w, r, "/?error=failed", http.StatusSeeOther)
				return
			}
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "Object Storage did not take the file."})
			return
		}
		if fromForm {
			http.Redirect(w, r, "/?sent="+url.QueryEscape(h.Filename), http.StatusSeeOther)
			return
		}
		writeJSON(w, http.StatusCreated, stored)
	})

	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		lang, t := texts(r.Header.Get("Accept-Language"))
		c := store.check(r.Context())
		var files []file
		if c.OK {
			files, _ = store.list(r.Context())
		}
		var note, problem string
		if sent := r.URL.Query().Get("sent"); sent != "" {
			note = fmt.Sprintf(t["sent"], sent)
		} else if e := r.URL.Query().Get("error"); e == "empty" {
			problem = t["empty"]
		} else if e != "" {
			problem = t["failed"]
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		err := page.Execute(w, map[string]any{
			"Lang": lang, "T": t, "Storage": c, "Files": files, "Note": note, "Problem": problem,
			"Go": runtime.Version(),
		})
		if err != nil {
			log.Printf("page: %v", err)
		}
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	// 0.0.0.0, not localhost: the platform reaches the app from outside the container.
	srv := &http.Server{Addr: "0.0.0.0:" + port, Handler: mux, ReadHeaderTimeout: 10 * time.Second}
	log.Printf("listening on :%s", port)
	log.Fatal(srv.ListenAndServe())
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
