package tray

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"fyne.io/systray"
	"github.com/ndzuma/probeTool/internal/version"
)

type Manager struct {
	serveCmd     *exec.Cmd
	menuItems    *MenuItems
	dashboardURL string
}

type MenuItems struct {
	openDashboard *systray.MenuItem
	checkUpdates  *systray.MenuItem
	version       *systray.MenuItem
	restart       *systray.MenuItem
	quit          *systray.MenuItem
}

func New() *Manager {
	return &Manager{
		dashboardURL: "http://localhost:37330",
	}
}

func (m *Manager) Start() {
	systray.Run(m.onReady, m.onExit)
}

func (m *Manager) onReady() {
	//systray.SetIcon(assets.Icon)
	systray.SetTitle("Probe")
	systray.SetTooltip("probeTool - Security Scanner")

	m.buildMenu()

	if err := m.startServer(); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		systray.Quit()
		return
	}

	go m.handleMenuActions()
}

func (m *Manager) buildMenu() {
	info := version.GetInfo()

	m.menuItems = &MenuItems{
		openDashboard: systray.AddMenuItem("Open Dashboard", "Open dashboard in browser"),
	}
	systray.AddSeparator()

	m.menuItems.checkUpdates = systray.AddMenuItem("Check for Updates", "Check if new version is available")
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

		case <-m.menuItems.checkUpdates.ClickedCh:
			m.checkForUpdates()

		case <-m.menuItems.restart.ClickedCh:
			m.restartServer()

		case <-m.menuItems.quit.ClickedCh:
			systray.Quit()
			return
		}
	}
}

func (m *Manager) startServer() error {
	fmt.Println("Starting dashboard server...")

	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	m.serveCmd = exec.Command(execPath, "serve", "--quiet")
	m.serveCmd.Stdout = os.Stdout
	m.serveCmd.Stderr = os.Stderr

	if err := m.serveCmd.Start(); err != nil {
		return fmt.Errorf("failed to start serve: %w", err)
	}

	if err := m.waitForServer(30); err != nil {
		if m.serveCmd.Process != nil {
			m.serveCmd.Process.Kill()
		}
		return err
	}

	fmt.Printf("Dashboard running at %s\n", m.dashboardURL)
	systray.SetTooltip("probeTool - Running")
	return nil
}

func (m *Manager) stopServer() error {
	if m.serveCmd == nil || m.serveCmd.Process == nil {
		return nil
	}

	fmt.Println("Stopping dashboard server...")

	if err := m.serveCmd.Process.Signal(os.Interrupt); err != nil {
		return m.serveCmd.Process.Kill()
	}

	m.serveCmd.Wait()
	return nil
}

func (m *Manager) restartServer() {
	fmt.Println("Restarting server...")

	systray.SetTooltip("probeTool - Restarting...")

	if err := m.stopServer(); err != nil {
		fmt.Printf("Warning: %v\n", err)
	}

	if err := m.startServer(); err != nil {
		fmt.Printf("Failed to restart: %v\n", err)
		systray.SetTooltip("probeTool - Server failed to start")
	} else {
		systray.SetTooltip("probeTool - Security Scanner")
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
		fmt.Printf("Please open: %s\n", url)
		return
	}

	if err := cmd.Start(); err != nil {
		fmt.Printf("Failed to open browser: %v\n", err)
		fmt.Printf("Please open: %s\n", url)
	}
}

func (m *Manager) checkForUpdates() {
	fmt.Println("Update check not implemented yet")
	info := version.GetInfo()
	fmt.Printf("Current version: %s\n", info.Version)
}

func (m *Manager) onExit() {
	fmt.Println("\nShutting down...")

	if err := m.stopServer(); err != nil {
		fmt.Printf("Warning during shutdown: %v\n", err)
	}

	fmt.Println("Goodbye!")
}
