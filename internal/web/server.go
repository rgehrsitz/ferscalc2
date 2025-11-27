package web

import (
	"embed"
	"fmt"
	"io/fs"
	"net/http"
)

//go:embed static/*
var staticFiles embed.FS

// StartServer starts the web server on the specified port
func StartServer(port int) error {
	mux := http.NewServeMux()

	// Static files
	staticFS, err := fs.Sub(staticFiles, "static")
	if err != nil {
		return err
	}
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(staticFS))))

	// Main page
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		content, _ := staticFiles.ReadFile("static/index.html")
		w.Header().Set("Content-Type", "text/html")
		w.Write(content)
	})

	// API
	mux.HandleFunc("/api/calculate", HandleCalculate)

	addr := fmt.Sprintf(":%d", port)
	return http.ListenAndServe(addr, mux)
}
