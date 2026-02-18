package cmd

import (
	"fmt"
	"os"

	"github.com/ndzuma/probeTool/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(configCmd)

	// Add subcommands
	configCmd.AddCommand(providersCmd)
	configCmd.AddCommand(addProviderCmd)
	configCmd.AddCommand(setKeyCmd)
	configCmd.AddCommand(addModelCmd)
	configCmd.AddCommand(setDefaultCmd)
	configCmd.AddCommand(listConfigCmd)
}

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage probe configuration",
	Long:  `Configure providers, API keys, and default settings for probe.`,
}

var providersCmd = &cobra.Command{
	Use:   "providers",
	Short: "List available providers",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("❌ Error loading config: %v\n", err)
			os.Exit(1)
		}

		if len(cfg.ListProviders()) == 0 {
			fmt.Println("No providers configured.")
			fmt.Println("Use 'probe config add-provider <name>' to add one.")
			return
		}

		fmt.Println("Configured Providers:")
		for _, name := range cfg.ListProviders() {
			provider, _ := cfg.GetProvider(name)
			marker := "  "
			if name == cfg.Default {
				marker = "* "
			}

			apiKeyStatus := "❌"
			if provider.APIKey != "" {
				apiKeyStatus = "✓"
			}

			fmt.Printf("%s%s (Base: %s, API: %s)\n", marker, name, provider.BaseURL, apiKeyStatus)
		}
	},
}

var addProviderCmd = &cobra.Command{
	Use:   "add-provider <name>",
	Short: "Add a new provider interactively",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		name := args[0]

		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("❌ Error loading config: %v\n", err)
			os.Exit(1)
		}

		if err := cfg.AddProvider(name); err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Provider '%s' added successfully\n", name)
	},
}

var setKeyCmd = &cobra.Command{
	Use:   "set-key <provider> <key>",
	Short: "Set API key for a provider",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		providerName := args[0]
		apiKey := args[1]

		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("❌ Error loading config: %v\n", err)
			os.Exit(1)
		}

		if err := cfg.SetAPIKey(providerName, apiKey); err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ API key set for provider '%s'\n", providerName)
	},
}

var addModelCmd = &cobra.Command{
	Use:   "add-model <provider> <model>",
	Short: "Add a model to a provider",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		providerName := args[0]
		model := args[1]

		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("❌ Error loading config: %v\n", err)
			os.Exit(1)
		}

		if err := cfg.AddModel(providerName, model); err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Model '%s' added to provider '%s'\n", model, providerName)
	},
}

var setDefaultCmd = &cobra.Command{
	Use:   "set-default <provider>",
	Short: "Set the default provider",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		providerName := args[0]

		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("❌ Error loading config: %v\n", err)
			os.Exit(1)
		}

		if err := cfg.SetDefault(providerName); err != nil {
			fmt.Printf("❌ Error: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("✅ Default provider set to '%s'\n", providerName)
	},
}

var listConfigCmd = &cobra.Command{
	Use:   "list",
	Short: "Show current configuration",
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Printf("❌ Error loading config: %v\n", err)
			os.Exit(1)
		}

		configPath := config.GetConfigPath()
		fmt.Printf("Configuration file: %s\n\n", configPath)

		if len(cfg.ListProviders()) == 0 {
			fmt.Println("No providers configured.")
			return
		}

		fmt.Printf("Default Provider: %s\n\n", cfg.Default)

		for _, name := range cfg.ListProviders() {
			provider, _ := cfg.GetProvider(name)

			fmt.Printf("Provider: %s\n", name)
			fmt.Printf("  Base URL: %s\n", provider.BaseURL)
			if provider.APIKey != "" {
				fmt.Printf("  API Key: *** (hidden)\n")
			} else {
				fmt.Printf("  API Key: (not set)\n")
			}
			fmt.Printf("  Models: %v\n", provider.Models)
			fmt.Printf("  Default Model: %s\n\n", provider.DefaultModel)
		}
	},
}
