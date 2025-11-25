#!/bin/bash

# 测试脚本 - Test Script
# 用于验证所有组件是否正确配置

echo "🧪 运行域名扫描器测试..."
echo "Running Domain Scanner Tests..."
echo ""

# 检查必要文件
echo "📁 检查文件结构..."
required_files=(
    "main.go"
    "cmd/webserver/main.go"
    "internal/database/database.go"
    "internal/web/server.go"
    "internal/web/static/index.html"
    "docker-compose.yml"
    "Dockerfile"
    "go.mod"
    "Makefile"
    "DEPLOYMENT.md"
)

missing_files=()
for file in "${required_files[@]}"; do
    if [ ! -f "$file" ]; then
        missing_files+=("$file")
    fi
done

if [ ${#missing_files[@]} -eq 0 ]; then
    echo "✅ 所有必要文件存在"
else
    echo "❌ 缺少以下文件:"
    printf '%s\n' "${missing_files[@]}"
    exit 1
fi

echo ""
echo "📦 检查 Go 模块..."
if [ -f "go.mod" ]; then
    echo "✅ go.mod 存在"
    echo "   依赖包："
    grep "^\s" go.mod | head -n 10
else
    echo "❌ go.mod 不存在"
    exit 1
fi

echo ""
echo "🐳 检查 Docker 配置..."
if [ -f "docker-compose.yml" ]; then
    echo "✅ docker-compose.yml 存在"
    echo "   服务列表："
    grep "^  [a-z]" docker-compose.yml
else
    echo "❌ docker-compose.yml 不存在"
    exit 1
fi

echo ""
echo "📄 检查文档..."
docs=("README.md" "README.zh.md" "DEPLOYMENT.md" "docs/CHANGELOG.md")
for doc in "${docs[@]}"; do
    if [ -f "$doc" ]; then
        echo "✅ $doc"
    else
        echo "❌ $doc 缺失"
    fi
done

echo ""
echo "🎉 测试完成！"
echo ""
echo "下一步："
echo "1. 确保已安装 Docker 和 Docker Compose"
echo "2. 运行 ./start.sh 启动服务"
echo "3. 访问 http://localhost:8080"
echo ""
echo "或者使用 CLI 版本："
echo "  go run main.go -l 3 -s .li -p D -workers 10"

