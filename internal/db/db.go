package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"github.com/ndzuma/probeTool/internal/paths"
)

// ProbesDir returns the path to the probes directory
// Deprecated: Use paths.GetProbesDir() instead
func ProbesDir() string {
	return paths.GetProbesDir()
}

// DBPath returns the full path to the SQLite database
// Deprecated: Use paths.GetDBPath() instead
func DBPath() string {
	return paths.GetDBPath()
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

	InitDB := func() error {
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
			return fmt.Errorf("failed to create probes table: %w", err)
		}

		findingsTableSQL := `CREATE TABLE IF NOT EXISTS findings (
			id TEXT PRIMARY KEY,
			probe_id TEXT NOT NULL,
			text TEXT NOT NULL,
			severity TEXT DEFAULT 'info',
			completed INTEGER DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (probe_id) REFERENCES probes(id) ON DELETE CASCADE
		);`

		_, err = db.Exec(findingsTableSQL)
		if err != nil {
			return fmt.Errorf("failed to create findings table: %w", err)
		}

		return nil
	}

	if err := InitDB(); err != nil {
		return nil, err
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

type Probe struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Target    string `json:"target"`
	FilePath  string `json:"file_path"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}

type Finding struct {
	ID        string `json:"id"`
	ProbeID   string `json:"probe_id"`
	Text      string `json:"text"`
	Severity  string `json:"severity"`
	Completed bool   `json:"completed"`
	CreatedAt string `json:"created_at"`
}

func InsertFinding(db *sql.DB, id, probeID, text, severity string) error {
	query := `INSERT INTO findings (id, probe_id, text, severity, completed) VALUES (?, ?, ?, ?, 0)`
	_, err := db.Exec(query, id, probeID, text, severity)
	return err
}

func GetFindingsByProbe(db *sql.DB, probeID string) ([]Finding, error) {
	query := `SELECT id, probe_id, text, severity, completed, created_at FROM findings WHERE probe_id = ? ORDER BY created_at ASC`
	rows, err := db.Query(query, probeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var findings []Finding
	for rows.Next() {
		var f Finding
		var completed int
		err := rows.Scan(&f.ID, &f.ProbeID, &f.Text, &f.Severity, &completed, &f.CreatedAt)
		if err != nil {
			return nil, err
		}
		f.Completed = completed == 1
		findings = append(findings, f)
	}

	return findings, nil
}

func ToggleFinding(db *sql.DB, id string) (bool, error) {
	var completed int
	err := db.QueryRow(`SELECT completed FROM findings WHERE id = ?`, id).Scan(&completed)
	if err != nil {
		return false, err
	}

	newCompleted := 1
	if completed == 1 {
		newCompleted = 0
	}

	_, err = db.Exec(`UPDATE findings SET completed = ? WHERE id = ?`, newCompleted, id)
	return newCompleted == 1, err
}

func DeleteFinding(db *sql.DB, id string) error {
	_, err := db.Exec(`DELETE FROM findings WHERE id = ?`, id)
	return err
}

func GetFinding(db *sql.DB, id string) (*Finding, error) {
	query := `SELECT id, probe_id, text, severity, completed, created_at FROM findings WHERE id = ?`
	row := db.QueryRow(query, id)

	var f Finding
	var completed int
	err := row.Scan(&f.ID, &f.ProbeID, &f.Text, &f.Severity, &completed, &f.CreatedAt)
	if err != nil {
		return nil, err
	}
	f.Completed = completed == 1

	return &f, nil
}
