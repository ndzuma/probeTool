package updater

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/ndzuma/probeTool/internal/paths"
	"github.com/ndzuma/probeTool/internal/version"
)

const (
	githubRepoURL  = "github.com/ndzuma/probeTool"
	requestTimeout = 30 * time.Second
)

var githubAPIURL = "https://api.github.com/repos/ndzuma/probeTool/releases/latest"

type Release struct {
	TagName string  `json:"tag_name"`
	Name    string  `json:"name"`
	Body    string  `json:"body"`
	Assets  []Asset `json:"assets"`
	HTMLURL string  `json:"html_url"`
}

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Size               int64  `json:"size"`
}

type UpdateInfo struct {
	CurrentVersion string
	LatestVersion  string
	HasUpdate      bool
	ReleaseNotes   string
	DownloadURL    string
	AssetName      string
	ReleasePageURL string
}

func getHTTPClient() *http.Client {
	return &http.Client{
		Timeout: requestTimeout,
	}
}

func FetchLatestRelease() (*Release, error) {
	client := getHTTPClient()
	resp, err := client.Get(githubAPIURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch release info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var release Release
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("failed to parse release info: %w", err)
	}

	return &release, nil
}

func CheckForUpdate() (*UpdateInfo, error) {
	currentVersion := version.GetInfo().Version

	release, err := FetchLatestRelease()
	if err != nil {
		return nil, err
	}

	latestVersion := strings.TrimPrefix(release.TagName, "v")
	currentVersionClean := strings.TrimPrefix(currentVersion, "v")

	hasUpdate := latestVersion != currentVersionClean

	downloadURL := findAssetURL(release)

	return &UpdateInfo{
		CurrentVersion: currentVersion,
		LatestVersion:  release.TagName,
		HasUpdate:      hasUpdate,
		ReleaseNotes:   release.Body,
		DownloadURL:    downloadURL,
		AssetName:      getExactAssetName(release.TagName),
		ReleasePageURL: release.HTMLURL,
	}, nil
}

func getAssetName() string {
	osName := runtime.GOOS
	arch := runtime.GOARCH

	switch osName {
	case "darwin":
		osName = "darwin"
	case "linux":
		osName = "linux"
	case "windows":
		osName = "windows"
	}

	return fmt.Sprintf("probeTool_*_%s_%s.tar.gz", osName, arch)
}

func getExactAssetName(tag string) string {
	osName := runtime.GOOS
	arch := runtime.GOARCH
	return fmt.Sprintf("probeTool_%s_%s_%s.tar.gz", tag, osName, arch)
}

func findAssetURL(release *Release) string {
	exactName := getExactAssetName(release.TagName)
	for _, asset := range release.Assets {
		if asset.Name == exactName {
			return asset.BrowserDownloadURL
		}
	}

	pattern := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)
	for _, asset := range release.Assets {
		if strings.Contains(asset.Name, pattern) && strings.HasSuffix(asset.Name, ".tar.gz") {
			return asset.BrowserDownloadURL
		}
	}

	return ""
}

func getExecutablePath() (string, error) {
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("failed to get executable path: %w", err)
	}

	resolvedPath, err := filepath.EvalSymlinks(execPath)
	if err != nil {
		return execPath, nil
	}
	return resolvedPath, nil
}

func DownloadAndInstall(downloadURL string) error {
	if downloadURL == "" {
		return fmt.Errorf("no download URL available for your platform (%s/%s)", runtime.GOOS, runtime.GOARCH)
	}

	bold := color.New(color.Bold).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	fmt.Printf("%s Downloading update...\n", yellow("⬇"))
	fmt.Printf("  URL: %s\n", cyan(downloadURL))

	client := getHTTPClient()
	resp, err := client.Get(downloadURL)
	if err != nil {
		return fmt.Errorf("failed to download update: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	tempDir, err := os.MkdirTemp("", "probe-update-*")
	if err != nil {
		return fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	archivePath := filepath.Join(tempDir, "probe.tar.gz")
	out, err := os.Create(archivePath)
	if err != nil {
		return fmt.Errorf("failed to create archive file: %w", err)
	}

	_, err = io.Copy(out, resp.Body)
	out.Close()
	if err != nil {
		return fmt.Errorf("failed to download archive: %w", err)
	}

	fmt.Printf("%s Extracting...\n", yellow("📦"))

	newBinaryPath := filepath.Join(tempDir, "probe")
	if err := extractBinary(archivePath, newBinaryPath); err != nil {
		return fmt.Errorf("failed to extract binary: %w", err)
	}

	currentExecPath, err := getExecutablePath()
	if err != nil {
		return err
	}

	fmt.Printf("%s Installing update...\n", yellow("🔧"))
	fmt.Printf("  Current binary: %s\n", cyan(currentExecPath))

	backupPath := currentExecPath + ".backup"
	if err := copyFile(currentExecPath, backupPath); err != nil {
		return fmt.Errorf("failed to backup current binary: %w", err)
	}

	if err := copyFile(newBinaryPath, currentExecPath); err != nil {
		fmt.Printf("%s Restore failed, please reinstall manually\n", color.RedString("❌"))
		return fmt.Errorf("failed to install new binary: %w", err)
	}

	os.Remove(backupPath)

	fmt.Printf("%s %s\n", color.GreenString("✅"), bold("Update installed successfully!"))
	fmt.Printf("\nYour data is preserved at: %s\n", paths.GetAppDir())
	fmt.Printf("Run 'probe version' to verify the update.\n")

	return nil
}

func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

func extractBinary(archivePath, destPath string) error {
	tempExtractDir := filepath.Dir(archivePath)

	if err := untar(archivePath, tempExtractDir); err != nil {
		return err
	}

	extractedBinary := filepath.Join(tempExtractDir, "probe")
	if _, err := os.Stat(extractedBinary); os.IsNotExist(err) {
		return fmt.Errorf("probe binary not found in archive")
	}

	return os.Rename(extractedBinary, destPath)
}

func untar(archivePath, destDir string) error {
	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	gzr, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		header, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		if header == nil {
			continue
		}

		target := filepath.Join(destDir, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return err
			}
		case tar.TypeReg:
			outFile, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		}
	}

	return nil
}
