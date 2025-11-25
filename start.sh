#!/bin/bash

# 颜色定义
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}"
echo "╔════════════════════════════════════════════════════════════╗"
echo "║              域名扫描器 - 快速启动脚本                      ║"
echo "║            Domain Scanner - Quick Start Script            ║"
echo "╚════════════════════════════════════════════════════════════╝"
echo -e "${NC}"

# 检查 Docker 是否安装
if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ Docker 未安装！${NC}"
    echo "请先安装 Docker: https://docs.docker.com/get-docker/"
    exit 1
fi

# 检查 Docker Compose 是否安装
if ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}❌ Docker Compose 未安装！${NC}"
    echo "请先安装 Docker Compose: https://docs.docker.com/compose/install/"
    exit 1
fi

echo -e "${GREEN}✅ Docker 环境检查通过${NC}"
echo ""

# 检查端口是否被占用
if lsof -Pi :8080 -sTCP:LISTEN -t >/dev/null 2>&1 ; then
    echo -e "${YELLOW}⚠️  警告: 端口 8080 已被占用${NC}"
    read -p "是否继续？(y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# 构建并启动服务
echo -e "${BLUE}🔨 正在构建 Docker 镜像...${NC}"
docker-compose build

if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Docker 镜像构建失败！${NC}"
    exit 1
fi

echo ""
echo -e "${BLUE}🚀 正在启动服务...${NC}"
docker-compose up -d

if [ $? -ne 0 ]; then
    echo -e "${RED}❌ 服务启动失败！${NC}"
    exit 1
fi

echo ""
echo -e "${GREEN}✅ 服务启动成功！${NC}"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${GREEN}📱 Web UI:${NC}       http://localhost:8080"
echo -e "${GREEN}🗄️  数据库:${NC}       localhost:5432"
echo -e "${GREEN}👤 数据库用户:${NC}     scanner"
echo -e "${GREEN}🔑 数据库密码:${NC}     scanner123"
echo -e "${GREEN}📊 数据库名:${NC}       domainscanner"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "💡 常用命令:"
echo "  • 查看日志:     docker-compose logs -f"
echo "  • 停止服务:     docker-compose down"
echo "  • 重启服务:     docker-compose restart"
echo "  • 查看状态:     docker-compose ps"
echo ""
echo "📚 更多信息请查看: DEPLOYMENT.md"
echo ""

# 等待服务启动
echo -e "${YELLOW}⏳ 等待服务启动...${NC}"
sleep 5

# 检查服务状态
if docker-compose ps | grep -q "Up"; then
    echo -e "${GREEN}✅ 所有服务运行正常！${NC}"
    echo ""
    echo -e "${BLUE}🌐 正在打开浏览器...${NC}"

    # 尝试打开浏览器
    if command -v open &> /dev/null; then
        open http://localhost:8080
    elif command -v xdg-open &> /dev/null; then
        xdg-open http://localhost:8080
    else
        echo -e "${YELLOW}请手动打开浏览器访问: http://localhost:8080${NC}"
    fi
else
    echo -e "${RED}❌ 服务启动异常，请检查日志${NC}"
    echo "运行以下命令查看日志:"
    echo "  docker-compose logs"
fi

