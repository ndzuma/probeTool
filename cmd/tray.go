package cmd

import (
	"fmt"

	"github.com/ndzuma/probeTool/internal/tray"
	"github.com/spf13/cobra"
)

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
}

func runTray() {
	fmt.Println("Starting probeTool in system tray...")
	fmt.Println("Right-click the tray icon to access menu")
	fmt.Println()

	manager := tray.New()
	manager.Start()
}
