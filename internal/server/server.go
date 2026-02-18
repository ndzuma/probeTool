package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/ndzuma/probeTool/internal/db"
)

var database *sql.DB

// StartServer starts the HTTP server on localhost:3030.
func StartServer(dbConn *sql.DB) {
	database = dbConn

	mux := http.NewServeMux()
	mux.HandleFunc("/api/probes", handleProbes)
	mux.HandleFunc("/api/probes/", handleProbeDetail)

	fmt.Println("🌐 Server starting on http://localhost:3030")
	err := http.ListenAndServe(":3030", mux)
	if err != nil {
		fmt.Printf("❌ Server error: %v\n", err)
	}
}

func handleProbes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	probes, err := db.GetAllProbes(database)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error fetching probes: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(probes)
}

func handleProbeDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from path /api/probes/{id}
	path := strings.TrimPrefix(r.URL.Path, "/api/probes/")
	if path == "" || path == r.URL.Path {
		http.Error(w, "Invalid path", http.StatusBadRequest)
		return
	}

	probe, err := db.GetProbe(database, path)
	if err != nil {
		http.Error(w, fmt.Sprintf("Probe not found: %v", err), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(probe)
}
