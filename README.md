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
# 必须先把 .env 中的占位值替换为真实密钥，否则 API 会拒绝启动
openssl rand -hex 32   # 生成 JWT_SECRET
docker compose up --build -d
```

- 博客首页：http://localhost:3000
- 后台管理：http://localhost:3000/admin
- API 健康检查：http://127.0.0.1:8080/api/v1/health（仅绑定本机回环）

首次启动会执行迁移和幂等初始化。管理员账号由 `.env` 中的 `ADMIN_EMAIL` 与 `ADMIN_PASSWORD` 决定：密码至少 12 位且需混合三类字符，仓库中发布过的默认密码会被拒绝。`SEED_DEMO=true` 时写入演示文章。

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
- 后台仪表盘、文章草稿/发布/定时发布、批量操作与回收站、分类标签管理、站点设置与密码修改
- Vditor Markdown 编辑，goldmark 服务端过滤，react-markdown 前台渲染与代码高亮
- JWT 短期访问令牌、HttpOnly 刷新 Cookie、Redis 会话旋转与注销
- Redis 公开内容缓存及写入失效，缓存不可用时自动回源 MySQL
- 响应式页面、加载/错误/空状态、表单校验和路由级代码分割

## 环境变量

所有配置见 `.env.example`。`JWT_SECRET` 与 `ADMIN_PASSWORD` 没有默认值，缺失或使用仓库中出现过的默认值时服务会直接退出并给出提示。

- `JWT_SECRET`：至少 32 个字符，建议 `openssl rand -hex 32`
- `ADMIN_PASSWORD`：至少 12 位，且包含小写、大写、数字、符号中的至少三类
- `TRUSTED_PROXIES`：允许提供 `X-Forwarded-For` 的反向代理网段，Compose 默认 `172.16.0.0/12`
- `COOKIE_SECURE`：仅在通过 HTTPS 提供服务时设为 `true`
- `CORS_ORIGINS`：允许的来源白名单；配置为 `*` 时不会同时开启凭据
- 若通过域名或非 `localhost` 访问，必须把该来源加入 `CORS_ORIGINS`，否则刷新与注销请求会被同源校验拒绝

生产环境还应关闭 `SEED_DEMO`，并把基础镜像固定到具体 digest（见 `deploy/*.Dockerfile` 顶部的 `ARG`）。

## 数据与安全

Markdown 原文是唯一内容源。后端禁用原始 HTML 并使用白名单过滤，前台不启用 `rehype-raw`。数据库迁移位于 `migrations/`，容器数据保存在 `mysql_data` 与 `redis_data` 命名卷中。

安全设计要点：

- 访问令牌只保存在内存，刷新令牌以哈希形式存 Redis 并每次轮换；注销会把当前访问令牌加入吊销名单
- 登录、刷新与公开接口按客户端 IP 分级限流，Redis 故障时回退到进程内计数而非放行
- 请求体限制 2 MiB，Markdown 预览正文限制 256 KiB
- Nginx 下发 CSP、`nosniff`、`Permissions-Policy`，并关闭版本号回显
- Vditor 运行时资源自行托管在 `/vditor`，后台不再从第三方 CDN 加载脚本
- 公开接口不返回作者邮箱，避免批量抓取
