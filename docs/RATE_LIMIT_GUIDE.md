# 限流功能使用指南

## 概述

LLM网关现已集成基于Redis的限流功能，可以有效防止恶意请求和保护后端LLM服务。

## 限流策略

### 1. 全局限流
- **1分钟窗口**: 默认20次请求
- **5分钟窗口**: 默认240次请求  
- **1小时窗口**: 默认1000次请求

### 2. 接口级限流
- **聊天接口** (`/v1/chat/completions`):
  - 1分钟: 10次
  - 5分钟: 120次
- **测试接口** (`/admin/api/test`):
  - 1分钟: 10次
- **其他接口**: 使用全局限流配置

## 配置说明

在 `.env` 文件中配置限流参数：

```env
# 限流总开关
RATE_LIMIT_ENABLED=true

# 全局限流配置
RATE_LIMIT_WINDOW_1M=20
RATE_LIMIT_WINDOW_5M=240
RATE_LIMIT_WINDOW_1H=1000

# 特定接口限流
RATE_LIMIT_CHAT_1M=10
RATE_LIMIT_CHAT_5M=120
RATE_LIMIT_TEST_1M=10
```

## 限流响应

当触发限流时，服务会返回HTTP 429状态码：

```json
{
  "error": "Rate limit exceeded",
  "message": "请求过于频繁，请稍后再试"
}
```

## 监控面板

访问 `http://localhost:8080/admin/` 查看限流状态：

![限流状态卡片](限流状态示例)
- **限流状态**: 显示是否启用
- **1分钟限制**: 当前1分钟窗口配置
- **5分钟限制**: 当前5分钟窗口配置

## 实现原理

### 滑动窗口算法
使用Redis的键过期机制实现滑动窗口：
- 每个时间窗口生成唯一键
- 请求计数自动递增
- 超过TTL自动清理

### Redis键格式
```
rate_limit:{path}:1m:{timestamp_minute}
rate_limit:{path}:5m:{timestamp_5min}
rate_limit:global:1h:{timestamp_hour}
```

## 性能影响

- **延迟增加**: < 2ms (Redis查询)
- **内存占用**: 极低 (仅存储计数器)
- **并发性能**: 不影响 (Pipeline操作)

## 最佳实践

1. **生产环境配置建议**:
   ```env
   RATE_LIMIT_CHAT_1M=30    # 适当放宽聊天限制
   RATE_LIMIT_CHAT_5M=120   # 防止突发流量
   RATE_LIMIT_WINDOW_1H=2000 # 合理的小时限制
   ```

2. **监控告警**:
   - 关注429响应码比例
   - 设置限流触发告警
   - 定期审查限流配置

3. **优雅降级**:
   - Redis故障时自动禁用限流
   - 不影响核心业务功能
   - 记录限流日志便于分析

## 测试限流

使用curl测试限流效果：

```bash
# 快速发送多个请求测试限流
for i in {1..30}; do
  curl -X POST http://localhost:8080/v1/chat/completions \
    -H "Content-Type: application/json" \
    -d '{"model":"gpt-3.5-turbo","messages":[{"role":"user","content":"test"}]}'
  echo ""
done
```

预期结果：
- 前10个请求正常响应
- 第11个请求开始返回429错误
- 等待1分钟后恢复正常

## 故障排查

1. **限流不生效**:
   - 检查 `RATE_LIMIT_ENABLED=true`
   - 确认Redis连接正常
   - 查看服务启动日志

2. **限流过于严格**:
   - 适当提高限制值
   - 检查是否有配置错误
   - 考虑接口差异化配置

3. **Redis连接问题**:
   - 检查Redis服务状态
   - 验证连接配置
   - 查看错误日志

## 总结

限流功能是保护LLM网关稳定运行的重要机制。通过合理配置，可以有效防止恶意请求，同时不影响正常用户使用。建议根据实际使用情况调整限流参数，找到安全性和可用性的平衡点。