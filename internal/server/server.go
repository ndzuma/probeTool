package server

import (
	"audit-tool/internal/db"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

var database *sql.DB

// StartServer starts the HTTP server on localhost:3030.
func StartServer(db *sql.DB) {
	database = db

	mux := http.NewServeMux()
	mux.HandleFunc("/api/audits", handleAudits)
	mux.HandleFunc("/api/audits/", handleAuditDetail)

	fmt.Println("🌐 Server starting on http://localhost:3030")
	err := http.ListenAndServe(":3030", mux)
	if err != nil {
		fmt.Printf("❌ Server error: %v\n", err)
	}
}

func handleAudits(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	audits, err := db.GetAllAudits(database)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching audits: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(audits)
}

func handleAuditDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path /api/audits/{id}
	path := strings.TrimPrefix(r.URL.Path, "/api/audits/")
	if path == "" || path == r.URL.Path {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	audit, err := db.GetAudit(database, path)
	if err != nil {
		http.Error(w, fmt.Sprintf("Audit not found: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(audit)
}
