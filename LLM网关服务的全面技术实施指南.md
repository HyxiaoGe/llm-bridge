# LLM网关服务技术实现指南

## 特殊要求

### 语言要求

- **所有回复必须使用中文**：包括代码注释、解释说明、错误信息等
- **Git 提交信息必须使用中文**：提交标题和描述都使用中文
- **文档和注释使用中文**：所有新创建的文档、代码注释都使用中文

### Git 提交规范

- **提交信息格式要求**：

  ```
  feat: 添加RAG检索增强功能
  
  - 实现向量数据库集成
  - 优化文档分块策略
  - 添加混合搜索支持
  
  🤖 Generated with [Claude Code](https://claude.ai/code)
  
  Co-Authored-By: Claude <noreply@anthropic.com>
  ```

- **必须包含 Co-author 信息**：每个提交都要包含 `Co-authored-by: Claude Code <claude-code@anthropic.com>`

- **使用中文提交类型**：

  - `feat`: 新功能
  - `fix`: 修复bug
  - `docs`: 文档更新
  - `style`: 代码格式调整
  - `refactor`: 重构代码
  - `test`: 测试相关
  - `chore`: 构建工具或辅助工具的变动

### 命令执行权限

- **常规 Linux 命令可直接执行**：
  - 文件操作：`ls`, `cd`, `cp`, `mv`, `mkdir`, `touch`, `cat`, `less`, `grep`, `find`
  - 文本处理：`sed`, `awk`, `sort`, `uniq`, `head`, `tail`, `wc`
  - 系统信息：`ps`, `top`, `df`, `du`, `free`, `whoami`, `pwd`
  - 网络工具：`ping`, `curl`, `wget`
  - 开发工具：`git status`, `git log`, `git diff`, `npm list`, `pip list`

- **需要确认的重要命令**：
  - 删除操作：`rm -rf`, `rmdir`
  - 系统级操作：`sudo`, `su`, `chmod 777`, `chown`
  - 网络配置：`iptables`, `netstat`, `ss`
  - 进程管理：`kill`, `killall`, `pkill`
  - 包管理：`apt install`, `yum install`, `npm install -g`
  - 数据库操作：`mysql`, `psql`, `mongo`
  - 服务管理：`systemctl`, `service`

### 个人开发偏好

- **代码风格**：使用4个空格缩进，不使用Tab
- **函数命名**：使用动词开头，如 `获取用户信息()`, `处理文档()`
- **错误处理**：优先使用 try-except，提供中文错误信息
- **日志格式**：使用中文日志信息，便于调试
- **注释语言**：所有代码注释使用中文
- **变量命名**：使用英文，但注释说明使用中文
- **函数设计**：单个函数不超过50行，职责单一
- **导入顺序**：标准库 → 第三方库 → 本地模块，每组之间空一行
- **字符串处理**：优先使用 f-string 格式化，避免使用 % 格式化
- **文件路径**：使用 `pathlib.Path` 而不是 `os.path`
- **配置管理**：使用 `.env` 文件管理环境变量，敏感信息不写入代码
- **依赖管理**：使用 `requirements.txt` 锁定版本，重要依赖添加中文注释说明用途

### 文档和注释偏好

- **函数文档**：所有函数必须有中文docstring，说明参数、返回值、异常
- **类文档**：类的作用、主要方法、使用示例都用中文描述
- **复杂逻辑**：超过5行的复杂逻辑必须添加中文注释解释
- **TODO标记**：使用中文 `# TODO: 待实现功能描述` 格式
- **代码示例**：在文档中提供中文注释的完整代码示例

### 测试和质量保证

- **测试覆盖**：重要函数必须有对应的测试用例
- **测试命名**：测试函数使用中文描述，如 `test_用户登录_成功场景()`
- **断言信息**：断言失败时提供中文错误信息
- **测试数据**：使用中文测试数据，更贴近实际使用场景
- **性能测试**：关键算法需要添加性能测试和基准测试

### 调试和日志偏好

- **调试信息**：使用中文debug信息，便于定位问题
- **日志级别**：开发环境使用DEBUG，生产环境使用INFO
- **异常捕获**：捕获异常时记录中文上下文信息
- **打印调试**：临时调试可以使用print，但正式代码必须使用logging
- **错误追踪**：重要错误必须记录完整的中文错误堆栈

### 安全和性能偏好

- **输入验证**：所有外部输入必须验证，提供中文错误提示
- **密码处理**：使用bcrypt等安全算法，不明文存储
- **API限流**：重要接口添加速率限制
- **缓存策略**：合理使用缓存，避免重复计算
- **资源清理**：及时关闭文件、数据库连接等资源

### 项目结构偏好

- **目录命名**：使用中文拼音或英文，避免中文目录名
- **文件分类**：工具函数放在 `utils/`，配置文件放在 `config/`
- **模块划分**：按功能模块划分，每个模块职责清晰
- **常量定义**：所有魔法数字和字符串定义为有意义的常量
- **环境隔离**：开发、测试、生产环境严格隔离

## 基于Go语言构建生产级LLM网关，支持多平台路由、安全加密、智能限流和低成本部署

本指南详细介绍如何使用Go语言构建一个生产级LLM网关服务，支持OpenAI、Claude、Gemini、Azure OpenAI等多个平台的统一接入。包含完整的安全架构、限流策略和部署方案。

## 1. 核心架构设计

### 统一API设计

LLM网关的核心是创建统一的请求响应格式，屏蔽各平台差异：

```go
// 统一请求结构
type UnifiedRequest struct {
    Model    string    `json:"model"`
    Messages []Message `json:"messages"`
    Parameters Parameters `json:"parameters"`
    Provider string    `json:"provider"`
    Metadata Metadata  `json:"metadata"`
}

type Message struct {
    Role    string `json:"role"`
    Content string `json:"content"`
}

type Parameters struct {
    Temperature      float64 `json:"temperature"`
    MaxTokens       int     `json:"max_tokens"`
    TopP            float64 `json:"top_p"`
    Stream          bool    `json:"stream"`
    FrequencyPenalty float64 `json:"frequency_penalty"`
    PresencePenalty  float64 `json:"presence_penalty"`
}
```

### 平台适配器实现

每个LLM平台都需要特定的请求格式转换：

```go
package providers

import (
    "bytes"
    "encoding/json"
    "net/http"
)

type ProviderAdapter interface {
    Transform(req *UnifiedRequest) ([]byte, error)
    CallAPI(data []byte) (*http.Response, error)
    ParseResponse(resp *http.Response) (*UnifiedResponse, error)
}

// OpenAI适配器
type OpenAIProvider struct {
    APIKey  string
    BaseURL string
}

func (p *OpenAIProvider) Transform(req *UnifiedRequest) ([]byte, error) {
    openaiReq := map[string]interface{}{
        "model":             req.Model,
        "messages":          req.Messages,
        "temperature":       req.Parameters.Temperature,
        "max_tokens":        req.Parameters.MaxTokens,
        "top_p":            req.Parameters.TopP,
        "stream":           req.Parameters.Stream,
        "frequency_penalty": req.Parameters.FrequencyPenalty,
        "presence_penalty":  req.Parameters.PresencePenalty,
    }
    
    return json.Marshal(openaiReq)
}

// Claude适配器
type ClaudeProvider struct {
    APIKey  string
    BaseURL string
}

func (p *ClaudeProvider) Transform(req *UnifiedRequest) ([]byte, error) {
    // 提取系统消息
    var systemMsg string
    var userMessages []Message
    
    for _, msg := range req.Messages {
        if msg.Role == "system" {
            systemMsg = msg.Content
        } else {
            userMessages = append(userMessages, msg)
        }
    }
    
    claudeReq := map[string]interface{}{
        "model":      req.Model,
        "max_tokens": req.Parameters.MaxTokens, // Claude必需参数
        "temperature": req.Parameters.Temperature,
        "top_p":      req.Parameters.TopP,
        "stream":     req.Parameters.Stream,
        "system":     systemMsg,
        "messages":   userMessages,
    }
    
    return json.Marshal(claudeReq)
}
```

### 流式响应处理

实现SSE流式响应，提供实时的token输出：

```go
package handlers

import (
    "bufio"
    "encoding/json"
    "fmt"
    "net/http"
    "strings"
)

type StreamHandler struct {
    providerManager *ProviderManager
}

func (h *StreamHandler) HandleStream(w http.ResponseWriter, r *http.Request) {
    // 设置SSE头部
    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")
    w.Header().Set("Access-Control-Allow-Origin", "*")
    
    var req UnifiedRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    
    provider := h.providerManager.GetProvider(req.Provider)
    resp, err := provider.StreamRequest(&req)
    if err != nil {
        fmt.Fprintf(w, "event: error\ndata: %s\n\n", err.Error())
        return
    }
    defer resp.Body.Close()
    
    scanner := bufio.NewScanner(resp.Body)
    for scanner.Scan() {
        line := scanner.Text()
        if strings.HasPrefix(line, "data: ") {
            data := strings.TrimPrefix(line, "data: ")
            if data == "[DONE]" {
                fmt.Fprint(w, "event: done\ndata: {\"finish_reason\": \"stop\"}\n\n")
                break
            }
            
            // 转换为统一格式
            unified := h.transformStreamChunk(data, req.Provider)
            fmt.Fprintf(w, "event: data\ndata: %s\n\n", unified)
            
            if f, ok := w.(http.Flusher); ok {
                f.Flush()
            }
        }
    }
}
```

## 2. 安全与加密实现

### AES-256-GCM密钥加密

使用Go标准库实现安全的API密钥加密传输：

```go
package security

import (
    "crypto/aes"
    "crypto/cipher"
    "crypto/rand"
    "encoding/base64"
    "errors"
)

type KeyManager struct {
    masterKey []byte
}

func NewKeyManager(masterKey string) *KeyManager {
    // 确保密钥长度为32字节（AES-256）
    key := make([]byte, 32)
    copy(key, []byte(masterKey))
    return &KeyManager{masterKey: key}
}

func (km *KeyManager) EncryptAPIKey(apiKey string) (string, error) {
    block, err := aes.NewCipher(km.masterKey)
    if err != nil {
        return "", err
    }
    
    aesGCM, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }
    
    // 生成随机nonce
    nonce := make([]byte, aesGCM.NonceSize())
    if _, err := rand.Read(nonce); err != nil {
        return "", err
    }
    
    // 加密
    ciphertext := aesGCM.Seal(nonce, nonce, []byte(apiKey), nil)
    return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (km *KeyManager) DecryptAPIKey(encryptedKey string) (string, error) {
    data, err := base64.StdEncoding.DecodeString(encryptedKey)
    if err != nil {
        return "", err
    }
    
    block, err := aes.NewCipher(km.masterKey)
    if err != nil {
        return "", err
    }
    
    aesGCM, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }
    
    nonceSize := aesGCM.NonceSize()
    if len(data) < nonceSize {
        return "", errors.New("密文长度无效")
    }
    
    nonce, ciphertext := data[:nonceSize], data[nonceSize:]
    plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
    if err != nil {
        return "", err
    }
    
    return string(plaintext), nil
}
```

### 防重放攻击机制

基于Redis实现nonce防重放验证：

```go
package security

import (
    "crypto/rand"
    "encoding/hex"
    "fmt"
    "time"
    
    "github.com/go-redis/redis/v8"
    "golang.org/x/net/context"
)

type NonceValidator struct {
    redis      *redis.Client
    expiration time.Duration
}

func NewNonceValidator(redisClient *redis.Client) *NonceValidator {
    return &NonceValidator{
        redis:      redisClient,
        expiration: 5 * time.Minute, // 5分钟过期
    }
}

func (nv *NonceValidator) GenerateNonce() (string, error) {
    bytes := make([]byte, 16)
    if _, err := rand.Read(bytes); err != nil {
        return "", err
    }
    
    timestamp := time.Now().Unix()
    nonce := fmt.Sprintf("%d-%s", timestamp, hex.EncodeToString(bytes))
    return nonce, nil
}

func (nv *NonceValidator) ValidateNonce(nonce string) error {
    ctx := context.Background()
    key := fmt.Sprintf("nonce:%s", nonce)
    
    // 检查nonce是否已存在
    exists, err := nv.redis.Exists(ctx, key).Result()
    if err != nil {
        return err
    }
    
    if exists > 0 {
        return errors.New("重放攻击：nonce已被使用")
    }
    
    // 存储nonce并设置过期时间
    err = nv.redis.Set(ctx, key, "used", nv.expiration).Err()
    if err != nil {
        return err
    }
    
    return nil
}
```

### 内存安全的凭证处理

实现安全的内存清理机制：

```go
package security

import (
    "crypto/rand"
    "runtime"
    "unsafe"
)

type SecureString struct {
    data []byte
    size int
}

func NewSecureString(value string) *SecureString {
    data := make([]byte, len(value))
    copy(data, []byte(value))
    
    return &SecureString{
        data: data,
        size: len(data),
    }
}

func (s *SecureString) Clear() {
    if s.data == nil {
        return
    }
    
    // 首先用随机数据覆盖
    rand.Read(s.data)
    
    // 然后用零值覆盖
    for i := range s.data {
        s.data[i] = 0
    }
    
    // 强制垃圾回收
    runtime.GC()
    
    s.data = nil
    s.size = 0
}

func (s *SecureString) String() string {
    if s.data == nil {
        return ""
    }
    return *(*string)(unsafe.Pointer(&s.data))
}

// 使用示例
func processAPIKey(encryptedKey string, keyManager *KeyManager) error {
    apiKey, err := keyManager.DecryptAPIKey(encryptedKey)
    if err != nil {
        return err
    }
    
    secureKey := NewSecureString(apiKey)
    defer secureKey.Clear() // 确保清理
    
    // 使用API密钥
    return callProviderAPI(secureKey.String())
}
```

## 3. 多层限流策略

### Redis分布式限流

使用Lua脚本实现原子性的多层限流：

```go
package ratelimit

import (
    "context"
    "time"
    
    "github.com/go-redis/redis/v8"
)

type RateLimiter struct {
    redis *redis.Client
    luaScript string
}

func NewRateLimiter(redisClient *redis.Client) *RateLimiter {
    script := `
    local global_key = KEYS[1]
    local user_key = KEYS[2]
    local ip_key = KEYS[3]
    
    local global_limit = tonumber(ARGV[1])
    local user_limit = tonumber(ARGV[2])
    local ip_limit = tonumber(ARGV[3])
    local window = tonumber(ARGV[4])
    
    -- 检查全局限流
    local global_count = tonumber(redis.call('GET', global_key) or 0)
    if global_count >= global_limit then
        return {0, "global_limit_exceeded", global_limit - global_count}
    end
    
    -- 检查用户限流
    local user_count = tonumber(redis.call('GET', user_key) or 0)
    if user_count >= user_limit then
        return {0, "user_limit_exceeded", user_limit - user_count}
    end
    
    -- 检查IP限流
    local ip_count = tonumber(redis.call('GET', ip_key) or 0)
    if ip_count >= ip_limit then
        return {0, "ip_limit_exceeded", ip_limit - ip_count}
    end
    
    -- 增加计数器
    redis.call('INCR', global_key)
    redis.call('EXPIRE', global_key, window)
    redis.call('INCR', user_key)
    redis.call('EXPIRE', user_key, window)
    redis.call('INCR', ip_key)
    redis.call('EXPIRE', ip_key, window)
    
    return {1, "allowed", math.min(
        global_limit - global_count - 1,
        user_limit - user_count - 1,
        ip_limit - ip_count - 1
    )}
    `
    
    return &RateLimiter{
        redis: redisClient,
        luaScript: script,
    }
}

type LimitResult struct {
    Allowed   bool   `json:"allowed"`
    Reason    string `json:"reason"`
    Remaining int    `json:"remaining"`
}

func (rl *RateLimiter) CheckLimit(ctx context.Context, userID, ip string) (*LimitResult, error) {
    now := time.Now()
    window := 3600 // 1小时窗口
    
    keys := []string{
        fmt.Sprintf("global:%d", now.Hour()),
        fmt.Sprintf("user:%s:%d", userID, now.Hour()),
        fmt.Sprintf("ip:%s:%d", ip, now.Hour()),
    }
    
    args := []interface{}{
        1000, // 全局限制：每小时1000次
        100,  // 用户限制：每小时100次
        50,   // IP限制：每小时50次
        window,
    }
    
    result, err := rl.redis.Eval(ctx, rl.luaScript, keys, args...).Result()
    if err != nil {
        return nil, err
    }
    
    res := result.([]interface{})
    allowed := res[0].(int64) == 1
    reason := res[1].(string)
    remaining := int(res[2].(int64))
    
    return &LimitResult{
        Allowed:   allowed,
        Reason:    reason,
        Remaining: remaining,
    }, nil
}
```

### 动态限流调整

根据上游API响应情况动态调整限流策略：

```go
package ratelimit

import (
    "sync"
    "time"
)

type AdaptiveRateLimiter struct {
    mu           sync.RWMutex
    baseLimit    int
    currentLimit int
    metrics      *MetricsCollector
}

type MetricsCollector struct {
    totalRequests int64
    errorCount    int64
    totalLatency  time.Duration
    window        time.Duration
}

func NewAdaptiveRateLimiter(baseLimit int) *AdaptiveRateLimiter {
    return &AdaptiveRateLimiter{
        baseLimit:    baseLimit,
        currentLimit: baseLimit,
        metrics:      &MetricsCollector{window: time.Minute},
    }
}

func (arl *AdaptiveRateLimiter) AdjustLimits() {
    arl.mu.Lock()
    defer arl.mu.Unlock()
    
    if arl.metrics.totalRequests == 0 {
        return
    }
    
    errorRate := float64(arl.metrics.errorCount) / float64(arl.metrics.totalRequests)
    avgLatency := arl.metrics.totalLatency / time.Duration(arl.metrics.totalRequests)
    
    adjustmentFactor := 1.0
    
    // 根据错误率调整
    switch {
    case errorRate > 0.1: // 错误率超过10%
        adjustmentFactor *= 0.5
    case errorRate > 0.05: // 错误率超过5%
        adjustmentFactor *= 0.7
    case errorRate < 0.01: // 错误率低于1%
        adjustmentFactor *= 1.2
    }
    
    // 根据延迟调整
    switch {
    case avgLatency > 5*time.Second:
        adjustmentFactor *= 0.6
    case avgLatency > 2*time.Second:
        adjustmentFactor *= 0.8
    case avgLatency < 500*time.Millisecond:
        adjustmentFactor *= 1.1
    }
    
    // 应用调整，设置边界
    newLimit := int(float64(arl.baseLimit) * adjustmentFactor)
    if newLimit < 10 {
        newLimit = 10
    } else if newLimit > arl.baseLimit*2 {
        newLimit = arl.baseLimit * 2
    }
    
    arl.currentLimit = newLimit
    
    // 重置指标
    arl.metrics.totalRequests = 0
    arl.metrics.errorCount = 0
    arl.metrics.totalLatency = 0
}

func (arl *AdaptiveRateLimiter) RecordRequest(latency time.Duration, isError bool) {
    arl.mu.Lock()
    defer arl.mu.Unlock()
    
    arl.metrics.totalRequests++
    arl.metrics.totalLatency += latency
    if isError {
        arl.metrics.errorCount++
    }
}
```

## 4. 低成本部署方案

### 推荐部署架构

**开发测试阶段（免费）**
- 平台：Oracle Cloud永久免费层
- 配置：ARM实例 4核24GB（可分割成多个实例）
- 成本：$0/月

**小规模生产（推荐）**
- 平台：Hetzner VPS
- 配置：CX31 (2vCPU, 8GB RAM, 20TB流量)
- 成本：€16.64/月 (~$18)

**大规模生产**
- 架构：Hetzner + Cloudflare Workers
- 成本：约$30-50/月（包含CDN和边缘缓存）

### Docker部署配置

```dockerfile
# 多阶段构建，最小化镜像大小
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-w -s' -o gateway ./cmd/server

# 运行时镜像
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /root/

COPY --from=builder /app/gateway .
COPY --from=builder /app/configs ./configs

EXPOSE 8080
CMD ["./gateway"]
```

### Docker Compose配置

```yaml
version: '3.8'

services:
  gateway:
    build: .
    ports:
      - "8080:8080"
    environment:
      - REDIS_URL=redis://redis:6379
      - MASTER_KEY=${MASTER_KEY}
      - LOG_LEVEL=info
    depends_on:
      - redis
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes
    volumes:
      - redis_data:/data
    ports:
      - "6379:6379"
    restart: unless-stopped

  prometheus:
    image: prom/prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./configs/prometheus.yml:/etc/prometheus/prometheus.yml
    restart: unless-stopped

volumes:
  redis_data:
```

## 5. Go框架选择与性能优化

### 框架对比

**Fiber（推荐）**
- 性能：单核5000+ RPS
- 特点：类Express API，零内存分配
- 适用：高性能网关服务

**Gin**
- 性能：单核3000+ RPS  
- 特点：生态丰富，中间件多
- 适用：快速开发

### 高性能服务器实现

```go
package main

import (
    "log"
    "runtime"
    
    "github.com/gofiber/fiber/v2"
    "github.com/gofiber/fiber/v2/middleware/cors"
    "github.com/gofiber/fiber/v2/middleware/logger"
    "github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
    // 设置使用所有CPU核心
    runtime.GOMAXPROCS(runtime.NumCPU())
    
    app := fiber.New(fiber.Config{
        Prefork:          true,  // 多进程模式
        CaseSensitive:    true,
        StrictRouting:    true,
        ServerHeader:     "ChatNexus/1.0",
        AppName:          "ChatNexus Gateway",
        BodyLimit:        10 * 1024 * 1024, // 10MB
        ReadTimeout:      time.Second * 30,
        WriteTimeout:     time.Second * 30,
        IdleTimeout:      time.Second * 120,
    })
    
    // 中间件
    app.Use(recover.New())
    app.Use(logger.New(logger.Config{
        Format: "${time} ${status} ${latency} ${method} ${path}\n",
    }))
    app.Use(cors.New())
    
    // 路由
    api := app.Group("/v1")
    api.Post("/chat/completions", handleChatCompletion)
    api.Get("/models", handleListModels)
    api.Get("/health", handleHealth)
    
    log.Fatal(app.Listen(":8080"))
}

func handleChatCompletion(c *fiber.Ctx) error {
    // 解析请求
    var req UnifiedRequest
    if err := c.BodyParser(&req); err != nil {
        return c.Status(400).JSON(fiber.Map{
            "error": "请求格式错误: " + err.Error(),
        })
    }
    
    // 验证和限流
    if err := validateAndRateLimit(c, &req); err != nil {
        return c.Status(429).JSON(fiber.Map{
            "error": err.Error(),
        })
    }
    
    // 路由到对应提供商
    provider := getProvider(req.Provider)
    response, err := provider.Process(&req)
    if err != nil {
        return c.Status(500).JSON(fiber.Map{
            "error": "处理请求失败: " + err.Error(),
        })
    }
    
    return c.JSON(response)
}
```

## 6. 生产环境考虑

### 负载均衡与故障转移

```go
package loadbalancer

import (
    "errors"
    "sort"
    "sync"
    "time"
)

type LoadBalancer struct {
    providers    []Provider
    healthStatus map[string]*HealthInfo
    mu           sync.RWMutex
}

type HealthInfo struct {
    IsHealthy     bool
    LastCheck     time.Time
    AverageLatency time.Duration
    ErrorRate     float64
}

func (lb *LoadBalancer) SelectProvider(request *UnifiedRequest) (Provider, error) {
    lb.mu.RLock()
    defer lb.mu.RUnlock()
    
    var healthyProviders []Provider
    for _, provider := range lb.providers {
        if health, exists := lb.healthStatus[provider.Name()]; exists && health.IsHealthy {
            healthyProviders = append(healthyProviders, provider)
        }
    }
    
    if len(healthyProviders) == 0 {
        return nil, errors.New("没有可用的健康提供商")
    }
    
    // 按延迟排序，选择最快的
    sort.Slice(healthyProviders, func(i, j int) bool {
        latencyI := lb.healthStatus[healthyProviders[i].Name()].AverageLatency
        latencyJ := lb.healthStatus[healthyProviders[j].Name()].AverageLatency
        return latencyI < latencyJ
    })
    
    return healthyProviders[0], nil
}

func (lb *LoadBalancer) StartHealthCheck() {
    ticker := time.NewTicker(30 * time.Second)
    go func() {
        for range ticker.C {
            lb.performHealthChecks()
        }
    }()
}

func (lb *LoadBalancer) performHealthChecks() {
    for _, provider := range lb.providers {
        go func(p Provider) {
            start := time.Now()
            err := p.HealthCheck()
            latency := time.Since(start)
            
            lb.mu.Lock()
            if lb.healthStatus[p.Name()] == nil {
                lb.healthStatus[p.Name()] = &HealthInfo{}
            }
            
            health := lb.healthStatus[p.Name()]
            health.IsHealthy = err == nil
            health.LastCheck = time.Now()
            health.AverageLatency = (health.AverageLatency + latency) / 2
            lb.mu.Unlock()
        }(provider)
    }
}
```

### 监控指标收集

```go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    RequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "llm_requests_total",
            Help: "LLM请求总数",
        },
        []string{"provider", "model", "status"},
    )
    
    RequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "llm_request_duration_seconds",
            Help: "LLM请求耗时（秒）",
            Buckets: []float64{0.1, 0.5, 1, 2.5, 5, 10, 30},
        },
        []string{"provider", "model"},
    )
    
    TokensProcessed = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "llm_tokens_processed_total", 
            Help: "处理的token总数",
        },
        []string{"provider", "model", "type"},
    )
    
    CostTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "llm_cost_usd_total",
            Help: "LLM调用总成本（美元）",
        },
        []string{"provider", "model"},
    )
)

// 记录请求指标
func RecordRequest(provider, model, status string, duration time.Duration, inputTokens, outputTokens int, cost float64) {
    RequestsTotal.WithLabelValues(provider, model, status).Inc()
    RequestDuration.WithLabelValues(provider, model).Observe(duration.Seconds())
    TokensProcessed.WithLabelValues(provider, model, "input").Add(float64(inputTokens))
    TokensProcessed.WithLabelValues(provider, model, "output").Add(float64(outputTokens))
    CostTotal.WithLabelValues(provider, model).Add(cost)
}
```

### 配置管理

```go
package config

import (
    "os"
    "strconv"
    "time"
)

type Config struct {
    Server   ServerConfig   `yaml:"server"`
    Redis    RedisConfig    `yaml:"redis"`
    Security SecurityConfig `yaml:"security"`
    Providers []ProviderConfig `yaml:"providers"`
}

type ServerConfig struct {
    Port         int           `yaml:"port"`
    ReadTimeout  time.Duration `yaml:"read_timeout"`
    WriteTimeout time.Duration `yaml:"write_timeout"`
    IdleTimeout  time.Duration `yaml:"idle_timeout"`
}

func LoadConfig() *Config {
    return &Config{
        Server: ServerConfig{
            Port:         getEnvInt("PORT", 8080),
            ReadTimeout:  time.Second * 30,
            WriteTimeout: time.Second * 30,
            IdleTimeout:  time.Second * 120,
        },
        Redis: RedisConfig{
            URL: getEnv("REDIS_URL", "redis://localhost:6379"),
        },
        Security: SecurityConfig{
            MasterKey: getEnv("MASTER_KEY", "your-secret-key"),
        },
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intValue, err := strconv.Atoi(value); err == nil {
            return intValue
        }
    }
    return defaultValue
}
```

## 项目结构建议

```
ChatNexus/
├── cmd/
│   └── server/
│       └── main.go              # 服务入口
├── internal/
│   ├── handlers/                # HTTP处理器
│   │   ├── chat.go
│   │   ├── models.go
│   │   └── health.go
│   ├── providers/               # LLM提供商
│   │   ├── openai.go
│   │   ├── claude.go
│   │   ├── gemini.go
│   │   └── base.go
│   ├── security/                # 安全模块
│   │   ├── encryption.go
│   │   └── nonce.go
│   ├── ratelimit/              # 限流模块
│   │   └── limiter.go
│   └── metrics/                # 监控指标
│       └── collector.go
├── pkg/
│   └── types/                  # 公共类型定义
│       └── requests.go
├── configs/
│   ├── config.yaml
│   └── prometheus.yml
├── docker-compose.yml
├── Dockerfile
├── go.mod
└── README.md
```

## 总结

本指南提供了构建生产级LLM网关的完整Go语言实现方案，主要特点：

**技术栈选择**
- Go + Fiber框架（5000+ RPS性能）
- Redis分布式缓存和限流
- AES-256-GCM加密传输
- Prometheus监控

**部署建议**
- 开发：Oracle Cloud免费层
- 生产：Hetzner VPS（最佳性价比）
- 扩展：Cloudflare Workers边缘缓存

**核心特性**
- 统一多平台API接口
- 多层自适应限流
- 端到端安全加密
- 实时监控告警

这个架构可以支持千万级请求，同时保持较低的运营成本和高可用性。