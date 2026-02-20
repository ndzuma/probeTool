package cmd

import (
	"fmt"

	"github.com/ndzuma/probeTool/internal/process"
	"github.com/spf13/cobra"
)

var stopAll bool

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop the probe server",
	Long:  `Stops the running probe dashboard server.`,
	Run: func(cmd *cobra.Command, args []string) {
		runStop()
	},
}

func init() {
	rootCmd.AddCommand(stopCmd)
	stopCmd.Flags().BoolVar(&stopAll, "all", false, "Stop both server and tray")
}

func runStop() {
	if stopAll {
		stopped := false

		if process.IsTrayRunning() {
			if err := process.StopTray(); err != nil {
				fmt.Printf("Failed to stop tray: %v\n", err)
			} else {
				fmt.Println("Tray stopped")
				stopped = true
			}
		}

		if process.IsServerRunning() {
			if err := process.StopServer(); err != nil {
				fmt.Printf("Failed to stop server: %v\n", err)
			} else {
				fmt.Println("Server stopped")
				stopped = true
			}
		}

		if !stopped {
			fmt.Println("No processes running")
		}
		return
	}

	if !process.IsServerRunning() {
		fmt.Println("Server is not running")
		return
	}

	if err := process.StopServer(); err != nil {
		fmt.Printf("Failed to stop server: %v\n", err)
		return
	}

	fmt.Println("Server stopped")
}
