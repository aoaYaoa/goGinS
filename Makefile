.PHONY: help build run test test-integration swagger clean dev docker-build docker-run

help: ## 显示帮助信息
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## 编译项目
	go build -o bin/server ./cmd/server

run: ## 运行服务
	go run ./cmd/server

dev: ## 启动热重载开发模式（需要安装 air）
	air

test: ## 运行所有测试
	go test -v ./...

test-integration: ## 运行集成测试
	go test -v ./internal/repositories/... -timeout 120s

test-coverage: ## 运行测试并生成覆盖率报告
	go test -cover ./...

swagger: ## 生成 Swagger 文档
	swag init -g cmd/server/main.go -o docs

clean: ## 清理构建产物
	rm -rf bin/ tmp/ build-errors.log

docker-build: ## 构建 Docker 镜像
	docker build -t go-gin-starter .

docker-run: ## 运行 Docker 容器
	docker run -p 8080:8080 --env-file .env go-gin-starter

docker-compose-up: ## 启动 docker-compose
	docker compose up -d

docker-compose-down: ## 停止 docker-compose
	docker compose down

install-tools: ## 安装开发工具
	go install github.com/swaggo/swag/cmd/swag@latest
	go install github.com/air-verse/air@latest

lint: ## 运行代码检查
	golangci-lint run

fmt: ## 格式化代码
	go fmt ./...
	goimports -w .
