package types

import "time"

// 统一请求结构 - 屏蔽各LLM平台差异
type UnifiedRequest struct {
	Model      string     `json:"model" validate:"required"`       // 模型名称
	Messages   []Message  `json:"messages" validate:"required"`    // 对话消息列表
	Parameters Parameters `json:"parameters"`                      // 请求参数
	Provider   string     `json:"provider" validate:"required"`    // 指定的LLM提供商
	Metadata   Metadata   `json:"metadata"`                        // 请求元数据
}

// 消息结构
type Message struct {
	Role    string `json:"role" validate:"required"`    // 角色: system, user, assistant
	Content string `json:"content" validate:"required"` // 消息内容
}

// 请求参数
type Parameters struct {
	Temperature      float64 `json:"temperature,omitempty"`       // 温度参数 (0.0-2.0)
	MaxTokens        int     `json:"max_tokens,omitempty"`        // 最大输出token数
	TopP             float64 `json:"top_p,omitempty"`             // 核采样参数
	Stream           bool    `json:"stream,omitempty"`            // 是否流式输出
	FrequencyPenalty float64 `json:"frequency_penalty,omitempty"` // 频率惩罚
	PresencePenalty  float64 `json:"presence_penalty,omitempty"`  // 存在惩罚
	Stop             []string `json:"stop,omitempty"`             // 停止序列
	Reasoning        bool    `json:"reasoning,omitempty"`         // 是否输出推理过程 (适用于o1、deepseek-reasoner等模型)
	ReasoningEffort  string  `json:"reasoning_effort,omitempty"`  // 推理强度: low, medium, high (适用于部分模型)
}

// 请求元数据
type Metadata struct {
	UserID    string            `json:"user_id,omitempty"`    // 用户ID
	SessionID string            `json:"session_id,omitempty"` // 会话ID
	ClientIP  string            `json:"client_ip,omitempty"`  // 客户端IP
	UserAgent string            `json:"user_agent,omitempty"` // 用户代理
	Headers   map[string]string `json:"headers,omitempty"`    // 自定义请求头
	Timestamp time.Time         `json:"timestamp"`            // 请求时间戳
}

// 统一响应结构
type UnifiedResponse struct {
	ID      string   `json:"id"`                // 响应ID
	Object  string   `json:"object"`            // 对象类型
	Created int64    `json:"created"`           // 创建时间戳
	Model   string   `json:"model"`             // 使用的模型
	Choices []Choice `json:"choices"`           // 生成的选择列表
	Usage   Usage    `json:"usage"`             // token使用统计
	Error   *Error   `json:"error,omitempty"`   // 错误信息
}

// 选择结构
type Choice struct {
	Index        int     `json:"index"`         // 选择索引
	Message      Message `json:"message"`       // 响应消息
	FinishReason string  `json:"finish_reason"` // 完成原因
}

// 使用统计
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`     // 输入token数
	CompletionTokens int `json:"completion_tokens"` // 输出token数
	TotalTokens      int `json:"total_tokens"`      // 总token数
}

// 错误信息
type Error struct {
	Code    string `json:"code"`    // 错误代码
	Message string `json:"message"` // 错误消息
	Type    string `json:"type"`    // 错误类型
}

// 流式响应结构
type StreamResponse struct {
	ID      string        `json:"id"`      // 响应ID
	Object  string        `json:"object"`  // 对象类型
	Created int64         `json:"created"` // 创建时间戳
	Model   string        `json:"model"`   // 使用的模型
	Choices []StreamChoice `json:"choices"` // 流式选择
}

// 流式选择结构
type StreamChoice struct {
	Index int          `json:"index"` // 选择索引
	Delta StreamDelta  `json:"delta"` // 增量内容
	FinishReason string `json:"finish_reason,omitempty"` // 完成原因
}

// 流式增量内容
type StreamDelta struct {
	Role      string `json:"role,omitempty"`      // 角色
	Content   string `json:"content,omitempty"`   // 内容片段
	Reasoning string `json:"reasoning,omitempty"` // 推理过程片段 (适用于推理模型)
}