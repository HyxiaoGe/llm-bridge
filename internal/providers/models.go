package providers

// ModelConfig 定义每个提供商支持的模型配置
type ModelConfig struct {
	Models       []string
	DefaultModel string
}

// SupportedModels 定义所有提供商支持的模型列表
var SupportedModels = map[string]ModelConfig{
	"openai": {
		Models: []string{
			"gpt-3.5-turbo",
			"gpt-4o-2024-08-06",
			"gpt-4.1-2025-04-14",
		},
		DefaultModel: "gpt-3.5-turbo",
	},
	"gemini": {
		Models: []string{
			"gemini-2.5-pro",
			"gemini-2.5-flash",
			"gemini-2.0-flash",
			"gemini-1.5-flash",
			"gemini-1.5-pro",
		},
		DefaultModel: "gemini-2.5-flash",
	},
	"deepseek": {
		Models: []string{
			"deepseek-reasoner",
			"deepseek-chat",
		},
		DefaultModel: "deepseek-chat",
	},
	"qwen": {
		Models: []string{
			"qwen-max",
			"qwen-plus",
			"qwq-plus",
		},
		DefaultModel: "qwen-plus",
	},
	"moonshot": {
		Models: []string{
			"moonshot-v1-8k",
			"moonshot-v1-32k",
			"moonshot-v1-128k",
			"kimi-k2-0711-preview",
		},
		DefaultModel: "moonshot-v1-8k",
	},
}

// GetProviderModels 获取提供商支持的模型列表
func GetProviderModels(provider string) []string {
	if config, exists := SupportedModels[provider]; exists {
		return config.Models
	}
	return []string{}
}

// GetDefaultModel 获取提供商的默认模型
func GetDefaultModel(provider string) string {
	if config, exists := SupportedModels[provider]; exists {
		return config.DefaultModel
	}
	return ""
}

// IsModelSupported 检查模型是否被提供商支持
func IsModelSupported(provider, model string) bool {
	models := GetProviderModels(provider)
	for _, m := range models {
		if m == model {
			return true
		}
	}
	return false
}