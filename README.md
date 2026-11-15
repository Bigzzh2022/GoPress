# GoPress

基于 Go、React 与 MySQL 的现代全栈博客，包含公开阅读站点与完整后台管理。

## 技术栈

- **后端**：Go、Gin、GORM、JWT、goldmark、bluemonday
- **前端**：React 19、Vite、TypeScript、Tailwind CSS、Ant Design 5
- **状态与编辑器**：Zustand、TanStack Query、Vditor、react-markdown、highlight.js
- **基础设施**：MySQL 8.0、Redis、Docker Compose、Nginx

## 快速启动

### Docker Compose

```bash
cp .env.example .env
# 部署前务必修改 .env 中的密码和 JWT_SECRET
docker compose up --build -d
```

- 博客首页：http://localhost:3000
- 后台管理：http://localhost:3000/admin
- API 健康检查：http://localhost:8080/api/v1/health

开发环境默认管理员：`admin@gopress.dev` / `GoPress123!`。首次启动会执行迁移和幂等初始化，`SEED_DEMO=true` 时写入演示文章。

### 本地开发

准备 MySQL 8.0 和 Redis，并创建 `gopress` 数据库，然后：

```bash
cp .env.example .env
set -a
source .env
set +a
go run ./cmd/seed
go run ./cmd/server

# 另一个终端
npm --prefix web install
npm --prefix web run dev
```

Vite 开发地址为 http://localhost:5173，已将 `/api` 代理到 `http://localhost:8080`。

## 常用命令

```bash
make test          # 后端测试与前端测试
make build         # 后端编译与前端生产构建
make docker-up     # 构建并启动完整环境
make docker-logs   # 查看容器日志
make docker-down   # 停止环境（保留数据卷）
```

## 功能

- 公开文章分页、关键词搜索、分类与标签筛选、相关文章、关于页
- 后台仪表盘、文章草稿/发布、分类标签管理、个人资料维护
- Vditor Markdown 编辑，goldmark 服务端过滤，react-markdown 前台渲染与代码高亮
- JWT 短期访问令牌、HttpOnly 刷新 Cookie、Redis 会话旋转与注销
- Redis 公开内容缓存及写入失效，缓存不可用时自动回源 MySQL
- 响应式页面、加载/错误/空状态、表单校验和路由级代码分割

## 环境变量

所有配置见 `.env.example`。生产环境至少修改 `MYSQL_PASSWORD`、`MYSQL_ROOT_PASSWORD`、`JWT_SECRET` 和 `ADMIN_PASSWORD`，并设置 `COOKIE_SECURE=true`、关闭 `SEED_DEMO`，通过 HTTPS 提供服务。

## 数据与安全

Markdown 原文是唯一内容源。后端禁用原始 HTML 并使用白名单过滤，前台不启用 `rehype-raw`。数据库迁移位于 `migrations/`，容器数据保存在 `mysql_data` 与 `redis_data` 命名卷中。
