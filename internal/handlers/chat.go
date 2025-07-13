package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/heyanxiao/llm-bridge/internal/providers"
	"github.com/heyanxiao/llm-bridge/internal/stats"
	"github.com/heyanxiao/llm-bridge/pkg/types"
)

// ChatHandler 聊天处理器
type ChatHandler struct {
	providerFactory *providers.ProviderFactory
	loadBalancer    providers.LoadBalancer
}

// NewChatHandler 创建聊天处理器实例
func NewChatHandler(factory *providers.ProviderFactory, balancer providers.LoadBalancer) *ChatHandler {
	return &ChatHandler{
		providerFactory: factory,
		loadBalancer:    balancer,
	}
}

// ChatCompletion 处理聊天补全请求
func (h *ChatHandler) ChatCompletion(c *fiber.Ctx) error {
	// 记录请求开始时间用于统计
	startTime := time.Now()

	// 解析请求体
	var req types.UnifiedRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "invalid_request",
				"message": "请求体格式错误: " + err.Error(),
				"type":    "invalid_request_error",
			},
		})
	}

	// 设置请求元数据
	req.Metadata.ClientIP = c.IP()
	req.Metadata.UserAgent = c.Get("User-Agent")
	req.Metadata.Timestamp = time.Now()

	// 处理提供商和模型的四种情况
	var provider providers.ProviderAdapter
	
	// 情况4：只指定了model但没有provider - 返回错误
	if req.Model != "" && req.Provider == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "missing_provider",
				"message": "指定模型时必须同时指定提供商(provider)参数",
				"type":    "invalid_request_error",
			},
		})
	}
	
	if req.Provider != "" {
		// 情况2&3：指定了提供商
		var exists bool
		provider, exists = h.providerFactory.GetProvider(req.Provider)
		if !exists {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "invalid_provider",
					"message": "不支持的LLM提供商: " + req.Provider,
					"type":    "invalid_request_error",
				},
			})
		}
		
		// 情况2：有provider但没有model - 使用默认模型
		if req.Model == "" {
			defaultModel := providers.GetDefaultModel(req.Provider)
			if defaultModel == "" {
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": fiber.Map{
						"code":    "no_default_model",
						"message": "提供商 " + req.Provider + " 没有配置默认模型",
						"type":    "internal_server_error",
					},
				})
			}
			req.Model = defaultModel
		}
		// 情况3：有provider和model - 正常处理（无需额外操作）
	} else {
		// 情况1：没有provider和model - 负载均衡选择
		allProviders := h.getAllProviders()
		provider = h.loadBalancer.SelectProvider(allProviders)
		if provider == nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "no_provider_available",
					"message": "当前没有可用的LLM提供商",
					"type":    "service_unavailable_error",
				},
			})
		}
		req.Provider = provider.GetProviderName()
		
		// 使用负载均衡选中提供商的默认模型
		defaultModel := providers.GetDefaultModel(req.Provider)
		if defaultModel == "" {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fiber.Map{
					"code":    "no_default_model",
					"message": "提供商 " + req.Provider + " 没有配置默认模型",
					"type":    "internal_server_error",
				},
			})
		}
		req.Model = defaultModel
	}

	// 验证请求参数
	if err := provider.ValidateRequest(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "invalid_request",
				"message": "请求参数验证失败: " + err.Error(),
				"type":    "invalid_request_error",
			},
		})
	}

	// 转换请求格式
	providerData, err := provider.Transform(&req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "transformation_error",
				"message": "请求格式转换失败: " + err.Error(),
				"type":    "internal_server_error",
			},
		})
	}

	// 创建请求上下文
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 调用LLM API
	resp, err := provider.CallAPI(ctx, providerData)
	if err != nil {
		// 更新提供商健康状态
		h.loadBalancer.UpdateHealth(provider.GetProviderName(), false)
		
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "api_call_failed",
				"message": "调用LLM API失败: " + err.Error(),
				"type":    "service_unavailable_error",
			},
		})
	}

	// 检查是否为流式请求
	if req.Parameters.Stream {
		return h.handleStreamResponse(c, provider, resp)
	}

	// 解析响应
	unifiedResp, err := provider.ParseResponse(resp)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fiber.Map{
				"code":    "response_parse_error",
				"message": "响应解析失败: " + err.Error(),
				"type":    "internal_server_error",
			},
		})
	}

	// 更新提供商健康状态为正常
	h.loadBalancer.UpdateHealth(provider.GetProviderName(), true)

	// 记录统计数据到Redis
	responseTime := time.Since(startTime)
	tokens := unifiedResp.Usage.TotalTokens
	
	// 记录统计
	if redisMetrics := stats.GetRedisMetrics(); redisMetrics != nil {
		redisMetrics.IncrementRequest(provider.GetProviderName(), responseTime, tokens)
	}

	// 返回统一格式的响应
	return c.JSON(unifiedResp)
}

// handleStreamResponse 处理流式响应
func (h *ChatHandler) handleStreamResponse(c *fiber.Ctx, provider providers.ProviderAdapter, resp *http.Response) error {
	// 设置流式响应头
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")
	c.Set("Access-Control-Allow-Origin", "*")
	c.Set("Access-Control-Allow-Headers", "Cache-Control")

	// 获取流式响应channel
	streamChan, err := provider.ParseStreamResponse(resp)
	if err != nil {
		// 发送错误事件
		errorEvent := fmt.Sprintf("data: {\"error\":{\"code\":\"stream_parse_error\",\"message\":\"流式响应解析失败: %s\"}}\n\n", err.Error())
		return c.SendString(errorEvent)
	}

	// 逐个发送流式事件
	for streamResp := range streamChan {
		// 将StreamResponse转换为JSON
		jsonData, err := json.Marshal(streamResp)
		if err != nil {
			continue // 跳过无法序列化的响应
		}
		
		// 发送SSE事件
		event := fmt.Sprintf("data: %s\n\n", string(jsonData))
		if err := c.SendString(event); err != nil {
			// 客户端断开连接
			break
		}
		
		// 检查是否完成
		if len(streamResp.Choices) > 0 && streamResp.Choices[0].FinishReason != "" {
			break
		}
	}

	// 发送完成事件
	return c.SendString("data: [DONE]\n\n")
}

// getAllProviders 获取所有可用的提供商
func (h *ChatHandler) getAllProviders() []providers.ProviderAdapter {
	providerNames := h.providerFactory.ListProviders()
	allProviders := make([]providers.ProviderAdapter, 0, len(providerNames))
	
	for _, name := range providerNames {
		if provider, exists := h.providerFactory.GetProvider(name); exists {
			allProviders = append(allProviders, provider)
		}
	}
	
	return allProviders
}

// Models 获取支持的模型列表
func (h *ChatHandler) Models(c *fiber.Ctx) error {
	// 返回支持的模型列表
	// 这里可以根据已注册的提供商动态生成
	models := []map[string]interface{}{
		{
			"id":      "gpt-3.5-turbo",
			"object":  "model",
			"created": time.Now().Unix(),
			"owned_by": "openai",
			"provider": "openai",
		},
		{
			"id":      "gpt-4",
			"object":  "model", 
			"created": time.Now().Unix(),
			"owned_by": "openai",
			"provider": "openai",
		},
		{
			"id":      "claude-3-sonnet",
			"object":  "model",
			"created": time.Now().Unix(),
			"owned_by": "anthropic",
			"provider": "claude",
		},
	}

	return c.JSON(fiber.Map{
		"object": "list",
		"data":   models,
	})
}