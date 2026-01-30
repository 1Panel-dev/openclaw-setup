package handlers

import (
	"net/http"
	"os"
	"path/filepath"
)

type ServerConfig struct {
	ComposeDir    string
	ConfigDir     string
	ContainerName string
	StaticDir     string
}

type Server struct {
	mux *http.ServeMux
}

func NewServer(cfg ServerConfig) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/api/config", NewConfigHandler(cfg))
	mux.Handle("/api/models", NewModelsHandler())

	if cfg.StaticDir != "" {
		fileServer := http.FileServer(http.Dir(cfg.StaticDir))
		mux.Handle("/assets/", fileServer)
		mux.Handle("/", spaHandler(cfg.StaticDir, "index.html"))
	}

	return mux
}

func spaHandler(staticDir, indexFile string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Join(staticDir, filepath.Clean(r.URL.Path))
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			http.ServeFile(w, r, path)
			return
		}
		http.ServeFile(w, r, filepath.Join(staticDir, indexFile))
	})
}
