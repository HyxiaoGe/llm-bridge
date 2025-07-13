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

// QwenProvider 通义千问适配器实现
type QwenProvider struct {
	BaseProvider
}

// QwenConfig 通义千问配置
type QwenConfig struct {
	APIKey  string `yaml:"api_key"`
	BaseURL string `yaml:"base_url"`
	Timeout int    `yaml:"timeout"`
	Retries int    `yaml:"retries"`
}

// NewQwenProvider 创建通义千问提供商实例
func NewQwenProvider(config *QwenConfig) *QwenProvider {
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://dashscope.aliyuncs.com/api/v1/services/aigc/text-generation"
	}
	
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30
	}
	
	retries := config.Retries
	if retries == 0 {
		retries = 3
	}

	return &QwenProvider{
		BaseProvider: BaseProvider{
			Name:    "qwen",
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
func (p *QwenProvider) GetProviderName() string {
	return p.Name
}

// ValidateRequest 验证请求参数
func (p *QwenProvider) ValidateRequest(req *types.UnifiedRequest) error {
	if req.Model == "" {
		return fmt.Errorf("模型名称不能为空")
	}
	
	if len(req.Messages) == 0 {
		return fmt.Errorf("消息列表不能为空")
	}
	
	return nil
}

// Transform 将统一请求转换为通义千问格式
func (p *QwenProvider) Transform(req *types.UnifiedRequest) ([]byte, error) {
	// 通义千问使用不同的API格式
	// 默认模型
	model := req.Model
	if model == "" || model == "qwen" {
		model = "qwen-turbo"
	}
	
	// 构建通义千问请求结构
	qwenReq := map[string]interface{}{
		"model": model,
		"input": map[string]interface{}{
			"messages": req.Messages,
		},
		"parameters": map[string]interface{}{},
	}
	
	// 添加参数
	params := qwenReq["parameters"].(map[string]interface{})
	
	if req.Parameters.Temperature > 0 {
		params["temperature"] = req.Parameters.Temperature
	}
	
	if req.Parameters.MaxTokens > 0 {
		params["max_tokens"] = req.Parameters.MaxTokens
	}
	
	if req.Parameters.TopP > 0 {
		params["top_p"] = req.Parameters.TopP
	}
	
	if len(req.Parameters.Stop) > 0 {
		params["stop"] = req.Parameters.Stop
	}
	
	// 通义千问支持增量输出
	if req.Parameters.Stream {
		params["incremental_output"] = true
	}
	
	return json.Marshal(qwenReq)
}

// CallAPI 调用通义千问 API
func (p *QwenProvider) CallAPI(ctx context.Context, data []byte) (*http.Response, error) {
	url := p.BaseURL + "/generation"
	
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
		return nil, fmt.Errorf("调用通义千问 API失败: %w", err)
	}
	
	return resp, nil
}

// ParseResponse 解析通义千问响应
func (p *QwenProvider) ParseResponse(resp *http.Response) (*types.UnifiedResponse, error) {
	defer resp.Body.Close()
	
	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		// 读取错误响应
		var errorResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err != nil {
			return nil, fmt.Errorf("通义千问 API返回错误状态码: %d", resp.StatusCode)
		}
		
		// 提取错误信息
		var errorMsg string
		if msg, exists := errorResp["message"]; exists {
			errorMsg = fmt.Sprintf("%v", msg)
		} else {
			errorMsg = fmt.Sprintf("API错误: %d", resp.StatusCode)
		}
		
		return &types.UnifiedResponse{
			Error: &types.Error{
				Code:    fmt.Sprintf("%d", resp.StatusCode),
				Message: errorMsg,
				Type:    "qwen_error",
			},
		}, nil
	}
	
	// 解析响应JSON
	var qwenResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&qwenResp); err != nil {
		return nil, fmt.Errorf("解析通义千问响应失败: %w", err)
	}
	
	// 转换为统一响应格式
	unifiedResp := &types.UnifiedResponse{
		ID:      fmt.Sprintf("qwen-%d", time.Now().Unix()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   "", // 从响应中获取
	}
	
	// 解析输出
	if outputData, exists := qwenResp["output"]; exists {
		outputMap := outputData.(map[string]interface{})
		
		// 获取文本内容
		var content string
		if textData, exists := outputMap["text"]; exists {
			content = textData.(string)
		} else if choicesData, exists := outputMap["choices"]; exists {
			// 有些情况下返回choices格式
			choicesArray := choicesData.([]interface{})
			if len(choicesArray) > 0 {
				choiceMap := choicesArray[0].(map[string]interface{})
				if messageData, exists := choiceMap["message"]; exists {
					messageMap := messageData.(map[string]interface{})
					content = messageMap["content"].(string)
				}
			}
		}
		
		// 获取结束原因
		finishReason := "stop"
		if reasonData, exists := outputMap["finish_reason"]; exists {
			finishReason = reasonData.(string)
		}
		
		unifiedResp.Choices = []types.Choice{
			{
				Index: 0,
				Message: types.Message{
					Role:    "assistant",
					Content: content,
				},
				FinishReason: finishReason,
			},
		}
	}
	
	// 解析使用情况
	if usageData, exists := qwenResp["usage"]; exists {
		usageMap := usageData.(map[string]interface{})
		
		var promptTokens, completionTokens, totalTokens int
		
		if val, exists := usageMap["input_tokens"]; exists {
			promptTokens = int(val.(float64))
		}
		if val, exists := usageMap["output_tokens"]; exists {
			completionTokens = int(val.(float64))
		}
		if val, exists := usageMap["total_tokens"]; exists {
			totalTokens = int(val.(float64))
		} else {
			totalTokens = promptTokens + completionTokens
		}
		
		unifiedResp.Usage = types.Usage{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      totalTokens,
		}
	}
	
	// 获取请求ID（如果有）
	if requestID, exists := qwenResp["request_id"]; exists {
		unifiedResp.ID = requestID.(string)
	}
	
	return unifiedResp, nil
}

// ParseStreamResponse 解析通义千问流式响应
func (p *QwenProvider) ParseStreamResponse(resp *http.Response) (<-chan *types.StreamResponse, error) {
	responseChan := make(chan *types.StreamResponse)
	
	go func() {
		defer close(responseChan)
		defer resp.Body.Close()
		
		// TODO: 实现通义千问的流式响应解析
		// 通义千问使用SSE格式，但数据格式与OpenAI略有不同
	}()
	
	return responseChan, nil
}