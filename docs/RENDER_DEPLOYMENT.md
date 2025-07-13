# Render.com 部署指南

## 🚀 快速部署到Render

Render是一个现代化的云平台，支持Docker部署，全球CDN，非常适合部署LLM网关服务。

## 📋 部署前准备

### 1. 注册Render账号
访问 [render.com](https://render.com) 注册免费账号

### 2. 准备API密钥
确保你有以下LLM提供商的API密钥：
- OpenAI API Key
- Google Gemini API Key  
- DeepSeek API Key
- 阿里通义千问 API Key (DASHSCOPE_API_KEY)
- 月之暗面 API Key (MOONSHOT_API_KEY)

## 🔧 部署步骤

### 1. 推送代码到GitHub
```bash
# 如果还没有推送到GitHub
git remote add origin https://github.com/yourusername/llm-bridge.git
git push -u origin master
```

### 2. 连接GitHub到Render

1. 登录Render Dashboard
2. 点击 "New +" → "Blueprint"
3. 连接你的GitHub账号
4. 选择 `llm-bridge` 仓库
5. Render会自动检测到 `render.yaml` 配置文件

### 3. 配置环境变量

在Render Dashboard中设置以下敏感环境变量：

#### 必需的API密钥
```bash
OPENAI_API_KEY=sk-your-openai-key-here
GEMINI_API_KEY=your-gemini-key-here
DEEPSEEK_API_KEY=sk-your-deepseek-key-here
DASHSCOPE_API_KEY=sk-your-qwen-key-here
MOONSHOT_API_KEY=sk-your-moonshot-key-here
```

#### 可选配置
```bash
# 如果需要自定义基础URL
OPENAI_BASE_URL=https://api.openai.com/v1
GEMINI_BASE_URL=https://generativelanguage.googleapis.com/v1beta
```

### 4. 部署

1. 点击 "Apply" 开始部署
2. 等待构建完成（约3-5分钟）
3. 获得你的部署URL: `https://your-app-name.onrender.com`

## 🌐 访问你的服务

部署完成后，你可以通过以下方式访问：

- **监控面板**: `https://your-app-name.onrender.com/admin/`
- **API端点**: `https://your-app-name.onrender.com/v1/chat/completions`
- **健康检查**: `https://your-app-name.onrender.com/health`

## 📊 免费额度说明

Render免费计划包括：
- ✅ **750小时/月**: 网关服务运行时间
- ✅ **25MB Redis**: 统计数据存储
- ✅ **全球CDN**: 自动优化访问速度
- ✅ **自动SSL**: HTTPS加密
- ✅ **自动重启**: 服务故障自动恢复

## 🔍 监控和调试

### 查看日志
1. 进入Render Dashboard
2. 选择你的服务
3. 点击 "Logs" 标签查看实时日志

### 查看指标
- CPU使用率
- 内存使用率
- 请求数量和响应时间

## ⚙️ 生产环境配置

如果需要升级到付费计划（$7/月），你将获得：
- ✅ **24/7运行**: 无休眠时间
- ✅ **更多资源**: 512MB内存，0.1CPU
- ✅ **更大Redis**: 付费Redis实例
- ✅ **自定义域名**: 绑定你的域名

## 🛠️ 故障排查

### 1. 服务无法启动
- 检查环境变量是否正确设置
- 查看构建日志是否有错误
- 确认API密钥格式正确

### 2. Redis连接失败
- 确认Redis服务已启动
- 检查Redis连接配置

### 3. API调用失败
- 验证API密钥有效性
- 检查网络连接
- 查看限流配置是否过于严格

## 🔒 安全建议

1. **API密钥安全**
   - 使用Render的环境变量功能
   - 不要在代码中硬编码密钥
   - 定期轮换API密钥

2. **限流配置**
   - 根据实际需求调整限流参数
   - 监控API使用量避免超额

3. **访问控制**
   - 考虑添加基础认证
   - 监控异常访问模式

## 📈 性能优化

1. **Redis优化**
   - 监控内存使用
   - 适当调整过期策略

2. **限流调优**
   - 根据用户反馈调整限制
   - 平衡安全性和可用性

## 🔄 更新部署

代码更新后自动重新部署：
```bash
git add .
git commit -m "更新功能"
git push origin master
```

Render会自动检测到推送并重新构建部署。

## 📞 获得帮助

- **Render官方文档**: https://render.com/docs
- **LLM网关项目**: 查看项目README和docs目录
- **问题反馈**: 创建GitHub Issue

部署成功后，你就拥有了一个全球可访问、高可用的LLM统一网关服务！🎉