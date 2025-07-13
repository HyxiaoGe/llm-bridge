package providers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/heyanxiao/llm-bridge/pkg/types"
)

// DeepSeekProvider DeepSeek适配器实现
type DeepSeekProvider struct {
	BaseProvider
}

// DeepSeekConfig DeepSeek配置
type DeepSeekConfig struct {
	APIKey  string `yaml:"api_key"`
	BaseURL string `yaml:"base_url"`
	Timeout int    `yaml:"timeout"`
	Retries int    `yaml:"retries"`
}

// NewDeepSeekProvider 创建DeepSeek提供商实例
func NewDeepSeekProvider(config *DeepSeekConfig) *DeepSeekProvider {
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.deepseek.com/v1"
	}
	
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30
	}
	
	retries := config.Retries
	if retries == 0 {
		retries = 3
	}

	return &DeepSeekProvider{
		BaseProvider: BaseProvider{
			Name:    "deepseek",
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
func (p *DeepSeekProvider) GetProviderName() string {
	return p.Name
}

// ValidateRequest 验证请求参数
func (p *DeepSeekProvider) ValidateRequest(req *types.UnifiedRequest) error {
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
	
	// 验证TopP参数范围
	if req.Parameters.TopP < 0 || req.Parameters.TopP > 1 {
		return fmt.Errorf("TopP参数必须在0-1之间")
	}
	
	return nil
}

// Transform 将统一请求转换为DeepSeek格式（与OpenAI兼容）
func (p *DeepSeekProvider) Transform(req *types.UnifiedRequest) ([]byte, error) {
	// DeepSeek API与OpenAI格式兼容
	// 如果没有指定模型，使用默认模型
	model := req.Model
	if model == "" || model == "deepseek" {
		model = GetDefaultModel("deepseek")
	}
	
	// 构建DeepSeek请求结构
	deepseekReq := map[string]interface{}{
		"model":    model,
		"messages": req.Messages,
	}
	
	// 添加可选参数
	if req.Parameters.Temperature > 0 {
		deepseekReq["temperature"] = req.Parameters.Temperature
	}
	
	if req.Parameters.MaxTokens > 0 {
		deepseekReq["max_tokens"] = req.Parameters.MaxTokens
	}
	
	if req.Parameters.TopP > 0 {
		deepseekReq["top_p"] = req.Parameters.TopP
	}
	
	if req.Parameters.Stream {
		deepseekReq["stream"] = true
	}
	
	if req.Parameters.FrequencyPenalty != 0 {
		deepseekReq["frequency_penalty"] = req.Parameters.FrequencyPenalty
	}
	
	if req.Parameters.PresencePenalty != 0 {
		deepseekReq["presence_penalty"] = req.Parameters.PresencePenalty
	}
	
	if len(req.Parameters.Stop) > 0 {
		deepseekReq["stop"] = req.Parameters.Stop
	}
	
	// 添加推理相关参数 (适用于deepseek-reasoner模型)
	if req.Parameters.Reasoning {
		deepseekReq["reasoning"] = true
	}
	
	if req.Parameters.ReasoningEffort != "" {
		deepseekReq["reasoning_effort"] = req.Parameters.ReasoningEffort
	}
	
	return json.Marshal(deepseekReq)
}

// CallAPI 调用DeepSeek API
func (p *DeepSeekProvider) CallAPI(ctx context.Context, data []byte) (*http.Response, error) {
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
	
	// 创建HTTP客户端
	client := &http.Client{
		Timeout: time.Duration(p.Timeout) * time.Second,
	}
	
	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("调用DeepSeek API失败: %w", err)
	}
	
	return resp, nil
}

// ParseResponse 解析DeepSeek响应（与OpenAI格式兼容）
func (p *DeepSeekProvider) ParseResponse(resp *http.Response) (*types.UnifiedResponse, error) {
	defer resp.Body.Close()
	
	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("DeepSeek API返回错误状态码: %d", resp.StatusCode)
	}
	
	// 解析响应JSON
	var deepseekResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&deepseekResp); err != nil {
		return nil, fmt.Errorf("解析DeepSeek响应失败: %w", err)
	}
	
	// 检查错误
	if errorData, exists := deepseekResp["error"]; exists {
		errorMap := errorData.(map[string]interface{})
		return &types.UnifiedResponse{
			Error: &types.Error{
				Code:    fmt.Sprintf("%v", errorMap["code"]),
				Message: fmt.Sprintf("%v", errorMap["message"]),
				Type:    "deepseek_error",
			},
		}, nil
	}
	
	// 转换为统一响应格式
	unifiedResp := &types.UnifiedResponse{
		ID:      deepseekResp["id"].(string),
		Object:  deepseekResp["object"].(string),
		Created: int64(deepseekResp["created"].(float64)),
		Model:   deepseekResp["model"].(string),
	}
	
	// 解析choices
	if choicesData, exists := deepseekResp["choices"]; exists {
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
	if usageData, exists := deepseekResp["usage"]; exists {
		usageMap := usageData.(map[string]interface{})
		unifiedResp.Usage = types.Usage{
			PromptTokens:     int(usageMap["prompt_tokens"].(float64)),
			CompletionTokens: int(usageMap["completion_tokens"].(float64)),
			TotalTokens:      int(usageMap["total_tokens"].(float64)),
		}
	}
	
	return unifiedResp, nil
}

// ParseStreamResponse 解析DeepSeek流式响应
func (p *DeepSeekProvider) ParseStreamResponse(resp *http.Response) (<-chan *types.StreamResponse, error) {
	responseChan := make(chan *types.StreamResponse)
	
	go func() {
		defer close(responseChan)
		defer resp.Body.Close()
		
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			
			// 跳过空行
			if line == "" {
				continue
			}
			
			// 检查是否是SSE数据行
			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimPrefix(line, "data: ")
				
				// 检查是否是结束标志
				if data == "[DONE]" {
					break
				}
				
				// 解析JSON数据
				var deepseekStreamResp map[string]interface{}
				if err := json.Unmarshal([]byte(data), &deepseekStreamResp); err != nil {
					continue // 跳过无效的JSON
				}
				
				// 转换为统一格式
				streamResp := p.convertDeepSeekStreamResponse(deepseekStreamResp)
				if streamResp != nil {
					responseChan <- streamResp
				}
			}
		}
	}()
	
	return responseChan, nil
}

// convertDeepSeekStreamResponse 转换DeepSeek流式响应为统一格式
func (p *DeepSeekProvider) convertDeepSeekStreamResponse(data map[string]interface{}) *types.StreamResponse {
	streamResp := &types.StreamResponse{
		Object:  "chat.completion.chunk",
		Model:   p.Name,
		Created: time.Now().Unix(),
	}
	
	// 提取ID
	if id, ok := data["id"].(string); ok {
		streamResp.ID = id
	}
	
	// 提取模型
	if model, ok := data["model"].(string); ok {
		streamResp.Model = model
	}
	
	// 提取选择
	if choices, ok := data["choices"].([]interface{}); ok && len(choices) > 0 {
		choice := choices[0].(map[string]interface{})
		
		streamChoice := types.StreamChoice{
			Index: 0,
		}
		
		// 提取增量内容
		if delta, ok := choice["delta"].(map[string]interface{}); ok {
			streamDelta := types.StreamDelta{}
			
			if role, ok := delta["role"].(string); ok {
				streamDelta.Role = role
			}
			
			if content, ok := delta["content"].(string); ok {
				streamDelta.Content = content
			}
			
			// 检查是否有推理内容 (适用于deepseek-reasoner模型)
			if reasoning, ok := delta["reasoning"].(string); ok {
				streamDelta.Reasoning = reasoning
			}
			
			streamChoice.Delta = streamDelta
		}
		
		// 提取完成原因
		if finishReason, ok := choice["finish_reason"].(string); ok {
			streamChoice.FinishReason = finishReason
		}
		
		streamResp.Choices = []types.StreamChoice{streamChoice}
	}
	
	return streamResp
}