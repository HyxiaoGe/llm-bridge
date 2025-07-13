# LLM网关服务 (LLM Bridge Gateway)

一个生产级的LLM API网关服务，支持多个LLM提供商统一接入、智能负载均衡、限流保护和实时监控。

## ✨ 功能特性

- 🚀 **统一API接口**: 兼容OpenAI API格式，无需修改现有代码
- 🔄 **多提供商支持**: OpenAI、Gemini、DeepSeek、通义千问、月之暗面
- ⚡ **智能负载均衡**: 轮询调度和故障转移
- 🛡️ **限流保护**: 多层限流机制，防止恶意请求
- 📊 **实时监控**: Web管理面板，统计分析和性能指标
- 🐳 **容器化部署**: Docker + 一键云部署
- 🌍 **全球访问**: 支持全球部署，无地域限制

## 🚀 一键部署到云端

### Render.com (推荐)
[![Deploy to Render](https://render.com/images/deploy-to-render-button.svg)](https://render.com/deploy)

1. 点击按钮连接GitHub
2. 配置API密钥环境变量
3. 一键部署，获得全球访问URL

**详细指南**: [📖 Render部署文档](docs/RENDER_DEPLOYMENT.md)

### 其他平台
- **Railway**: 支持Docker，$5/月
- **Fly.io**: 全球边缘网络部署
- **自建服务器**: VPS + Docker部署

## 🛠️ 本地开发

### 使用Docker Compose (推荐)

```bash
# 1. 克隆项目
git clone https://github.com/heyanxiao/llm-bridge.git
cd llm-bridge

# 2. 配置环境变量
cp .env.example .env
# 编辑.env文件，填入你的API密钥

# 3. 启动服务
docker-compose up -d

# 4. 访问服务
# 监控面板: http://localhost:8080/ (自动跳转到管理面板)
# API端点: http://localhost:8080/v1/chat/completions
```

### 手动编译运行

```bash
# 环境要求: Go 1.21+, Redis
make deps    # 下载依赖
make build   # 编译
make run     # 运行
```

## 📡 API使用

### 聊天完成接口

```bash
curl -X POST https://your-app.onrender.com/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [
      {"role": "user", "content": "Hello!"}
    ]
  }'
```

### 负载均衡使用示例

系统支持四种调用方式，具备智能负载均衡和默认模型选择功能：

```bash
# 情况1: 负载均衡模式 - 不指定provider和model
curl -X POST https://your-app.onrender.com/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "messages": [
      {"role": "user", "content": "Hello!"}
    ]
  }'
# 系统自动轮询: OpenAI(gpt-3.5-turbo) → Gemini(gemini-2.5-flash) → DeepSeek(deepseek-chat) → 通义千问(qwen-plus) → 月之暗面(moonshot-v1-8k)

# 情况2: 指定提供商，使用默认模型
curl -X POST https://your-app.onrender.com/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "provider": "openai",
    "messages": [
      {"role": "user", "content": "使用OpenAI的默认模型(gpt-3.5-turbo)"}
    ]
  }'

# 情况3: 完全指定提供商和模型
curl -X POST https://your-app.onrender.com/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o-2024-08-06",
    "provider": "openai",
    "messages": [
      {"role": "user", "content": "使用指定的GPT-4模型"}
    ]
  }'

# 情况4: 错误示例 - 只指定model不指定provider (会返回错误)
curl -X POST https://your-app.onrender.com/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-3.5-turbo",
    "messages": [
      {"role": "user", "content": "这会返回错误：需要指定provider"}
    ]
  }'
# 返回错误: "指定模型时必须同时指定提供商(provider)参数"
```

**负载均衡特性**:
- 🔄 轮询算法: 自动在健康提供商间轮询
- 🛡️ 故障转移: 自动跳过不健康的提供商
- 🎯 智能选择: 自动使用提供商的默认模型
- 📊 健康监控: 实时检测提供商API状态
- ⚡ 高可用性: 单点故障不影响整体服务
- 🚫 参数校验: 防止无效的模型/提供商组合

### 支持的模型

| 提供商 | 模型列表 |
|--------|---------|
| **OpenAI** | gpt-3.5-turbo, gpt-4o-2024-08-06, gpt-4.1-2025-04-14 |
| **Gemini** | gemini-2.5-pro, gemini-2.5-flash, gemini-2.0-flash |
| **DeepSeek** | deepseek-reasoner, deepseek-chat |
| **通义千问** | qwen-max, qwen-plus, qwq-plus |
| **月之暗面** | moonshot-v1-8k, moonshot-v1-32k, kimi-k2-0711-preview |

### 其他接口

```bash
# 获取可用模型
curl https://your-app.onrender.com/v1/models

# 健康检查
curl https://your-app.onrender.com/health
```

## 📊 监控面板

访问服务根目录自动跳转到管理面板，查看：
- 🔋 **提供商状态**: 实时健康状态和响应时间
- 📈 **请求统计**: 总请求数、成功率、平均响应时间  
- 💰 **成本分析**: Token消耗统计和费用估算
- 🧪 **在线测试**: API接口测试工具
- 🛡️ **限流监控**: 当前限流配置和触发统计

![监控面板截图](docs/monitor-dashboard.png)

## ⚙️ 限流保护

内置多层限流机制防止恶意请求：

- **全局限流**: 60次/分钟, 300次/5分钟, 2000次/小时
- **聊天接口**: 30次/分钟, 150次/5分钟  
- **测试接口**: 20次/分钟
- **基于Redis**: 滑动窗口算法，持久化存储

**配置指南**: [🛡️ 限流功能文档](docs/RATE_LIMIT_GUIDE.md)

## 🔧 项目结构

```
llm-bridge/
├── cmd/server/           # 应用入口
├── internal/
│   ├── handlers/         # HTTP处理器
│   ├── providers/        # LLM提供商适配器
│   ├── middleware/       # 中间件(限流等)
│   └── stats/           # Redis统计服务
├── static/              # 监控面板前端
├── docs/                # 项目文档
├── docker-compose.yml   # Docker编排
├── render.yaml         # Render部署配置
└── Dockerfile          # Docker镜像
```

## 📚 文档

- [🔑 API密钥配置指南](API_KEYS_GUIDE.md)
- [🚀 Render部署指南](docs/RENDER_DEPLOYMENT.md)
- [🛡️ 限流功能文档](docs/RATE_LIMIT_GUIDE.md)
- [📊 项目进度总结](docs/PROJECT_PROGRESS.md)
- [🎨 界面优化记录](docs/UI_IMPROVEMENTS.md)
- [🔧 模型配置说明](docs/MODEL_CONFIGURATION.md)

## 🌟 项目亮点

### 生产就绪
- ✅ 核心功能完整稳定
- ✅ 完善的错误处理和重试机制
- ✅ 详细的监控和日志
- ✅ 安全的限流保护
- ✅ Redis持久化统计

### 易于使用
- 🎯 统一的API接口，无需修改现有代码
- 🔄 自动负载均衡和故障转移
- 📱 响应式监控面板
- 🐳 一键Docker/云端部署
- 🛠️ 丰富的开发工具

### 高性能
- ⚡ Go + Fiber高性能框架 (5000+ RPS)
- 🗄️ Redis缓存和统计存储
- 🌐 全球CDN加速
- 📊 实时性能监控

## 🚀 部署选择

| 平台 | 免费额度 | 优势 | 适用场景 |
|------|----------|------|----------|
| **Render** | 750h/月 | 全球CDN，自动SSL | 推荐，生产使用 |
| **Railway** | $5/月 | Docker原生支持 | 简单快速 |
| **Fly.io** | 3个应用 | 全球边缘网络 | 低延迟需求 |
| **自建** | 服务器成本 | 完全控制 | 高级用户 |

## 🤝 贡献

欢迎提交Issue和Pull Request！

1. Fork项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 打开Pull Request

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE) 文件

---

⭐ **如果这个项目对你有帮助，请给一个Star支持！**

🌐 **立即部署**: [一键部署到Render](https://render.com/deploy)