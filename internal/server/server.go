package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/ndzuma/probeTool/internal/config"
	"github.com/ndzuma/probeTool/internal/db"
)

var database *sql.DB

// StartServer starts the API server on localhost:3030.
// The Next.js frontend runs separately on :3000 and proxies API calls here.
func StartServer(dbConn *sql.DB) {
	database = dbConn

	mux := http.NewServeMux()

	mux.HandleFunc("/api/probes", cors(handleProbes))
	mux.HandleFunc("/api/probes/", cors(handleProbeDetail))
	mux.HandleFunc("/api/findings/", cors(handleFindings))
	mux.HandleFunc("/api/config", cors(handleConfig))
	mux.HandleFunc("/api/file-tree/", cors(handleFileTree))

	fmt.Println("🌐 API server starting on http://localhost:3030")
	fmt.Println("📡 Frontend: cd web && npm run dev (http://localhost:3000)")
	err := http.ListenAndServe(":3030", mux)
	if err != nil {
		fmt.Printf("❌ Server error: %v\n", err)
	}
}

func RegisterRoutes(mux *http.ServeMux, dbConn *sql.DB) {
	database = dbConn

	mux.HandleFunc("/api/probes", cors(handleProbes))
	mux.HandleFunc("/api/probes/", cors(handleProbeDetail))
	mux.HandleFunc("/api/findings/", cors(handleFindings))
	mux.HandleFunc("/api/config", cors(handleConfig))
	mux.HandleFunc("/api/file-tree/", cors(handleFileTree))
}

// ─── Middleware ──────────────────────────────────────────────────────────────

func cors(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, r)
	}
}

// ─── Helpers ────────────────────────────────────────────────────────────────

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// ─── GET /api/probes ────────────────────────────────────────────────────────

func handleProbes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	probes, err := db.GetAllProbes(database)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Error fetching probes: %v", err))
		return
	}

	if probes == nil {
		probes = []db.Probe{}
	}

	writeJSON(w, http.StatusOK, probes)
}

// ─── GET /api/probes/{id}  ·  GET /api/probes/{id}/content ──────────────────

func handleProbeDetail(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/probes/")
	if path == "" || path == r.URL.Path {
		writeError(w, http.StatusBadRequest, "Invalid path")
		return
	}

	parts := strings.SplitN(path, "/", 2)
	probeID := parts[0]

	// Sub-route: /api/probes/{id}/content
	if len(parts) > 1 && parts[1] == "content" {
		handleProbeContent(w, r, probeID)
		return
	}

	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	probe, err := db.GetProbe(database, probeID)
	if err != nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("Probe not found: %v", err))
		return
	}

	response := map[string]interface{}{
		"id":         probe.ID,
		"type":       probe.Type,
		"target":     probe.Target,
		"file_path":  probe.FilePath,
		"status":     probe.Status,
		"created_at": probe.CreatedAt,
	}

	// Attach markdown content if the file exists
	if probe.FilePath != "" {
		content, err := os.ReadFile(probe.FilePath)
		if err == nil {
			response["content"] = string(content)
		}
	}

	writeJSON(w, http.StatusOK, response)
}

func handleProbeContent(w http.ResponseWriter, r *http.Request, probeID string) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	probe, err := db.GetProbe(database, probeID)
	if err != nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("Probe not found: %v", err))
		return
	}

	if probe.FilePath == "" {
		writeError(w, http.StatusNotFound, "No file path for this probe")
		return
	}

	content, err := os.ReadFile(probe.FilePath)
	if err != nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("Could not read probe file: %v", err))
		return
	}

	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(content)
}

// ─── PATCH /api/findings/{id} ───────────────────────────────────────────────

func handleFindings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	findingID := strings.TrimPrefix(r.URL.Path, "/api/findings/")
	if findingID == "" || findingID == r.URL.Path {
		writeError(w, http.StatusBadRequest, "Invalid finding ID")
		return
	}

	// Placeholder — ready for when findings are stored in the DB
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"id":        findingID,
		"completed": true,
		"message":   "Finding toggled",
	})
}

// ─── GET/PUT /api/config ────────────────────────────────────────────────────

func handleConfig(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		handleGetConfig(w)
	case http.MethodPut:
		handleUpdateConfig(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func handleGetConfig(w http.ResponseWriter) {
	cfg, err := config.Load()
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Error loading config: %v", err))
		return
	}
	writeJSON(w, http.StatusOK, cfg)
}

func handleUpdateConfig(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Could not read request body")
		return
	}
	defer r.Body.Close()

	var cfg config.Config
	if err := json.Unmarshal(body, &cfg); err != nil {
		writeError(w, http.StatusBadRequest, fmt.Sprintf("Invalid JSON: %v", err))
		return
	}

	if cfg.Providers == nil {
		cfg.Providers = make(map[string]config.Provider)
	}

	if err := cfg.Save(); err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Error saving config: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, cfg)
}

// ─── GET /api/file-tree/{id} ────────────────────────────────────────────────

func handleFileTree(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	probeID := strings.TrimPrefix(r.URL.Path, "/api/file-tree/")
	if probeID == "" || probeID == r.URL.Path {
		writeError(w, http.StatusBadRequest, "Invalid probe ID")
		return
	}

	probe, err := db.GetProbe(database, probeID)
	if err != nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("Probe not found: %v", err))
		return
	}

	targetDir := probe.Target
	if info, err := os.Stat(targetDir); err != nil || !info.IsDir() {
		writeJSON(w, http.StatusOK, []interface{}{})
		return
	}

	type FileNode struct {
		Name     string      `json:"name"`
		Path     string      `json:"path"`
		Type     string      `json:"type"`
		Children []*FileNode `json:"children,omitempty"`
	}

	var walkDir func(dir string, depth int) []*FileNode
	walkDir = func(dir string, depth int) []*FileNode {
		if depth > 3 {
			return nil
		}

		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil
		}

		var nodes []*FileNode
		for _, entry := range entries {
			name := entry.Name()
			if strings.HasPrefix(name, ".") || name == "node_modules" || name == "__pycache__" || name == "vendor" {
				continue
			}

			relPath, _ := filepath.Rel(targetDir, filepath.Join(dir, name))
			node := &FileNode{
				Name: name,
				Path: relPath,
			}

			if entry.IsDir() {
				node.Type = "directory"
				node.Children = walkDir(filepath.Join(dir, name), depth+1)
			} else {
				node.Type = "file"
			}

			nodes = append(nodes, node)
		}
		return nodes
	}

	tree := walkDir(targetDir, 0)
	if tree == nil {
		tree = []*FileNode{}
	}

	writeJSON(w, http.StatusOK, tree)
}
