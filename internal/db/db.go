package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// ProbesDir returns the path to the probes directory (inside probeTool repo)
func ProbesDir() string {
	exe, _ := os.Executable()
	dir := filepath.Dir(exe)
	return filepath.Join(dir, "probes")
}

// DBPath returns the full path to the SQLite database
func DBPath() string {
	return filepath.Join(ProbesDir(), "probes.db")
}

// InitDB initializes the SQLite database and creates the necessary tables.
func InitDB(dbPath string) (*sql.DB, error) {
	// Create the directory if it doesn't exist
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Create table
	createTableSQL := `CREATE TABLE IF NOT EXISTS probes (
		id TEXT PRIMARY KEY,
		type TEXT NOT NULL,
		target TEXT NOT NULL,
		file_path TEXT,
		status TEXT DEFAULT 'running',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = db.Exec(createTableSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return db, nil
}

// InsertProbe inserts a new probe record into the database.
func InsertProbe(db *sql.DB, id, probeType, target, filePath string) error {
	query := `INSERT INTO probes (id, type, target, file_path, status) VALUES (?, ?, ?, ?, 'running')`
	_, err := db.Exec(query, id, probeType, target, filePath)
	return err
}

// UpdateProbeStatus updates the status of a probe.
func UpdateProbeStatus(db *sql.DB, id, status string) error {
	query := `UPDATE probes SET status = ? WHERE id = ?`
	_, err := db.Exec(query, status, id)
	return err
}

// GetProbe retrieves a single probe by ID.
func GetProbe(db *sql.DB, id string) (*Probe, error) {
	query := `SELECT id, type, target, file_path, status, created_at FROM probes WHERE id = ?`
	row := db.QueryRow(query, id)

	var probe Probe
	err := row.Scan(&probe.ID, &probe.Type, &probe.Target, &probe.FilePath, &probe.Status, &probe.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &probe, nil
}

// GetAllProbes retrieves all probes from the database.
func GetAllProbes(db *sql.DB) ([]Probe, error) {
	query := `SELECT id, type, target, file_path, status, created_at FROM probes ORDER BY created_at DESC`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var probes []Probe
	for rows.Next() {
		var probe Probe
		err := rows.Scan(&probe.ID, &probe.Type, &probe.Target, &probe.FilePath, &probe.Status, &probe.CreatedAt)
		if err != nil {
			return nil, err
		}
		probes = append(probes, probe)
	}

	return probes, nil
}

// Probe represents a probe record.
type Probe struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Target    string `json:"target"`
	FilePath  string `json:"file_path"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}
