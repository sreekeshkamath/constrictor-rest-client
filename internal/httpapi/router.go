package httpapi

import (
	"github.com/gorilla/mux"
)

// SetupRoutes configures API routes
func SetupRoutes(r *mux.Router, handlers *Handlers) {
	api := r.PathPrefix("/api").Subrouter()
	api.HandleFunc("/workspace", handlers.HandleGetWorkspace).Methods("GET")
	api.HandleFunc("/workspace", handlers.HandlePutWorkspace).Methods("PUT")
	api.HandleFunc("/execute", handlers.HandleExecute).Methods("POST")

	// Google Drive backup endpoints
	api.HandleFunc("/gdrive/backup", handlers.HandleBackupToGDrive).Methods("POST")
	api.HandleFunc("/gdrive/backups", handlers.HandleListGDriveBackups).Methods("GET")
	api.HandleFunc("/gdrive/restore", handlers.HandleRestoreFromGDrive).Methods("POST")
}
