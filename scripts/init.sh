#!/bin/bash

set -e

echo "🚀 初始化 go-gin-starter 项目..."

# 检查 .env 文件
if [ ! -f .env ]; then
    echo "📝 创建 .env 文件..."
    cp .env.example .env
    
    # 生成随机 JWT_SECRET
    JWT_SECRET=$(openssl rand -base64 32)
    if [[ "$OSTYPE" == "darwin"* ]]; then
        sed -i '' "s|JWT_SECRET=.*|JWT_SECRET=$JWT_SECRET|" .env
    else
        sed -i "s|JWT_SECRET=.*|JWT_SECRET=$JWT_SECRET|" .env
    fi
    echo "✅ 已生成随机 JWT_SECRET"
else
    echo "✅ .env 文件已存在"
fi

# 检查 Go 版本
if ! command -v go &> /dev/null; then
    echo "❌ 未检测到 Go，请先安装 Go 1.26+"
    exit 1
fi

GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
echo "✅ Go 版本: $GO_VERSION"

# 安装依赖
echo "📦 安装 Go 依赖..."
go mod download

# 安装开发工具（可选）
read -p "是否安装开发工具 (swag, air)? [y/N] " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    echo "🔧 安装 swag..."
    go install github.com/swaggo/swag/cmd/swag@latest
    
    echo "🔧 安装 air..."
    go install github.com/air-verse/air@latest
    
    echo "✅ 开发工具安装完成"
fi

# 生成 Swagger 文档
if command -v swag &> /dev/null; then
    echo "📚 生成 Swagger 文档..."
    swag init -g cmd/server/main.go -o docs
    echo "✅ Swagger 文档已生成"
fi

# 检查 Docker
if command -v docker &> /dev/null; then
    echo "✅ Docker 已安装"
    
    read -p "是否启动 docker-compose (PostgreSQL + 可选 Redis)? [y/N] " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "🐳 启动 docker-compose..."
        docker compose up -d
        echo "⏳ 等待数据库就绪..."
        sleep 5
        echo "✅ 服务已启动"
    fi
else
    echo "⚠️  未检测到 Docker，请手动配置 PostgreSQL（Redis 可选）"
fi

echo ""
echo "🎉 初始化完成！"
echo ""
echo "下一步："
echo "  1. 编辑 .env 配置数据库连接"
echo "  2. 运行服务:"
echo "     - 开发模式: make dev (或 air)"
echo "     - 普通模式: make run (或 go run ./cmd/server)"
echo "  3. 访问 API 文档: http://localhost:8080/swagger/index.html"
echo ""
