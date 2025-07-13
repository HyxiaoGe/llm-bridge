package handlers

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/heyanxiao/llm-bridge/internal/providers"
	"github.com/heyanxiao/llm-bridge/internal/stats"
	"github.com/heyanxiao/llm-bridge/pkg/types"
)

// AdminHandler 管理面板处理器
type AdminHandler struct {
	providerFactory *providers.ProviderFactory
	loadBalancer    providers.LoadBalancer
	startTime       time.Time
}

// NewAdminHandler 创建管理面板处理器实例
func NewAdminHandler(factory *providers.ProviderFactory, balancer providers.LoadBalancer) *AdminHandler {
	return &AdminHandler{
		providerFactory: factory,
		loadBalancer:    balancer,
		startTime:       time.Now(),
	}
}

// Dashboard 返回监控面板首页
func (h *AdminHandler) Dashboard(c *fiber.Ctx) error {
	return c.SendFile("./static/index.html")
}

// GetProvidersStatus 获取所有提供商状态
func (h *AdminHandler) GetProvidersStatus(c *fiber.Ctx) error {
	providerNames := h.providerFactory.ListProviders()
	providersStatus := make([]map[string]interface{}, 0)

	for _, name := range providerNames {
		provider, exists := h.providerFactory.GetProvider(name)
		if !exists {
			continue
		}

		// 获取基础信息
		baseProvider := getBaseProvider(provider)
		
		// 从Redis获取提供商统计
		var requests, avgResponseTime, tokens int64
		if redisMetrics := stats.GetRedisMetrics(); redisMetrics != nil {
			requests, avgResponseTime, tokens = redisMetrics.GetProviderStats(name)
		}
		
		status := map[string]interface{}{
			"name":             name,
			"status":           "unknown", // 默认状态
			"baseUrl":          baseProvider.BaseURL,
			"timeout":          baseProvider.Timeout,
			"retries":          baseProvider.Retries,
			"models":           getProviderModels(name),
			"lastTest":         nil,
			"requests":         requests,
			"avgResponseTime":  avgResponseTime,
			"tokens":           tokens,
		}

		// 检查健康状态（简单检查，可以扩展为实际的健康检查）
		if baseProvider.APIKey != "" && baseProvider.BaseURL != "" {
			status["status"] = "healthy"
		} else {
			status["status"] = "unhealthy"
		}

		providersStatus = append(providersStatus, status)
	}

	return c.JSON(fiber.Map{
		"success":   true,
		"providers": providersStatus,
		"timestamp": time.Now().Unix(),
	})
}

// TestProvider 测试特定提供商
func (h *AdminHandler) TestProvider(c *fiber.Ctx) error {
	var req struct {
		Provider string `json:"provider"`
		Model    string `json:"model"`
		Message  string `json:"message"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"success": false,
			"error":   "请求参数解析失败: " + err.Error(),
		})
	}

	// 默认值
	if req.Message == "" {
		req.Message = "这是一个健康检查测试消息"
	}
	if req.Model == "" {
		req.Model = getDefaultModelForProvider(req.Provider)
	}

	// 构建测试请求
	testReq := &types.UnifiedRequest{
		Model:    req.Model,
		Provider: req.Provider,
		Messages: []types.Message{
			{
				Role:    "user",
				Content: req.Message,
			},
		},
		Parameters: types.Parameters{
			Temperature: 0.1,
			MaxTokens:   50,
		},
		Metadata: types.Metadata{
			UserID:    "health_check",
			Timestamp: time.Now(),
		},
	}

	// 获取提供商
	var provider providers.ProviderAdapter
	if req.Provider != "" {
		var exists bool
		provider, exists = h.providerFactory.GetProvider(req.Provider)
		if !exists {
			return c.Status(404).JSON(fiber.Map{
				"success": false,
				"error":   "提供商不存在: " + req.Provider,
			})
		}
	} else {
		// 使用负载均衡选择
		allProviders := h.getAllProviders()
		provider = h.loadBalancer.SelectProvider(allProviders)
		if provider == nil {
			return c.Status(503).JSON(fiber.Map{
				"success": false,
				"error":   "没有可用的提供商",
			})
		}
		req.Provider = provider.GetProviderName()
	}

	// 记录开始时间
	startTime := time.Now()

	// 验证请求
	if err := provider.ValidateRequest(testReq); err != nil {
		return c.JSON(fiber.Map{
			"success": false,
			"error":   "请求验证失败: " + err.Error(),
			"provider": req.Provider,
			"duration": time.Since(startTime).Milliseconds(),
		})
	}

	// 转换请求格式
	providerData, err := provider.Transform(testReq)
	if err != nil {
		return c.JSON(fiber.Map{
			"success": false,
			"error":   "请求转换失败: " + err.Error(),
			"provider": req.Provider,
			"duration": time.Since(startTime).Milliseconds(),
		})
	}

	// 调用API
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := provider.CallAPI(ctx, providerData)
	if err != nil {
		return c.JSON(fiber.Map{
			"success": false,
			"error":   "API调用失败: " + err.Error(),
			"provider": req.Provider,
			"duration": time.Since(startTime).Milliseconds(),
		})
	}

	// 解析响应
	unifiedResp, err := provider.ParseResponse(resp)
	if err != nil {
		return c.JSON(fiber.Map{
			"success": false,
			"error":   "响应解析失败: " + err.Error(),
			"provider": req.Provider,
			"duration": time.Since(startTime).Milliseconds(),
		})
	}

	duration := time.Since(startTime).Milliseconds()

	// 检查是否有错误
	if unifiedResp.Error != nil {
		return c.JSON(fiber.Map{
			"success": false,
			"error":   unifiedResp.Error.Message,
			"provider": req.Provider,
			"duration": duration,
		})
	}

	return c.JSON(fiber.Map{
		"success":  true,
		"provider": req.Provider,
		"model":    unifiedResp.Model,
		"duration": duration,
		"response": unifiedResp,
		"message":  "测试成功",
	})
}

// GetSystemStats 获取系统统计信息
func (h *AdminHandler) GetSystemStats(c *fiber.Ctx) error {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// 从Redis获取统计数据
	var totalRequests, avgResponseTime, totalTokens int64
	var uptime time.Duration

	redisMetrics := stats.GetRedisMetrics()
	if redisMetrics != nil {
		totalRequests, avgResponseTime, totalTokens, uptime = redisMetrics.GetStats()
	} else {
		// 如果Redis不可用，使用本地时间
		uptime = time.Since(h.startTime)
	}

	statsData := fiber.Map{
		"service": fiber.Map{
			"version": "v1.0.0",
			"uptime":  formatDuration(uptime),
			"port":    8080, // 可以从配置中获取
		},
		"metrics": fiber.Map{
			"total_requests":      totalRequests,
			"avg_response_time":   avgResponseTime,
			"total_tokens":        totalTokens,
			"active_connections":  0, // TODO: 实现活跃连接统计
			"memory_usage":        formatBytes(m.Alloc),
			"memory_total":        formatBytes(m.Sys),
			"cpu_usage":          "0%", // TODO: 实现CPU使用率统计
		},
		"providers": fiber.Map{
			"total":   len(h.providerFactory.ListProviders()),
			"healthy": h.getHealthyProvidersCount(),
		},
		"timestamp": time.Now().Unix(),
	}

	return c.JSON(fiber.Map{
		"success": true,
		"data":    statsData,
	})
}

// GetProviderModels 获取提供商支持的模型列表
func (h *AdminHandler) GetProviderModels(c *fiber.Ctx) error {
	providerName := c.Params("provider")
	
	models := getProviderModels(providerName)
	
	return c.JSON(fiber.Map{
		"success": true,
		"provider": providerName,
		"models":   models,
	})
}

// 辅助函数
func (h *AdminHandler) getAllProviders() []providers.ProviderAdapter {
	providerNames := h.providerFactory.ListProviders()
	allProviders := make([]providers.ProviderAdapter, 0, len(providerNames))
	
	for _, name := range providerNames {
		if provider, exists := h.providerFactory.GetProvider(name); exists {
			allProviders = append(allProviders, provider)
		}
	}
	
	return allProviders
}

func (h *AdminHandler) getHealthyProvidersCount() int {
	// 简单的健康检查逻辑
	count := 0
	for _, name := range h.providerFactory.ListProviders() {
		if provider, exists := h.providerFactory.GetProvider(name); exists {
			baseProvider := getBaseProvider(provider)
			if baseProvider.APIKey != "" && baseProvider.BaseURL != "" {
				count++
			}
		}
	}
	return count
}

// 获取提供商的基础信息
func getBaseProvider(provider providers.ProviderAdapter) *providers.BaseProvider {
	// 使用类型断言获取BaseProvider
	switch p := provider.(type) {
	case *providers.OpenAIProvider:
		return &p.BaseProvider
	case *providers.GeminiProvider:
		return &p.BaseProvider
	case *providers.DeepSeekProvider:
		return &p.BaseProvider
	case *providers.QwenProvider:
		return &p.BaseProvider
	case *providers.MoonshotProvider:
		return &p.BaseProvider
	default:
		// 返回默认值
		return &providers.BaseProvider{
			Name:    provider.GetProviderName(),
			BaseURL: "unknown",
			Timeout: 30,
			Retries: 3,
		}
	}
}

// 获取提供商支持的模型
func getProviderModels(providerName string) []string {
	switch providerName {
	case "openai":
		return []string{"gpt-3.5-turbo", "gpt-4", "gpt-4-turbo"}
	case "gemini":
		return []string{"gemini-pro", "gemini-pro-vision"}
	case "deepseek":
		return []string{"deepseek-chat", "deepseek-coder"}
	case "qwen":
		return []string{"qwen-turbo", "qwen-plus", "qwen-max"}
	case "moonshot":
		return []string{"moonshot-v1-8k", "moonshot-v1-32k", "moonshot-v1-128k"}
	default:
		return []string{}
	}
}

// 获取提供商的默认模型
func getDefaultModelForProvider(providerName string) string {
	switch providerName {
	case "openai":
		return "gpt-3.5-turbo"
	case "gemini":
		return "gemini-pro"
	case "deepseek":
		return "deepseek-chat"
	case "qwen":
		return "qwen-turbo"
	case "moonshot":
		return "moonshot-v1-8k"
	default:
		return ""
	}
}

// 格式化字节数
func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// 格式化持续时间
func formatDuration(d time.Duration) string {
	days := int(d.Hours()) / 24
	hours := int(d.Hours()) % 24
	minutes := int(d.Minutes()) % 60
	
	if days > 0 {
		return fmt.Sprintf("%d天%d小时%d分钟", days, hours, minutes)
	} else if hours > 0 {
		return fmt.Sprintf("%d小时%d分钟", hours, minutes)
	} else {
		return fmt.Sprintf("%d分钟", minutes)
	}
}