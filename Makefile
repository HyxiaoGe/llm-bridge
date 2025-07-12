# LLM网关服务 Makefile

.PHONY: help build run test clean docker-build docker-run docker-stop deps lint fmt

# 默认目标
help:
	@echo "LLM网关服务 - 可用命令:"
	@echo "  build        - 编译应用程序"
	@echo "  run          - 运行应用程序"
	@echo "  test         - 运行测试"
	@echo "  clean        - 清理构建文件"
	@echo "  docker-build - 构建Docker镜像"
	@echo "  docker-run   - 运行Docker容器"
	@echo "  docker-stop  - 停止Docker容器"
	@echo "  deps         - 下载依赖"
	@echo "  lint         - 运行代码检查"
	@echo "  fmt          - 格式化代码"

# 构建应用程序
build:
	@echo "正在编译应用程序..."
	go build -o bin/gateway ./cmd/server

# 运行应用程序
run:
	@echo "正在启动LLM网关服务..."
	go run ./cmd/server/main.go

# 运行测试
test:
	@echo "正在运行测试..."
	go test -v ./...

# 清理构建文件
clean:
	@echo "正在清理构建文件..."
	rm -rf bin/
	go clean

# 构建Docker镜像
docker-build:
	@echo "正在构建Docker镜像..."
	docker build -t llm-gateway:latest .

# 运行Docker容器
docker-run:
	@echo "正在启动Docker容器..."
	docker-compose up -d

# 停止Docker容器
docker-stop:
	@echo "正在停止Docker容器..."
	docker-compose down

# 下载依赖
deps:
	@echo "正在下载依赖..."
	go mod download
	go mod tidy

# 代码检查
lint:
	@echo "正在运行代码检查..."
	go vet ./...
	go fmt ./...

# 格式化代码
fmt:
	@echo "正在格式化代码..."
	go fmt ./...

# 创建目录
bin:
	mkdir -p bin