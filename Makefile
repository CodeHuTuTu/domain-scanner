.PHONY: help build run-cli run-web docker-build docker-up docker-down clean

help: ## 显示帮助信息
	@echo "域名扫描器 - 可用命令:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

build: ## 编译所有程序
	@echo "编译 CLI 工具..."
	go build -o bin/domain-scanner main.go
	@echo "编译 Web 服务器..."
	go build -o bin/webserver cmd/webserver/main.go
	@echo "✅ 编译完成！"

run-cli: ## 运行 CLI 工具（示例）
	go run main.go -l 3 -s .li -p D -workers 10

run-web: ## 本地运行 Web 服务器
	go run cmd/webserver/main.go

docker-build: ## 构建 Docker 镜像
	docker-compose build

docker-up: ## 启动 Docker 容器
	docker-compose up -d
	@echo ""
	@echo "🚀 服务已启动！"
	@echo "📱 Web UI: http://localhost:8080"
	@echo "🗄️  PostgreSQL: localhost:5432"
	@echo ""
	@echo "查看日志: make docker-logs"

docker-down: ## 停止 Docker 容器
	docker-compose down

docker-logs: ## 查看 Docker 日志
	docker-compose logs -f

docker-restart: ## 重启 Docker 容器
	docker-compose restart

docker-clean: ## 停止容器并删除数据卷
	docker-compose down -v
	@echo "✅ 已清理所有容器和数据卷"

deps: ## 下载 Go 依赖
	go mod download
	go mod tidy

test: ## 运行测试
	go test -v ./...

clean: ## 清理构建文件
	rm -rf bin/
	rm -f *.txt
	@echo "✅ 清理完成！"

install: ## 安装到系统
	go install ./...

.DEFAULT_GOAL := help

