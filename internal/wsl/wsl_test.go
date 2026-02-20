package wsl

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsWSL(t *testing.T) {
	_ = IsWSL()
}

func TestGetTrayDisabledNoFile(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "probeTool-test")
	defer os.RemoveAll(tmpDir)

	SetSettingsFilePath(filepath.Join(tmpDir, "settings.json"))

	if GetTrayDisabled() {
		t.Error("GetTrayDisabled should return false when file doesn't exist")
	}
}

func TestSetAndGetTrayDisabled(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "probeTool-test")
	defer os.RemoveAll(tmpDir)

	SetSettingsFilePath(filepath.Join(tmpDir, "settings.json"))

	if err := SetTrayDisabled(true); err != nil {
		t.Errorf("SetTrayDisabled(true) failed: %v", err)
	}

	if !GetTrayDisabled() {
		t.Error("GetTrayDisabled should return true after SetTrayDisabled(true)")
	}

	if err := SetTrayDisabled(false); err != nil {
		t.Errorf("SetTrayDisabled(false) failed: %v", err)
	}

	if GetTrayDisabled() {
		t.Error("GetTrayDisabled should return false after SetTrayDisabled(false)")
	}
}

func TestNeedsSetupNoFile(t *testing.T) {
	tmpDir, _ := os.MkdirTemp("", "probeTool-test")
	defer os.RemoveAll(tmpDir)

	SetSettingsFilePath(filepath.Join(tmpDir, "settings.json"))

	if IsWSL() {
		if !NeedsSetup() {
			t.Error("NeedsSetup should return true when running on WSL with no settings file")
		}
	} else {
		if NeedsSetup() {
			t.Error("NeedsSetup should return false when not running on WSL")
		}
	}
}
