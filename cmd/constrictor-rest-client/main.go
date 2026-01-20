package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/constrictor/constrictor-rest-client/internal/config"
	"github.com/constrictor/constrictor-rest-client/internal/executor"
	"github.com/constrictor/constrictor-rest-client/internal/httpapi"
	"github.com/constrictor/constrictor-rest-client/internal/storage"
	"github.com/gorilla/mux"
)

func main() {
	cfg := config.Load()

	// Initialize dependencies
	store := storage.NewFileStore(cfg.WorkspacePath())
	exec := executor.NewHTTPExecutor(executor.Config{
		Timeout:     time.Duration(cfg.Timeout) * time.Second,
		MaxBodySize: cfg.MaxBodySize,
	})
	handlers := httpapi.NewHandlers(store, exec)

	r := mux.NewRouter()
	
	// API routes
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/health", handleHealth).Methods("GET")
	httpapi.SetupRoutes(r, handlers)

	// Serve static files from web/ directory (will be added later)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./web/dist")))

	log.Printf("Server starting on port %s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatal(err)
	}
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"ok": true})
}
