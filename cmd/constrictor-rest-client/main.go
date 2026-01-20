package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/constrictor/constrictor-rest-client/internal/config"
	"github.com/gorilla/mux"
)

func main() {
	cfg := config.Load()

	r := mux.NewRouter()
	
	// API routes
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/health", handleHealth).Methods("GET")

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
