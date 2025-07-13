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

// OpenAIProvider OpenAI适配器实现
type OpenAIProvider struct {
	BaseProvider
}

// NewOpenAIProvider 创建OpenAI提供商实例
func NewOpenAIProvider(config *OpenAIConfig) *OpenAIProvider {
	baseURL := config.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	
	timeout := config.Timeout
	if timeout == 0 {
		timeout = 30
	}
	
	retries := config.Retries
	if retries == 0 {
		retries = 3
	}

	return &OpenAIProvider{
		BaseProvider: BaseProvider{
			Name:    "openai",
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
func (p *OpenAIProvider) GetProviderName() string {
	return p.Name
}

// ValidateRequest 验证请求参数
func (p *OpenAIProvider) ValidateRequest(req *types.UnifiedRequest) error {
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

// Transform 将统一请求转换为OpenAI格式
func (p *OpenAIProvider) Transform(req *types.UnifiedRequest) ([]byte, error) {
	// 构建OpenAI请求结构
	openaiReq := map[string]interface{}{
		"model":    req.Model,
		"messages": req.Messages,
	}
	
	// 添加可选参数
	if req.Parameters.Temperature > 0 {
		openaiReq["temperature"] = req.Parameters.Temperature
	}
	
	if req.Parameters.MaxTokens > 0 {
		openaiReq["max_tokens"] = req.Parameters.MaxTokens
	}
	
	if req.Parameters.TopP > 0 {
		openaiReq["top_p"] = req.Parameters.TopP
	}
	
	if req.Parameters.Stream {
		openaiReq["stream"] = true
	}
	
	if req.Parameters.FrequencyPenalty != 0 {
		openaiReq["frequency_penalty"] = req.Parameters.FrequencyPenalty
	}
	
	if req.Parameters.PresencePenalty != 0 {
		openaiReq["presence_penalty"] = req.Parameters.PresencePenalty
	}
	
	if len(req.Parameters.Stop) > 0 {
		openaiReq["stop"] = req.Parameters.Stop
	}
	
	// 添加推理相关参数 (适用于o1等推理模型)
	if req.Parameters.Reasoning {
		openaiReq["reasoning"] = true
	}
	
	if req.Parameters.ReasoningEffort != "" {
		openaiReq["reasoning_effort"] = req.Parameters.ReasoningEffort
	}
	
	// 添加用户ID（如果存在）
	if req.Metadata.UserID != "" {
		openaiReq["user"] = req.Metadata.UserID
	}
	
	return json.Marshal(openaiReq)
}

// CallAPI 调用OpenAI API
func (p *OpenAIProvider) CallAPI(ctx context.Context, data []byte) (*http.Response, error) {
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
		return nil, fmt.Errorf("调用OpenAI API失败: %w", err)
	}
	
	return resp, nil
}

// ParseResponse 解析OpenAI响应
func (p *OpenAIProvider) ParseResponse(resp *http.Response) (*types.UnifiedResponse, error) {
	defer resp.Body.Close()
	
	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OpenAI API返回错误状态码: %d", resp.StatusCode)
	}
	
	// 解析响应JSON
	var openaiResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&openaiResp); err != nil {
		return nil, fmt.Errorf("解析OpenAI响应失败: %w", err)
	}
	
	// 检查错误
	if errorData, exists := openaiResp["error"]; exists {
		errorMap := errorData.(map[string]interface{})
		return &types.UnifiedResponse{
			Error: &types.Error{
				Code:    errorMap["code"].(string),
				Message: errorMap["message"].(string),
				Type:    errorMap["type"].(string),
			},
		}, nil
	}
	
	// 转换为统一响应格式
	unifiedResp := &types.UnifiedResponse{
		ID:      openaiResp["id"].(string),
		Object:  openaiResp["object"].(string),
		Created: int64(openaiResp["created"].(float64)),
		Model:   openaiResp["model"].(string),
	}
	
	// 解析choices
	if choicesData, exists := openaiResp["choices"]; exists {
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
				FinishReason: choiceMap["finish_reason"].(string),
			}
		}
		unifiedResp.Choices = choices
	}
	
	// 解析usage
	if usageData, exists := openaiResp["usage"]; exists {
		usageMap := usageData.(map[string]interface{})
		unifiedResp.Usage = types.Usage{
			PromptTokens:     int(usageMap["prompt_tokens"].(float64)),
			CompletionTokens: int(usageMap["completion_tokens"].(float64)),
			TotalTokens:      int(usageMap["total_tokens"].(float64)),
		}
	}
	
	return unifiedResp, nil
}

// ParseStreamResponse 解析OpenAI流式响应
func (p *OpenAIProvider) ParseStreamResponse(resp *http.Response) (<-chan *types.StreamResponse, error) {
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
				var openaiStreamResp map[string]interface{}
				if err := json.Unmarshal([]byte(data), &openaiStreamResp); err != nil {
					continue // 跳过无效的JSON
				}
				
				// 转换为统一格式
				streamResp := p.convertOpenAIStreamResponse(openaiStreamResp)
				if streamResp != nil {
					responseChan <- streamResp
				}
			}
		}
	}()
	
	return responseChan, nil
}

// convertOpenAIStreamResponse 转换OpenAI流式响应为统一格式
func (p *OpenAIProvider) convertOpenAIStreamResponse(data map[string]interface{}) *types.StreamResponse {
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
			
			// 检查是否有推理内容 (适用于o1等模型)
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