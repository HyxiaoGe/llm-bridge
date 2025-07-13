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

// MoonshotProvider 月之暗面适配器实现
type MoonshotProvider struct {
	BaseProvider
}

// MoonshotConfig 月之暗面配置
type MoonshotConfig struct {
	APIKey  string `yaml:"api_key"`
	BaseURL string `yaml:"base_url"`
	Timeout int    `yaml:"timeout"`
	Retries int    `yaml:"retries"`
}

// NewMoonshotProvider 创建月之暗面提供商实例
func NewMoonshotProvider(config *MoonshotConfig) *MoonshotProvider {
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.moonshot.cn/v1"
	}
	
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 60 // 月之暗面支持长文本，增加超时时间
	}
	
	retries := config.Retries
	if retries == 0 {
		retries = 3
	}

	return &MoonshotProvider{
		BaseProvider: BaseProvider{
			Name:    "moonshot",
			APIKey:  config.APIKey,
			BaseURL: baseURL,
			Headers: map[string]string{
				"Content-Type":  "application/json",
				"Authorization": "Bearer " + config.APIKey,
			},
			Timeout: timeout,
			Retries: retries,
		},
	}
}

// GetProviderName 获取提供商名称
func (p *MoonshotProvider) GetProviderName() string {
	return p.Name
}

// ValidateRequest 验证请求参数
func (p *MoonshotProvider) ValidateRequest(req *types.UnifiedRequest) error {
	if req.Model == "" {
		return fmt.Errorf("模型名称不能为空")
	}
	
	if len(req.Messages) == 0 {
		return fmt.Errorf("消息列表不能为空")
	}
	
	// 验证温度参数范围
	if req.Parameters.Temperature < 0 || req.Parameters.Temperature > 1 {
		return fmt.Errorf("温度参数必须在0-1之间")
	}
	
	return nil
}

// Transform 将统一请求转换为月之暗面格式（与OpenAI兼容）
func (p *MoonshotProvider) Transform(req *types.UnifiedRequest) ([]byte, error) {
	// 月之暗面API与OpenAI格式兼容
	// 如果没有指定模型，使用默认模型
	model := req.Model
	if model == "" || model == "moonshot" {
		model = "moonshot-v1-8k" // 默认使用8k上下文模型
	}
	
	// 模型名称映射
	switch model {
	case "moonshot-8k":
		model = "moonshot-v1-8k"
	case "moonshot-32k":
		model = "moonshot-v1-32k"
	case "moonshot-128k":
		model = "moonshot-v1-128k"
	}
	
	// 构建月之暗面请求结构
	moonshotReq := map[string]interface{}{
		"model":    model,
		"messages": req.Messages,
	}
	
	// 添加可选参数
	if req.Parameters.Temperature > 0 {
		moonshotReq["temperature"] = req.Parameters.Temperature
	}
	
	if req.Parameters.MaxTokens > 0 {
		moonshotReq["max_tokens"] = req.Parameters.MaxTokens
	}
	
	if req.Parameters.TopP > 0 {
		moonshotReq["top_p"] = req.Parameters.TopP
	}
	
	if req.Parameters.Stream {
		moonshotReq["stream"] = true
	}
	
	if req.Parameters.FrequencyPenalty != 0 {
		moonshotReq["frequency_penalty"] = req.Parameters.FrequencyPenalty
	}
	
	if req.Parameters.PresencePenalty != 0 {
		moonshotReq["presence_penalty"] = req.Parameters.PresencePenalty
	}
	
	if len(req.Parameters.Stop) > 0 {
		moonshotReq["stop"] = req.Parameters.Stop
	}
	
	// 添加用户ID（如果存在）
	if req.Metadata.UserID != "" {
		moonshotReq["user"] = req.Metadata.UserID
	}
	
	return json.Marshal(moonshotReq)
}

// CallAPI 调用月之暗面 API
func (p *MoonshotProvider) CallAPI(ctx context.Context, data []byte) (*http.Response, error) {
	url := p.BaseURL + "/chat/completions"
	
	// 创建HTTP请求
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}
	
	// 设置请求头
	for key, value := range p.Headers {
		req.Header.Set(key, value)
	}
	
	// 创建HTTP客户端（月之暗面支持长文本，需要更长的超时时间）
	client := &http.Client{
		Timeout: time.Duration(p.Timeout) * time.Second,
	}
	
	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("调用月之暗面 API失败: %w", err)
	}
	
	return resp, nil
}

// ParseResponse 解析月之暗面响应（与OpenAI格式兼容）
func (p *MoonshotProvider) ParseResponse(resp *http.Response) (*types.UnifiedResponse, error) {
	defer resp.Body.Close()
	
	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		// 读取错误响应
		var errorResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			return nil, fmt.Errorf("月之暗面 API返回错误状态码: %d", resp.StatusCode)
		}
		
		// 解析错误信息
		if errorData, exists := errorResp["error"]; exists {
			errorMap := errorData.(map[string]interface{})
			return &types.UnifiedResponse{
				Error: &types.Error{
					Code:    fmt.Sprintf("%v", errorMap["code"]),
					Message: fmt.Sprintf("%v", errorMap["message"]),
					Type:    "moonshot_error",
				},
			}, nil
		}
		
		return nil, fmt.Errorf("月之暗面 API返回错误状态码: %d", resp.StatusCode)
	}
	
	// 解析响应JSON
	var moonshotResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&moonshotResp); err != nil {
		return nil, fmt.Errorf("解析月之暗面响应失败: %w", err)
	}
	
	// 检查响应中的错误
	if errorData, exists := moonshotResp["error"]; exists {
		errorMap := errorData.(map[string]interface{})
		return &types.UnifiedResponse{
			Error: &types.Error{
				Code:    fmt.Sprintf("%v", errorMap["code"]),
				Message: fmt.Sprintf("%v", errorMap["message"]),
				Type:    "moonshot_error",
			},
		}, nil
	}
	
	// 转换为统一响应格式
	unifiedResp := &types.UnifiedResponse{
		ID:      moonshotResp["id"].(string),
		Object:  moonshotResp["object"].(string),
		Created: int64(moonshotResp["created"].(float64)),
		Model:   moonshotResp["model"].(string),
	}
	
	// 解析choices
	if choicesData, exists := moonshotResp["choices"]; exists {
		choicesArray := choicesData.([]interface{})
		choices := make([]types.Choice, len(choicesArray))
		
		for i, choiceData := range choicesArray {
			choiceMap := choiceData.(map[string]interface{})
			messageMap := choiceMap["message"].(map[string]interface{})
			
			choices[i] = types.Choice{
				Index: int(choiceMap["index"].(float64)),
				Message: types.Message{
					Role:    messageMap["role"].(string),
					Content: messageMap["content"].(string),
				},
				FinishReason: fmt.Sprintf("%v", choiceMap["finish_reason"]),
			}
		}
		unifiedResp.Choices = choices
	}
	
	// 解析usage
	if usageData, exists := moonshotResp["usage"]; exists {
		usageMap := usageData.(map[string]interface{})
		unifiedResp.Usage = types.Usage{
			PromptTokens:     int(usageMap["prompt_tokens"].(float64)),
			CompletionTokens: int(usageMap["completion_tokens"].(float64)),
			TotalTokens:      int(usageMap["total_tokens"].(float64)),
		}
	}
	
	return unifiedResp, nil
}

// ParseStreamResponse 解析月之暗面流式响应
func (p *MoonshotProvider) ParseStreamResponse(resp *http.Response) (<-chan *types.StreamResponse, error) {
	responseChan := make(chan *types.StreamResponse)
	
	go func() {
		defer close(responseChan)
		defer resp.Body.Close()
		
		// TODO: 实现SSE(Server-Sent Events)解析逻辑
		// 月之暗面的流式响应格式与OpenAI兼容
	}()
	
	return responseChan, nil
}