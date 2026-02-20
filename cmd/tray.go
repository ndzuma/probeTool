package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/ndzuma/probeTool/internal/process"
	"github.com/ndzuma/probeTool/internal/tray"
	"github.com/ndzuma/probeTool/internal/wsl"
	"github.com/spf13/cobra"
)

var trayDaemonMode bool
var skipWSLCheck bool

var trayCmd = &cobra.Command{
	Use:   "tray",
	Short: "Run probeTool in system tray",
	Long:  `Starts probeTool as a background system tray application with menu controls.`,
	Run: func(cmd *cobra.Command, args []string) {
		runTray()
	},
}

func init() {
	rootCmd.AddCommand(trayCmd)
	trayCmd.Flags().BoolVar(&trayDaemonMode, "daemon", false, "Run as daemon (internal use)")
	trayCmd.Flags().BoolVar(&skipWSLCheck, "skip-wsl-check", false, "Skip WSL detection check")
}

func runTray() {
	if trayDaemonMode {
		runTrayForeground()
		return
	}

	if !skipWSLCheck && wsl.IsWSL() {
		handleWSLDetection()
		return
	}

	if wsl.GetTrayDisabled() {
		fmt.Println("Tray is disabled for WSL environment.")
		fmt.Println("Use 'probe serve --quiet' to start the server.")
		fmt.Println()
		fmt.Println("To re-enable tray, run: probe config set-tray true")
		return
	}

	if process.IsTrayRunning() {
		fmt.Println("Tray is already running")
		return
	}

	execPath, err := os.Executable()
	if err != nil {
		fmt.Printf("Error getting executable path: %v\n", err)
		os.Exit(1)
	}

	cmd := exec.Command(execPath, "tray", "--daemon", "--skip-wsl-check")
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	startDaemon(cmd)

	if err := cmd.Start(); err != nil {
		fmt.Printf("Failed to start tray: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Starting probeTool in system tray...")
	fmt.Println("Right-click the tray icon to access menu")
}

func handleWSLDetection() {
	fmt.Println("WSL environment detected!")
	fmt.Println()
	fmt.Println("System tray support in WSL requires additional setup.")
	fmt.Println("Choose an option:")
	fmt.Println()
	fmt.Println("  1) Run with tray (requires WSLg or X server)")
	fmt.Println("  2) Run without tray (recommended for WSL)")
	fmt.Println()

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter choice [1-2]: ")
	input, _ := reader.ReadString('\n')
	choice := strings.TrimSpace(input)

	switch choice {
	case "1":
		wsl.SetTrayDisabled(false)
		fmt.Println()
		fmt.Println("Starting with tray enabled...")
		skipWSLCheck = true
		runTray()
	case "2":
		wsl.SetTrayDisabled(true)
		fmt.Println()
		fmt.Println("Tray disabled. Starting server only...")
		startServerOnly()
	default:
		fmt.Println("Invalid choice. Tray disabled by default.")
		wsl.SetTrayDisabled(true)
		startServerOnly()
	}
}

func startServerOnly() {
	if process.IsServerRunning() {
		fmt.Println("Server is already running")
		return
	}

	execPath, err := os.Executable()
	if err != nil {
		fmt.Printf("Error getting executable path: %v\n", err)
		os.Exit(1)
	}

	cmd := exec.Command(execPath, "serve", "--quiet", "--daemon")
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	startDaemon(cmd)

	if err := cmd.Start(); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Server started in background")
	fmt.Println("Use 'probe stop' to stop the server")
}

func runTrayForeground() {
	fmt.Println("Starting probeTool in system tray...")
	fmt.Println("Right-click the tray icon to access menu")
	fmt.Println()

	if err := process.WriteTrayPID(os.Getpid()); err != nil {
		fmt.Printf("Warning: could not write tray PID file: %v\n", err)
	}

	manager := tray.New()
	manager.Start()
}
