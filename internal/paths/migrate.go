package paths

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Migrate moves data from old ~/.probe to new OS-standard location
func Migrate() error {
	oldPath := GetOldProbePath()
	newPath := GetAppDir()

	if oldPath == "" {
		return fmt.Errorf("could not determine home directory")
	}

	fmt.Printf("📦 Migrating from %s to %s\n", oldPath, newPath)

	// Ensure new directories exist
	if err := EnsureAppDirs(); err != nil {
		return fmt.Errorf("failed to create new directories: %w", err)
	}

	// Migrate config.json
	if err := migrateFile(
		filepath.Join(oldPath, "config.json"),
		GetConfigPath(),
	); err != nil {
		fmt.Printf("⚠️  Config migration: %v\n", err)
	}

	// Migrate probes directory
	if err := migrateDirectory(
		filepath.Join(oldPath, "probes"),
		GetProbesDir(),
	); err != nil {
		fmt.Printf("⚠️  Probes migration: %v\n", err)
	}

	// Migrate agent directory
	if err := migrateDirectory(
		filepath.Join(oldPath, "agent"),
		GetAgentDir(),
	); err != nil {
		fmt.Printf("⚠️  Agent migration: %v\n", err)
	}

	fmt.Println("✅ Migration complete!")
	fmt.Printf("💡 Old files remain at %s (you can delete them manually)\n", oldPath)

	return nil
}

func migrateFile(src, dst string) error {
	// Check if source exists
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return nil // Nothing to migrate
	}

	// Check if destination already exists
	if _, err := os.Stat(dst); err == nil {
		return nil // Already migrated
	}

	// Copy file
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	fmt.Printf("  ✓ Migrated: %s\n", filepath.Base(src))
	return nil
}

func migrateDirectory(src, dst string) error {
	// Check if source exists
	if _, err := os.Stat(src); os.IsNotExist(err) {
		return nil // Nothing to migrate
	}

	// Create destination
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := migrateDirectory(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := migrateFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}
