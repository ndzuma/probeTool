package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

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
	createTableSQL := `CREATE TABLE IF NOT EXISTS audits (
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

// InsertAudit inserts a new audit record into the database.
func InsertAudit(db *sql.DB, id, auditType, target, filePath string) error {
	query := `INSERT INTO audits (id, type, target, file_path, status) VALUES (?, ?, ?, ?, 'running')`
	_, err := db.Exec(query, id, auditType, target, filePath)
	return err
}

// UpdateAuditStatus updates the status of an audit.
func UpdateAuditStatus(db *sql.DB, id, status string) error {
	query := `UPDATE audits SET status = ? WHERE id = ?`
	_, err := db.Exec(query, status, id)
	return err
}

// GetAudit retrieves a single audit by ID.
func GetAudit(db *sql.DB, id string) (*Audit, error) {
	query := `SELECT id, type, target, file_path, status, created_at FROM audits WHERE id = ?`
	row := db.QueryRow(query, id)

	var audit Audit
	err := row.Scan(&audit.ID, &audit.Type, &audit.Target, &audit.FilePath, &audit.Status, &audit.CreatedAt)
	if err != nil {
		return nil, err
	}

	return &audit, nil
}

// GetAllAudits retrieves all audits from the database.
func GetAllAudits(db *sql.DB) ([]Audit, error) {
	query := `SELECT id, type, target, file_path, status, created_at FROM audits ORDER BY created_at DESC`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var audits []Audit
	for rows.Next() {
		var audit Audit
		err := rows.Scan(&audit.ID, &audit.Type, &audit.Target, &audit.FilePath, &audit.Status, &audit.CreatedAt)
		if err != nil {
			return nil, err
		}
		audits = append(audits, audit)
	}

	return audits, nil
}

// Audit represents an audit record.
type Audit struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Target    string `json:"target"`
	FilePath  string `json:"file_path"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}
