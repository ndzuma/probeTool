//go:build !nodedebug

package runtime

import (
	"embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
)

// NodePath returns the path to the bundled Node.js binary
func NodePath() (string, error) {
	runtimeDir, err := getRuntimeDir()
	if err != nil {
		return "", err
	}

	nodePath := filepath.Join(runtimeDir, "bin", "node")

	// Check if already extracted
	if _, err := os.Stat(nodePath); err == nil {
		return nodePath, nil
	}

	// Extract Node.js runtime
	if err := extractNodeRuntime(runtimeDir); err != nil {
		return "", fmt.Errorf("failed to extract node runtime: %w", err)
	}

	return nodePath, nil
}

// NpmPath returns the path to npm
func NpmPath() (string, error) {
	runtimeDir, err := getRuntimeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(runtimeDir, "bin", "npm"), nil
}

// WebPath returns the path to the bundled web directory
func WebPath() (string, error) {
	runtimeDir, err := getRuntimeDir()
	if err != nil {
		return "", err
	}

	webDir := filepath.Join(runtimeDir, "web")

	// Check if already extracted
	if _, err := os.Stat(webDir); err == nil {
		return webDir, nil
	}

	// Extract web directory
	if err := extractWebDir(webDir); err != nil {
		return "", fmt.Errorf("failed to extract web directory: %w", err)
	}

	return webDir, nil
}

func getRuntimeDir() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	runtimeDir := filepath.Join(cacheDir, "probeTool", "runtime")
	if err := os.MkdirAll(runtimeDir, 0755); err != nil {
		return "", err
	}

	return runtimeDir, nil
}

func extractNodeRuntime(destDir string) error {
	archiveDir := fmt.Sprintf("node-%s-%s", runtime.GOOS, runtime.GOARCH)

	// Check if this platform is bundled
	if _, err := nodeFS.ReadDir(archiveDir); err != nil {
		return fmt.Errorf("node runtime not bundled for %s/%s", runtime.GOOS, runtime.GOARCH)
	}

	return extractDir(nodeFS, archiveDir, destDir)
}

func extractWebDir(destDir string) error {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return err
	}

	return extractDir(webFS, "web", destDir)
}

func extractDir(fsys embed.FS, src, dest string) error {
	entries, err := fsys.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		destPath := filepath.Join(dest, entry.Name())

		if entry.IsDir() {
			if err := os.MkdirAll(destPath, 0755); err != nil {
				return err
			}
			if err := extractDir(fsys, srcPath, destPath); err != nil {
				return err
			}
		} else {
			if err := extractFile(fsys, srcPath, destPath); err != nil {
				return err
			}
		}
	}

	return nil
}

func extractFile(fsys embed.FS, src, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}

	srcFile, err := fsys.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	perm := os.FileMode(0644)
	if filepath.Base(filepath.Dir(dest)) == "bin" {
		perm = 0755
	}

	destFile, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	return err
}
