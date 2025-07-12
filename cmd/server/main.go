package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/heyanxiao/llm-bridge/internal/handlers"
	"github.com/heyanxiao/llm-bridge/internal/providers"
)

func main() {
	// 创建Fiber应用实例
	app := fiber.New(fiber.Config{
		ServerHeader: "LLM-Bridge-Gateway",
		AppName:      "LLM网关服务 v1.0.0",
		ErrorHandler: customErrorHandler,
	})

	// 添加中间件
	setupMiddleware(app)

	// 初始化提供商工厂和负载均衡器
	providerFactory := providers.NewProviderFactory()
	loadBalancer := providers.NewRoundRobinBalancer()

	// 注册LLM提供商
	registerProviders(providerFactory)

	// 设置路由
	setupRoutes(app, providerFactory, loadBalancer)

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
func setupMiddleware(app *fiber.App) {
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
}

// setupRoutes 设置路由
func setupRoutes(app *fiber.App, factory *providers.ProviderFactory, balancer providers.LoadBalancer) {
	// 创建处理器实例
	chatHandler := handlers.NewChatHandler(factory, balancer)
	healthHandler := handlers.NewHealthHandler()

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

	// 根路径
	app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"name":        "LLM网关服务",
			"version":     "1.0.0",
			"description": "统一的LLM API网关，支持多个提供商",
			"endpoints": fiber.Map{
				"chat":   "/v1/chat/completions",
				"models": "/v1/models",
				"health": "/health",
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

	// TODO: 注册其他提供商(Claude, Gemini等)
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