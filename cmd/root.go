package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ndzuma/probeTool/internal/db"
	"github.com/ndzuma/probeTool/internal/prober"
	"github.com/ndzuma/probeTool/internal/server"
	"github.com/spf13/cobra"
)

var (
	fullFlag  bool
	quickFlag bool
	modelFlag string
)

var rootCmd = &cobra.Command{
	Use:   "probe",
	Short: "Probe tool for code analysis",
	Long:  `A CLI tool to perform probes on codebases and view results via web interface.`,
	Run: func(cmd *cobra.Command, args []string) {
		runProbe()
	},
}

func init() {
	rootCmd.Flags().BoolVar(&fullFlag, "full", false, "Run a full probe (default)")
	rootCmd.Flags().BoolVar(&quickFlag, "quick", false, "Run a quick probe")
	rootCmd.Flags().StringVar(&modelFlag, "model", "", "Override the default model")

	// Make --full the default if no other flag is set
	rootCmd.PreRun = func(cmd *cobra.Command, args []string) {
		if !fullFlag && !quickFlag {
			fullFlag = true
		}
	}
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runProbe() {
	// Determine probe type
	probeType := "full"
	if quickFlag {
		probeType = "quick"
	}

	// Initialize database
	database, err := db.InitDB(db.DBPath())
	if err != nil {
		fmt.Printf("❌ Error initializing database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	// Start server in background
	go server.StartServer(database)

	// Setup context with signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
		os.Exit(0)
	}()

	// Run probe with new API
	args := prober.ProbeArgs{
		Type:  probeType,
		Model: modelFlag,
	}

	_, err = prober.RunProbe(ctx, args)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		os.Exit(1)
	}

	// Keep the program running
	select {}
}
