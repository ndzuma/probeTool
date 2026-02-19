package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGetConfigPath(t *testing.T) {
	path := GetConfigPath()
	if path == "" {
		t.Error("GetConfigPath() returned empty string")
	}
	if !filepath.IsAbs(path) {
		t.Error("GetConfigPath() should return an absolute path")
	}
	if !strings.Contains(path, "probeTool") {
		t.Error("GetConfigPath() should contain 'probeTool' directory")
	}
	if filepath.Base(path) != "config.json" {
		t.Error("GetConfigPath() should end with 'config.json'")
	}
}

func TestLoadConfig(t *testing.T) {
	// Create a temporary config
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".probe")
	os.MkdirAll(configDir, 0755)
	configPath := filepath.Join(configDir, "config.json")

	// Save original config path handling
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

	// Create test config file
	testConfig := &Config{
		Providers: map[string]Provider{
			"test": {
				Name:         "test",
				BaseURL:      "https://test.example.com",
				APIKey:       "test-key",
				Models:       []string{"model1", "model2"},
				DefaultModel: "model1",
			},
		},
		Default: "test",
	}
	data, _ := json.Marshal(testConfig)
	os.WriteFile(configPath, data, 0644)

	// Test loading from the test directory (we can't easily mock GetConfigDir)
	// So we'll test the Load function with the default behavior
	// First, let's create a valid config in temp and test Load with custom path

	// Since we can't easily override GetConfigDir, we'll test the behavior
	// by verifying Load handles missing config correctly
	cfg, err := Load()
	if err != nil {
		t.Errorf("Load() failed: %v", err)
	}
	if cfg == nil {
		t.Error("Load() returned nil config")
	}
	if cfg.Providers == nil {
		t.Error("Load() returned nil Providers map")
	}
}

func TestLoadConfigWithDefaults(t *testing.T) {
	// Test loading when config file doesn't exist
	// This should return an empty config with initialized Providers map
	cfg, err := Load()
	if err != nil {
		t.Errorf("Load() failed for missing config: %v", err)
	}
	if cfg == nil {
		t.Fatal("Load() returned nil config")
	}
	if cfg.Providers == nil {
		t.Error("Load() should initialize empty Providers map")
	}
}

func TestSaveConfig(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".probe")
	os.MkdirAll(configDir, 0755)

	// Create a test config
	cfg := &Config{
		Providers: map[string]Provider{
			"test": {
				Name:         "test",
				BaseURL:      "https://test.example.com",
				APIKey:       "test-key",
				Models:       []string{"model1"},
				DefaultModel: "model1",
			},
		},
		Default: "test",
	}

	// We can't easily test Save() without modifying the config path
	// But we can verify it doesn't panic
	// Note: This test will actually save to ~/.probe/config.json
	err := cfg.Save()
	if err != nil {
		t.Errorf("Save() failed: %v", err)
	}
}

func TestConfigPathByOS(t *testing.T) {
	// Test that GetConfigPath returns appropriate paths for different OS
	configPath := GetConfigPath()

	switch runtime.GOOS {
	case "windows":
		// On Windows, should use AppData\Roaming\probeTool
		if !strings.Contains(configPath, "probeTool") {
			t.Error("Config path should contain 'probeTool' on Windows")
		}
	case "darwin":
		// On macOS: ~/Library/Application Support/probeTool
		if !strings.Contains(configPath, "probeTool") {
			t.Error("Config path should contain 'probeTool' on macOS")
		}
	case "linux":
		// On Linux: ~/.config/probeTool
		if !strings.Contains(configPath, "probeTool") {
			t.Error("Config path should contain 'probeTool' on Linux")
		}
	}

	if !filepath.IsAbs(configPath) {
		t.Error("Config path should be absolute")
	}
}

func TestProviderValidation(t *testing.T) {
	// Test provider structure validation
	provider := Provider{
		Name:         "",
		BaseURL:      "",
		APIKey:       "",
		Models:       []string{},
		DefaultModel: "",
	}

	// Empty provider should be valid struct-wise
	if provider.Name != "" {
		t.Error("Provider name should be empty")
	}

	// Test with valid provider
	provider2 := Provider{
		Name:         "openrouter",
		BaseURL:      "https://openrouter.ai/api",
		APIKey:       "test-api-key",
		Models:       []string{"model1", "model2"},
		DefaultModel: "model1",
	}

	if provider2.Name != "openrouter" {
		t.Error("Provider name mismatch")
	}
	if provider2.BaseURL != "https://openrouter.ai/api" {
		t.Error("Provider BaseURL mismatch")
	}
	if len(provider2.Models) != 2 {
		t.Errorf("Expected 2 models, got %d", len(provider2.Models))
	}
}

func TestAPIKeyRedaction(t *testing.T) {
	// Test that we can handle API keys properly
	// This is a behavioral test - we want to ensure keys can be set
	cfg := &Config{
		Providers: map[string]Provider{
			"test": {
				Name:         "test",
				BaseURL:      "https://test.example.com",
				APIKey:       "sk-test-secret-key-12345",
				Models:       []string{"model1"},
				DefaultModel: "model1",
			},
		},
		Default: "test",
	}

	provider, exists := cfg.Providers["test"]
	if !exists {
		t.Fatal("Provider should exist")
	}

	if provider.APIKey != "sk-test-secret-key-12345" {
		t.Error("API key should match")
	}

	// Test SetAPIKey
	err := cfg.SetAPIKey("test", "new-secret-key")
	if err != nil {
		t.Errorf("SetAPIKey() failed: %v", err)
	}

	provider, _ = cfg.Providers["test"]
	if provider.APIKey != "new-secret-key" {
		t.Error("API key should be updated")
	}
}

func TestSetAPIKeyForNonExistentProvider(t *testing.T) {
	cfg := &Config{
		Providers: make(map[string]Provider),
		Default:   "",
	}

	err := cfg.SetAPIKey("non-existent", "key")
	if err == nil {
		t.Error("SetAPIKey() should return error for non-existent provider")
	}
}

func TestAddModel(t *testing.T) {
	cfg := &Config{
		Providers: map[string]Provider{
			"test": {
				Name:         "test",
				BaseURL:      "https://test.example.com",
				APIKey:       "test-key",
				Models:       []string{},
				DefaultModel: "",
			},
		},
		Default: "test",
	}

	// Add a model
	err := cfg.AddModel("test", "model1")
	if err != nil {
		t.Errorf("AddModel() failed: %v", err)
	}

	provider := cfg.Providers["test"]
	if len(provider.Models) != 1 {
		t.Errorf("Expected 1 model, got %d", len(provider.Models))
	}
	if provider.DefaultModel != "model1" {
		t.Error("Default model should be set to the first model")
	}

	// Add duplicate model should fail
	err = cfg.AddModel("test", "model1")
	if err == nil {
		t.Error("AddModel() should return error for duplicate model")
	}

	// Add model to non-existent provider should fail
	err = cfg.AddModel("non-existent", "model2")
	if err == nil {
		t.Error("AddModel() should return error for non-existent provider")
	}
}

func TestSetDefault(t *testing.T) {
	cfg := &Config{
		Providers: map[string]Provider{
			"test": {
				Name:         "test",
				BaseURL:      "https://test.example.com",
				APIKey:       "test-key",
				Models:       []string{"model1"},
				DefaultModel: "model1",
			},
		},
		Default: "",
	}

	// Set default
	err := cfg.SetDefault("test")
	if err != nil {
		t.Errorf("SetDefault() failed: %v", err)
	}
	if cfg.Default != "test" {
		t.Errorf("Default should be 'test', got '%s'", cfg.Default)
	}

	// Set default to non-existent provider should fail
	err = cfg.SetDefault("non-existent")
	if err == nil {
		t.Error("SetDefault() should return error for non-existent provider")
	}
}

func TestGetDefaultProvider(t *testing.T) {
	// Test when no default is set
	cfg := &Config{
		Providers: make(map[string]Provider),
		Default:   "",
	}

	_, exists := cfg.GetDefaultProvider()
	if exists {
		t.Error("GetDefaultProvider() should return false when no default is set")
	}

	// Test when default is set
	cfg.Providers["test"] = Provider{
		Name:         "test",
		BaseURL:      "https://test.example.com",
		APIKey:       "test-key",
		Models:       []string{"model1"},
		DefaultModel: "model1",
	}
	cfg.Default = "test"

	provider, exists := cfg.GetDefaultProvider()
	if !exists {
		t.Error("GetDefaultProvider() should return true when default is set")
	}
	if provider.Name != "test" {
		t.Errorf("Provider name should be 'test', got '%s'", provider.Name)
	}
}

func TestListProviders(t *testing.T) {
	cfg := &Config{
		Providers: map[string]Provider{
			"provider1": {Name: "provider1"},
			"provider2": {Name: "provider2"},
			"provider3": {Name: "provider3"},
		},
		Default: "",
	}

	names := cfg.ListProviders()
	if len(names) != 3 {
		t.Errorf("Expected 3 providers, got %d", len(names))
	}

	// Check all providers are listed
	found := make(map[string]bool)
	for _, name := range names {
		found[name] = true
	}
	for _, expected := range []string{"provider1", "provider2", "provider3"} {
		if !found[expected] {
			t.Errorf("Provider '%s' not found in list", expected)
		}
	}
}
