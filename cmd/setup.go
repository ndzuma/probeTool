package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/fatih/color"
	"github.com/ndzuma/probeTool/internal/agent"
	"github.com/ndzuma/probeTool/internal/paths"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Install probe agent files",
	Long:  "Copies agent files to OS-standard location and installs dependencies",
	Run: func(cmd *cobra.Command, args []string) {
		runSetup()
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

func runSetup() {
	cyan := color.New(color.FgCyan).SprintFunc()
	green := color.New(color.FgGreen).SprintFunc()
	red := color.New(color.FgRed).SprintFunc()

	fmt.Printf("%s Setting up probe agent...\n\n", cyan("⚙️"))

	if err := paths.EnsureAppDirs(); err != nil {
		fmt.Printf("%s Failed to create directories: %v\n", red("❌"), err)
		os.Exit(1)
	}

	agentDir := paths.GetAgentDir()

	fmt.Println("Extracting agent files...")
	if err := agent.Extract(agentDir); err != nil {
		fmt.Printf("%s Failed to extract agent files: %v\n", red("❌"), err)
		os.Exit(1)
	}

	fmt.Println("Installing npm dependencies...")
	npmCmd := exec.Command("npm", "install", "--production")
	npmCmd.Dir = agentDir
	npmCmd.Stdout = os.Stdout
	npmCmd.Stderr = os.Stderr

	if err := npmCmd.Run(); err != nil {
		fmt.Printf("%s Failed to install dependencies: %v\n", red("❌"), err)
		fmt.Println("Make sure Node.js and npm are installed")
		os.Exit(1)
	}

	fmt.Printf("\n%s Probe agent installed successfully!\n", green("✅"))
	fmt.Printf("Agent location: %s\n", agentDir)
	fmt.Printf("\nRun %s to configure providers\n", cyan("probe config add-provider openrouter"))
}
