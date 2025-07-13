# API密钥配置指南

## 问题说明

如果你在监控面板中看到API测试失败（401、400错误），这是正常现象，通常有以下原因：

### 常见错误类型

1. **401 Unauthorized** - API密钥无效或已过期
2. **400 Bad Request** - 请求格式错误或参数无效  
3. **403 Forbidden** - 账户余额不足或权限限制
4. **Invalid API-key provided** - API密钥格式错误

## 解决方案

### 1. 更新API密钥

编辑 `.env` 文件，替换为你的有效API密钥：

```bash
# 复制示例文件
cp .env.example .env

# 编辑.env文件，填入你的真实API密钥
vim .env
```

### 2. 获取API密钥

#### OpenAI
- 访问: https://platform.openai.com/api-keys
- 创建新的API密钥
- 确保账户有足够余额

#### DeepSeek
- 访问: https://platform.deepseek.com/api_keys
- 注册并获取API密钥
- 新用户通常有免费额度

#### 通义千问 (阿里云)
- 访问: https://dashscope.console.aliyun.com/apiKey
- 开通DashScope服务
- 创建API密钥

#### 月之暗面 (Moonshot)
- 访问: https://platform.moonshot.cn/console/api-keys
- 注册并创建API密钥

#### Google Gemini
- 访问: https://aistudio.google.com/app/apikey
- 创建API密钥
- 注意地区限制

### 3. 重启服务

更新API密钥后重启Docker服务：

```bash
# 重启服务
docker compose restart

# 或者重新构建
docker compose down
docker compose up -d
```

### 4. 验证配置

1. 访问监控面板: http://localhost:8080/admin/
2. 查看提供商状态 - 有效密钥会显示"健康"状态
3. 使用测试工具验证连接

## 注意事项

1. **安全性**: 不要将API密钥提交到版本控制系统
2. **余额**: 确保账户有足够余额
3. **限制**: 某些提供商有地区或使用限制
4. **可选性**: 即使某些提供商配置失败，其他提供商仍可正常使用

## 当前测试结果

基于目前的测试，只有OpenAI的密钥是有效的，其他提供商的密钥需要更新。这不影响系统的整体功能，监控面板和API网关仍然正常工作。