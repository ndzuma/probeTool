package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ndzuma/probeTool/internal/db"
	"github.com/ndzuma/probeTool/internal/paths"
	"github.com/ndzuma/probeTool/internal/prober"
	"github.com/ndzuma/probeTool/internal/process"
	"github.com/ndzuma/probeTool/internal/version"
	"github.com/spf13/cobra"
)

var (
	fullFlag     bool
	quickFlag    bool
	modelFlag    string
	verboseFlag  bool
	versionFlag  bool
	overrideFlag bool
)

var rootCmd = &cobra.Command{
	Use:   "probe",
	Short: "Probe tool for code analysis",
	Long:  `A CLI tool to perform probes on codebases and view results via web interface.`,
	Run: func(cmd *cobra.Command, args []string) {
		if versionFlag {
			info := version.GetInfo()
			fmt.Println(info.String())
			return
		}
		runProbe()
	},
}

func init() {
	rootCmd.Flags().BoolVar(&fullFlag, "full", false, "Run a full probe (default)")
	rootCmd.Flags().BoolVar(&quickFlag, "quick", false, "Run a quick probe")
	rootCmd.Flags().StringVar(&modelFlag, "model", "", "Override the default model")
	rootCmd.Flags().BoolVar(&verboseFlag, "verbose", false, "Enable verbose output")
	rootCmd.Flags().BoolVarP(&versionFlag, "version", "v", false, "Show version")
	rootCmd.Flags().BoolVarP(&overrideFlag, "override", "o", false, "Run probe without server running")

	rootCmd.PreRun = func(cmd *cobra.Command, args []string) {
		if !fullFlag && !quickFlag {
			fullFlag = true
		}
	}
}

func Execute() {
	if paths.NeedsMigration() {
		fmt.Println("First run with new version - migrating config...")
		if err := paths.Migrate(); err != nil {
			fmt.Printf("Migration warning: %v\n", err)
			fmt.Println("You can run 'probe migrate' manually if needed")
		}
		fmt.Println()
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func runProbe() {
	if !process.IsServerRunning() && !overrideFlag {
		fmt.Println("Server is not running.")
		fmt.Println()
		fmt.Println("Start the server with:")
		fmt.Println("  probe serve --quiet")
		fmt.Println()
		fmt.Println("Or run in system tray:")
		fmt.Println("  probe tray")
		fmt.Println()
		fmt.Println("To run probe without the server, use --override or -o")
		os.Exit(1)
	}

	probeType := "full"
	if quickFlag {
		probeType = "quick"
	}

	database, err := db.InitDB(db.DBPath())
	if err != nil {
		fmt.Printf("Error initializing database: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
		os.Exit(0)
	}()

	args := prober.ProbeArgs{
		Type:    probeType,
		Model:   modelFlag,
		Verbose: verboseFlag,
	}

	probeID, err := prober.RunProbe(ctx, args)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Security audit complete!")
	if process.IsServerRunning() {
		fmt.Printf("View: http://localhost:%s/probes/%s\n", process.ServerPort, probeID)
	}
}
