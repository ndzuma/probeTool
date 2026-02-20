package cmd

import (
	"testing"

	"github.com/ndzuma/probeTool/internal/config"
	"github.com/ndzuma/probeTool/internal/process"
)

func TestRootCommandExists(t *testing.T) {
	// Verify root command is initialized
	if rootCmd == nil {
		t.Error("rootCmd should be initialized")
	}
	if rootCmd.Use != "probe" {
		t.Errorf("rootCmd.Use = %v, want probe", rootCmd.Use)
	}
}

func TestConfigCommandExists(t *testing.T) {
	// Verify config command is initialized
	if configCmd == nil {
		t.Error("configCmd should be initialized")
	}
	if configCmd.Use != "config" {
		t.Errorf("configCmd.Use = %v, want config", configCmd.Use)
	}
}

func TestConfigSubcommands(t *testing.T) {
	// Check that all config subcommands exist
	expectedCommands := map[string]bool{
		"providers":    false,
		"add-provider": false,
		"set-key":      false,
		"add-model":    false,
		"set-default":  false,
		"list":         false,
	}

	for _, cmd := range configCmd.Commands() {
		name := cmd.Name()
		if _, exists := expectedCommands[name]; exists {
			expectedCommands[name] = true
		}
	}

	for name, found := range expectedCommands {
		if !found {
			t.Errorf("Config subcommand '%s' not found", name)
		}
	}
}

func TestServeCommandExists(t *testing.T) {
	// Verify serve command exists
	cmd, _, err := rootCmd.Find([]string{"serve"})
	if err != nil {
		t.Errorf("serve command not found: %v", err)
	}
	if cmd == nil {
		t.Error("serve command should exist")
	}
}

func TestSetupCommandExists(t *testing.T) {
	// Verify setup command exists
	cmd, _, err := rootCmd.Find([]string{"setup"})
	if err != nil {
		t.Errorf("setup command not found: %v", err)
	}
	if cmd == nil {
		t.Error("setup command should exist")
	}
}

func TestMigrateCommandExists(t *testing.T) {
	// Verify migrate command exists
	cmd, _, err := rootCmd.Find([]string{"migrate"})
	if err != nil {
		t.Errorf("migrate command not found: %v", err)
	}
	if cmd == nil {
		t.Error("migrate command should exist")
	}
}

func TestCleanCommandExists(t *testing.T) {
	// Verify clean command exists
	cmd, _, err := rootCmd.Find([]string{"clean"})
	if err != nil {
		t.Errorf("clean command not found: %v", err)
	}
	if cmd == nil {
		t.Error("clean command should exist")
	}
}

func TestProviderStruct(t *testing.T) {
	// Test Provider struct fields
	provider := config.Provider{
		Name:         "test",
		BaseURL:      "https://test.example.com",
		APIKey:       "test-key",
		Models:       []string{"model1", "model2"},
		DefaultModel: "model1",
	}

	if provider.Name != "test" {
		t.Errorf("Provider.Name = %v, want test", provider.Name)
	}
	if provider.BaseURL != "https://test.example.com" {
		t.Errorf("Provider.BaseURL = %v, want https://test.example.com", provider.BaseURL)
	}
	if len(provider.Models) != 2 {
		t.Errorf("len(Provider.Models) = %v, want 2", len(provider.Models))
	}
}

func TestConfigStruct(t *testing.T) {
	// Test Config struct
	cfg := config.Config{
		Providers: map[string]config.Provider{
			"test": {
				Name:    "test",
				BaseURL: "https://test.example.com",
			},
		},
		Default: "test",
	}

	if len(cfg.Providers) != 1 {
		t.Errorf("len(Config.Providers) = %v, want 1", len(cfg.Providers))
	}
	if cfg.Default != "test" {
		t.Errorf("Config.Default = %v, want test", cfg.Default)
	}
}

func TestRootFlags(t *testing.T) {
	flags := rootCmd.Flags()

	if flags.Lookup("full") == nil {
		t.Error("rootCmd should have --full flag")
	}

	if flags.Lookup("quick") == nil {
		t.Error("rootCmd should have --quick flag")
	}

	if flags.Lookup("model") == nil {
		t.Error("rootCmd should have --model flag")
	}

	if flags.Lookup("verbose") == nil {
		t.Error("rootCmd should have --verbose flag")
	}

	if flags.Lookup("override") == nil {
		t.Error("rootCmd should have --override flag")
	}
}

func TestStopCommandExists(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"stop"})
	if err != nil {
		t.Errorf("stop command not found: %v", err)
	}
	if cmd == nil {
		t.Error("stop command should exist")
	}
}

func TestTrayCommandExists(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"tray"})
	if err != nil {
		t.Errorf("tray command not found: %v", err)
	}
	if cmd == nil {
		t.Error("tray command should exist")
	}
}

func TestServeQuietFlag(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"serve"})
	if err != nil {
		t.Errorf("serve command not found: %v", err)
	}

	if cmd.Flags().Lookup("quiet") == nil {
		t.Error("serve command should have --quiet flag")
	}
}

func TestStopAllFlag(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"stop"})
	if err != nil {
		t.Errorf("stop command not found: %v", err)
	}

	if cmd.Flags().Lookup("all") == nil {
		t.Error("stop command should have --all flag")
	}
}

func TestServerConstants(t *testing.T) {
	if process.ServerPort != "37330" {
		t.Errorf("ServerPort = %v, want 37330", process.ServerPort)
	}
	if process.NextJSPort != "37331" {
		t.Errorf("NextJSPort = %v, want 37331", process.NextJSPort)
	}
}

func TestStatusCommandExists(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"status"})
	if err != nil {
		t.Errorf("status command not found: %v", err)
	}
	if cmd == nil {
		t.Error("status command should exist")
	}
}

func TestTrayWSLCheckFlag(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"tray"})
	if err != nil {
		t.Errorf("tray command not found: %v", err)
	}

	if cmd.Flags().Lookup("skip-wsl-check") == nil {
		t.Error("tray command should have --skip-wsl-check flag")
	}
}

func TestUpdateCommandExists(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"update"})
	if err != nil {
		t.Errorf("update command not found: %v", err)
	}
	if cmd == nil {
		t.Error("update command should exist")
	}
	if cmd.Use != "update" {
		t.Errorf("update command Use = %v, want update", cmd.Use)
	}
}

func TestUpdateCommandFlags(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"update"})
	if err != nil {
		t.Fatalf("update command not found: %v", err)
	}

	if cmd.Flags().Lookup("check") == nil {
		t.Error("update command should have --check flag")
	}

	if cmd.Flags().Lookup("yes") == nil {
		t.Error("update command should have --yes flag")
	}
}

func TestUpdateCommandShortFlag(t *testing.T) {
	cmd, _, err := rootCmd.Find([]string{"update"})
	if err != nil {
		t.Fatalf("update command not found: %v", err)
	}

	yesFlag := cmd.Flags().Lookup("yes")
	if yesFlag == nil {
		t.Fatal("update command should have --yes flag")
	}

	if yesFlag.Shorthand != "y" {
		t.Errorf("update --yes shorthand = %v, want y", yesFlag.Shorthand)
	}
}
