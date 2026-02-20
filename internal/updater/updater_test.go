package updater

import (
	"archive/tar"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ndzuma/probeTool/internal/version"
)

func TestGetAssetName(t *testing.T) {
	assetName := getAssetName()

	if assetName == "" {
		t.Error("getAssetName() should not return empty string")
	}

	if !strings.Contains(assetName, runtime.GOOS) {
		t.Errorf("getAssetName() should contain OS (%s), got: %s", runtime.GOOS, assetName)
	}

	if !strings.Contains(assetName, runtime.GOARCH) {
		t.Errorf("getAssetName() should contain arch (%s), got: %s", runtime.GOARCH, assetName)
	}

	if !strings.HasSuffix(assetName, ".tar.gz") {
		t.Errorf("getAssetName() should end with .tar.gz, got: %s", assetName)
	}
}

func TestGetExactAssetName(t *testing.T) {
	tag := "v1.0.0"
	assetName := getExactAssetName(tag)

	expected := "probeTool_v1.0.0_" + runtime.GOOS + "_" + runtime.GOARCH + ".tar.gz"
	if assetName != expected {
		t.Errorf("getExactAssetName() = %s, want %s", assetName, expected)
	}
}

func TestReleaseStruct(t *testing.T) {
	release := Release{
		TagName: "v1.0.0",
		Name:    "Release 1.0.0",
		Body:    "Release notes",
		Assets: []Asset{
			{Name: "probe.tar.gz", BrowserDownloadURL: "https://example.com/probe.tar.gz"},
		},
		HTMLURL: "https://github.com/example/repo/releases/tag/v1.0.0",
	}

	if release.TagName != "v1.0.0" {
		t.Error("Release.TagName mismatch")
	}
	if len(release.Assets) != 1 {
		t.Error("Release.Assets should have 1 element")
	}
}

func TestAssetStruct(t *testing.T) {
	asset := Asset{
		Name:               "probe.tar.gz",
		BrowserDownloadURL: "https://example.com/probe.tar.gz",
		Size:               12345,
	}

	if asset.Name != "probe.tar.gz" {
		t.Error("Asset.Name mismatch")
	}
	if asset.Size != 12345 {
		t.Error("Asset.Size mismatch")
	}
}

func TestUpdateInfoStruct(t *testing.T) {
	info := UpdateInfo{
		CurrentVersion: "1.0.0",
		LatestVersion:  "v1.1.0",
		HasUpdate:      true,
		ReleaseNotes:   "New features",
		DownloadURL:    "https://example.com/probe.tar.gz",
		AssetName:      "probe.tar.gz",
		ReleasePageURL: "https://github.com/example/repo/releases/tag/v1.1.0",
	}

	if !info.HasUpdate {
		t.Error("UpdateInfo.HasUpdate should be true")
	}
	if info.CurrentVersion != "1.0.0" {
		t.Error("UpdateInfo.CurrentVersion mismatch")
	}
}

func TestFetchLatestRelease(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"tag_name": "v1.0.0",
			"name": "Release 1.0.0",
			"body": "Release notes",
			"html_url": "https://github.com/ndzuma/probeTool/releases/tag/v1.0.0",
			"assets": [
				{
					"name": "probeTool_v1.0.0_darwin_arm64.tar.gz",
					"browser_download_url": "https://example.com/probe.tar.gz",
					"size": 12345
				}
			]
		}`))
	}))
	defer server.Close()

	originalURL := githubAPIURL
	githubAPIURL = server.URL
	defer func() { githubAPIURL = originalURL }()

	release, err := FetchLatestRelease()
	if err != nil {
		t.Fatalf("FetchLatestRelease() returned error: %v", err)
	}

	if release.TagName != "v1.0.0" {
		t.Errorf("Release.TagName = %s, want v1.0.0", release.TagName)
	}
	if len(release.Assets) != 1 {
		t.Errorf("len(Release.Assets) = %d, want 1", len(release.Assets))
	}
}

func TestFetchLatestReleaseError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	originalURL := githubAPIURL
	githubAPIURL = server.URL
	defer func() { githubAPIURL = originalURL }()

	_, err := FetchLatestRelease()
	if err == nil {
		t.Error("FetchLatestRelease() should return error on 500 status")
	}
}

func TestFetchLatestReleaseInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`invalid json`))
	}))
	defer server.Close()

	originalURL := githubAPIURL
	githubAPIURL = server.URL
	defer func() { githubAPIURL = originalURL }()

	_, err := FetchLatestRelease()
	if err == nil {
		t.Error("FetchLatestRelease() should return error on invalid JSON")
	}
}

func TestCheckForUpdate(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"tag_name": "v99.99.99",
			"name": "Release 99.99.99",
			"body": "New release",
			"html_url": "https://github.com/ndzuma/probeTool/releases/tag/v99.99.99",
			"assets": [
				{
					"name": "probeTool_v99.99.99_` + runtime.GOOS + `_` + runtime.GOARCH + `.tar.gz",
					"browser_download_url": "https://example.com/probe.tar.gz",
					"size": 12345
				}
			]
		}`))
	}))
	defer server.Close()

	originalURL := githubAPIURL
	githubAPIURL = server.URL
	defer func() { githubAPIURL = originalURL }()

	info, err := CheckForUpdate()
	if err != nil {
		t.Fatalf("CheckForUpdate() returned error: %v", err)
	}

	if info.LatestVersion != "v99.99.99" {
		t.Errorf("LatestVersion = %s, want v99.99.99", info.LatestVersion)
	}

	currentVersion := version.GetInfo().Version
	if info.CurrentVersion != currentVersion {
		t.Errorf("CurrentVersion = %s, want %s", info.CurrentVersion, currentVersion)
	}
}

func TestCheckForUpdateNoUpdateAvailable(t *testing.T) {
	currentVersion := version.GetInfo().Version
	versionWithoutV := strings.TrimPrefix(currentVersion, "v")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"tag_name": "` + versionWithoutV + `",
			"name": "Current Release",
			"body": "Current",
			"html_url": "https://github.com/ndzuma/probeTool/releases/tag/v` + versionWithoutV + `",
			"assets": []
		}`))
	}))
	defer server.Close()

	originalURL := githubAPIURL
	githubAPIURL = server.URL
	defer func() { githubAPIURL = originalURL }()

	info, err := CheckForUpdate()
	if err != nil {
		t.Fatalf("CheckForUpdate() returned error: %v", err)
	}

	if info.HasUpdate {
		t.Error("HasUpdate should be false when versions match")
	}
}

func TestFindAssetURL(t *testing.T) {
	release := &Release{
		TagName: "v1.0.0",
		Assets: []Asset{
			{Name: "probeTool_v1.0.0_linux_amd64.tar.gz", BrowserDownloadURL: "https://example.com/linux.tar.gz"},
			{Name: "probeTool_v1.0.0_darwin_arm64.tar.gz", BrowserDownloadURL: "https://example.com/darwin.tar.gz"},
			{Name: "probeTool_v1.0.0_windows_amd64.tar.gz", BrowserDownloadURL: "https://example.com/windows.tar.gz"},
		},
	}

	url := findAssetURL(release)
	if url == "" {
		t.Error("findAssetURL() should return a URL for the current platform")
	}

	if !strings.Contains(url, ".tar.gz") {
		t.Errorf("findAssetURL() should return a .tar.gz URL, got: %s", url)
	}
}

func TestFindAssetURLNoMatch(t *testing.T) {
	release := &Release{
		TagName: "v1.0.0",
		Assets:  []Asset{},
	}

	url := findAssetURL(release)
	if url != "" {
		t.Errorf("findAssetURL() should return empty string when no assets, got: %s", url)
	}
}

func TestUntar(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "probe-untar-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	archivePath := filepath.Join(tempDir, "test.tar.gz")

	file, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("Failed to create archive: %v", err)
	}

	gzw := gzip.NewWriter(file)
	tw := tar.NewWriter(gzw)

	content := []byte("#!/bin/sh\necho 'probe'\n")
	hdr := &tar.Header{
		Name: "probe",
		Mode: 0755,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("Failed to write tar header: %v", err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatalf("Failed to write tar content: %v", err)
	}

	tw.Close()
	gzw.Close()
	file.Close()

	extractDir := filepath.Join(tempDir, "extracted")
	if err := os.MkdirAll(extractDir, 0755); err != nil {
		t.Fatalf("Failed to create extract dir: %v", err)
	}

	if err := untar(archivePath, extractDir); err != nil {
		t.Fatalf("untar() returned error: %v", err)
	}

	extractedPath := filepath.Join(extractDir, "probe")
	if _, err := os.Stat(extractedPath); os.IsNotExist(err) {
		t.Error("untar() should extract 'probe' file")
	}
}

func TestCopyFile(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "probe-copy-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	srcPath := filepath.Join(tempDir, "source.txt")
	dstPath := filepath.Join(tempDir, "dest.txt")

	content := []byte("test content")
	if err := os.WriteFile(srcPath, content, 0644); err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	if err := copyFile(srcPath, dstPath); err != nil {
		t.Fatalf("copyFile() returned error: %v", err)
	}

	dstContent, err := os.ReadFile(dstPath)
	if err != nil {
		t.Fatalf("Failed to read dest file: %v", err)
	}

	if string(dstContent) != string(content) {
		t.Errorf("copyFile() content = %s, want %s", dstContent, content)
	}
}

func TestCopyFileNonexistentSource(t *testing.T) {
	err := copyFile("/nonexistent/file.txt", "/tmp/dest.txt")
	if err == nil {
		t.Error("copyFile() should return error for nonexistent source")
	}
}

func TestGetExecutablePath(t *testing.T) {
	path, err := getExecutablePath()
	if err != nil {
		t.Fatalf("getExecutablePath() returned error: %v", err)
	}

	if path == "" {
		t.Error("getExecutablePath() should return non-empty path")
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("getExecutablePath() returned nonexistent path: %s", path)
	}
}

func TestExtractBinary(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "probe-extract-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	archivePath := filepath.Join(tempDir, "probe.tar.gz")

	file, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("Failed to create archive: %v", err)
	}

	gzw := gzip.NewWriter(file)
	tw := tar.NewWriter(gzw)

	content := []byte("#!/bin/sh\necho 'probe'\n")
	hdr := &tar.Header{
		Name: "probe",
		Mode: 0755,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("Failed to write tar header: %v", err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatalf("Failed to write tar content: %v", err)
	}

	tw.Close()
	gzw.Close()
	file.Close()

	destPath := filepath.Join(tempDir, "extracted-probe")

	if err := extractBinary(archivePath, destPath); err != nil {
		t.Fatalf("extractBinary() returned error: %v", err)
	}

	if _, err := os.Stat(destPath); os.IsNotExist(err) {
		t.Error("extractBinary() should create destination file")
	}
}

func TestExtractBinaryNoProbeInArchive(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "probe-extract-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	archivePath := filepath.Join(tempDir, "empty.tar.gz")

	file, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("Failed to create archive: %v", err)
	}

	gzw := gzip.NewWriter(file)
	tw := tar.NewWriter(gzw)

	hdr := &tar.Header{
		Name: "other-file.txt",
		Mode: 0644,
		Size: 0,
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("Failed to write tar header: %v", err)
	}

	tw.Close()
	gzw.Close()
	file.Close()

	destPath := filepath.Join(tempDir, "extracted-probe")

	err = extractBinary(archivePath, destPath)
	if err == nil {
		t.Error("extractBinary() should return error when 'probe' not in archive")
	}
}

func TestGetHTTPClient(t *testing.T) {
	client := getHTTPClient()
	if client == nil {
		t.Error("getHTTPClient() should return non-nil client")
	}

	if client.Timeout != requestTimeout {
		t.Errorf("getHTTPClient().Timeout = %v, want %v", client.Timeout, requestTimeout)
	}
}
