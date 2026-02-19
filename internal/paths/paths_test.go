package paths

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestGetAppDir(t *testing.T) {
	dir := GetAppDir()
	if dir == "" {
		t.Error("GetAppDir() returned empty string")
	}
	if !filepath.IsAbs(dir) {
		t.Error("GetAppDir() should return absolute path")
	}

	// Verify OS-specific paths
	switch runtime.GOOS {
	case "darwin":
		home, _ := os.UserHomeDir()
		expected := filepath.Join(home, "Library", "Application Support", appName)
		if dir != expected {
			t.Errorf("macOS app dir = %v, want %v", dir, expected)
		}
	case "windows":
		// Should contain probeTool
		if !contains(dir, appName) {
			t.Error("Windows app dir should contain appName")
		}
	default:
		// Linux: should use XDG_CONFIG_HOME or ~/.config
		home, _ := os.UserHomeDir()
		expected := filepath.Join(home, ".config", appName)
		if dir != expected {
			xdgConfig := os.Getenv("XDG_CONFIG_HOME")
			if xdgConfig != "" {
				expected = filepath.Join(xdgConfig, appName)
			}
			if dir != expected {
				t.Errorf("Linux app dir = %v, want %v", dir, expected)
			}
		}
	}
}

func TestGetConfigPath(t *testing.T) {
	configPath := GetConfigPath()
	if configPath == "" {
		t.Error("GetConfigPath() returned empty string")
	}
	if !filepath.IsAbs(configPath) {
		t.Error("GetConfigPath() should return absolute path")
	}
	if filepath.Base(configPath) != "config.json" {
		t.Error("GetConfigPath() should end with 'config.json'")
	}
}

func TestGetProbesDir(t *testing.T) {
	probesDir := GetProbesDir()
	if probesDir == "" {
		t.Error("GetProbesDir() returned empty string")
	}
	if !filepath.IsAbs(probesDir) {
		t.Error("GetProbesDir() should return absolute path")
	}
	if filepath.Base(probesDir) != "probes" {
		t.Error("GetProbesDir() should end with 'probes'")
	}
}

func TestGetDBPath(t *testing.T) {
	dbPath := GetDBPath()
	if dbPath == "" {
		t.Error("GetDBPath() returned empty string")
	}
	if !filepath.IsAbs(dbPath) {
		t.Error("GetDBPath() should return absolute path")
	}
	if filepath.Ext(dbPath) != ".db" {
		t.Error("GetDBPath() should have .db extension")
	}
}

func TestGetCacheDir(t *testing.T) {
	cacheDir := GetCacheDir()
	if cacheDir == "" {
		t.Error("GetCacheDir() returned empty string")
	}
	if !filepath.IsAbs(cacheDir) {
		t.Error("GetCacheDir() should return absolute path")
	}

	// Verify OS-specific paths
	switch runtime.GOOS {
	case "darwin":
		home, _ := os.UserHomeDir()
		expected := filepath.Join(home, "Library", "Caches", appName)
		if cacheDir != expected {
			t.Errorf("macOS cache dir = %v, want %v", cacheDir, expected)
		}
	case "windows":
		// Should contain probeTool and Cache
		if !contains(cacheDir, appName) {
			t.Error("Windows cache dir should contain appName")
		}
	case "linux":
		// Should use XDG_CACHE_HOME or ~/.cache
		home, _ := os.UserHomeDir()
		expected := filepath.Join(home, ".cache", appName)
		if cacheDir != expected {
			xdgCache := os.Getenv("XDG_CACHE_HOME")
			if xdgCache != "" {
				expected = filepath.Join(xdgCache, appName)
			}
			if cacheDir != expected {
				t.Errorf("Linux cache dir = %v, want %v", cacheDir, expected)
			}
		}
	}
}

func TestGetAgentDir(t *testing.T) {
	agentDir := GetAgentDir()
	if agentDir == "" {
		t.Error("GetAgentDir() returned empty string")
	}
	if !filepath.IsAbs(agentDir) {
		t.Error("GetAgentDir() should return absolute path")
	}
	if filepath.Base(agentDir) != "agent" {
		t.Error("GetAgentDir() should end with 'agent'")
	}
}

func TestGetAgentPath(t *testing.T) {
	agentPath := GetAgentPath()
	if agentPath == "" {
		t.Error("GetAgentPath() returned empty string")
	}
	if !filepath.IsAbs(agentPath) {
		t.Error("GetAgentPath() should return absolute path")
	}
	if filepath.Base(agentPath) != "probe-runner.js" {
		t.Error("GetAgentPath() should end with 'probe-runner.js'")
	}
}

func TestGetLogDir(t *testing.T) {
	logDir := GetLogDir()
	if logDir == "" {
		t.Error("GetLogDir() returned empty string")
	}
	if !filepath.IsAbs(logDir) {
		t.Error("GetLogDir() should return absolute path")
	}

	// Verify OS-specific paths
	switch runtime.GOOS {
	case "darwin":
		home, _ := os.UserHomeDir()
		expected := filepath.Join(home, "Library", "Logs", appName)
		if logDir != expected {
			t.Errorf("macOS log dir = %v, want %v", logDir, expected)
		}
	case "windows":
		// Should contain probeTool and Logs
		if !contains(logDir, appName) {
			t.Error("Windows log dir should contain appName")
		}
	case "linux":
		// Should use ~/.local/share
		home, _ := os.UserHomeDir()
		expected := filepath.Join(home, ".local", "share", appName, "logs")
		if logDir != expected {
			t.Errorf("Linux log dir = %v, want %v", logDir, expected)
		}
	}
}

func TestEnsureAppDirs(t *testing.T) {
	// Save original HOME
	origHome := os.Getenv("HOME")
	if runtime.GOOS == "windows" {
		origHome = os.Getenv("USERPROFILE")
	}
	defer func() {
		if runtime.GOOS == "windows" {
			os.Setenv("USERPROFILE", origHome)
		} else {
			os.Setenv("HOME", origHome)
		}
	}()

	// Set a temp directory as HOME
	tmpDir := t.TempDir()
	if runtime.GOOS == "windows" {
		os.Setenv("USERPROFILE", tmpDir)
	} else {
		os.Setenv("HOME", tmpDir)
	}

	// Create all directories
	err := EnsureAppDirs()
	if err != nil {
		t.Fatalf("EnsureAppDirs() failed: %v", err)
	}

	// Verify directories were created
	dirs := []string{
		GetAppDir(),
		GetProbesDir(),
		GetAgentDir(),
		GetCacheDir(),
		GetLogDir(),
	}

	for _, dir := range dirs {
		info, err := os.Stat(dir)
		if err != nil {
			t.Errorf("Directory %s was not created: %v", dir, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("%s is not a directory", dir)
		}
	}
}

func TestGetOldProbePath(t *testing.T) {
	oldPath := GetOldProbePath()
	if oldPath == "" {
		t.Error("GetOldProbePath() returned empty string")
	}
	if !filepath.IsAbs(oldPath) {
		t.Error("GetOldProbePath() should return absolute path")
	}
	if !contains(oldPath, ".probe") {
		t.Error("GetOldProbePath() should contain '.probe'")
	}
}

func TestNeedsMigration(t *testing.T) {
	// This test assumes no old ~/.probe exists in test environment
	// So it should return false
	needs := NeedsMigration()
	// We can't make strong assertions here without creating test fixtures
	// Just verify the function doesn't panic
	_ = needs
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
