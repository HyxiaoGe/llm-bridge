# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview
This is a production-grade LLM Gateway service (LLM网关服务) built in Go that provides unified API access to multiple LLM providers including OpenAI, Claude, Gemini, and Azure OpenAI. The project implements enterprise features including AES-256-GCM encryption, multi-layer rate limiting, load balancing, and comprehensive monitoring.

## Language and Communication Requirements
- **All code comments, documentation, and Git commit messages must be in Chinese**
- **All error messages and debug output should be in Chinese**
- Git commits must follow Chinese format: `feat: 添加RAG检索增强功能` with Co-authored-by info
- Code documentation and docstrings must be in Chinese

## Project Structure
```
llm-bridge/
├── cmd/
│   └── server/
│       └── main.go              # 服务入口
├── internal/
│   ├── handlers/                # HTTP处理器
│   ├── providers/               # LLM提供商适配器
│   ├── security/                # 安全模块(AES-256-GCM)
│   ├── ratelimit/              # 限流模块
│   └── metrics/                # 监控指标
├── pkg/
│   └── types/                  # 公共类型定义
├── configs/
│   ├── config.yaml
│   └── prometheus.yml
├── docker-compose.yml
├── Dockerfile
└── go.mod
```

## Technology Stack
- **Framework**: Fiber (high-performance web framework, targets 5000+ RPS)
- **Cache/Rate Limiting**: Redis with Lua scripts for atomic operations
- **Security**: AES-256-GCM encryption with anti-replay protection
- **Monitoring**: Prometheus metrics collection
- **Deployment**: Docker and Docker Compose

## Development Commands
```bash
# Initialize Go module (if not exists)
go mod init github.com/username/llm-bridge

# Download dependencies
go mod download

# Build the application
go build -o gateway ./cmd/server

# Run the server locally
go run cmd/server/main.go

# Run tests with Chinese test names
go test ./... -v

# Build optimized binary for production
CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-w -s' -o gateway ./cmd/server

# Run with Docker
docker-compose up -d

# View logs
docker-compose logs -f gateway
```

## Code Style Requirements
- Use 4 spaces for indentation (no tabs)
- Functions should not exceed 50 lines, single responsibility
- Use English for variable names but Chinese for comments
- Prefer f-string formatting style where applicable
- Use pathlib.Path equivalent (filepath package) for file operations
- Import order: standard library → third-party → local modules

## Core Architecture Components
- **UnifiedRequest/Response**: Standardized API format across all LLM providers
- **ProviderAdapter Interface**: Transforms requests between unified format and provider-specific APIs
- **Multi-layer Rate Limiting**: Global, user, and IP-based limits using Redis Lua scripts
- **Security Module**: Request encryption, nonce validation, and anti-replay protection
- **Metrics Collection**: Comprehensive monitoring with Prometheus integration

## Testing Requirements
- Test functions must use Chinese descriptions: `test_用户登录_成功场景()`
- Important functions require corresponding test cases
- Use Chinese test data when possible
- Performance benchmarks for critical algorithms

## Configuration Management
- Use environment variables via `.env` files
- YAML configuration files for service settings
- Never hardcode sensitive information (API keys, secrets)
- Support for multiple environments (dev, test, prod)

## Key Dependencies (Planned)
```go
"github.com/gofiber/fiber/v2"
"github.com/go-redis/redis/v8"
"github.com/prometheus/client_golang"
"gopkg.in/yaml.v3"
```

## Security Considerations
- All external input must be validated with Chinese error messages
- Use bcrypt for password hashing
- Implement request/response encryption
- Add rate limiting to all important endpoints
- Log security events in Chinese for debugging

## Current Project State
This project is in the initial planning/setup phase. The comprehensive technical implementation guide (`LLM网关服务的全面技术实施指南.md`) contains detailed specifications but no Go code has been implemented yet. The next steps involve setting up the Go module structure and implementing core functionality according to the technical specifications.