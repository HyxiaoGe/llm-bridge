# LLM网关服务 (LLM Bridge Gateway)

一个生产级的LLM API网关服务，提供多个LLM提供商的统一接入、智能负载均衡、安全加密和速率限制功能。

## 功能特性

- **多提供商支持**: 统一接入OpenAI、Claude、Gemini、Azure OpenAI等多个LLM平台
- **统一API格式**: 屏蔽各平台差异，提供标准化的请求响应接口
- **智能负载均衡**: 支持轮询、权重、最少连接等多种负载均衡策略
- **安全加密**: AES-256-GCM加密传输，防重放攻击保护
- **多层限流**: 全局、用户、IP三级限流保护
- **实时监控**: Prometheus指标收集，完整的健康检查机制
- **高性能**: 基于Fiber框架，支持5000+ RPS

## 快速开始

### 环境要求

- Go 1.21+
- Redis (用于限流和缓存)
- Docker & Docker Compose (可选)

### 本地开发

1. **克隆项目**
```bash
git clone https://github.com/heyanxiao/llm-bridge.git
cd llm-bridge
```

2. **配置环境变量**
```bash
cp .env.example .env
# 编辑 .env 文件，填入各LLM提供商的API密钥
```

3. **下载依赖**
```bash
make deps
```

4. **运行服务**
```bash
make run
```

服务将在 `http://localhost:8080` 启动

### Docker部署

1. **使用Docker Compose**
```bash
# 配置环境变量
cp .env.example .env

# 启动服务
make docker-run
```

2. **或者手动构建**
```bash
make docker-build
docker run -p 8080:8080 --env-file .env llm-gateway:latest
```

## API使用示例

### 聊天补全

```bash
curl -X POST http://localhost:8080/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-3.5-turbo",
    "provider": "openai",
    "messages": [
      {"role": "user", "content": "你好！"}
    ],
    "parameters": {
      "temperature": 0.7,
      "max_tokens": 100
    }
  }'
```

### 获取可用模型

```bash
curl http://localhost:8080/v1/models
```

### 健康检查

```bash
curl http://localhost:8080/health
```

## 项目结构

```
llm-bridge/
├── cmd/server/          # 应用入口
├── internal/
│   ├── handlers/        # HTTP处理器
│   ├── providers/       # LLM提供商适配器
│   ├── security/        # 安全模块
│   ├── ratelimit/      # 限流模块
│   └── metrics/        # 监控指标
├── pkg/types/          # 公共类型定义
├── configs/            # 配置文件
├── docker-compose.yml  # Docker编排
├── Dockerfile         # Docker镜像
└── Makefile          # 构建脚本
```

## 配置说明

主要配置文件位于 `configs/config.yaml`，支持通过环境变量覆盖配置项。

关键配置：
- **providers**: 各LLM提供商的API配置
- **rate_limit**: 限流策略配置
- **security**: 安全和加密配置
- **monitoring**: 监控和日志配置

## 开发指南

### 常用命令

```bash
make help          # 查看所有可用命令
make build         # 编译应用
make test          # 运行测试
make lint          # 代码检查
make fmt           # 代码格式化
```

### 添加新的LLM提供商

1. 在 `internal/providers/` 目录下创建新的适配器文件
2. 实现 `ProviderAdapter` 接口
3. 在 `cmd/server/main.go` 中注册新提供商
4. 更新配置文件添加相应配置项

## 监控和运维

- **健康检查**: `/health`, `/health/ready`, `/health/live`
- **指标监控**: `/metrics` (Prometheus格式)
- **日志**: 结构化JSON日志输出

## 许可证

MIT License