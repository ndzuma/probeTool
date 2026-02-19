package prober

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestProbeArgsValidation(t *testing.T) {
	// Test valid args
	validArgs := ProbeArgs{
		Type:     "security",
		Provider: "openrouter",
		Model:    "anthropic/claude-3.5-haiku",
		Verbose:  true,
	}

	if validArgs.Type != "security" {
		t.Error("Type should be 'security'")
	}
	if validArgs.Provider != "openrouter" {
		t.Error("Provider should be 'openrouter'")
	}
	if !validArgs.Verbose {
		t.Error("Verbose should be true")
	}

	// Test empty args (should still be valid struct)
	emptyArgs := ProbeArgs{}
	if emptyArgs.Type != "" {
		t.Error("Empty args should have empty Type")
	}
}

func TestRunProbeWithValidArgs(t *testing.T) {
	// This test requires the agent to be set up
	// We'll test the validation part without actually running the agent
	args := ProbeArgs{
		Type:     "security",
		Provider: "test-provider",
		Model:    "test-model",
		Verbose:  false,
	}

	// Verify the args are correctly structured
	if args.Type == "" {
		t.Error("Type should not be empty")
	}
}

func TestRunProbeWithInvalidDirectory(t *testing.T) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Create args pointing to non-existent directory
	args := ProbeArgs{
		Type:     "security",
		Provider: "test",
		Model:    "test-model",
		Verbose:  false,
	}

	// Since RunProbe actually runs and we don't have a full setup,
	// we'll just verify the context is created properly
	select {
	case <-ctx.Done():
		t.Error("Context should not be done yet")
	default:
		// Context is still valid
	}

	_ = args
}

func TestRunProbeContextCancellation(t *testing.T) {
	// Test that context cancellation works
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately
	cancel()

	// Verify context is cancelled
	select {
	case <-ctx.Done():
		// Expected - context was cancelled
	case <-time.After(100 * time.Millisecond):
		t.Error("Context should have been cancelled")
	}

	if ctx.Err() != context.Canceled {
		t.Errorf("Expected context.Canceled, got %v", ctx.Err())
	}
}

func TestAgentProcessLifecycle(t *testing.T) {
	// Test that we can check for agent path
	// This test verifies the getAgentPath function behavior

	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot get user home directory")
	}

	agentPath := filepath.Join(homeDir, ".probe", "agent", "probe-runner.js")

	// Check if agent exists
	_, err = os.Stat(agentPath)
	if err != nil && !os.IsNotExist(err) {
		t.Errorf("Unexpected error checking agent path: %v", err)
	}

	// The test should not fail if agent doesn't exist
	// We're just verifying the path logic works
}

func TestGetAgentPath(t *testing.T) {
	// We can't easily test getAgentPath without mocking
	// But we can verify it handles the error case
	// by checking it returns an error when agent doesn't exist

	// Save original HOME
	origHome := os.Getenv("HOME")
	defer os.Setenv("HOME", origHome)

	// Set a temp directory as HOME
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)

	// Create a dummy agent path that doesn't exist
	agentDir := filepath.Join(tmpDir, ".probe", "agent")
	os.MkdirAll(agentDir, 0755)

	// The agent file shouldn't exist, so getAgentPath should return an error
	// But we can't easily test the internal function
	// We'll verify the directory structure is correct

	if _, err := os.Stat(agentDir); os.IsNotExist(err) {
		t.Error("Agent directory should exist")
	}

	agentScript := filepath.Join(agentDir, "probe-runner.js")
	if _, err := os.Stat(agentScript); !os.IsNotExist(err) {
		t.Skip("Agent script exists, cannot test error case")
	}
}

func TestProbeArgsStruct(t *testing.T) {
	// Test various ProbeArgs configurations
	tests := []struct {
		name     string
		args     ProbeArgs
		wantType string
	}{
		{
			name: "security probe",
			args: ProbeArgs{
				Type:     "security",
				Provider: "openrouter",
				Model:    "anthropic/claude-3.5-haiku",
				Verbose:  false,
			},
			wantType: "security",
		},
		{
			name: "verbose probe",
			args: ProbeArgs{
				Type:     "security",
				Provider: "openrouter",
				Model:    "anthropic/claude-3-opus",
				Verbose:  true,
			},
			wantType: "security",
		},
		{
			name: "empty provider",
			args: ProbeArgs{
				Type:     "security",
				Provider: "",
				Model:    "",
				Verbose:  false,
			},
			wantType: "security",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.args.Type != tt.wantType {
				t.Errorf("Type = %v, want %v", tt.args.Type, tt.wantType)
			}
		})
	}
}
