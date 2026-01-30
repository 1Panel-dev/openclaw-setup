package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"os/exec"
	"path/filepath"
	"strings"

	"openclaw-setup/internal/config"
)

type ConfigRequest struct {
	Model        string               `json:"model"`
	GatewayToken string               `json:"gatewayToken"`
	Providers    []config.ProviderKey `json:"providers"`
}

type ConfigResponse struct {
	OK           bool   `json:"ok"`
	Restarted    bool   `json:"restarted"`
	Message      string `json:"message"`
	RestartError string `json:"restartError,omitempty"`
}

type ConfigHandler struct {
	composeDir    string
	configDir     string
	containerName string
}

func NewConfigHandler(cfg ServerConfig) http.Handler {
	return &ConfigHandler{
		composeDir:    cfg.ComposeDir,
		configDir:     cfg.ConfigDir,
		containerName: cfg.ContainerName,
	}
}

func (h *ConfigHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeJSON(w, http.StatusMethodNotAllowed, ConfigResponse{
			OK:      false,
			Message: "method not allowed",
		})
		return
	}

	var req ConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, ConfigResponse{
			OK:      false,
			Message: "invalid json",
		})
		return
	}

	model := strings.TrimSpace(req.Model)
	if model == "" {
		writeJSON(w, http.StatusBadRequest, ConfigResponse{
			OK:      false,
			Message: "model is required",
		})
		return
	}

	token := strings.TrimSpace(req.GatewayToken)
	if token == "" {
		token = generateToken()
	}

	if err := config.WriteConfigAndEnv(config.WriteOptions{
		ConfigDir:    h.configDir,
		Model:        model,
		GatewayToken: token,
		Providers:    req.Providers,
	}); err != nil {
		writeJSON(w, http.StatusInternalServerError, ConfigResponse{
			OK:      false,
			Message: err.Error(),
		})
		return
	}

	restarted, restartErr := restartContainer(h.containerName, h.composeDir)
	resp := ConfigResponse{
		OK:        restartErr == nil,
		Restarted: restarted,
		Message:   "配置已保存",
	}
	if restartErr != nil {
		resp.OK = false
		resp.Message = "配置已保存，但重启失败"
		resp.RestartError = restartErr.Error()
	}

	writeJSON(w, http.StatusOK, resp)
}

func restartContainer(_ string, composeDir string) (bool, error) {
	if strings.TrimSpace(composeDir) == "" {
		return false, nil
	}

	if err := chownDataDir(composeDir); err != nil {
		return false, err
	}

	downCmd := exec.Command("docker", "compose", "down")
	downCmd.Dir = composeDir
	if err := downCmd.Run(); err != nil {
		return false, err
	}

	upCmd := exec.Command("docker", "compose", "up", "-d")
	upCmd.Dir = composeDir
	if err := upCmd.Run(); err != nil {
		return false, err
	}

	return true, nil
}

func chownDataDir(composeDir string) error {
	dataDir := filepath.Join(composeDir, "data")
	chownCmd := exec.Command("chown", "-R", "1000:1000", dataDir)
	if err := chownCmd.Run(); err != nil {
		return err
	}
	return nil
}

func generateToken() string {
	bytes := make([]byte, 24)
	if _, err := rand.Read(bytes); err != nil {
		return ""
	}
	return hex.EncodeToString(bytes)
}

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
