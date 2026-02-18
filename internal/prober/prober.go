package prober

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/ndzuma/probeTool/internal/db"
)

// GetTarget returns the target directory for the probe.
// If specificPath is provided, it uses that; otherwise, it uses the current working directory.
func GetTarget(specificPath string) (string, error) {
	if specificPath != "" {
		return specificPath, nil
	}

	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	return cwd, nil
}

// RunProbe executes the probe and returns the path to the generated markdown file.
// This spawns the node agent to perform the actual probe.
func RunProbe(ctx context.Context, probeType string) (string, error) {
	cwd, err := os.Getwd() // Repo being probed
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	probesDir := db.ProbesDir() // probeTool/probes/

	// Create probes dir if needed
	if err := os.MkdirAll(probesDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create probes directory: %w", err)
	}

	// Generate ID
	id := fmt.Sprintf("%s-%s", time.Now().Format("2006-01-02-150405"), probeType)
	mdPath := filepath.Join(probesDir, id+".md")

	// Spawn node agent/probe-runner.js --full --target=cwd --out=mdPath
	// TODO: Implement actual node agent spawning
	cmd := exec.CommandContext(ctx, "node", "agent/probe-runner.js",
		"--"+probeType,
		"--target="+cwd,
		"--out="+mdPath,
	)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// For now, just return the path (stubbed)
	_ = cmd

	return mdPath, nil
}
