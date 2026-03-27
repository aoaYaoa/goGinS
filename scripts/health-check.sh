#!/bin/bash

# 健康检查脚本，用于生产环境监控

set -e

HOST=${1:-localhost:8080}
MAX_RETRIES=${2:-30}
RETRY_INTERVAL=${3:-2}

echo "🔍 检查服务健康状态: $HOST"

for i in $(seq 1 $MAX_RETRIES); do
    if curl -f -s "http://$HOST/healthz" > /dev/null; then
        echo "✅ 服务健康检查通过 (尝试 $i/$MAX_RETRIES)"
        
        # 检查就绪状态
        if curl -f -s "http://$HOST/readyz" > /dev/null; then
            echo "✅ 服务就绪检查通过"
            exit 0
        else
            echo "⚠️  服务未就绪（数据库可能未连接）"
            exit 1
        fi
    fi
    
    echo "⏳ 等待服务启动... ($i/$MAX_RETRIES)"
    sleep $RETRY_INTERVAL
done

echo "❌ 服务健康检查失败"
exit 1
