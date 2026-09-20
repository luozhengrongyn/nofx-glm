package api

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"nofx/logger"
	"nofx/mcp"
	"nofx/wallet"

	gethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-gonic/gin"
)

type beginnerOnboardingResponse struct {
	Address           string `json:"address"`
	PrivateKey        string `json:"private_key"`
	Chain             string `json:"chain"`
	Asset             string `json:"asset"`
	Provider          string `json:"provider"`
	DefaultModel      string `json:"default_model"`
	ConfiguredModelID string `json:"configured_model_id"`
	BalanceUSDC       string `json:"balance_usdc"`
	BalanceStatus     string `json:"balance_status,omitempty"`
	EnvSaved          bool   `json:"env_saved"`
	EnvPath           string `json:"env_path,omitempty"`
	ReusedExisting    bool   `json:"reused_existing"`
	EnvWarning        string `json:"env_warning,omitempty"`
}

type currentBeginnerWalletResponse struct {
	Found         bool   `json:"found"`
	Address       string `json:"address,omitempty"`
	BalanceUSDC   string `json:"balance_usdc,omitempty"`
	BalanceStatus string `json:"balance_status,omitempty"`
	Source        string `json:"source,omitempty"`
	Claw402Status string `json:"claw402_status"`
}

// queryBeginnerWalletBalance returns the wallet balance plus a status flag so
// the UI can distinguish "RPC unreachable" from a genuinely empty wallet.
// Uses the 30s cache — safe for UI polling.
func queryBeginnerWalletBalance(address string) (balanceUSDC string, balanceStatus string) {
	balance, err := wallet.QueryUSDCBalanceCached(address)
	if err != nil {
		return "", "unknown"
	}
	return fmt.Sprintf("%.2f", balance), "ok"
}

func (s *Server) handleBeginnerOnboarding(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user context"})
		return
	}

	// Use GLM direct API key instead of EVM wallet
	glmKey := mcp.DefaultGLMBaseURL // placeholder, real key set by user in UI
	configuredModelID, err := s.findConfiguredGLMModelID(userID)
	if err != nil {
		// Create GLM model entry for user
		if err := s.store.AIModel().Update(userID, "glm", true, glmKey, mcp.DefaultGLMBaseURL, mcp.DefaultGLMModel); err != nil {
			logger.Errorf("Failed to save beginner GLM config for user %s: %v", userID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save beginner model configuration"})
			return
		}
		configuredModelID, _ = s.findConfiguredGLMModelID(userID)
	}

	resp := beginnerOnboardingResponse{
		Provider:          "glm",
		DefaultModel:      mcp.DefaultGLMModel,
		ConfiguredModelID: configuredModelID,
		ReusedExisting:    true,
	}

	c.JSON(http.StatusOK, resp)
}

func (s *Server) handleCurrentBeginnerWallet(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user context"})
		return
	}

	models, err := s.store.AIModel().List(userID)
	if err != nil {
		logger.Errorf("Failed to load current beginner config for user %s: %v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load current config"})
		return
	}

	for _, model := range models {
		if model == nil || model.Provider != "glm" {
			continue
		}
		c.JSON(http.StatusOK, currentBeginnerWalletResponse{
			Found:         true,
			Source:         "model",
			Claw402Status: "ok",
		})
		return
	}

	c.JSON(http.StatusOK, currentBeginnerWalletResponse{
		Found:         false,
		Claw402Status: "ok",
	})
}

// resolveBeginnerWallet is kept for backward compat but no longer generates EVM wallets.
func (s *Server) resolveBeginnerWallet(userID string) (privateKey string, address string, configuredModelID string, reused bool, err error) {
	return "", "", "", false, nil
}

func (s *Server) findConfiguredGLMModelID(userID string) (string, error) {
	models, err := s.store.AIModel().List(userID)
	if err != nil {
		return "", err
	}

	for _, model := range models {
		if model != nil && model.Provider == "glm" {
			return model.ID, nil
		}
	}

	return "", fmt.Errorf("glm model not found")
}

func walletAddressFromPrivateKey(privateKey string) (string, error) {
	key := strings.TrimSpace(privateKey)
	if !strings.HasPrefix(key, "0x") {
		return "", fmt.Errorf("private key must start with 0x")
	}
	if len(key) != 66 {
		return "", fmt.Errorf("private key must be 66 characters")
	}

	privateKeyObj, err := gethcrypto.HexToECDSA(strings.TrimPrefix(key, "0x"))
	if err != nil {
		return "", err
	}

	return gethcrypto.PubkeyToAddress(privateKeyObj.PublicKey).Hex(), nil
}

func persistBeginnerWalletEnv(privateKey string, address string) (bool, string, error) {
	// No longer needed — GLM uses API key, not EVM wallet
	return true, "", nil
}

func uniqueEnvPaths(paths []string) []string {
	seen := make(map[string]struct{}, len(paths))
	result := make([]string, 0, len(paths))
	for _, path := range paths {
		clean := filepath.Clean(path)
		if _, ok := seen[clean]; ok {
			continue
		}
		seen[clean] = struct{}{}
		result = append(result, clean)
	}
	return result
}

func upsertEnvFile(path string, values map[string]string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	existingLines := make([]string, 0)
	if file, err := os.Open(path); err == nil {
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			existingLines = append(existingLines, scanner.Text())
		}
		file.Close()
		if err := scanner.Err(); err != nil {
			return err
		}
	} else if !os.IsNotExist(err) {
		return err
	}

	remaining := make(map[string]string, len(values))
	for key, value := range values {
		remaining[key] = value
	}

	updatedLines := make([]string, 0, len(existingLines)+len(values))
	for _, line := range existingLines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") || !strings.Contains(line, "=") {
			updatedLines = append(updatedLines, line)
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		key := strings.TrimSpace(parts[0])
		value, ok := remaining[key]
		if !ok {
			updatedLines = append(updatedLines, line)
			continue
		}

		updatedLines = append(updatedLines, fmt.Sprintf("%s=%s", key, value))
		delete(remaining, key)
	}

	for key, value := range remaining {
		updatedLines = append(updatedLines, fmt.Sprintf("%s=%s", key, value))
	}

	content := strings.Join(updatedLines, "\n")
	if content != "" && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		return err
	}

	return nil
}
