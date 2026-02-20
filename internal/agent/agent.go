//go:build !nodedebug

package agent

import (
	"embed"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

func Extract(destDir string) error {
	return extractDir(Files, "files", destDir)
}

func extractDir(fsys embed.FS, src, dest string) error {
	entries, err := fs.ReadDir(fsys, src)
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

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	return err
}
