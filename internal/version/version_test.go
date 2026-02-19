package version

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestGetInfo(t *testing.T) {
	info := GetInfo()

	if info.Version == "" {
		t.Error("Version should not be empty")
	}
	if info.Commit == "" {
		t.Error("Commit should not be empty")
	}
	if info.BuildDate == "" {
		t.Error("BuildDate should not be empty")
	}
	if info.GoVersion == "" {
		t.Error("GoVersion should not be empty")
	}
	if info.Platform == "" {
		t.Error("Platform should not be empty")
	}
	if info.DashboardURL == "" {
		t.Error("DashboardURL should not be empty")
	}
	if info.ConfigPath == "" {
		t.Error("ConfigPath should not be empty")
	}
}

func TestInfoString(t *testing.T) {
	info := GetInfo()
	str := info.String()

	if !strings.Contains(str, "probeTool") {
		t.Error("String() should contain 'probeTool'")
	}
	if !strings.Contains(str, info.Version) {
		t.Error("String() should contain version")
	}
}

func TestInfoDetailed(t *testing.T) {
	info := GetInfo()
	detailed := info.Detailed()

	if !strings.Contains(detailed, "probeTool") {
		t.Error("Detailed() should contain 'probeTool'")
	}
	if !strings.Contains(detailed, info.Version) {
		t.Error("Detailed() should contain version")
	}
	if !strings.Contains(detailed, "Version") {
		t.Error("Detailed() should contain 'Version' label")
	}
	if !strings.Contains(detailed, "Build Date") {
		t.Error("Detailed() should contain 'Build Date' label")
	}
	if !strings.Contains(detailed, "Git Commit") {
		t.Error("Detailed() should contain 'Git Commit' label")
	}
	if !strings.Contains(detailed, "Go Version") {
		t.Error("Detailed() should contain 'Go Version' label")
	}
	if !strings.Contains(detailed, "Platform") {
		t.Error("Detailed() should contain 'Platform' label")
	}
	if !strings.Contains(detailed, "Dashboard") {
		t.Error("Detailed() should contain 'Dashboard' label")
	}
	if !strings.Contains(detailed, "Config") {
		t.Error("Detailed() should contain 'Config' label")
	}
}

func TestInfoJSON(t *testing.T) {
	info := GetInfo()
	jsonStr, err := info.JSON()

	if err != nil {
		t.Errorf("JSON() returned error: %v", err)
	}

	if jsonStr == "" {
		t.Error("JSON() should return non-empty string")
	}

	// Verify it's valid JSON
	var parsed Info
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		t.Errorf("JSON() returned invalid JSON: %v", err)
	}

	// Verify fields match
	if parsed.Version != info.Version {
		t.Error("JSON version mismatch")
	}
	if parsed.Commit != info.Commit {
		t.Error("JSON commit mismatch")
	}
}

func TestCommitShort(t *testing.T) {
	// Save original commit
	originalCommit := Commit
	defer func() { Commit = originalCommit }()

	// Test with long commit
	Commit = "abc123def456789"
	info := GetInfo()
	if len(info.CommitShort) > 7 {
		t.Error("CommitShort should be truncated to 7 characters")
	}
	if info.CommitShort != "abc123d" {
		t.Errorf("CommitShort = %s, want abc123d", info.CommitShort)
	}

	// Test with short commit
	Commit = "abc12"
	info = GetInfo()
	if info.CommitShort != "abc12" {
		t.Errorf("CommitShort = %s, want abc12", info.CommitShort)
	}
}

func TestPlatform(t *testing.T) {
	info := GetInfo()

	// Platform should be in format os/arch
	parts := strings.Split(info.Platform, "/")
	if len(parts) != 2 {
		t.Errorf("Platform format should be os/arch, got: %s", info.Platform)
	}
}

func TestInfoStruct(t *testing.T) {
	info := Info{
		Version:      "1.0.0",
		Commit:       "abc123",
		CommitShort:  "abc123",
		BuildDate:    "2026-02-19T20:04:32Z",
		GoVersion:    "go1.22.1",
		Platform:     "darwin/arm64",
		DashboardURL: "http://localhost:37330",
		ConfigPath:   "/test/config",
	}

	if info.Version != "1.0.0" {
		t.Error("Version mismatch")
	}
	if info.Platform != "darwin/arm64" {
		t.Error("Platform mismatch")
	}
}
