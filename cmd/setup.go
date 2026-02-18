package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Install probe agent files",
	Long:  "Copies agent files to ~/.probe/agent/ and installs dependencies",
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

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("%s Failed to get home directory: %v\n", red("❌"), err)
		os.Exit(1)
	}

	probeDir := filepath.Join(homeDir, ".probe")
	agentDir := filepath.Join(probeDir, "agent")
	// FIX: .claude goes INSIDE agent directory (where SDK expects it with settingSources: ['project'])
	skillsDir := filepath.Join(agentDir, ".claude", "skills")

	// Create directories
	fmt.Println("Creating directories...")
	os.MkdirAll(agentDir, 0755)
	os.MkdirAll(skillsDir, 0755)

	// Check if we're in development (source available)
	devSourceDir := "."
	if _, err := os.Stat("agent/probe-runner.js"); err != nil {
		// Try one level up (if running from cmd/)
		devSourceDir = ".."
	}

	// Copy agent files
	fmt.Println("Copying agent files...")
	agentFiles := []string{"probe-runner.js", "prompts.js", "package.json"}
	for _, file := range agentFiles {
		src := filepath.Join(devSourceDir, "agent", file)
		dst := filepath.Join(agentDir, file)

		if err := copyFile(src, dst); err != nil {
			fmt.Printf("%s Failed to copy %s: %v\n", red("❌"), file, err)
			os.Exit(1)
		}
	}

	// Copy skills
	fmt.Println("Copying skills...")
	skillSrc := filepath.Join(devSourceDir, ".claude", "skills", "security-audit")
	skillDst := filepath.Join(skillsDir, "security-audit")
	if err := copyDir(skillSrc, skillDst); err != nil {
		fmt.Printf("%s Failed to copy skills: %v\n", red("❌"), err)
		os.Exit(1)
	}

	// Install npm dependencies
	fmt.Println("Installing npm dependencies...")
	npmCmd := exec.Command("npm", "install")
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

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

func copyDir(src, dst string) error {
	os.MkdirAll(dst, 0755)

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return err
			}
		} else {
			if err := copyFile(srcPath, dstPath); err != nil {
				return err
			}
		}
	}

	return nil
}
