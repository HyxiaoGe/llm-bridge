package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/heyanxiao/llm-bridge/pkg/types"
)

// GeminiProvider Gemini适配器实现
type GeminiProvider struct {
	BaseProvider
}

// NewGeminiProvider 创建Gemini提供商实例
func NewGeminiProvider(config *GeminiConfig) *GeminiProvider {
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com/v1beta"
	}
	
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30
	}
	
	retries := config.Retries
	if retries == 0 {
		retries = 3
	}

	return &GeminiProvider{
		BaseProvider: BaseProvider{
			Name:    "gemini",
			APIKey:  config.APIKey,
			BaseURL: baseURL,
			Headers: map[string]string{
				"Content-Type":    "application/json",
				"x-goog-api-key": config.APIKey,
			},
			Timeout: timeout,
			Retries: retries,
		},
	}
}

// GetProviderName 获取提供商名称
func (p *GeminiProvider) GetProviderName() string {
	return p.Name
}

// ValidateRequest 验证请求参数
func (p *GeminiProvider) ValidateRequest(req *types.UnifiedRequest) error {
	if req.Model == "" {
		return fmt.Errorf("模型名称不能为空")
	}
	
	if len(req.Messages) == 0 {
		return fmt.Errorf("消息列表不能为空")
	}
	
	// 验证温度参数范围
	if req.Parameters.Temperature < 0 || req.Parameters.Temperature > 2 {
		return fmt.Errorf("温度参数必须在0-2之间")
	}
	
	return nil
}

// Transform 将统一请求转换为Gemini格式
func (p *GeminiProvider) Transform(req *types.UnifiedRequest) ([]byte, error) {
	// Gemini使用简化的消息格式，只取最后一条用户消息
	var text string
	for i := len(req.Messages) - 1; i >= 0; i-- {
		msg := req.Messages[i]
		if msg.Role == "user" {
			text = msg.Content
			break
		}
	}
	
	if text == "" && len(req.Messages) > 0 {
		text = req.Messages[len(req.Messages)-1].Content
	}
	
	// 构建Gemini请求结构
	geminiReq := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{"text": text},
				},
			},
		},
	}
	
	// 添加生成配置
	generationConfig := make(map[string]interface{})
	
	if req.Parameters.Temperature > 0 {
		generationConfig["temperature"] = req.Parameters.Temperature
	}
	
	if req.Parameters.MaxTokens > 0 {
		generationConfig["maxOutputTokens"] = req.Parameters.MaxTokens
	}
	
	if req.Parameters.TopP > 0 {
		generationConfig["topP"] = req.Parameters.TopP
	}
	
	if len(req.Parameters.Stop) > 0 {
		generationConfig["stopSequences"] = req.Parameters.Stop
	}
	
	if len(generationConfig) > 0 {
		geminiReq["generationConfig"] = generationConfig
	}
	
	return json.Marshal(geminiReq)
}

// CallAPI 调用Gemini API
func (p *GeminiProvider) CallAPI(ctx context.Context, data []byte) (*http.Response, error) {
	// Gemini API URL格式: /v1beta/models/{model}:generateContent
	model := "gemini-2.0-flash-exp" // 使用可用的模型
	url := fmt.Sprintf("%s/models/%s:generateContent", p.BaseURL, model)
	
	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}
	
	// 设置请求头
	for key, value := range p.Headers {
		req.Header.Set(key, value)
	}
	
	// 创建HTTP客户端
	client := &http.Client{
		Timeout: time.Duration(p.Timeout) * time.Second,
	}
	
	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("调用Gemini API失败: %w", err)
	}
	
	return resp, nil
}

// ParseResponse 解析Gemini响应
func (p *GeminiProvider) ParseResponse(resp *http.Response) (*types.UnifiedResponse, error) {
	defer resp.Body.Close()
	
	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Gemini API返回错误状态码: %d", resp.StatusCode)
	}
	
	// 解析响应JSON
	var geminiResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&geminiResp); err != nil {
		return nil, fmt.Errorf("解析Gemini响应失败: %w", err)
	}
	
	// 检查错误
	if errorData, exists := geminiResp["error"]; exists {
		errorMap := errorData.(map[string]interface{})
		return &types.UnifiedResponse{
			Error: &types.Error{
				Code:    fmt.Sprintf("%v", errorMap["code"]),
				Message: fmt.Sprintf("%v", errorMap["message"]),
				Type:    "gemini_error",
			},
		}, nil
	}
	
	// 转换为统一响应格式
	unifiedResp := &types.UnifiedResponse{
		ID:      fmt.Sprintf("gemini-%d", time.Now().Unix()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   "gemini-pro",
	}
	
	// 解析candidates
	if candidatesData, exists := geminiResp["candidates"]; exists {
		candidatesArray := candidatesData.([]interface{})
		choices := make([]types.Choice, len(candidatesArray))
		
		for i, candidateData := range candidatesArray {
			candidateMap := candidateData.(map[string]interface{})
			
			var content string
			if contentData, exists := candidateMap["content"]; exists {
				contentMap := contentData.(map[string]interface{})
				if partsData, exists := contentMap["parts"]; exists {
					partsArray := partsData.([]interface{})
					if len(partsArray) > 0 {
						partMap := partsArray[0].(map[string]interface{})
						if text, exists := partMap["text"]; exists {
							content = text.(string)
						}
					}
				}
			}
			
			finishReason := "stop"
			if reasonData, exists := candidateMap["finishReason"]; exists {
				finishReason = reasonData.(string)
			}
			
			choices[i] = types.Choice{
				Index: i,
				Message: types.Message{
					Role:    "assistant",
					Content: content,
				},
				FinishReason: finishReason,
			}
		}
		unifiedResp.Choices = choices
	}
	
	// Gemini通常不返回详细的token使用统计
	unifiedResp.Usage = types.Usage{
		PromptTokens:     0, // Gemini API可能不提供此信息
		CompletionTokens: 0,
		TotalTokens:      0,
	}
	
	return unifiedResp, nil
}

// ParseStreamResponse 解析Gemini流式响应
func (p *GeminiProvider) ParseStreamResponse(resp *http.Response) (<-chan *types.StreamResponse, error) {
	responseChan := make(chan *types.StreamResponse)
	
	go func() {
		defer close(responseChan)
		defer resp.Body.Close()
		
		// TODO: 实现Gemini流式响应解析
		// Gemini的流式API格式与OpenAI不同，需要专门处理
	}()
	
	return responseChan, nil
}