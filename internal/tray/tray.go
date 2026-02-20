package tray

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"fyne.io/systray"
	"github.com/ndzuma/probeTool/internal/process"
	"github.com/ndzuma/probeTool/internal/updater"
	"github.com/ndzuma/probeTool/internal/version"
)

const updateCheckInterval = 4 * time.Hour

type Manager struct {
	serveCmd     *exec.Cmd
	menuItems    *MenuItems
	dashboardURL string
	updateInfo   *updater.UpdateInfo
	stopPolling  chan struct{}
}

type MenuItems struct {
	openDashboard *systray.MenuItem
	update        *systray.MenuItem
	version       *systray.MenuItem
	restart       *systray.MenuItem
	quit          *systray.MenuItem
}

func New() *Manager {
	return &Manager{
		dashboardURL: fmt.Sprintf("http://localhost:%s", process.ServerPort),
		stopPolling:  make(chan struct{}),
	}
}

func (m *Manager) Start() {
	systray.Run(m.onReady, m.onExit)
}

func (m *Manager) onReady() {
	systray.SetTitle("Probe")
	systray.SetTooltip("probeTool - Security Scanner")

	m.buildMenu()

	if !process.IsServerRunning() {
		if err := m.startServer(); err != nil {
			systray.Quit()
			return
		}
	}

	go m.handleMenuActions()
	go m.startUpdatePolling()
	go m.checkInitialUpdate()
}

func (m *Manager) buildMenu() {
	info := version.GetInfo()

	m.menuItems = &MenuItems{
		openDashboard: systray.AddMenuItem("Open Dashboard", "Open dashboard in browser"),
	}
	systray.AddSeparator()

	m.menuItems.update = systray.AddMenuItem("Check for Updates", "Check if new version is available")
	systray.AddSeparator()

	versionText := fmt.Sprintf("Version%s%s", strings.Repeat(" ", 15), info.Version)
	m.menuItems.version = systray.AddMenuItem(versionText, "Current version")
	m.menuItems.version.Disable()

	m.menuItems.restart = systray.AddMenuItem("Restart Server", "Restart the dashboard server")
	m.menuItems.quit = systray.AddMenuItem("Quit", "Quit probeTool")
}

func (m *Manager) handleMenuActions() {
	for {
		select {
		case <-m.menuItems.openDashboard.ClickedCh:
			m.openBrowser(m.dashboardURL)

		case <-m.menuItems.update.ClickedCh:
			m.handleUpdateClick()

		case <-m.menuItems.restart.ClickedCh:
			m.restartServer()

		case <-m.menuItems.quit.ClickedCh:
			systray.Quit()
			return
		}
	}
}

func (m *Manager) handleUpdateClick() {
	if m.updateInfo != nil && m.updateInfo.HasUpdate {
		m.runUpdate()
	} else {
		m.checkForUpdatesNow()
	}
}

func (m *Manager) startUpdatePolling() {
	ticker := time.NewTicker(updateCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			m.checkForUpdates()
		case <-m.stopPolling:
			return
		}
	}
}

func (m *Manager) checkInitialUpdate() {
	time.Sleep(2 * time.Second)
	m.checkForUpdates()
}

func (m *Manager) checkForUpdates() {
	systray.SetTooltip("probeTool - Running")

	info, err := updater.CheckForUpdateCached()
	if err != nil {
		return
	}

	m.updateInfo = info

	if info.HasUpdate {
		m.menuItems.update.SetTitle("Update Available (" + info.LatestVersion + ")")
		m.menuItems.update.SetTooltip("Click to install update " + info.LatestVersion)
		systray.SetTooltip("probeTool - Update available: " + info.LatestVersion)
	} else {
		m.menuItems.update.SetTitle("Check for Updates")
		m.menuItems.update.SetTooltip("Check if new version is available")
	}
}

func (m *Manager) checkForUpdatesNow() {
	systray.SetTooltip("probeTool - Checking for updates...")

	info, err := updater.CheckForUpdateWithCache()
	if err != nil {
		systray.SetTooltip("probeTool - Running")
		return
	}

	m.updateInfo = info

	if info.HasUpdate {
		m.menuItems.update.SetTitle("Update Available (" + info.LatestVersion + ")")
		m.menuItems.update.SetTooltip("Click to install update " + info.LatestVersion)
		systray.SetTooltip("probeTool - Update available: " + info.LatestVersion)
	} else {
		m.menuItems.update.SetTitle("Check for Updates")
		m.menuItems.update.SetTooltip("Check if new version is available")
		systray.SetTooltip("probeTool - Running")
	}
}

func (m *Manager) runUpdate() {
	systray.SetTooltip("probeTool - Verifying update...")
	m.menuItems.update.SetTitle("Verifying...")

	info, err := updater.CheckForUpdate()
	if err != nil {
		systray.SetTooltip("probeTool - Update check failed")
		m.menuItems.update.SetTitle("Check Failed - Retry")
		return
	}

	m.updateInfo = info

	if !info.HasUpdate || info.DownloadURL == "" {
		systray.SetTooltip("probeTool - Running")
		m.menuItems.update.SetTitle("Check for Updates")
		m.menuItems.update.SetTooltip("Check if new version is available")
		return
	}

	systray.SetTooltip("probeTool - Updating...")
	m.menuItems.update.SetTitle("Updating...")
	m.menuItems.update.Disable()

	err = updater.DownloadAndInstall(info.DownloadURL)
	if err != nil {
		errStr := err.Error()
		if strings.Contains(errStr, "permission denied") || strings.Contains(errStr, "Permission denied") {
			systray.SetTooltip("probeTool - Needs sudo to update")
			m.menuItems.update.SetTitle("Run: sudo probe update")
			m.menuItems.update.SetTooltip("Update requires administrator privileges")
			m.menuItems.update.Enable()
			return
		}
		systray.SetTooltip("probeTool - Update failed")
		m.menuItems.update.SetTitle("Update Failed - Retry")
		m.menuItems.update.Enable()
		return
	}

	m.menuItems.update.SetTitle("Updated - Restart to Apply")
	m.menuItems.update.Disable()
	systray.SetTooltip("probeTool - Restart to complete update")
}

func (m *Manager) startServer() error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	m.serveCmd = exec.Command(execPath, "serve", "--quiet", "--daemon")
	m.serveCmd.Stdout = nil
	m.serveCmd.Stderr = nil

	if err := m.serveCmd.Start(); err != nil {
		return fmt.Errorf("failed to start serve: %w", err)
	}

	if err := m.waitForServer(30); err != nil {
		if m.serveCmd.Process != nil {
			m.serveCmd.Process.Kill()
		}
		return err
	}

	systray.SetTooltip("probeTool - Running")
	return nil
}

func (m *Manager) stopServer() error {
	if m.serveCmd == nil || m.serveCmd.Process == nil {
		return nil
	}

	if err := m.serveCmd.Process.Signal(os.Interrupt); err != nil {
		return m.serveCmd.Process.Kill()
	}

	m.serveCmd.Wait()
	return nil
}

func (m *Manager) restartServer() {
	systray.SetTooltip("probeTool - Restarting...")

	m.stopServer()

	if err := m.startServer(); err != nil {
		systray.SetTooltip("probeTool - Server failed to start")
	} else {
		systray.SetTooltip("probeTool - Running")
	}
}

func (m *Manager) openBrowser(url string) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return
	}

	cmd.Start()
}

func (m *Manager) onExit() {
	close(m.stopPolling)
	m.stopServer()
	process.RemoveTrayPID()
}
