package middleware

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

type RateLimiter struct {
	client  *redis.Client
	enabled bool
	
	// 全局限流配置
	window1m  int
	window5m  int
	window1h  int
	
	// 特定接口限流配置
	chatLimit1m int
	chatLimit5m int
	testLimit1m int
}

func NewRateLimiter(client *redis.Client) *RateLimiter {
	rl := &RateLimiter{
		client: client,
	}
	
	// 从环境变量加载配置
	rl.loadConfig()
	
	return rl
}

func (rl *RateLimiter) loadConfig() {
	// 是否启用限流
	rl.enabled = os.Getenv("RATE_LIMIT_ENABLED") == "true"
	
	// 全局限流窗口
	rl.window1m = getEnvAsInt("RATE_LIMIT_WINDOW_1M", 60)
	rl.window5m = getEnvAsInt("RATE_LIMIT_WINDOW_5M", 240)
	rl.window1h = getEnvAsInt("RATE_LIMIT_WINDOW_1H", 1000)
	
	// 特定接口限流
	rl.chatLimit1m = getEnvAsInt("RATE_LIMIT_CHAT_1M", 30)
	rl.chatLimit5m = getEnvAsInt("RATE_LIMIT_CHAT_5M", 120)
	rl.testLimit1m = getEnvAsInt("RATE_LIMIT_TEST_1M", 10)
	
	// 启动日志
	if rl.enabled {
		fmt.Printf("[RateLimit] 限流功能已启用 - 全局:%d/1m %d/5m, 聊天:%d/1m, 测试:%d/1m\n", 
			rl.window1m, rl.window5m, rl.chatLimit1m, rl.testLimit1m)
	}
}

func getEnvAsInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}

// Middleware 返回限流中间件
func (rl *RateLimiter) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// 如果未启用限流，直接放行
		if !rl.enabled || rl.client == nil {
			return c.Next()
		}
		
		path := c.Path()
		ctx := context.Background()
		
		// 检查限流
		allowed, err := rl.checkRateLimit(ctx, path)
		if err != nil {
			// 发生错误时记录日志但不阻塞请求
			fmt.Printf("[RateLimit] 检查错误: %v\n", err)
			return c.Next()
		}
		
		if !allowed {
			// 返回429状态码
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Rate limit exceeded",
				"message": "请求过于频繁，请稍后再试",
			})
		}
		
		return c.Next()
	}
}

func (rl *RateLimiter) checkRateLimit(ctx context.Context, path string) (bool, error) {
	now := time.Now()
	
	// 获取路径特定的限制
	limit1m, limit5m := rl.getPathLimits(path)
	
	// 构建Redis键
	minute := now.Truncate(time.Minute).Unix()
	fiveMin := now.Truncate(5 * time.Minute).Unix()
	hour := now.Truncate(time.Hour).Unix()
	
	// 检查1分钟窗口
	if limit1m > 0 {
		key1m := fmt.Sprintf("rate_limit:%s:1m:%d", path, minute)
		count, err := rl.incrementAndCheck(ctx, key1m, limit1m, 65) // 65秒过期
		if err != nil {
			return false, err
		}
		if count > limit1m {
			return false, nil
		}
	}
	
	// 检查5分钟窗口
	if limit5m > 0 {
		key5m := fmt.Sprintf("rate_limit:%s:5m:%d", path, fiveMin)
		count, err := rl.incrementAndCheck(ctx, key5m, limit5m, 310) // 310秒过期
		if err != nil {
			return false, err
		}
		if count > limit5m {
			return false, nil
		}
	}
	
	// 检查全局1小时窗口
	if rl.window1h > 0 {
		keyGlobal := fmt.Sprintf("rate_limit:global:1h:%d", hour)
		count, err := rl.incrementAndCheck(ctx, keyGlobal, rl.window1h, 3660) // 61分钟过期
		if err != nil {
			return false, err
		}
		if count > rl.window1h {
			return false, nil
		}
	}
	
	return true, nil
}

func (rl *RateLimiter) incrementAndCheck(ctx context.Context, key string, limit int, ttl int) (int, error) {
	pipe := rl.client.Pipeline()
	
	// 增加计数
	incr := pipe.Incr(ctx, key)
	// 设置过期时间（仅在键不存在时）
	pipe.Expire(ctx, key, time.Duration(ttl)*time.Second)
	
	_, err := pipe.Exec(ctx)
	if err != nil {
		return 0, err
	}
	
	count := int(incr.Val())
	return count, nil
}

func (rl *RateLimiter) getPathLimits(path string) (limit1m, limit5m int) {
	switch path {
	case "/v1/chat/completions":
		return rl.chatLimit1m, rl.chatLimit5m
	case "/admin/api/test":
		return rl.testLimit1m, 0 // 测试接口只限制1分钟
	default:
		// 其他接口使用全局限制
		return rl.window1m, rl.window5m
	}
}

// GetStats 获取限流统计信息
func (rl *RateLimiter) GetStats(ctx context.Context) map[string]interface{} {
	if !rl.enabled || rl.client == nil {
		return map[string]interface{}{
			"enabled": false,
		}
	}
	
	stats := map[string]interface{}{
		"enabled": true,
		"config": map[string]interface{}{
			"window_1m": rl.window1m,
			"window_5m": rl.window5m,
			"window_1h": rl.window1h,
		},
	}
	
	// 可以添加当前限流计数等信息
	
	return stats
}