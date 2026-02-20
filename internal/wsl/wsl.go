package wsl

import (
	"bufio"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var settingsFile string

func init() {
	settingsFile = getSettingsFilePath()
}

func getSettingsFilePath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		home, _ := os.UserHomeDir()
		configDir = filepath.Join(home, ".config", "probeTool")
	} else {
		configDir = filepath.Join(configDir, "probeTool")
	}
	return filepath.Join(configDir, "settings.json")
}

func SetSettingsFilePath(path string) {
	settingsFile = path
}

func IsWSL() bool {
	if runtime.GOOS != "linux" {
		return false
	}

	if os.Getenv("WSL_DISTRO_NAME") != "" {
		return true
	}

	if _, err := os.Stat("/proc/version"); err == nil {
		file, err := os.Open("/proc/version")
		if err != nil {
			return false
		}
		defer file.Close()

		scanner := bufio.NewScanner(file)
		if scanner.Scan() {
			version := strings.ToLower(scanner.Text())
			return strings.Contains(version, "microsoft") || strings.Contains(version, "wsl")
		}
	}

	return false
}

func GetTrayDisabled() bool {
	data, err := os.ReadFile(settingsFile)
	if err != nil {
		return false
	}
	content := strings.ToLower(string(data))
	return strings.Contains(content, `"tray_disabled"`) && strings.Contains(content, "true")
}

func SetTrayDisabled(disabled bool) error {
	os.MkdirAll(filepath.Dir(settingsFile), 0755)

	content := "{\n"
	if disabled {
		content += "  \"tray_disabled\": true\n"
	} else {
		content += "  \"tray_disabled\": false\n"
	}
	content += "}\n"

	return os.WriteFile(settingsFile, []byte(content), 0644)
}

func NeedsSetup() bool {
	if !IsWSL() {
		return false
	}

	if _, err := os.Stat(settingsFile); os.IsNotExist(err) {
		return true
	}

	return false
}
