# 域名扫描器 Web 版部署指南

## 🚀 快速开始

### 使用 Docker Compose 部署（推荐）

1. **克隆仓库**
```bash
git clone https://github.com/xuemian168/domain-scanner.git
cd domain-scanner
```

2. **启动服务**
```bash
docker-compose up -d
```

3. **访问 Web 界面**
打开浏览器访问: http://localhost:8080

4. **查看日志**
```bash
docker-compose logs -f web
```

5. **停止服务**
```bash
docker-compose down
```

6. **停止并删除数据**
```bash
docker-compose down -v
```

## 📋 功能特性

### Web UI 功能
- ✅ 图形化界面启动域名扫描
- ✅ 实时查看扫描统计数据
- ✅ 域名结果列表展示（可用/已注册）
- ✅ 域名搜索和过滤功能
- ✅ 扫描历史记录查看
- ✅ 自动刷新数据（30秒间隔）

### 数据库功能
- ✅ PostgreSQL 数据持久化存储
- ✅ 扫描会话管理
- ✅ 域名记录去重
- ✅ 完整的查询统计

## 🔧 配置说明

### 环境变量

Web 服务支持以下环境变量：

- `DATABASE_URL`: PostgreSQL 连接字符串
  - 默认: `postgres://scanner:scanner123@postgres:5432/domainscanner?sslmode=disable`
  
### 端口配置

默认端口映射：
- Web UI: `8080:8080`
- PostgreSQL: `5432:5432`

如需修改端口，编辑 `docker-compose.yml`:

```yaml
services:
  web:
    ports:
      - "9000:8080"  # 将 Web UI 映射到 9000 端口
```

## 📊 数据库结构

### 表: scan_sessions
存储扫描会话信息
- `id`: 会话 ID
- `pattern`: 域名模式 (d/D/a)
- `length`: 域名长度
- `suffix`: 域名后缀
- `total_domains`: 总域名数
- `available_count`: 可用域名数
- `registered_count`: 已注册域名数
- `started_at`: 开始时间
- `completed_at`: 完成时间
- `status`: 状态 (running/completed)

### 表: domain_records
存储域名检查结果
- `id`: 记录 ID
- `session_id`: 关联的会话 ID
- `domain`: 域名
- `available`: 是否可用
- `signatures`: 验证签名数组
- `checked_at`: 检查时间
- `pattern`: 域名模式
- `length`: 域名长度
- `suffix`: 域名后缀

## 🛠️ 本地开发

### 前置要求
- Go 1.22+
- PostgreSQL 16+

### 安装依赖
```bash
go mod download
```

### 启动 PostgreSQL
```bash
docker run -d \
  --name domain-scanner-db \
  -e POSTGRES_DB=domainscanner \
  -e POSTGRES_USER=scanner \
  -e POSTGRES_PASSWORD=scanner123 \
  -p 5432:5432 \
  postgres:16-alpine
```

### 运行 Web 服务器
```bash
go run cmd/webserver/main.go
```

### 运行原有 CLI 工具
```bash
go run main.go -l 3 -s .li -p D -workers 20
```

## 🔌 API 文档

### GET /api/stats
获取统计信息
```json
{
  "success": true,
  "data": {
    "total_domains": 1000,
    "available_domains": 150,
    "registered_domains": 850
  }
}
```

### GET /api/domains
获取域名列表

查询参数:
- `available`: true/false (可选，过滤可用/已注册)
- `limit`: 返回数量限制 (默认: 50)
- `offset`: 偏移量 (默认: 0)

### GET /api/domains/search
搜索域名

查询参数:
- `q`: 搜索关键词 (必需)
- `available`: true/false (可选)
- `limit`: 返回数量限制 (默认: 50)
- `offset`: 偏移量 (默认: 0)

### GET /api/sessions
获取扫描会话列表

查询参数:
- `limit`: 返回数量限制 (默认: 20)
- `offset`: 偏移量 (默认: 0)

### POST /api/scan
启动新的域名扫描

请求体:
```json
{
  "length": 3,
  "suffix": ".li",
  "pattern": "D",
  "regex_filter": "",
  "delay": 1000,
  "workers": 10
}
```

## 🔐 安全建议

### 生产环境部署

1. **修改数据库密码**
编辑 `docker-compose.yml`:
```yaml
environment:
  POSTGRES_PASSWORD: your_secure_password
```

2. **使用环境变量文件**
创建 `.env` 文件:
```env
POSTGRES_PASSWORD=your_secure_password
DATABASE_URL=postgres://scanner:your_secure_password@postgres:5432/domainscanner?sslmode=disable
```

然后在 `docker-compose.yml` 中引用:
```yaml
services:
  postgres:
    env_file:
      - .env
```

3. **启用 HTTPS**
使用反向代理（如 Nginx）配置 SSL 证书

4. **限制访问**
配置防火墙规则，仅允许必要的 IP 访问

## 📈 性能优化

### 数据库优化
```sql
-- 创建额外索引以提升查询性能
CREATE INDEX idx_domain_records_checked_at ON domain_records(checked_at DESC);
CREATE INDEX idx_scan_sessions_created_at ON scan_sessions(started_at DESC);
```

### 扫描性能
- 增加 `workers` 数量以提高扫描速度
- 降低 `delay` 值（注意可能触发速率限制）
- 使用正则过滤器缩小扫描范围

## 🐛 故障排查

### 容器无法启动
```bash
# 查看容器日志
docker-compose logs web
docker-compose logs postgres

# 检查容器状态
docker-compose ps
```

### 数据库连接失败
```bash
# 确认 PostgreSQL 已启动
docker-compose ps postgres

# 测试数据库连接
docker-compose exec postgres psql -U scanner -d domainscanner
```

### Web UI 无法访问
```bash
# 检查端口占用
lsof -i :8080

# 确认防火墙设置
sudo ufw status
```

## 📝 更新日志

### v1.4.0 - 2025-11-25
- ✨ 新增 Web UI 界面
- ✨ 新增 PostgreSQL 数据库存储
- ✨ 新增 Docker Compose 部署支持
- ✨ 新增 RESTful API
- ✨ 新增实时统计和搜索功能

## 📄 许可证

本项目采用 AGPL-3.0 许可证。详见 [LICENSE](LICENSE) 文件。

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📧 联系方式

- 网站: www.ict.run
- GitHub: https://github.com/xuemian168/domain-scanner

