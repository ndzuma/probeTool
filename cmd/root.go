package cmd

import (
	"fmt"
	"os"
	"time"

	"audit-tool/internal/auditor"
	"audit-tool/internal/db"
	"audit-tool/internal/server"
	"github.com/spf13/cobra"
)

var (
	fullFlag     bool
	quickFlag    bool
	specificPath string
	changesFlag  bool
)

var rootCmd = &cobra.Command{
	Use:   "audit",
	Short: "Audit tool for code analysis",
	Long:  `A CLI tool to perform audits on codebases and view results via web interface.`,
	Run: func(cmd *cobra.Command, args []string) {
		runAudit()
	},
}

func init() {
	rootCmd.Flags().BoolVar(&fullFlag, "full", true, "Run a full audit (default)")
	rootCmd.Flags().BoolVar(&quickFlag, "quick", false, "Run a quick audit")
	rootCmd.Flags().StringVar(&specificPath, "specific", "", "Audit a specific path")
	rootCmd.Flags().BoolVar(&changesFlag, "changes", false, "Audit changes only")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runAudit() {
	// Determine audit type
	auditType := "full"
	if quickFlag {
		auditType = "quick"
	}
	if changesFlag {
		auditType = "changes"
	}

	// Get target directory
	target, err := auditor.GetTarget(specificPath)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		os.Exit(1)
	}

	// Generate audit ID
	auditID := time.Now().Format("2006-01-02-15-04-05") + "-" + auditType

	// Initialize database
	database, err := db.InitDB("./audits/audits.db")
	if err != nil {
		fmt.Printf("❌ Error initializing database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	// Insert audit record
	err = db.InsertAudit(database, auditID, auditType, target, "")
	if err != nil {
		fmt.Printf("❌ Error creating audit record: %v\n", err)
		os.Exit(1)
	}

	// Start server in background
	go server.StartServer(database)

	// Print progress
	fmt.Println("🔄 Starting audit...")
	fmt.Printf("📦 Target: %s\n", target)
	fmt.Printf("📊 Status: %s\n", auditID)
	fmt.Printf("🚀 View: http://localhost:3030/audits/%s\n", auditID)

	// Simulate audit work
	time.Sleep(1 * time.Second)

	// Mark as completed
	err = db.UpdateAuditStatus(database, auditID, "completed")
	if err != nil {
		fmt.Printf("❌ Error updating audit status: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Audit stored in SQLite")

	// Keep the program running
	select {}
}
