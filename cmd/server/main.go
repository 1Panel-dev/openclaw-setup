package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"

	"openclaw-setup/internal/handlers"
)

func main() {
	addr := getenvDefault("SETUP_LISTEN_ADDR", "0.0.0.0:8188")
	composeDir := getenvDefault("OPENCLAW_COMPOSE_DIR", os.Getenv("MOLTBOT_COMPOSE_DIR"))
	containerName := getenvDefault("OPENCLAW_CONTAINER_NAME", os.Getenv("MOLTBOT_CONTAINER_NAME"))

	if len(os.Args) > 1 && os.Args[1] == "init" {
		if err := runInit(initOptions{composeDir: composeDir}); err != nil {
			log.Fatal(err)
		}
		log.Print("openclaw.json generated")
		return
	}

	if composeDir == "" {
		log.Fatal("OPENCLAW_COMPOSE_DIR is required")
	}

	configDir := filepath.Join(composeDir, "data", "conf")

	handler := handlers.NewServer(handlers.ServerConfig{
		ComposeDir:    composeDir,
		ConfigDir:     configDir,
		ContainerName: containerName,
		StaticDir:     "web/dist",
	})

	log.Printf("OpenClaw setup listening on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatal(err)
	}
}

func getenvDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
