package auditor

import (
	"os"
)

// GetTarget returns the target directory for the audit.
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
