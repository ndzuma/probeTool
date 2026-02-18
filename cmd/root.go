package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/ndzuma/probeTool/internal/db"
	"github.com/ndzuma/probeTool/internal/prober"
	"github.com/ndzuma/probeTool/internal/server"
	"github.com/spf13/cobra"
)

var (
	fullFlag     bool
	quickFlag    bool
	specificPath string
	changesFlag  bool
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
	rootCmd.Flags().BoolVar(&fullFlag, "full", true, "Run a full probe (default)")
	rootCmd.Flags().BoolVar(&quickFlag, "quick", false, "Run a quick probe")
	rootCmd.Flags().StringVar(&specificPath, "specific", "", "Probe a specific path")
	rootCmd.Flags().BoolVar(&changesFlag, "changes", false, "Probe changes only")
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
	if changesFlag {
		probeType = "changes"
	}

	// Get target directory
	target, err := prober.GetTarget(specificPath)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		os.Exit(1)
	}

	// Generate probe ID
	probeID := time.Now().Format("2006-01-02-15-04-05") + "-" + probeType

	// Initialize database (probesDir and dbPath from db package)
	database, err := db.InitDB(db.DBPath())
	if err != nil {
		fmt.Printf("❌ Error initializing database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	// Insert probe record
	err = db.InsertProbe(database, probeID, probeType, target, "")
	if err != nil {
		fmt.Printf("❌ Error creating probe record: %v\n", err)
		os.Exit(1)
	}

	// Start server in background
	go server.StartServer(database)

	// Print progress
	fmt.Println("🔄 Starting probe...")
	fmt.Printf("📦 Target: %s\n", target)
	fmt.Printf("📊 Status: %s\n", probeID)
	fmt.Printf("🚀 View: http://localhost:3030/probes/%s\n", probeID)

	// Simulate probe work
	time.Sleep(1 * time.Second)

	// Mark as completed
	err = db.UpdateProbeStatus(database, probeID, "completed")
	if err != nil {
		fmt.Printf("❌ Error updating probe status: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Probe stored in SQLite")

	// Keep the program running
	select {}
}
