package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type ModelsRequest struct {
	Provider string `json:"provider"`
	ApiKey   string `json:"apiKey"`
}

type ModelsResponse struct {
	Models  []string `json:"models"`
	Message string   `json:"message,omitempty"`
}

func NewModelsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			writeJSON(w, http.StatusMethodNotAllowed, ModelsResponse{Message: "method not allowed"})
			return
		}

		var req ModelsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, ModelsResponse{Message: "invalid json"})
			return
		}

		provider := strings.TrimSpace(req.Provider)
		apiKey := strings.TrimSpace(req.ApiKey)
		if provider == "" || apiKey == "" {
			writeJSON(w, http.StatusBadRequest, ModelsResponse{Message: "provider and apiKey required"})
			return
		}

		models, message, err := fetchModels(provider, apiKey)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, ModelsResponse{Message: err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, ModelsResponse{Models: models, Message: message})
	})
}

type openAIModelsResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

type geminiModelsResponse struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

func fetchModels(provider, apiKey string) ([]string, string, error) {
	provider = strings.ToLower(provider)
	client := &http.Client{Timeout: 15 * time.Second}

	switch provider {
	case "openai":
		return fetchOpenAICompatible(client, "https://api.openai.com/v1/models", apiKey)
	case "groq":
		return fetchOpenAICompatible(client, "https://api.groq.com/openai/v1/models", apiKey)
	case "mistral":
		return fetchOpenAICompatible(client, "https://api.mistral.ai/v1/models", apiKey)
	case "moonshot":
		return fetchOpenAICompatible(client, "https://api.moonshot.cn/v1/models", apiKey)
	case "deepseek":
		return fetchOpenAICompatible(client, "https://api.deepseek.com/v1/models", apiKey)
	case "qwen":
		return fetchOpenAICompatible(client, "https://dashscope.aliyuncs.com/compatible-mode/v1/models", apiKey)
	case "anthropic":
		return fetchAnthropic(client, apiKey)
	case "gemini":
		return fetchGemini(client, apiKey)
	case "cohere", "minimax", "zai", "custom":
		return nil, "该提供商暂不支持自动拉取模型，请手动填写", nil
	default:
		return nil, "", fmt.Errorf("unknown provider")
	}
}

func fetchOpenAICompatible(client *http.Client, url, apiKey string) ([]string, string, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("provider error: %s", strings.TrimSpace(string(body)))
	}

	var payload openAIModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, "", err
	}

	models := make([]string, 0, len(payload.Data))
	for _, item := range payload.Data {
		if strings.TrimSpace(item.ID) == "" {
			continue
		}
		models = append(models, item.ID)
	}
	return models, "", nil
}

func fetchAnthropic(client *http.Client, apiKey string) ([]string, string, error) {
	req, err := http.NewRequest(http.MethodGet, "https://api.anthropic.com/v1/models", nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("provider error: %s", strings.TrimSpace(string(body)))
	}

	var payload openAIModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, "", err
	}

	models := make([]string, 0, len(payload.Data))
	for _, item := range payload.Data {
		if strings.TrimSpace(item.ID) == "" {
			continue
		}
		models = append(models, item.ID)
	}
	return models, "", nil
}

func fetchGemini(client *http.Client, apiKey string) ([]string, string, error) {
	url := "https://generativelanguage.googleapis.com/v1beta/models?key=" + apiKey
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, "", err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("provider error: %s", strings.TrimSpace(string(body)))
	}

	var payload geminiModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, "", err
	}

	models := make([]string, 0, len(payload.Models))
	for _, item := range payload.Models {
		name := strings.TrimSpace(item.Name)
		name = strings.TrimPrefix(name, "models/")
		if name == "" {
			continue
		}
		models = append(models, name)
	}
	return models, "", nil
}
