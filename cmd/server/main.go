package main

import (
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/heyanxiao/llm-bridge/internal/handlers"
	"github.com/heyanxiao/llm-bridge/internal/middleware"
	"github.com/heyanxiao/llm-bridge/internal/providers"
	"github.com/heyanxiao/llm-bridge/internal/stats"
)

func main() {
	// 初始化Redis统计服务
	if err := stats.InitRedisMetrics(); err != nil {
		log.Printf("Redis统计服务初始化失败 (将使用内存统计): %v", err)
	} else {
		log.Println("Redis统计服务初始化成功")
	}

	// 创建Fiber应用实例
	app := fiber.New(fiber.Config{
		ServerHeader: "LLM-Bridge-Gateway",
		AppName:      "LLM网关服务 v1.0.0",
		ErrorHandler: customErrorHandler,
	})

	// 初始化限流器
	var rateLimiter *middleware.RateLimiter
	if redisClient := stats.GetRedisClient(); redisClient != nil {
		rateLimiter = middleware.NewRateLimiter(redisClient)
		log.Println("限流服务初始化成功")
	}

	// 添加中间件
	setupMiddleware(app, rateLimiter)

	// 初始化提供商工厂和负载均衡器
	providerFactory := providers.NewProviderFactory()
	loadBalancer := providers.NewRoundRobinBalancer()

	// 注册LLM提供商
	registerProviders(providerFactory)

	// 设置路由
	setupRoutes(app, providerFactory, loadBalancer, rateLimiter)

	// 获取端口配置
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 启动服务器
	log.Printf("LLM网关服务启动，监听端口: %s", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}

// setupMiddleware 设置中间件
func setupMiddleware(app *fiber.App, rateLimiter *middleware.RateLimiter) {
	// 恢复中间件 - 捕获panic
	app.Use(recover.New(recover.Config{
		EnableStackTrace: true,
	}))

	// 日志中间件
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path} - ${ip} - ${latency}\n",
	}))

	// CORS中间件
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE, OPTIONS",
	}))
	
	// 限流中间件
	if rateLimiter != nil {
		app.Use(rateLimiter.Middleware())
	}
}

// setupRoutes 设置路由
func setupRoutes(app *fiber.App, factory *providers.ProviderFactory, balancer providers.LoadBalancer, rateLimiter *middleware.RateLimiter) {
	// 创建处理器实例
	chatHandler := handlers.NewChatHandler(factory, balancer)
	healthHandler := handlers.NewHealthHandler()
	adminHandler := handlers.NewAdminHandler(factory, balancer)
	
	// 设置限流器
	if rateLimiter != nil {
		adminHandler.SetRateLimiter(rateLimiter)
	}

	// 静态文件服务 - 监控面板
	app.Static("/static", "./static")
	
	// 管理面板路由
	admin := app.Group("/admin")
	admin.Get("/", adminHandler.Dashboard)
	
	// 管理API路由
	adminAPI := admin.Group("/api")
	adminAPI.Get("/providers", adminHandler.GetProvidersStatus)
	adminAPI.Post("/test", adminHandler.TestProvider)
	adminAPI.Get("/stats", adminHandler.GetSystemStats)
	adminAPI.Get("/providers/:provider/models", adminHandler.GetProviderModels)
	adminAPI.Get("/models-config", adminHandler.GetAllModelsConfig)
	
	// 添加简单的限流测试接口
	adminAPI.Get("/rate-limit-test", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "限流测试成功",
			"timestamp": time.Now().Unix(),
		})
	})

	// API v1 路由组
	v1 := app.Group("/v1")

	// 聊天相关路由
	v1.Post("/chat/completions", chatHandler.ChatCompletion)
	v1.Get("/models", chatHandler.Models)

	// 健康检查路由
	health := app.Group("/health")
	health.Get("/", healthHandler.Health)
	health.Get("/ready", healthHandler.Ready)
	health.Get("/live", healthHandler.Live)

	// 根路径 - 重定向到管理面板
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("/admin/", fiber.StatusMovedPermanently)
	})

	// API信息接口
	app.Get("/api/info", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"name":        "LLM网关服务",
			"version":     "1.0.0",
			"description": "统一的LLM API网关，支持多个提供商",
			"endpoints": fiber.Map{
				"chat":     "/v1/chat/completions",
				"models":   "/v1/models",
				"health":   "/health",
				"admin":    "/admin",
				"monitor":  "/admin",
			},
		})
	})
}

// registerProviders 注册LLM提供商
func registerProviders(factory *providers.ProviderFactory) {
	// TODO: 从配置文件读取提供商配置
	
	// 示例：注册OpenAI提供商
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		openaiConfig := &providers.OpenAIConfig{
			APIKey:  apiKey,
			BaseURL: os.Getenv("OPENAI_BASE_URL"),
			Timeout: 30,
			Retries: 3,
		}
		openaiProvider := providers.NewOpenAIProvider(openaiConfig)
		factory.RegisterProvider("openai", openaiProvider)
		log.Println("已注册OpenAI提供商")
	}

	// 注册Gemini提供商
	if apiKey := os.Getenv("GEMINI_API_KEY"); apiKey != "" {
		geminiConfig := &providers.GeminiConfig{
			APIKey:   apiKey,
			BaseURL:  os.Getenv("GEMINI_BASE_URL"),
			Project:  os.Getenv("GEMINI_PROJECT"),
			Location: os.Getenv("GEMINI_LOCATION"),
			Timeout:  30,
			Retries:  3,
		}
		geminiProvider := providers.NewGeminiProvider(geminiConfig)
		factory.RegisterProvider("gemini", geminiProvider)
		log.Println("已注册Gemini提供商")
	}

	// 注册DeepSeek提供商
	if apiKey := os.Getenv("DEEPSEEK_API_KEY"); apiKey != "" {
		deepseekConfig := &providers.DeepSeekConfig{
			APIKey:  apiKey,
			BaseURL: os.Getenv("DEEPSEEK_BASE_URL"),
			Timeout: 30,
			Retries: 3,
		}
		deepseekProvider := providers.NewDeepSeekProvider(deepseekConfig)
		factory.RegisterProvider("deepseek", deepseekProvider)
		log.Println("已注册DeepSeek提供商")
	}

	// 注册通义千问提供商
	if apiKey := os.Getenv("DASHSCOPE_API_KEY"); apiKey != "" {
		qwenConfig := &providers.QwenConfig{
			APIKey:  apiKey,
			BaseURL: os.Getenv("QWEN_BASE_URL"),
			Timeout: 30,
			Retries: 3,
		}
		qwenProvider := providers.NewQwenProvider(qwenConfig)
		factory.RegisterProvider("qwen", qwenProvider)
		log.Println("已注册通义千问提供商")
	}

	// 注册月之暗面提供商
	if apiKey := os.Getenv("MOONSHOT_API_KEY"); apiKey != "" {
		moonshotConfig := &providers.MoonshotConfig{
			APIKey:  apiKey,
			BaseURL: os.Getenv("MOONSHOT_BASE_URL"),
			Timeout: 60,
			Retries: 3,
		}
		moonshotProvider := providers.NewMoonshotProvider(moonshotConfig)
		factory.RegisterProvider("moonshot", moonshotProvider)
		log.Println("已注册月之暗面提供商")
	}

	// TODO: 注册其他提供商(Claude, Azure等)
	// 这里可以根据环境变量或配置文件动态注册
}

// customErrorHandler 自定义错误处理器
func customErrorHandler(c *fiber.Ctx, err error) error {
	// 默认500状态码
	code := fiber.StatusInternalServerError
	message := "内部服务器错误"

	// 检查是否为Fiber错误
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		message = e.Message
	}

	// 返回错误响应
	return c.Status(code).JSON(fiber.Map{
		"error": fiber.Map{
			"code":    "internal_server_error",
			"message": message,
			"type":    "server_error",
		},
	})
}