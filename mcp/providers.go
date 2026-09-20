package mcp

// Provider name constants — kept in the mcp package so that client.go can
// reference them for default configuration without importing sub-packages.
// Provider sub-packages re-use these same values.
const (
	ProviderDeepSeek = "deepseek"
	ProviderOpenAI   = "openai"
	ProviderClaude   = "claude"
	ProviderQwen     = "qwen"
	ProviderGemini   = "gemini"
	ProviderGrok     = "grok"
	ProviderKimi     = "kimi"
	ProviderMiniMax  = "minimax"

	ProviderClaw402 = "claw402"
	ProviderGLM      = "glm"

	// Default GLM configuration (OpenAI-compatible endpoint)
	DefaultGLMBaseURL = "https://ai.871717.xyz/v1"
	DefaultGLMModel    = "glm-5.2"

	// Default DeepSeek configuration (used as fallback in NewClient)
	DefaultDeepSeekBaseURL = "https://api.deepseek.com"
	DefaultDeepSeekModel   = "deepseek-chat"

	// Default Qwen configuration (used by WithQwenConfig convenience option)
	DefaultQwenBaseURL = "https://dashscope.aliyuncs.com/compatible-mode/v1"
	DefaultQwenModel   = "qwen3-max"

	// Default MiniMax configuration (used by WithMiniMaxConfig convenience option)
	DefaultMiniMaxBaseURL = "https://api.minimax.io/v1"
	DefaultMiniMaxModel   = "MiniMax-M2.7"
)
