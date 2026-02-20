package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/ndzuma/probeTool/internal/process"
	"github.com/ndzuma/probeTool/internal/tray"
	"github.com/spf13/cobra"
)

var trayDaemonMode bool

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
}

func runTray() {
	if trayDaemonMode {
		runTrayForeground()
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

	cmd := exec.Command(execPath, "tray", "--daemon")
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
