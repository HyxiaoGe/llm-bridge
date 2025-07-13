# LLM网关模型配置指南

## 概述

LLM网关的模型配置现已集中管理，支持动态配置各提供商的模型列表。本文档说明如何添加、修改或删除支持的模型。

## 模型配置文件位置

所有模型配置集中在: `internal/providers/models.go`

## 当前支持的模型

### OpenAI
- gpt-3.5-turbo (默认)
- gpt-4o-2024-08-06
- gpt-4.1-2025-04-14

### Gemini
- gemini-2.5-pro
- gemini-2.5-flash (默认)
- gemini-2.0-flash
- gemini-1.5-flash
- gemini-1.5-pro

### DeepSeek
- deepseek-reasoner
- deepseek-chat (默认)

### 通义千问 (Qwen)
- qwen-max
- qwen-plus (默认)
- qwq-plus

### 月之暗面 (Moonshot)
- moonshot-v1-8k (默认)
- moonshot-v1-32k
- moonshot-v1-128k
- kimi-k2-0711-preview

## 如何修改模型配置

### 1. 添加新模型到现有提供商

编辑 `internal/providers/models.go` 文件，在对应提供商的 `Models` 数组中添加新模型：

```go
"openai": {
    Models: []string{
        "gpt-3.5-turbo",
        "gpt-4o-2024-08-06",
        "gpt-4.1-2025-04-14",
        "gpt-4-turbo", // 新增模型
    },
    DefaultModel: "gpt-3.5-turbo",
},
```

### 2. 修改默认模型

更改 `DefaultModel` 字段的值：

```go
"gemini": {
    Models: []string{
        "gemini-2.5-pro",
        "gemini-2.5-flash",
        // ...
    },
    DefaultModel: "gemini-2.5-pro", // 修改默认模型
},
```

### 3. 添加新的提供商

在 `SupportedModels` map 中添加新的提供商配置：

```go
"claude": {
    Models: []string{
        "claude-3-opus",
        "claude-3-sonnet",
        "claude-3-haiku",
    },
    DefaultModel: "claude-3-sonnet",
},
```

## API接口

### 获取所有模型配置

```bash
GET /admin/api/models-config
```

返回示例：
```json
{
    "success": true,
    "modelsConfig": {
        "openai": {
            "models": ["gpt-3.5-turbo", "gpt-4o-2024-08-06"],
            "defaultModel": "gpt-3.5-turbo"
        },
        // ...
    }
}
```

### 获取特定提供商的模型

```bash
GET /admin/api/providers/:provider/models
```

返回示例：
```json
{
    "success": true,
    "provider": "openai",
    "models": ["gpt-3.5-turbo", "gpt-4o-2024-08-06"],
    "defaultModel": "gpt-3.5-turbo"
}
```

## 前端集成

监控面板会自动从API获取最新的模型配置，无需手动更新前端代码。如果API调用失败，前端会降级使用内置的默认配置。

## 注意事项

1. **模型可用性**: 添加模型到配置中并不保证该模型实际可用，需要确保：
   - 提供商确实支持该模型
   - API密钥有权访问该模型
   - 模型名称拼写正确

2. **向后兼容**: 修改模型配置时，请确保不要删除正在使用的模型，以免影响现有的API调用。

3. **测试**: 添加新模型后，请通过监控面板的测试功能验证模型是否可以正常调用。

## 最佳实践

1. 定期检查各提供商的官方文档，及时更新支持的模型列表
2. 为每个提供商选择一个稳定、性价比高的模型作为默认模型
3. 在生产环境更新模型配置前，先在测试环境验证
4. 记录模型变更日志，方便追踪历史记录