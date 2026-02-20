package version

import (
	"encoding/json"
	"fmt"
	"runtime"
	"time"

	"github.com/fatih/color"
	"github.com/ndzuma/probeTool/internal/paths"
)

// These will be set via -ldflags during build
var (
	Version   = "dev"
	Commit    = "unknown"
	BuildDate = "unknown"
	GoVersion = runtime.Version()
)

type Info struct {
	Version      string `json:"version"`
	Commit       string `json:"commit"`
	CommitShort  string `json:"commitShort"`
	BuildDate    string `json:"buildDate"`
	GoVersion    string `json:"goVersion"`
	Platform     string `json:"platform"`
	DashboardURL string `json:"dashboardURL"`
	ConfigPath   string `json:"configPath"`
}

func GetInfo() Info {
	commitShort := Commit
	if len(commitShort) > 7 {
		commitShort = commitShort[:7]
	}

	return Info{
		Version:      Version,
		Commit:       Commit,
		CommitShort:  commitShort,
		BuildDate:    BuildDate,
		GoVersion:    GoVersion,
		Platform:     fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH),
		DashboardURL: "http://localhost:37330",
		ConfigPath:   paths.GetAppDir(),
	}
}

func (i Info) String() string {
	return fmt.Sprintf("probeTool %s", i.Version)
}

func (i Info) Detailed() string {
	cyan := color.New(color.FgCyan).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()

	// Parse and format build date nicely
	buildDateFormatted := i.BuildDate
	if t, err := time.Parse(time.RFC3339, i.BuildDate); err == nil {
		buildDateFormatted = t.Format("January 2, 2006 at 3:04 PM")
	}

	// Pretty platform name
	platformName := i.Platform
	switch i.Platform {
	case "darwin/amd64":
		platformName = "macOS (Intel)"
	case "darwin/arm64":
		platformName = "macOS (Apple Silicon)"
	case "linux/amd64":
		platformName = "Linux (AMD64)"
	case "linux/arm64":
		platformName = "Linux (ARM64)"
	case "windows/amd64":
		platformName = "Windows (AMD64)"
	}

	return fmt.Sprintf(`%s %s

  %s  %s
  %s  %s
  %s  %s
  %s  %s
  %s  %s

  %s  %s
  %s  %s`,
		cyan("🔧"), bold(fmt.Sprintf("probeTool %s", i.Version)),
		bold("Version:    "), i.Version,
		bold("Build Date: "), buildDateFormatted,
		bold("Git Commit: "), i.Commit,
		bold("Go Version: "), i.GoVersion,
		bold("Platform:   "), platformName,
		bold("Dashboard:  "), cyan(i.DashboardURL),
		bold("Config:     "), i.ConfigPath,
	)
}

func (i Info) JSON() (string, error) {
	data, err := json.MarshalIndent(i, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
