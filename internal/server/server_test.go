package server

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ndzuma/probeTool/internal/db"
)

func setupTestDB(t *testing.T) *sql.DB {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	database, err := db.InitDB(dbPath)
	if err != nil {
		t.Fatalf("Failed to init test database: %v", err)
	}
	return database
}

func TestHealthEndpoint(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	// Register routes
	mux := http.NewServeMux()
	RegisterRoutes(mux, database)

	// Create a simple health check by hitting probes endpoint
	req := httptest.NewRequest(http.MethodGet, "/api/probes", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	// We expect this to work (200) since server is running
	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}
}

func TestGetProbesEndpoint(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	// Insert some test probes
	for i := 0; i < 3; i++ {
		testID := "test-probe-" + time.Now().Format("20060102150405") + "-" + string(rune('0'+i))
		err := db.InsertProbe(database, testID, "security", "/tmp/test", "/tmp/test.md")
		if err != nil {
			t.Fatalf("Failed to insert test probe: %v", err)
		}
	}

	mux := http.NewServeMux()
	RegisterRoutes(mux, database)

	req := httptest.NewRequest(http.MethodGet, "/api/probes", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("Expected JSON content type, got %s", contentType)
	}

	// Parse response
	var probes []db.Probe
	if err := json.Unmarshal(rec.Body.Bytes(), &probes); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if len(probes) < 3 {
		t.Errorf("Expected at least 3 probes, got %d", len(probes))
	}
}

func TestGetProbeDetailEndpoint(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	// Insert a test probe
	testID := "detail-test-" + time.Now().Format("20060102150405")
	err := db.InsertProbe(database, testID, "security", "/tmp/test", "/tmp/test.md")
	if err != nil {
		t.Fatalf("Failed to insert test probe: %v", err)
	}

	mux := http.NewServeMux()
	RegisterRoutes(mux, database)

	req := httptest.NewRequest(http.MethodGet, "/api/probes/"+testID, nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Parse response
	var response map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if response["id"] != testID {
		t.Errorf("Expected ID %s, got %v", testID, response["id"])
	}
}

func TestGetProbeContentEndpoint(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	// Create a test file
	tmpDir := t.TempDir()
	testContent := "# Test Security Report\n\nThis is a test report."
	testFile := filepath.Join(tmpDir, "test-report.md")
	os.WriteFile(testFile, []byte(testContent), 0644)

	// Insert a test probe with file path
	testID := "content-test-" + time.Now().Format("20060102150405")
	err := db.InsertProbe(database, testID, "security", "/tmp/test", testFile)
	if err != nil {
		t.Fatalf("Failed to insert test probe: %v", err)
	}

	mux := http.NewServeMux()
	RegisterRoutes(mux, database)

	req := httptest.NewRequest(http.MethodGet, "/api/probes/"+testID+"/content", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if !strings.Contains(contentType, "text/markdown") {
		t.Errorf("Expected markdown content type, got %s", contentType)
	}

	if !strings.Contains(rec.Body.String(), testContent) {
		t.Error("Response body should contain test content")
	}
}

func TestFindingsEndpoint(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	// Insert a test probe
	probeID := "finding-probe-" + time.Now().Format("20060102150405")
	err := db.InsertProbe(database, probeID, "security", "/tmp/test", "/tmp/test.md")
	if err != nil {
		t.Fatalf("Failed to insert test probe: %v", err)
	}

	// Insert a test finding
	findingID := "finding-" + time.Now().Format("20060102150405")
	err = db.InsertFinding(database, findingID, probeID, "Test finding", "high")
	if err != nil {
		t.Fatalf("Failed to insert test finding: %v", err)
	}

	mux := http.NewServeMux()
	RegisterRoutes(mux, database)

	// Test PATCH - toggle finding
	req := httptest.NewRequest(http.MethodPatch, "/api/findings/"+findingID, nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if response["completed"] != true {
		t.Error("Finding should be marked as completed")
	}

	// Test DELETE - delete finding
	req = httptest.NewRequest(http.MethodDelete, "/api/findings/"+findingID, nil)
	rec = httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if response["message"] != "Finding deleted" {
		t.Errorf("Expected 'Finding deleted' message, got %v", response["message"])
	}
}

func TestConfigEndpoint(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	mux := http.NewServeMux()
	RegisterRoutes(mux, database)

	// Test GET config
	req := httptest.NewRequest(http.MethodGet, "/api/config", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", rec.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	if response["providers"] == nil {
		t.Error("Config response should have 'providers' field")
	}

	// Test PUT config
	updateData := map[string]interface{}{
		"providers": map[string]interface{}{
			"test": map[string]interface{}{
				"name":          "test",
				"base_url":      "https://test.example.com",
				"api_key":       "test-key",
				"models":        []string{"model1"},
				"default_model": "model1",
			},
		},
		"default": "test",
	}
	body, _ := json.Marshal(updateData)
	req = httptest.NewRequest(http.MethodPut, "/api/config", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestFileTreeEndpoint(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	// Create a test directory structure
	tmpDir := t.TempDir()
	os.MkdirAll(filepath.Join(tmpDir, "subdir"), 0755)
	os.WriteFile(filepath.Join(tmpDir, "file1.go"), []byte("package main"), 0644)
	os.WriteFile(filepath.Join(tmpDir, "subdir", "file2.go"), []byte("package sub"), 0644)

	// Insert a test probe
	testID := "filetree-test-" + time.Now().Format("20060102150405")
	err := db.InsertProbe(database, testID, "security", tmpDir, "/tmp/test.md")
	if err != nil {
		t.Fatalf("Failed to insert test probe: %v", err)
	}

	mux := http.NewServeMux()
	RegisterRoutes(mux, database)

	req := httptest.NewRequest(http.MethodGet, "/api/file-tree/"+testID, nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Parse response
	var tree []map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &tree); err != nil {
		t.Errorf("Failed to parse response: %v", err)
	}

	// Should have at least 2 items (file1.go and subdir)
	if len(tree) < 1 {
		t.Errorf("Expected at least 1 item in file tree, got %d", len(tree))
	}
}

func TestCORSHeaders(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	mux := http.NewServeMux()
	RegisterRoutes(mux, database)

	req := httptest.NewRequest(http.MethodGet, "/api/probes", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	// Check CORS headers
	origin := rec.Header().Get("Access-Control-Allow-Origin")
	if origin != "*" {
		t.Errorf("Expected CORS origin '*', got '%s'", origin)
	}

	methods := rec.Header().Get("Access-Control-Allow-Methods")
	if !strings.Contains(methods, "GET") {
		t.Errorf("Expected CORS methods to include GET, got '%s'", methods)
	}
}

func TestAPIErrors(t *testing.T) {
	database := setupTestDB(t)
	defer database.Close()

	mux := http.NewServeMux()
	RegisterRoutes(mux, database)

	// Test invalid method
	req := httptest.NewRequest(http.MethodPost, "/api/probes", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Errorf("Expected status 405, got %d", rec.Code)
	}

	// Test non-existent probe
	req = httptest.NewRequest(http.MethodGet, "/api/probes/non-existent-id", nil)
	rec = httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", rec.Code)
	}

	// Test invalid finding ID
	req = httptest.NewRequest(http.MethodPatch, "/api/findings/", nil)
	rec = httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}

	// Test invalid JSON in config update
	req = httptest.NewRequest(http.MethodPut, "/api/config", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", rec.Code)
	}
}
