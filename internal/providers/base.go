package providers

import (
	"context"
	"net/http"

	"github.com/heyanxiao/llm-bridge/pkg/types"
)

// ProviderAdapter 定义LLM提供商适配器接口
// 每个LLM平台(OpenAI、Claude、Gemini等)都需要实现此接口
type ProviderAdapter interface {
	// Transform 将统一请求格式转换为特定平台的请求格式
	Transform(req *types.UnifiedRequest) ([]byte, error)
	
	// CallAPI 调用特定平台的API
	CallAPI(ctx context.Context, data []byte) (*http.Response, error)
	
	// ParseResponse 解析平台响应并转换为统一格式
	ParseResponse(resp *http.Response) (*types.UnifiedResponse, error)
	
	// ParseStreamResponse 解析流式响应并转换为统一格式
	ParseStreamResponse(resp *http.Response) (<-chan *types.StreamResponse, error)
	
	// GetProviderName 获取提供商名称
	GetProviderName() string
	
	// ValidateRequest 验证请求参数是否符合该平台要求
	ValidateRequest(req *types.UnifiedRequest) error
}

// BaseProvider 基础提供商结构，包含通用字段
type BaseProvider struct {
	Name     string            // 提供商名称
	APIKey   string            // API密钥
	BaseURL  string            // API基础URL
	Headers  map[string]string // 默认请求头
	Timeout  int               // 请求超时时间(秒)
	Retries  int               // 重试次数
}

// ProviderConfig 提供商配置结构
type ProviderConfig struct {
	OpenAI *OpenAIConfig `yaml:"openai"`
	Claude *ClaudeConfig `yaml:"claude"`
	Gemini *GeminiConfig `yaml:"gemini"`
	Azure  *AzureConfig  `yaml:"azure"`
}

// OpenAI配置
type OpenAIConfig struct {
	APIKey   string `yaml:"api_key"`
	BaseURL  string `yaml:"base_url"`
	OrgID    string `yaml:"org_id"`
	Timeout  int    `yaml:"timeout"`
	Retries  int    `yaml:"retries"`
}

// Claude配置
type ClaudeConfig struct {
	APIKey  string `yaml:"api_key"`
	BaseURL string `yaml:"base_url"`
	Version string `yaml:"version"`
	Timeout int    `yaml:"timeout"`
	Retries int    `yaml:"retries"`
}

// Gemini配置
type GeminiConfig struct {
	APIKey   string `yaml:"api_key"`
	BaseURL  string `yaml:"base_url"`
	Project  string `yaml:"project"`
	Location string `yaml:"location"`
	Timeout  int    `yaml:"timeout"`
	Retries  int    `yaml:"retries"`
}

// Azure OpenAI配置
type AzureConfig struct {
	APIKey      string `yaml:"api_key"`
	Endpoint    string `yaml:"endpoint"`
	Deployment  string `yaml:"deployment"`
	APIVersion  string `yaml:"api_version"`
	Timeout     int    `yaml:"timeout"`
	Retries     int    `yaml:"retries"`
}

// ProviderFactory 提供商工厂
type ProviderFactory struct {
	providers map[string]ProviderAdapter
}

// NewProviderFactory 创建提供商工厂实例
func NewProviderFactory() *ProviderFactory {
	return &ProviderFactory{
		providers: make(map[string]ProviderAdapter),
	}
}

// RegisterProvider 注册提供商适配器
func (f *ProviderFactory) RegisterProvider(name string, provider ProviderAdapter) {
	f.providers[name] = provider
}

// GetProvider 根据名称获取提供商适配器
func (f *ProviderFactory) GetProvider(name string) (ProviderAdapter, bool) {
	provider, exists := f.providers[name]
	return provider, exists
}

// ListProviders 获取所有已注册的提供商名称
func (f *ProviderFactory) ListProviders() []string {
	names := make([]string, 0, len(f.providers))
	for name := range f.providers {
		names = append(names, name)
	}
	return names
}

// LoadBalancer 负载均衡器接口
type LoadBalancer interface {
	// SelectProvider 根据负载均衡策略选择提供商
	SelectProvider(providers []ProviderAdapter) ProviderAdapter
	
	// UpdateHealth 更新提供商健康状态
	UpdateHealth(providerName string, isHealthy bool)
}

// RoundRobinBalancer 轮询负载均衡器
type RoundRobinBalancer struct {
	current int
	health  map[string]bool
}

// NewRoundRobinBalancer 创建轮询负载均衡器
func NewRoundRobinBalancer() *RoundRobinBalancer {
	return &RoundRobinBalancer{
		current: 0,
		health:  make(map[string]bool),
	}
}

// SelectProvider 轮询选择健康的提供商
func (rb *RoundRobinBalancer) SelectProvider(providers []ProviderAdapter) ProviderAdapter {
	if len(providers) == 0 {
		return nil
	}
	
	// 过滤健康的提供商
	healthyProviders := make([]ProviderAdapter, 0)
	for _, provider := range providers {
		if isHealthy, exists := rb.health[provider.GetProviderName()]; !exists || isHealthy {
			healthyProviders = append(healthyProviders, provider)
		}
	}
	
	if len(healthyProviders) == 0 {
		// 如果没有健康的提供商，返回第一个
		return providers[0]
	}
	
	// 轮询选择
	selected := healthyProviders[rb.current%len(healthyProviders)]
	rb.current++
	return selected
}

// UpdateHealth 更新提供商健康状态
func (rb *RoundRobinBalancer) UpdateHealth(providerName string, isHealthy bool) {
	rb.health[providerName] = isHealthy
}