package cmd

import (
	"database/sql"
	"fmt"

	"github.com/ndzuma/probeTool/internal/db"
	"github.com/ndzuma/probeTool/internal/process"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show probe server and report status",
	Long:  `Displays information about the running server and probe reports.`,
	Run: func(cmd *cobra.Command, args []string) {
		runStatus()
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus() {
	fmt.Println("probeTool Status")
	fmt.Println("================")
	fmt.Println()

	if process.IsServerRunning() {
		pid, _ := process.ReadServerPID()
		fmt.Printf("Server:     Running (PID: %d)\n", pid)
		fmt.Printf("Dashboard:  http://localhost:%s\n", process.ServerPort)
	} else {
		fmt.Println("Server:     Not running")
	}

	if process.IsTrayRunning() {
		pid, _ := process.ReadTrayPID()
		fmt.Printf("Tray:       Running (PID: %d)\n", pid)
	} else {
		fmt.Println("Tray:       Not running")
	}

	fmt.Println()

	database, err := db.InitDB(db.DBPath())
	if err != nil {
		fmt.Printf("Reports:    Database unavailable\n")
		return
	}
	defer database.Close()

	probeCount := getProbeCount(database)
	completedCount := getCompletedProbeCount(database)
	findingCount := getFindingCount(database)

	fmt.Printf("Reports:    %d total (%d completed)\n", probeCount, completedCount)
	fmt.Printf("Findings:   %d total\n", findingCount)
}

func getProbeCount(database *sql.DB) int {
	var count int
	database.QueryRow("SELECT COUNT(*) FROM probes").Scan(&count)
	return count
}

func getCompletedProbeCount(database *sql.DB) int {
	var count int
	database.QueryRow("SELECT COUNT(*) FROM probes WHERE status = ?", "completed").Scan(&count)
	return count
}

func getFindingCount(database *sql.DB) int {
	var count int
	database.QueryRow("SELECT COUNT(*) FROM findings").Scan(&count)
	return count
}
