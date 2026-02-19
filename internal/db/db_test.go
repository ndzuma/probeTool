package db

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func TestGetAppDir(t *testing.T) {
	dir := ProbesDir()
	if dir == "" {
		t.Error("ProbesDir() returned empty string")
	}
}

func TestProbesDir(t *testing.T) {
	probesDir := ProbesDir()
	if probesDir == "" {
		t.Error("ProbesDir() returned empty string")
	}
	if !filepath.IsAbs(probesDir) {
		t.Error("ProbesDir() should return an absolute path")
	}
}

func TestDBPath(t *testing.T) {
	dbPath := DBPath()
	if dbPath == "" {
		t.Error("DBPath() returned empty string")
	}
	if !filepath.IsAbs(dbPath) {
		t.Error("DBPath() should return an absolute path")
	}
	if filepath.Ext(dbPath) != ".db" {
		t.Error("DBPath() should have .db extension")
	}
}

func TestInitDB(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB() failed: %v", err)
	}
	defer db.Close()

	// Verify the database file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Database file was not created")
	}

	// Test that tables exist by trying to insert a probe
	testID := "test-probe-" + time.Now().Format("20060102150405")
	err = InsertProbe(db, testID, "security", "/tmp/test", "/tmp/test.md")
	if err != nil {
		t.Errorf("Failed to insert test probe: %v", err)
	}
}

func TestInsertProbe(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB() failed: %v", err)
	}
	defer db.Close()

	testID := "test-insert-" + time.Now().Format("20060102150405")
	err = InsertProbe(db, testID, "security", "/tmp/test", "/tmp/test.md")
	if err != nil {
		t.Errorf("InsertProbe() failed: %v", err)
	}

	// Verify probe was inserted
	probe, err := GetProbe(db, testID)
	if err != nil {
		t.Errorf("Failed to get inserted probe: %v", err)
	}
	if probe == nil {
		t.Error("Inserted probe is nil")
	} else {
		if probe.ID != testID {
			t.Errorf("Probe ID mismatch: got %s, want %s", probe.ID, testID)
		}
		if probe.Status != "running" {
			t.Errorf("Probe status should be 'running', got %s", probe.Status)
		}
	}
}

func TestGetProbe(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB() failed: %v", err)
	}
	defer db.Close()

	// Test getting non-existent probe
	_, err = GetProbe(db, "non-existent-id")
	if err == nil {
		t.Error("GetProbe() should return error for non-existent probe")
	}

	// Insert and retrieve a probe
	testID := "test-get-" + time.Now().Format("20060102150405")
	err = InsertProbe(db, testID, "security", "/tmp/test", "/tmp/test.md")
	if err != nil {
		t.Fatalf("InsertProbe() failed: %v", err)
	}

	probe, err := GetProbe(db, testID)
	if err != nil {
		t.Errorf("GetProbe() failed: %v", err)
	}
	if probe == nil {
		t.Fatal("GetProbe() returned nil")
	}
	if probe.ID != testID {
		t.Errorf("Probe ID mismatch: got %s, want %s", probe.ID, testID)
	}
}

func TestGetAllProbes(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB() failed: %v", err)
	}
	defer db.Close()

	// Get all probes from empty database
	probes, err := GetAllProbes(db)
	if err != nil {
		t.Errorf("GetAllProbes() failed on empty db: %v", err)
	}
	if len(probes) != 0 {
		t.Errorf("GetAllProbes() should return empty slice for empty db, got %d items", len(probes))
	}

	// Insert multiple probes
	for i := 0; i < 3; i++ {
		testID := "test-all-" + time.Now().Format("20060102150405") + "-" + string(rune('0'+i))
		err = InsertProbe(db, testID, "security", "/tmp/test", "/tmp/test.md")
		if err != nil {
			t.Fatalf("InsertProbe() failed: %v", err)
		}
	}

	// Get all probes
	probes, err = GetAllProbes(db)
	if err != nil {
		t.Errorf("GetAllProbes() failed: %v", err)
	}
	if len(probes) < 3 {
		t.Errorf("GetAllProbes() should return at least 3 probes, got %d", len(probes))
	}
}

func TestUpdateProbeStatus(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB() failed: %v", err)
	}
	defer db.Close()

	testID := "test-update-" + time.Now().Format("20060102150405")
	err = InsertProbe(db, testID, "security", "/tmp/test", "/tmp/test.md")
	if err != nil {
		t.Fatalf("InsertProbe() failed: %v", err)
	}

	// Update status
	err = UpdateProbeStatus(db, testID, "completed")
	if err != nil {
		t.Errorf("UpdateProbeStatus() failed: %v", err)
	}

	// Verify status was updated
	probe, err := GetProbe(db, testID)
	if err != nil {
		t.Errorf("GetProbe() failed: %v", err)
	}
	if probe.Status != "completed" {
		t.Errorf("Probe status not updated: got %s, want completed", probe.Status)
	}
}

func TestInsertFinding(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB() failed: %v", err)
	}
	defer db.Close()

	// First insert a probe
	probeID := "test-finding-probe-" + time.Now().Format("20060102150405")
	err = InsertProbe(db, probeID, "security", "/tmp/test", "/tmp/test.md")
	if err != nil {
		t.Fatalf("InsertProbe() failed: %v", err)
	}

	// Insert a finding
	findingID := "finding-" + time.Now().Format("20060102150405")
	err = InsertFinding(db, findingID, probeID, "Test finding", "high")
	if err != nil {
		t.Errorf("InsertFinding() failed: %v", err)
	}

	// Verify finding was inserted
	findings, err := GetFindingsByProbe(db, probeID)
	if err != nil {
		t.Errorf("GetFindingsByProbe() failed: %v", err)
	}
	if len(findings) != 1 {
		t.Errorf("Expected 1 finding, got %d", len(findings))
	}
}

func TestGetFindingsByProbe(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB() failed: %v", err)
	}
	defer db.Close()

	probeID := "test-finding-" + time.Now().Format("20060102150405")
	err = InsertProbe(db, probeID, "security", "/tmp/test", "/tmp/test.md")
	if err != nil {
		t.Fatalf("InsertProbe() failed: %v", err)
	}

	// Insert multiple findings
	severities := []string{"critical", "high", "medium", "low"}
	for i, severity := range severities {
		findingID := "finding-" + time.Now().Format("20060102150405") + "-" + string(rune('0'+i))
		err = InsertFinding(db, findingID, probeID, "Test finding "+severity, severity)
		if err != nil {
			t.Fatalf("InsertFinding() failed: %v", err)
		}
	}

	// Get findings
	findings, err := GetFindingsByProbe(db, probeID)
	if err != nil {
		t.Errorf("GetFindingsByProbe() failed: %v", err)
	}
	if len(findings) != len(severities) {
		t.Errorf("Expected %d findings, got %d", len(severities), len(findings))
	}
}

func TestToggleFinding(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB() failed: %v", err)
	}
	defer db.Close()

	probeID := "test-toggle-" + time.Now().Format("20060102150405")
	err = InsertProbe(db, probeID, "security", "/tmp/test", "/tmp/test.md")
	if err != nil {
		t.Fatalf("InsertProbe() failed: %v", err)
	}

	findingID := "toggle-finding-" + time.Now().Format("20060102150405")
	err = InsertFinding(db, findingID, probeID, "Test finding", "high")
	if err != nil {
		t.Fatalf("InsertFinding() failed: %v", err)
	}

	// Toggle to completed
	completed, err := ToggleFinding(db, findingID)
	if err != nil {
		t.Errorf("ToggleFinding() failed: %v", err)
	}
	if !completed {
		t.Error("ToggleFinding() should return true after first toggle")
	}

	// Toggle back to not completed
	completed, err = ToggleFinding(db, findingID)
	if err != nil {
		t.Errorf("ToggleFinding() failed on second toggle: %v", err)
	}
	if completed {
		t.Error("ToggleFinding() should return false after second toggle")
	}
}

func TestDeleteFinding(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB() failed: %v", err)
	}
	defer db.Close()

	probeID := "test-delete-" + time.Now().Format("20060102150405")
	err = InsertProbe(db, probeID, "security", "/tmp/test", "/tmp/test.md")
	if err != nil {
		t.Fatalf("InsertProbe() failed: %v", err)
	}

	findingID := "delete-finding-" + time.Now().Format("20060102150405")
	err = InsertFinding(db, findingID, probeID, "Test finding to delete", "high")
	if err != nil {
		t.Fatalf("InsertFinding() failed: %v", err)
	}

	// Delete the finding
	err = DeleteFinding(db, findingID)
	if err != nil {
		t.Errorf("DeleteFinding() failed: %v", err)
	}

	// Verify finding was deleted
	findings, err := GetFindingsByProbe(db, probeID)
	if err != nil {
		t.Errorf("GetFindingsByProbe() failed: %v", err)
	}
	if len(findings) != 0 {
		t.Errorf("Expected 0 findings after deletion, got %d", len(findings))
	}
}

func TestConcurrentAccess(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := InitDB(dbPath)
	if err != nil {
		t.Fatalf("InitDB() failed: %v", err)
	}
	defer db.Close()

	// Insert a probe
	probeID := "concurrent-test-" + time.Now().Format("20060102150405")
	err = InsertProbe(db, probeID, "security", "/tmp/test", "/tmp/test.md")
	if err != nil {
		t.Fatalf("InsertProbe() failed: %v", err)
	}

	// Concurrent inserts
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(idx int) {
			findingID := "concurrent-finding-" + time.Now().Format("20060102150405") + "-" + string(rune('0'+idx))
			err := InsertFinding(db, findingID, probeID, "Concurrent finding", "medium")
			if err != nil {
				t.Errorf("Concurrent InsertFinding() failed: %v", err)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Verify all findings were inserted
	findings, err := GetFindingsByProbe(db, probeID)
	if err != nil {
		t.Errorf("GetFindingsByProbe() failed: %v", err)
	}
	if len(findings) != 10 {
		t.Errorf("Expected 10 findings, got %d", len(findings))
	}
}
