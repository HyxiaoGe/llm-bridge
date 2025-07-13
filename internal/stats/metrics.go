package stats

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisMetrics Redis统计服务
type RedisMetrics struct {
	client    *redis.Client
	startTime time.Time
}

var globalMetrics *RedisMetrics

// InitRedisMetrics 初始化Redis统计服务
func InitRedisMetrics() error {
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost"
	}
	
	port := os.Getenv("REDIS_PORT")
	if port == "" {
		port = "6379"
	}
	
	password := os.Getenv("REDIS_PASSWORD")
	
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       0, // 使用默认数据库
	})
	
	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if err := rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis连接失败: %w", err)
	}
	
	globalMetrics = &RedisMetrics{
		client:    rdb,
		startTime: time.Now(),
	}
	
	// 初始化服务启动时间
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	globalMetrics.client.Set(ctx2, "service:start_time", time.Now().Unix(), 0)
	
	return nil
}

// GetRedisMetrics 获取Redis统计实例
func GetRedisMetrics() *RedisMetrics {
	return globalMetrics
}

// IncrementRequest 增加请求计数并记录响应时间
func (m *RedisMetrics) IncrementRequest(provider string, responseTime time.Duration, tokens int) {
	if m.client == nil {
		return
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	// 使用Pipeline提高性能
	pipe := m.client.Pipeline()
	
	// 全局统计
	pipe.Incr(ctx, "stats:total_requests")
	pipe.IncrBy(ctx, "stats:total_response_time", responseTime.Milliseconds())
	pipe.IncrBy(ctx, "stats:total_tokens", int64(tokens))
	
	// 按提供商统计
	if provider != "" {
		pipe.Incr(ctx, fmt.Sprintf("stats:provider:%s:requests", provider))
		pipe.IncrBy(ctx, fmt.Sprintf("stats:provider:%s:response_time", provider), responseTime.Milliseconds())
		pipe.IncrBy(ctx, fmt.Sprintf("stats:provider:%s:tokens", provider), int64(tokens))
	}
	
	// 按日期统计
	today := time.Now().Format("2006-01-02")
	pipe.Incr(ctx, fmt.Sprintf("stats:daily:%s:requests", today))
	pipe.IncrBy(ctx, fmt.Sprintf("stats:daily:%s:tokens", today), int64(tokens))
	
	// 执行Pipeline
	pipe.Exec(ctx)
}

// GetStats 获取统计数据
func (m *RedisMetrics) GetStats() (totalRequests int64, avgResponseTime int64, totalTokens int64, uptime time.Duration) {
	if m.client == nil {
		return 0, 0, 0, time.Since(m.startTime)
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	// 获取统计数据
	requests, _ := m.client.Get(ctx, "stats:total_requests").Int64()
	responseTime, _ := m.client.Get(ctx, "stats:total_response_time").Int64()
	tokens, _ := m.client.Get(ctx, "stats:total_tokens").Int64()
	
	// 获取服务启动时间
	startTimeUnix, err := m.client.Get(ctx, "service:start_time").Int64()
	if err == nil {
		m.startTime = time.Unix(startTimeUnix, 0)
	}
	
	uptime = time.Since(m.startTime)
	totalRequests = requests
	totalTokens = tokens
	
	if requests > 0 {
		avgResponseTime = responseTime / requests
	}
	
	return
}

// GetProviderStats 获取提供商统计
func (m *RedisMetrics) GetProviderStats(provider string) (requests int64, avgResponseTime int64, tokens int64) {
	if m.client == nil {
		return 0, 0, 0
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	requests, _ = m.client.Get(ctx, fmt.Sprintf("stats:provider:%s:requests", provider)).Int64()
	responseTime, _ := m.client.Get(ctx, fmt.Sprintf("stats:provider:%s:response_time", provider)).Int64()
	tokens, _ = m.client.Get(ctx, fmt.Sprintf("stats:provider:%s:tokens", provider)).Int64()
	
	if requests > 0 {
		avgResponseTime = responseTime / requests
	}
	
	return
}

// GetDailyStats 获取每日统计
func (m *RedisMetrics) GetDailyStats(date string) (requests int64, tokens int64) {
	if m.client == nil {
		return 0, 0
	}
	
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	
	requests, _ = m.client.Get(ctx, fmt.Sprintf("stats:daily:%s:requests", date)).Int64()
	tokens, _ = m.client.Get(ctx, fmt.Sprintf("stats:daily:%s:tokens", date)).Int64()
	
	return
}

// Close 关闭Redis连接
func (m *RedisMetrics) Close() error {
	if m.client != nil {
		return m.client.Close()
	}
	return nil
}