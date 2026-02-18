package prober

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/ndzuma/probeTool/internal/config"
)

// Color definitions
var (
	cyan   = color.New(color.FgCyan).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
	blue   = color.New(color.FgBlue).SprintFunc()
)

type ProbeArgs struct {
	Type     string // full (only full for now)
	Provider string // Optional: override provider
	Model    string // Optional: override model
}

func RunProbe(ctx context.Context, args ProbeArgs) (string, error) {
	// Get current directory (repo being audited)
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return "", fmt.Errorf("config load failed: %w\nRun: probe config add-provider openrouter", err)
	}

	// Determine provider (default to openrouter)
	provider := args.Provider
	if provider == "" {
		if cfg.Default != "" {
			provider = cfg.Default
		} else {
			provider = "openrouter"
		}
	}

	// Get provider config
	providerCfg, ok := cfg.Providers[provider]
	if !ok {
		return "", fmt.Errorf("provider '%s' not configured\nRun: probe config add-provider %s", provider, provider)
	}

	if providerCfg.APIKey == "" {
		return "", fmt.Errorf("API key missing for provider '%s'\nRun: probe config set-key %s <key>", provider, provider)
	}

	// Determine model
	model := args.Model
	if model == "" {
		model = providerCfg.DefaultModel
		if model == "" {
			model = "anthropic/claude-3.5-haiku" // Default to Haiku 4.5
		}
	}

	// Generate probe ID
	id := fmt.Sprintf("%s-%s", time.Now().Format("2006-01-02-150405"), args.Type)
	probesDir := "./probes"
	os.MkdirAll(probesDir, 0755)
	mdPath := filepath.Join(probesDir, id+".md")

	// Resolve absolute path
	absPath, _ := filepath.Abs(mdPath)

	fmt.Printf("%s Starting probe audit...\n", cyan("🔍"))
	fmt.Printf("  Target: %s\n", cwd)
	fmt.Printf("  Provider: %s\n", provider)
	fmt.Printf("  Model: %s\n", model)
	fmt.Println()

	// Build Node.js command
	cmd := exec.CommandContext(ctx, "node",
		"agent/probe-runner.js",
		"--target="+cwd,
		"--out="+absPath,
		"--model="+model,
	)

	// CRITICAL: Set OpenRouter env vars per docs
	// https://openrouter.ai/docs/guides/community/anthropic-agent-sdk
	cmd.Env = append(os.Environ(),
		"ANTHROPIC_BASE_URL=https://openrouter.ai/api",
		"ANTHROPIC_AUTH_TOKEN="+providerCfg.APIKey,
		"ANTHROPIC_API_KEY=", // Must be explicitly empty
	)

	// Set working directory to agent/
	cmd.Dir = "."

	// Stream stdout
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	// Progress handler
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()

			if strings.HasPrefix(line, "PROGRESS:") {
				parts := strings.Split(strings.TrimPrefix(line, "PROGRESS:"), ":")
				stage := parts[0]

				switch stage {
				case "init":
					fmt.Printf("%s Initializing audit engine...\n", cyan("⚙️ "))
				case "reading_files":
					fmt.Printf("%s Reading codebase...\n", blue("📂"))
				case "security":
					fmt.Printf("%s Analyzing security...\n", red("🔒"))
				case "performance":
					fmt.Printf("%s Checking performance...\n", yellow("⚡"))
				case "architecture":
					fmt.Printf("%s Reviewing architecture...\n", blue("🏗️ "))
				case "quality":
					fmt.Printf("%s Assessing code quality...\n", cyan("✨"))
				case "finalizing":
					fmt.Printf("%s Generating report...\n", green("📝"))
				}
			} else if strings.HasPrefix(line, "SUCCESS:") {
				fmt.Println()
				fmt.Printf("%s Audit complete!\n", green("✅"))
			} else if strings.HasPrefix(line, "ERROR:") {
				fmt.Printf("%s %s\n", red("❌"), strings.TrimPrefix(line, "ERROR:"))
			}
		}
	}()

	// Error handler
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			fmt.Fprintf(os.Stderr, "%s %s\n", red("Error:"), scanner.Text())
		}
	}()

	// Run
	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start probe: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return "", fmt.Errorf("probe failed: %w", err)
	}

	// Save to DB (uncomment when DB ready)
	// db.SaveProbe(id, args.Type, cwd, absPath, provider, model)

	url := fmt.Sprintf("http://localhost:3030/probes/%s", id)
	fmt.Println()
	fmt.Printf("%s View assessment: %s\n", green("🔗"), cyan(url))

	return url, nil
}
