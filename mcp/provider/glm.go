package provider

import (
	"net/http"

	"nofx/mcp"
)

func init() {
	mcp.RegisterProvider(mcp.ProviderGLM, func(opts ...mcp.ClientOption) mcp.AIClient {
		return NewGLMClientWithOptions(opts...)
	})
}

type GLMClient struct {
	*mcp.Client
}

func (c *GLMClient) BaseClient() *mcp.Client { return c.Client }

// NewGLMClient creates GLM client (backward compatible)
func NewGLMClient() mcp.AIClient {
	return NewGLMClientWithOptions()
}

// NewGLMClientWithOptions creates GLM client (supports options pattern)
func NewGLMClientWithOptions(opts ...mcp.ClientOption) mcp.AIClient {
	glmOpts := []mcp.ClientOption{
		mcp.WithProvider(mcp.ProviderGLM),
		mcp.WithModel(mcp.DefaultGLMModel),
		mcp.WithBaseURL(mcp.DefaultGLMBaseURL),
	}

	allOpts := append(glmOpts, opts...)
	baseClient := mcp.NewClient(allOpts...).(*mcp.Client)

	glmClient := &GLMClient{
		Client: baseClient,
	}

	baseClient.Hooks = glmClient
	return glmClient
}

func (glmClient *GLMClient) SetAPIKey(apiKey string, customURL string, customModel string) {
	glmClient.APIKey = apiKey

	if len(apiKey) > 8 {
		glmClient.Log.Infof("🔧 [MCP] GLM API Key: %s...%s", apiKey[:4], apiKey[len(apiKey)-4:])
	}
	if customURL != "" {
		glmClient.BaseURL = customURL
		glmClient.Log.Infof("🔧 [MCP] GLM using custom BaseURL: %s", customURL)
	} else {
		glmClient.Log.Infof("🔧 [MCP] GLM using default BaseURL: %s", glmClient.BaseURL)
	}
	if customModel != "" {
		glmClient.Model = customModel
		glmClient.Log.Infof("🔧 [MCP] GLM using custom Model: %s", customModel)
	} else {
		glmClient.Log.Infof("🔧 [MCP] GLM using default Model: %s", glmClient.Model)
	}
}

func (glmClient *GLMClient) SetAuthHeader(reqHeaders http.Header) {
	glmClient.Client.SetAuthHeader(reqHeaders)
}
