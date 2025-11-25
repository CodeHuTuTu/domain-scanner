# 快速参考指南 - Quick Reference Guide

## 🚀 快速启动 Quick Start

### Web UI 版本
```bash
./start.sh                    # 一键启动 / One-click start
# 或 Or
docker-compose up -d          # Docker Compose 启动
```

访问 / Visit: http://localhost:8080

### CLI 版本
```bash
go run main.go -l 3 -s .li -p D -workers 20
```

---

## 📋 常用命令 Common Commands

### Docker 管理
```bash
docker-compose up -d          # 启动服务 (后台运行)
docker-compose down           # 停止服务
docker-compose restart        # 重启服务
docker-compose logs -f        # 查看日志
docker-compose ps             # 查看状态
docker-compose down -v        # 停止并删除数据
```

### Make 命令
```bash
make help                     # 显示帮助
make build                    # 编译程序
make run-web                  # 本地运行 Web 服务器
make run-cli                  # 运行 CLI 工具
make docker-up                # 启动 Docker
make docker-down              # 停止 Docker
make docker-logs              # 查看日志
make clean                    # 清理构建文件
```

---

## 🔧 CLI 参数 CLI Parameters

| 参数 | 说明 | 默认值 | 示例 |
|------|------|--------|------|
| `-l` | 域名长度 | 3 | `-l 4` |
| `-s` | 域名后缀 | .li | `-s .com` |
| `-p` | 域名模式<br>d: 纯数字<br>D: 纯字母<br>a: 字母数字 | D | `-p d` |
| `-r` | 正则过滤器 | - | `-r "^abc"` |
| `-dict` | 字典文件 | - | `-dict words.txt` |
| `-delay` | 查询延迟 (毫秒) | 1000 | `-delay 500` |
| `-workers` | 工作线程数 | 10 | `-workers 20` |
| `-show-registered` | 显示已注册域名 | false | `-show-registered` |
| `-force` | 跳过警告 | false | `-force` |
| `-h` | 显示帮助 | - | `-h` |

---

## 📊 API 端点 API Endpoints

### 统计信息
```bash
GET /api/stats
# 返回域名统计数据
```

### 获取域名列表
```bash
GET /api/domains?available=true&limit=50&offset=0
# 参数：
#   available: true/false (可选)
#   limit: 返回数量
#   offset: 偏移量
```

### 搜索域名
```bash
GET /api/domains/search?q=example&available=true
# 参数：
#   q: 搜索关键词 (必需)
#   available: true/false (可选)
#   limit: 返回数量
#   offset: 偏移量
```

### 获取扫描历史
```bash
GET /api/sessions?limit=20&offset=0
# 返回扫描会话列表
```

### 启动扫描
```bash
POST /api/scan
Content-Type: application/json

{
  "length": 3,
  "suffix": ".li",
  "pattern": "D",
  "regex_filter": "",
  "delay": 1000,
  "workers": 10
}
```

---

## 🗄️ 数据库 Database

### 连接信息
- **主机**: localhost
- **端口**: 5432
- **数据库**: domainscanner
- **用户**: scanner
- **密码**: scanner123

### 连接命令
```bash
# Docker 内连接
docker-compose exec postgres psql -U scanner -d domainscanner

# 本地连接 (需要安装 psql)
psql -h localhost -p 5432 -U scanner -d domainscanner
```

### 常用 SQL
```sql
-- 查看可用域名数量
SELECT COUNT(*) FROM domain_records WHERE available = true;

-- 查看最近扫描的域名
SELECT * FROM domain_records ORDER BY checked_at DESC LIMIT 10;

-- 查看扫描会话统计
SELECT id, pattern, length, suffix, available_count, registered_count, status 
FROM scan_sessions ORDER BY started_at DESC;

-- 搜索特定域名
SELECT * FROM domain_records WHERE domain LIKE '%example%';
```

---

## 🔍 使用示例 Usage Examples

### Web UI 使用
1. 访问 http://localhost:8080
2. 在左侧面板填写扫描参数
3. 点击"开始扫描"
4. 在右侧查看实时结果
5. 使用搜索框过滤结果
6. 查看底部的扫描历史

### CLI 示例

#### 基础扫描
```bash
# 扫描 3 位字母 .li 域名
go run main.go -l 3 -s .li -p D

# 扫描 4 位数字 .com 域名
go run main.go -l 4 -s .com -p d
```

#### 高级过滤
```bash
# 查找以 "abc" 开头的域名
go run main.go -l 5 -s .com -p D -r "^abc"

# 查找包含特定模式的域名
go run main.go -l 3 -s .li -p D -r "^[a-z]{2}[0-9]$"
```

#### 字典模式
```bash
# 从字典文件检查域名
go run main.go -dict words.txt -s .com

# 字典 + 正则过滤
go run main.go -dict words.txt -s .com -r "^[a-z]{4,8}$"
```

#### 性能调优
```bash
# 使用更多工作线程
go run main.go -l 3 -s .li -p D -workers 50

# 减少查询延迟
go run main.go -l 3 -s .li -p D -delay 500

# 组合使用
go run main.go -l 3 -s .li -p D -workers 30 -delay 300
```

---

## 🔐 安全建议 Security Tips

### 生产环境
1. 修改默认数据库密码
2. 使用环境变量存储敏感信息
3. 配置防火墙规则
4. 使用反向代理 + SSL
5. 定期备份数据库

### 配置 .env 文件
```bash
cp .env.example .env
# 编辑 .env 修改密码
```

---

## 🐛 故障排查 Troubleshooting

### 端口被占用
```bash
# 查看端口占用
lsof -i :8080
lsof -i :5432

# 修改端口（编辑 docker-compose.yml）
ports:
  - "9000:8080"  # 使用 9000 端口
```

### 容器启动失败
```bash
# 查看详细日志
docker-compose logs web
docker-compose logs postgres

# 重新构建
docker-compose down
docker-compose build --no-cache
docker-compose up -d
```

### 数据库连接失败
```bash
# 检查容器状态
docker-compose ps

# 重启数据库
docker-compose restart postgres

# 查看数据库日志
docker-compose logs postgres
```

---

## 📚 更多资源 More Resources

- [完整部署文档](DEPLOYMENT.md)
- [更新日志](docs/CHANGELOG.md)
- [贡献指南](CONTRIBUTING.md)
- [GitHub 仓库](https://github.com/xuemian168/domain-scanner)

---

## 💡 提示 Tips

1. **首次运行**: 建议使用小范围参数测试 (如 `-l 2`)
2. **性能优化**: 根据网络情况调整 workers 和 delay
3. **大规模扫描**: 使用 `-force` 跳过警告
4. **结果导出**: CLI 版本会自动保存结果到文本文件
5. **Web 版本**: 扫描在后台运行，可同时启动多个扫描

---

**版本**: v1.4.0  
**更新时间**: 2025-11-25

