package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type ProviderKey struct {
	Key   string
	Value string
}

type WriteOptions struct {
	ConfigDir    string
	Model        string
	GatewayToken string
	Providers    []ProviderKey
}

type WriteConfigOnlyOptions struct {
	ConfigDir      string
	Model          string
	GatewayToken   string
	ProviderID     string
	ProviderEnvKey string
	ProviderApiKey string
	WriteEnv       bool
}

type openclawConfig struct {
	Gateway gatewayConfig `json:"gateway"`
	Agents  agentsConfig  `json:"agents"`
	Models  *modelsConfig `json:"models,omitempty"`
}

type gatewayConfig struct {
	Mode      string           `json:"mode"`
	Bind      string           `json:"bind"`
	Port      int              `json:"port"`
	Auth      gatewayAuth      `json:"auth"`
	ControlUi gatewayControlUi `json:"controlUi"`
}

type gatewayControlUi struct {
	AllowInsecureAuth bool `json:"allowInsecureAuth"`
}

type gatewayAuth struct {
	Mode  string `json:"mode"`
	Token string `json:"token"`
}

type agentsConfig struct {
	Defaults agentDefaults `json:"defaults"`
}

type agentDefaults struct {
	Model modelRef `json:"model"`
}

type modelRef struct {
	Primary string `json:"primary"`
}

type modelsConfig struct {
	Mode      string                   `json:"mode,omitempty"`
	Providers map[string]modelProvider `json:"providers,omitempty"`
}

type modelProvider struct {
	ApiKey  string       `json:"apiKey,omitempty"`
	BaseUrl string       `json:"baseUrl,omitempty"`
	Api     string       `json:"api,omitempty"`
	Models  []modelEntry `json:"models,omitempty"`
}

type modelEntry struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Reasoning     bool     `json:"reasoning"`
	Input         []string `json:"input"`
	ContextWindow int      `json:"contextWindow"`
	MaxTokens     int      `json:"maxTokens"`
}

func WriteConfigAndEnv(opts WriteOptions) error {
	if strings.TrimSpace(opts.ConfigDir) == "" {
		return fmt.Errorf("config dir is required")
	}
	if strings.TrimSpace(opts.Model) == "" {
		return fmt.Errorf("model is required")
	}
	if strings.TrimSpace(opts.GatewayToken) == "" {
		return fmt.Errorf("gateway token is required")
	}

	if err := os.MkdirAll(opts.ConfigDir, 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	configPath := filepath.Join(opts.ConfigDir, "openclaw.json")
	envPath := filepath.Join(opts.ConfigDir, ".env")

	cfg := openclawConfig{
		Gateway: gatewayConfig{
			Mode: "local",
			Bind: "lan",
			Port: 18789,
			Auth: gatewayAuth{
				Mode:  "token",
				Token: opts.GatewayToken,
			},
			ControlUi: gatewayControlUi{
				AllowInsecureAuth: true,
			},
		},
		Agents: agentsConfig{
			Defaults: agentDefaults{
				Model: modelRef{
					Primary: opts.Model,
				},
			},
		},
	}

	payload, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, payload, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	envLines := []string{fmt.Sprintf("OPENCLAW_GATEWAY_TOKEN=%s", opts.GatewayToken)}
	for _, provider := range opts.Providers {
		key := strings.TrimSpace(provider.Key)
		value := strings.TrimSpace(provider.Value)
		if key == "" || value == "" {
			continue
		}
		envLines = append(envLines, fmt.Sprintf("%s=%s", key, value))
	}

	envContent := strings.Join(envLines, "\n") + "\n"
	if err := os.WriteFile(envPath, []byte(envContent), 0o600); err != nil {
		return fmt.Errorf("write env: %w", err)
	}

	return nil
}

func WriteConfigOnly(opts WriteConfigOnlyOptions) error {
	if strings.TrimSpace(opts.ConfigDir) == "" {
		return fmt.Errorf("config dir is required")
	}
	if strings.TrimSpace(opts.Model) == "" {
		return fmt.Errorf("model is required")
	}
	if strings.TrimSpace(opts.GatewayToken) == "" {
		return fmt.Errorf("gateway token is required")
	}

	if err := os.MkdirAll(opts.ConfigDir, 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	configPath := filepath.Join(opts.ConfigDir, "openclaw.json")

	cfg := openclawConfig{
		Gateway: gatewayConfig{
			Mode: "local",
			Bind: "lan",
			Port: 18789,
			Auth: gatewayAuth{
				Mode:  "token",
				Token: opts.GatewayToken,
			},
			ControlUi: gatewayControlUi{
				AllowInsecureAuth: true,
			},
		},
		Agents: agentsConfig{
			Defaults: agentDefaults{
				Model: modelRef{
					Primary: opts.Model,
				},
			},
		},
	}

	providerID := strings.ToLower(strings.TrimSpace(opts.ProviderID))
	if providerID == "deepseek" {
		cfg.Models = &modelsConfig{
			Mode: "merge",
			Providers: map[string]modelProvider{
				"deepseek": {
					ApiKey:  "${DEEPSEEK_API_KEY}",
					BaseUrl: "https://api.deepseek.com/v1",
					Api:     "openai-completions",
					Models: []modelEntry{
						{
							ID:            "deepseek-chat",
							Name:          "DeepSeek Chat",
							Reasoning:     false,
							Input:         []string{"text"},
							ContextWindow: 128000,
							MaxTokens:     8192,
						},
					},
				},
			},
		}
	}

	payload, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, payload, 0o600); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	if opts.WriteEnv {
		envPath := filepath.Join(opts.ConfigDir, ".env")
		lines := []string{fmt.Sprintf("OPENCLAW_GATEWAY_TOKEN=%s", opts.GatewayToken)}
		if strings.TrimSpace(opts.ProviderEnvKey) != "" && strings.TrimSpace(opts.ProviderApiKey) != "" {
			lines = append(lines, fmt.Sprintf("%s=%s", opts.ProviderEnvKey, opts.ProviderApiKey))
		}
		content := strings.Join(lines, "\n") + "\n"
		if err := os.WriteFile(envPath, []byte(content), 0o600); err != nil {
			return fmt.Errorf("write env: %w", err)
		}
	}

	return nil
}
