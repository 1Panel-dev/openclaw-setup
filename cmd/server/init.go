package main

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"openclaw-setup/internal/config"
)

type initOptions struct {
	composeDir string
}

func runInit(opts initOptions) error {
	composeDir, err := resolveComposeDir(opts.composeDir)
	if err != nil {
		return err
	}

	envPath := filepath.Join(composeDir, ".env")
	envMap, err := readDotEnv(envPath)
	if err != nil {
		return err
	}

	provider := normalizeEnvValue(envMap["PROVIDER"])
	apiKey := normalizeEnvValue(envMap["API_KEY"])
	model := normalizeEnvValue(envMap["MODEL"])
	baseUrl := normalizeEnvValue(envMap["BASE_URL"])
	provider = strings.ToLower(provider)
	if provider == "" || model == "" {
		return fmt.Errorf(".env must include PROVIDER and MODEL")
	}
	if provider != "ollama" && apiKey == "" {
		return fmt.Errorf(".env must include API_KEY for provider %s", provider)
	}
	if provider == "ollama" && baseUrl == "" {
		return fmt.Errorf(".env must include BASE_URL for provider ollama")
	}

	providerEnvKey, err := providerEnvKey(provider)
	if err != nil {
		return err
	}

	token := randomToken()
	configDir := filepath.Join(composeDir, "data", "conf")
	if err := config.WriteConfigOnly(config.WriteConfigOnlyOptions{
		ConfigDir:      configDir,
		Model:          model,
		GatewayToken:   token,
		ProviderID:     provider,
		ProviderEnvKey: providerEnvKey,
		ProviderApiKey: apiKey,
		BaseUrl:        baseUrl,
		WriteEnv:       true,
	}); err != nil {
		return err
	}

	if err := writeComposeToken(envPath, token); err != nil {
		return err
	}

	return nil
}

func resolveComposeDir(explicit string) (string, error) {
	if strings.TrimSpace(explicit) != "" {
		return explicit, nil
	}
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Clean(wd), nil
}

func providerEnvKey(provider string) (string, error) {
	switch provider {
	case "openai":
		return "OPENAI_API_KEY", nil
	case "anthropic":
		return "ANTHROPIC_API_KEY", nil
	case "gemini":
		return "GEMINI_API_KEY", nil
	case "groq":
		return "GROQ_API_KEY", nil
	case "mistral":
		return "MISTRAL_API_KEY", nil
	case "cohere":
		return "COHERE_API_KEY", nil
	case "minimax":
		return "MINIMAX_API_KEY", nil
	case "deepseek":
		return "DEEPSEEK_API_KEY", nil
	case "moonshot":
		return "MOONSHOT_API_KEY", nil
	case "qwen":
		return "QWEN_API_KEY", nil
	case "zai":
		return "ZAI_API_KEY", nil
	case "ollama":
		return "", nil
	default:
		return "", fmt.Errorf("unsupported PROVIDER: %s", provider)
	}
}

func readDotEnv(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("read .env: %w", err)
	}
	defer file.Close()

	result := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := normalizeEnvValue(parts[1])
		if key == "" {
			continue
		}
		result[key] = value
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("read .env: %w", err)
	}
	return result, nil
}

func writeComposeToken(path string, token string) error {
	if strings.TrimSpace(token) == "" {
		return fmt.Errorf("generated token is empty")
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read .env: %w", err)
	}

	lines := strings.Split(string(content), "\n")
	updated := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		parts := strings.SplitN(trimmed, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		if key == "OPENCLAW_GATEWAY_TOKEN" || key == "CLAWDBOT_GATEWAY_TOKEN" {
			lines[i] = fmt.Sprintf("OPENCLAW_GATEWAY_TOKEN=%s", token)
			updated = true
		}
	}
	if !updated {
		lines = append(lines, fmt.Sprintf("OPENCLAW_GATEWAY_TOKEN=%s", token))
	}

	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o600)
}

func normalizeEnvValue(value string) string {
	trimmed := strings.TrimSpace(value)
	trimmed = strings.Trim(trimmed, "\"")
	trimmed = strings.Trim(trimmed, "'")
	return strings.TrimSpace(trimmed)
}

func randomToken() string {
	bytes := make([]byte, 24)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)
}

func chownDataDir(composeDir string) error {
	dataDir := filepath.Join(composeDir, "data")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return fmt.Errorf("create data dir: %w", err)
	}
	cmd := exec.Command("chown", "-R", "1000:1000", dataDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("chown data: %w", err)
	}
	return nil
}
