package paths

import (
	"os"
	"path/filepath"
	"runtime"
)

const appName = "probeTool"

// GetAppDir returns the OS-appropriate application directory
func GetAppDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory
		return "."
	}

	switch runtime.GOOS {
	case "darwin":
		// macOS: ~/Library/Application Support/probeTool
		return filepath.Join(home, "Library", "Application Support", appName)
	case "windows":
		// Windows: %APPDATA%\probeTool
		appData := os.Getenv("APPDATA")
		if appData != "" {
			return filepath.Join(appData, appName)
		}
		return filepath.Join(home, "AppData", "Roaming", appName)
	default:
		// Linux/Unix: ~/.config/probeTool
		xdgConfig := os.Getenv("XDG_CONFIG_HOME")
		if xdgConfig != "" {
			return filepath.Join(xdgConfig, appName)
		}
		return filepath.Join(home, ".config", appName)
	}
}

// GetConfigPath returns the full path to config.json
func GetConfigPath() string {
	return filepath.Join(GetAppDir(), "config.json")
}

// GetProbesDir returns the directory where probe reports are stored
func GetProbesDir() string {
	return filepath.Join(GetAppDir(), "probes")
}

// GetDBPath returns the full path to the SQLite database
func GetDBPath() string {
	return filepath.Join(GetProbesDir(), "probes.db")
}

// GetAgentDir returns the directory where agent files are stored
func GetAgentDir() string {
	return filepath.Join(GetAppDir(), "agent")
}

// GetAgentPath returns the full path to the agent script
func GetAgentPath() string {
	return filepath.Join(GetAgentDir(), "probe-runner.js")
}

// GetCacheDir returns the directory for temporary/cache files
func GetCacheDir() string {
	switch runtime.GOOS {
	case "darwin":
		// macOS: ~/Library/Caches/probeTool
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Caches", appName)
	case "windows":
		// Windows: %LOCALAPPDATA%\probeTool\Cache
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			home, _ := os.UserHomeDir()
			localAppData = filepath.Join(home, "AppData", "Local")
		}
		return filepath.Join(localAppData, appName, "Cache")
	default:
		// Linux: ~/.cache/probeTool
		home, _ := os.UserHomeDir()
		xdgCache := os.Getenv("XDG_CACHE_HOME")
		if xdgCache != "" {
			return filepath.Join(xdgCache, appName)
		}
		return filepath.Join(home, ".cache", appName)
	}
}

// GetLogDir returns the directory for log files
func GetLogDir() string {
	switch runtime.GOOS {
	case "darwin":
		home, _ := os.UserHomeDir()
		return filepath.Join(home, "Library", "Logs", appName)
	case "windows":
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			home, _ := os.UserHomeDir()
			localAppData = filepath.Join(home, "AppData", "Local")
		}
		return filepath.Join(localAppData, appName, "Logs")
	default:
		home, _ := os.UserHomeDir()
		return filepath.Join(home, ".local", "share", appName, "logs")
	}
}

// EnsureAppDirs creates all necessary application directories
func EnsureAppDirs() error {
	dirs := []string{
		GetAppDir(),
		GetProbesDir(),
		GetAgentDir(),
		GetCacheDir(),
		GetLogDir(),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}

	return nil
}

// GetOldProbePath returns the old ~/.probe path for migration
func GetOldProbePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".probe")
}

// NeedsMigration checks if old config exists and new doesn't
func NeedsMigration() bool {
	oldPath := GetOldProbePath()
	newPath := GetAppDir()

	// Check if old path exists
	if _, err := os.Stat(oldPath); os.IsNotExist(err) {
		return false
	}

	// Check if new path doesn't exist or is empty
	if _, err := os.Stat(newPath); os.IsNotExist(err) {
		return true
	}

	// New path exists - check if it has content
	entries, err := os.ReadDir(newPath)
	return err != nil || len(entries) == 0
}
