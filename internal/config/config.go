package config

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"

	"github.com/ndzuma/probeTool/internal/paths"
)

type Config struct {
	Providers map[string]Provider `json:"providers"`
	Default   string              `json:"default"`
}

type Provider struct {
	Name         string   `json:"name"`
	BaseURL      string   `json:"base_url"`
	APIKey       string   `json:"api_key"`
	Models       []string `json:"models"`
	DefaultModel string   `json:"default_model"`
}

// GetConfigDir returns the application directory path
// Deprecated: Use paths.GetAppDir() instead
func GetConfigDir() string {
	return paths.GetAppDir()
}

// GetConfigPath returns the full path to config.json
// Deprecated: Use paths.GetConfigPath() instead
func GetConfigPath() string {
	return paths.GetConfigPath()
}

// Load loads the configuration from disk
func Load() (*Config, error) {
	configPath := GetConfigPath()

	data, err := os.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{
				Providers: make(map[string]Provider),
				Default:   "",
			}, nil
		}
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	if config.Providers == nil {
		config.Providers = make(map[string]Provider)
	}

	return &config, nil
}

// Save saves the configuration to disk
func (c *Config) Save() error {
	configDir := GetConfigDir()
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(GetConfigPath(), data, 0600)
}

// AddProvider adds a new provider interactively
func (c *Config) AddProvider(name string) error {
	if _, exists := c.Providers[name]; exists {
		return fmt.Errorf("provider '%s' already exists", name)
	}

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Base URL: ")
	scanner.Scan()
	baseURL := scanner.Text()

	fmt.Print("API Key: ")
	scanner.Scan()
	apiKey := scanner.Text()

	fmt.Print("Models (comma-separated): ")
	scanner.Scan()
	modelsStr := scanner.Text()
	models := []string{}
	if modelsStr != "" {
		// Simple split on comma
		for _, m := range splitModels(modelsStr) {
			if m != "" {
				models = append(models, m)
			}
		}
	}

	defaultModel := ""
	if len(models) > 0 {
		defaultModel = models[0]
	}

	c.Providers[name] = Provider{
		Name:         name,
		BaseURL:      baseURL,
		APIKey:       apiKey,
		Models:       models,
		DefaultModel: defaultModel,
	}

	// If this is the first provider, set it as default
	if c.Default == "" {
		c.Default = name
	}

	return c.Save()
}

// SetAPIKey sets the API key for a provider
func (c *Config) SetAPIKey(providerName, apiKey string) error {
	provider, exists := c.Providers[providerName]
	if !exists {
		return fmt.Errorf("provider '%s' not found", providerName)
	}

	provider.APIKey = apiKey
	c.Providers[providerName] = provider
	return c.Save()
}

// AddModel adds a model to a provider
func (c *Config) AddModel(providerName, model string) error {
	provider, exists := c.Providers[providerName]
	if !exists {
		return fmt.Errorf("provider '%s' not found", providerName)
	}

	// Check if model already exists
	for _, m := range provider.Models {
		if m == model {
			return fmt.Errorf("model '%s' already exists for provider '%s'", model, providerName)
		}
	}

	provider.Models = append(provider.Models, model)

	// If no default model set, use this one
	if provider.DefaultModel == "" {
		provider.DefaultModel = model
	}

	c.Providers[providerName] = provider
	return c.Save()
}

// SetDefault sets the default provider
func (c *Config) SetDefault(providerName string) error {
	if _, exists := c.Providers[providerName]; !exists {
		return fmt.Errorf("provider '%s' not found", providerName)
	}

	c.Default = providerName
	return c.Save()
}

// ListProviders returns a list of provider names
func (c *Config) ListProviders() []string {
	names := make([]string, 0, len(c.Providers))
	for name := range c.Providers {
		names = append(names, name)
	}
	return names
}

// GetProvider returns a provider by name
func (c *Config) GetProvider(name string) (Provider, bool) {
	provider, exists := c.Providers[name]
	return provider, exists
}

// GetDefaultProvider returns the default provider
func (c *Config) GetDefaultProvider() (Provider, bool) {
	if c.Default == "" {
		return Provider{}, false
	}
	return c.GetProvider(c.Default)
}

// splitModels splits a comma-separated string into slice
func splitModels(s string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			if i > start {
				result = append(result, s[start:i])
			}
			start = i + 1
		}
	}
	if start < len(s) {
		result = append(result, s[start:])
	}
	return result
}
