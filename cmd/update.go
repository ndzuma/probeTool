package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/ndzuma/probeTool/internal/updater"
	"github.com/spf13/cobra"
)

var (
	checkOnly bool
	forceYes  bool
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update probe to the latest version",
	Long: `Check for updates and install the latest version from GitHub.

Your configuration and data are preserved during the update process.
Data is stored separately from the binary in your application directory.`,
	Run: func(cmd *cobra.Command, args []string) {
		runUpdate()
	},
}

func init() {
	rootCmd.AddCommand(updateCmd)
	updateCmd.Flags().BoolVar(&checkOnly, "check", false, "Only check for updates, don't install")
	updateCmd.Flags().BoolVarP(&forceYes, "yes", "y", false, "Skip confirmation prompt")
}

func runUpdate() {
	bold := color.New(color.Bold).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()

	fmt.Printf("%s Checking for updates...\n\n", yellow("🔍"))

	info, err := updater.CheckForUpdateWithCache()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error checking for updates: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("  %s %s\n", bold("Current version:"), info.CurrentVersion)
	fmt.Printf("  %s %s\n", bold("Latest version:"), info.LatestVersion)
	fmt.Println()

	if !info.HasUpdate {
		fmt.Printf("%s You're already on the latest version!\n", color.GreenString("✅"))
		return
	}

	if checkOnly {
		fmt.Printf("%s An update is available.\n", yellow("⚠"))
		fmt.Printf("\nRun 'probe update' to install the update.\n")
		fmt.Printf("Release notes: %s\n", cyan(info.ReleasePageURL))
		return
	}

	fmt.Printf("%s A new version is available!\n\n", yellow("⚠"))

	if info.ReleaseNotes != "" {
		fmt.Printf("%s\n", bold("Release Notes:"))
		lines := strings.Split(info.ReleaseNotes, "\n")
		maxLines := 10
		for i, line := range lines {
			if i >= maxLines {
				fmt.Printf("  ... (see %s for more)\n", cyan(info.ReleasePageURL))
				break
			}
			if strings.TrimSpace(line) != "" {
				fmt.Printf("  %s\n", line)
			}
		}
		fmt.Println()
	}

	if !forceYes {
		fmt.Printf("Do you want to update? [Y/n]: ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response == "n" || response == "no" {
			fmt.Println("Update cancelled.")
			return
		}
	}

	if info.DownloadURL == "" {
		fmt.Fprintf(os.Stderr, "No download available for your platform\n")
		fmt.Printf("Please download manually from: %s\n", cyan(info.ReleasePageURL))
		os.Exit(1)
	}

	fmt.Printf("%s Verifying latest version before update...\n", yellow("🔍"))
	latestInfo, err := updater.CheckForUpdate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error verifying update: %v\n", err)
		os.Exit(1)
	}

	if latestInfo.DownloadURL == "" {
		fmt.Fprintf(os.Stderr, "No download available for your platform\n")
		fmt.Printf("Please download manually from: %s\n", cyan(latestInfo.ReleasePageURL))
		os.Exit(1)
	}

	if err := updater.DownloadAndInstall(latestInfo.DownloadURL); err != nil {
		fmt.Fprintf(os.Stderr, "\nUpdate failed: %v\n", err)
		fmt.Printf("\nYou can download manually from: %s\n", cyan(latestInfo.ReleasePageURL))
		os.Exit(1)
	}
}
