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

var (
	cyan   = color.New(color.FgCyan).SprintFunc()
	green  = color.New(color.FgGreen).SprintFunc()
	yellow = color.New(color.FgYellow).SprintFunc()
	red    = color.New(color.FgRed).SprintFunc()
	blue   = color.New(color.FgBlue).SprintFunc()
)

type ProbeArgs struct {
	Type     string
	Provider string
	Model    string
	Verbose  bool
}

func getAgentPath() (string, error) {
	// Check if agent is installed in ~/.probe/agent/
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	agentDir := filepath.Join(homeDir, ".probe", "agent")
	agentScript := filepath.Join(agentDir, "probe-runner.js")

	if _, err := os.Stat(agentScript); os.IsNotExist(err) {
		return "", fmt.Errorf("agent not installed. Run: probe setup")
	}

	return agentScript, nil
}

func RunProbe(ctx context.Context, args ProbeArgs) (string, error) {
	// Get current directory (repo being audited)
	cwd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get working directory: %w", err)
	}

	// Get agent script path
	agentScript, err := getAgentPath()
	if err != nil {
		return "", err
	}

	// Load config
	cfg, err := config.Load()
	if err != nil {
		return "", fmt.Errorf("config load failed: %w\nRun: probe config add-provider openrouter", err)
	}

	// Determine provider
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
			model = "anthropic/claude-3.5-haiku"
		}
	}

	// Generate probe ID
	id := fmt.Sprintf("%s-%s", time.Now().Format("2006-01-02-150405"), args.Type)

	// Probes directory
	homeDir, _ := os.UserHomeDir()
	probesDir := filepath.Join(homeDir, ".probe", "probes")
	os.MkdirAll(probesDir, 0755)
	mdPath := filepath.Join(probesDir, id+".md")

	absPath, _ := filepath.Abs(mdPath)

	fmt.Printf("%s Starting probe audit...\n", cyan("🔍"))
	fmt.Printf("  Target: %s\n", cwd)
	fmt.Printf("  Provider: %s\n", provider)
	fmt.Printf("  Model: %s\n", model)
	fmt.Println()

	cmd := exec.CommandContext(ctx, "node",
		agentScript,
		"--target="+cwd,
		"--out="+absPath,
		"--model="+model,
		"--verbose="+fmt.Sprintf("%t", args.Verbose),
	)

	// Set OpenRouter env vars
	cmd.Env = append(os.Environ(),
		"ANTHROPIC_BASE_URL=https://openrouter.ai/api",
		"ANTHROPIC_AUTH_TOKEN="+providerCfg.APIKey,
		"ANTHROPIC_API_KEY=",
	)

	cmd.Dir = cwd

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
					fmt.Printf("%s Initializing security scanner...\n", cyan("⚙️ "))
				case "reading_files":
					fmt.Printf("%s Scanning codebase...\n", blue("📂"))
				case "critical":
					fmt.Printf("%s Analyzing critical vulnerabilities...\n", red("🔴"))
				case "high":
					fmt.Printf("%s Checking high severity issues...\n", yellow("🟠"))
				case "medium":
					fmt.Printf("%s Reviewing medium risks...\n", yellow("🟡"))
				case "finalizing":
					fmt.Printf("%s Compiling security report...\n", green("📝"))
				}
			} else if strings.HasPrefix(line, "VERBOSE:") {
				// Handle verbose logs
				if args.Verbose {
					msg := strings.TrimPrefix(line, "VERBOSE:")
					fmt.Printf("%s %s\n", blue("🔍"), msg)
				}
			} else if strings.HasPrefix(line, "SUCCESS:") {
				fmt.Println()
				fmt.Printf("%s Security audit complete!\n", green("✅"))
			} else if strings.HasPrefix(line, "ERROR:") {
				fmt.Printf("%s %s\n", red("❌"), strings.TrimPrefix(line, "ERROR:"))
			}
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			fmt.Fprintf(os.Stderr, "%s %s\n", red("Error:"), scanner.Text())
		}
	}()

	if err := cmd.Start(); err != nil {
		return "", fmt.Errorf("failed to start probe: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		return "", fmt.Errorf("probe failed: %w", err)
	}

	url := fmt.Sprintf("http://localhost:3030/probes/%s", id)
	fmt.Println()
	fmt.Printf("%s View assessment: %s\n", green("🔗"), cyan(url))
	fmt.Printf("%s Report saved: %s\n", green("📄"), absPath)

	return url, nil
}
