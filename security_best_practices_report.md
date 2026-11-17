# GoPress 安全最佳实践审计报告

审计范围：`/workspace`（Go + Gin + GORM 后端，React 19 + Vite 前端，MySQL 8.0、Redis、JWT、Markdown 渲染）。
审计依据：`security-best-practices` skill 的 Go 后端与 TypeScript/React 前端参考规则，以及当前工作区实际代码。

## 执行摘要

项目的安全基础总体扎实：访问令牌仅存内存、刷新令牌哈希后存 Redis 并轮换、密码使用 bcrypt、Markdown 经 goldmark + bluemonday 双重处理、GORM 查询全部参数化、前台 `skipHtml` 且预览 iframe `sandbox=""`。

但存在两类可导致**完全接管后台**的问题：公开可推断的默认管理员口令会自动写入数据库，以及 JWT 密钥存在硬编码默认值且未做强度校验。其次，登录限流在反向代理后会退化为全局单桶并 fail-open，请求体无大小上限，Markdown 编辑器依赖第三方 CDN 资源，这些都会显著放大被攻击面。建议优先修复 C-1、C-2。

---

## Critical

### C-1 默认管理员凭据会被自动创建
影响：按文档默认配置部署后，任何知道默认口令的人可直接登录后台，发布、修改或删除全部内容，并获得完整的站点控制权。

- [config.go](file:///workspace/internal/config/config.go#L36-L38) 中 `ADMIN_EMAIL` 默认为 `admin@gopress.dev`，`ADMIN_PASSWORD` 默认为 `GoPress123!`。
- [seed.go](file:///workspace/internal/platform/database/seed.go#L73-L90) 在不存在管理员时使用该默认值创建账号，仅校验长度 ≥ 8。
- [docker-compose.yml](file:///workspace/docker-compose.yml#L49-L51) 与 [.env.example](file:///workspace/.env.example#L17-L18) 同样以明文默认值提供。
- [README.md](file:///workspace/README.md) 直接公布了该开发账号与密码。

建议：生产环境禁止使用默认值——启动时若检测到默认口令或默认邮箱则拒绝启动；首次初始化改为从环境变量强制读取一次性口令，并强制改密；文档区分“演示”与“生产”配置。

### C-2 JWT 密钥存在硬编码默认值，且只校验非空
影响：若沿用默认密钥，攻击者可离线伪造任意包含 `role=admin` 的访问令牌，完全绕过认证与授权。

- [config.go](file:///workspace/internal/config/config.go#L34) 默认值 `gopress-development-secret-change-me`。
- [config.go](file:///workspace/internal/config/config.go#L60-L62) 仅检查 `JWT_SECRET` 非空，没有最小长度或熵要求。
- [docker-compose.yml](file:///workspace/docker-compose.yml#L44) 又以另一已知默认值兜底，容易在未创建 `.env` 时静默使用。
- 令牌使用 HS256，见 [token.go](file:///workspace/internal/modules/auth/token.go)。

建议：强制 `JWT_SECRET` ≥ 32 字节随机值；拒绝等于任一已知默认值；未设置时直接启动失败而不是回退；可考虑启动时记录密钥指纹（非密钥本身）便于排查。

---

## High

### H-1 登录限流在反向代理后退化为全局单桶，且 Redis 故障时完全放行
影响：攻击者每分钟 10 次失败登录即可让**所有用户**无法登录（可用性拒绝服务）；Redis 不可用时暴力破解完全不受限制。

- [router.go](/workspace/internal/http/router.go#L23) 调用 `SetTrustedProxies(nil)`，[handler.go](/workspace/internal/modules/auth/handler.go#L99) 使用 `c.ClientIP()`；经 Nginx 转发后该值恒为代理地址。
- [handler.go](/workspace/internal/modules/auth/handler.go#L205-L220) 在 `redis == nil` 或 `Incr` 出错时 `return true`（fail-open）。
- `/auth/refresh` 与 `/auth/logout` 未做任何限流。

建议：显式配置可信代理网段（或在 Nginx 后改为信任 `X-Real-IP` 并限制来源）；Redis 故障时对登录改为 fail-closed 或本地内存兜底限流；为 refresh 增加独立限流并返回 `Retry-After`。

### H-2 缺少请求体大小限制，Markdown 预览输入无上限
影响：直接暴露的 API 端口可接收超大 JSON，配合 goldmark 渲染造成内存与 CPU 耗尽。

- [router.go](/workspace/internal/http/router.go#L22-L24) 未挂载 `http.MaxBytesReader` 或等效中间件。
- [handler.go](/workspace/internal/modules/posts/handler.go#L38-L40) 的 `previewInput.Content` 无 `max` 约束；[Preview](/workspace/internal/modules/posts/handler.go#L193) 直接解析并渲染。
- Nginx 仅对经其转发的请求限制 2 MB（[nginx.conf](/workspace/deploy/nginx.conf#L6)），而 Compose 将 API 8080 直接发布到宿主机（[docker-compose.yml](/workspace/docker-compose.yml#L53-L54)），可绕过该限制。

建议：统一在 Gin 层加 body 上限（如 1–2 MB）；预览正文单独限制（如 256 KB）；生产环境不要直接发布 API 端口，仅经反代访问。

### H-3 Vditor 未配置 `cdn`，默认从第三方 CDN 运行时加载资源
影响：后台编辑器会从 unpkg 拉取脚本/样式，缺少 SRI 与 CSP 时，第三方篡改即可在已登录的管理员会话中执行任意脚本（可直接调用后台接口）。

- [VditorEditor.tsx](/workspace/web/src/components/VditorEditor.tsx#L20-L39) 的初始化选项未设置 `cdn`，Vditor 默认使用 `https://unpkg.com/vditor@<version>` 按需加载资源。
- `package.json` 仅锁定 npm 依赖版本（[package.json](/workspace/web/package.json#L28)），无法约束 CDN 运行时资源。

建议：将 Vditor 静态资源纳入 `web/public` 自托管并显式配置 `cdn: '/vditor'`；为页面补充 CSP（至少 `script-src 'self'`）。

---

## Medium

### M-1 公开接口批量泄露作者邮箱
- [dto.go](/workspace/internal/model/dto.go#L8) 的 `UserDTO` 包含 `Email`；[ToPostSummary](/workspace/internal/model/dto.go#L86) 将其嵌入每篇文章，公开列表/详情均可获取。
- 影响：邮箱可被抓取用于钓鱼与垃圾邮件，并直接暴露管理员账号名。

建议：公开 DTO 去掉邮箱，单独提供“联系”字段或联系表单。

### M-2 CORS 允许 `*` 且同时开启凭据
- [cors.go](/workspace/internal/http/middleware/cors.go#L15-L18) 接受 `*`，[L26-L32](/workspace/internal/http/middleware/cors.go#L26-L32) 随后回显任意 Origin 并设置 `Allow-Credentials: true`。
- 影响：运维误配 `CORS_ORIGINS=*` 时，任意站点可带凭据跨域访问。

建议：`*` 与 credentials 互斥，配置为 `*` 时强制去掉凭据头；或在启动时直接拒绝该组合。

### M-3 公开搜索可被放大为数据库 DoS
- [service.go](/workspace/internal/modules/posts/service.go#L122-L124) 对 `title`、`excerpt`、`content_md` 使用 `%keyword%` 全表扫描；[L114](/workspace/internal/modules/posts/service.go#L114) 的 `page` 无上限，大 offset 进一步放大成本；公开接口无限流。

建议：给 `page` 设上限或改用游标分页；对搜索加独立限流；正文搜索可改为 MySQL 全文索引或专用搜索服务。

### M-4 缺少 CSP 与 API 侧安全响应头
- [nginx.conf](/workspace/deploy/nginx.conf#L11-L13) 仅设置 `X-Content-Type-Options`、`X-Frame-Options`、`Referrer-Policy`，无 `Content-Security-Policy`、`Permissions-Policy`。
- Go 侧无安全响应头中间件（[router.go](/workspace/internal/http/router.go#L24)）。

建议：为 SPA 添加贴合实际资源来源的 CSP（注意 Vditor 的内联样式需求），API 响应补充 `nosniff` 与缓存控制。

### M-5 Cookie 鉴权接口缺少 CSRF 纵深
- [Refresh](/workspace/internal/modules/auth/handler.go#L136-L178) 与 [Logout](/workspace/internal/modules/auth/handler.go#L180-L186) 仅依赖 `SameSite=Lax`（[L255](/workspace/internal/modules/auth/handler.go#L255)），未校验 `Origin`/`Referer`。

建议：对 cookie 鉴权的写操作增加 Origin 白名单校验，保持 `SameSite=Lax` 作为第二层。

### M-6 访问令牌无法吊销，权限变更延迟生效
- [Logout](/workspace/internal/modules/auth/handler.go#L180-L186) 只删除 refresh 会话，access token 在 15 分钟内仍有效；角色声明在签发时固化。

建议：高敏感操作（改密、改邮箱、删文章）前二次校验数据库角色；接受 15 分钟窗口时在文档中明确说明。

### M-7 部署默认值与端口暴露
- [docker-compose.yml](/workspace/docker-compose.yml#L6-L9) 使用 `root`/`gopress` 默认数据库口令，[L53-L54](/workspace/docker-compose.yml#L53-L54) 直接发布 API 8080，[L72-L73](/workspace/docker-compose.yml#L72-L73) 发布 3000。
- MySQL/Redis 端口未对外（良好），但同一 Docker 网络内任意容器可用默认口令连接。

建议：Compose 中移除弱默认值或改为必填（`${VAR:?err}`）；生产仅发布 Web 端口。

---

## Low

### L-1 登录存在账号枚举时间差
- [handler.go](/workspace/internal/modules/auth/handler.go#L105-L112)：用户不存在时直接返回，不执行 bcrypt 比较，响应时间明显更短。

建议：用户不存在时仍与固定 dummy hash 比较。

### L-2 管理员口令策略偏弱
- [seed.go](/workspace/internal/platform/database/seed.go#L73-L75) 仅要求 ≥ 8 位，无复杂度或常见口令检查。

### L-3 健康检查与公开接口无速率限制
- [router.go](/workspace/internal/http/router.go#L37-L45)，可被用于低成本探测或放大负载。

### L-4 Gin 访问日志记录完整查询串
- [router.go](/workspace/internal/http/router.go#L24) 使用 `gin.Logger()`，搜索关键词等会进入日志；无敏感信息但涉及用户输入留存。

### L-5 容器基础镜像使用浮动 tag
- [backend.Dockerfile](/workspace/deploy/backend.Dockerfile#L1)、[frontend.Dockerfile](/workspace/deploy/frontend.Dockerfile#L1-L10)、[nginx.conf](/workspace/deploy/nginx.conf) 依赖 `alpine:3.21`、`node:24-alpine`、`nginx:1.27-alpine` 等，未固定 digest，存在供应链漂移风险。

### L-6 Nginx 未关闭版本信息
- [nginx.conf](/workspace/deploy/nginx.conf#L1-L3) 未设置 `server_tokens off`。

---

## 已确认的良好实践

- 刷新令牌使用 32 字节随机值、仅以 SHA-256 哈希形式存 Redis，并在刷新时 `GetDel` 轮换（[handler.go](/workspace/internal/modules/auth/handler.go#L147-L176)）。
- Cookie 设置 `HttpOnly`、`SameSite=Lax`、限定 `Path=/api/v1/auth`，`Secure` 可配置（[handler.go](/workspace/internal/modules/auth/handler.go#L252-L256)）。
- access token 仅保存在 Zustand 内存，不落 `localStorage`（[authStore.ts](/workspace/web/src/stores/authStore.ts)）。
- 密码使用 bcrypt，登录失败返回统一错误（[handler.go](/workspace/internal/modules/auth/handler.go#L110-L112)）。
- Markdown 后端禁用原始 HTML 并经 bluemonday 白名单过滤（[markdown.go](/workspace/internal/platform/markdown/markdown.go)）。
- 前台 `skipHtml` + `defaultUrlTransform`，不使用 `dangerouslySetInnerHTML`（[MarkdownContent.tsx](/workspace/web/src/components/MarkdownContent.tsx)）；后台预览使用 `sandbox=""` 的 iframe（[PostEditorPage.tsx](/workspace/web/src/pages/admin/PostEditorPage.tsx#L102)）。
- GORM 全部使用参数化查询，LIKE 通配符已转义（[service.go](/workspace/internal/modules/posts/service.go#L140)）。
- HTTP 服务器设置 `ReadHeaderTimeout`/`ReadTimeout`/`WriteTimeout`/`IdleTimeout`（[main.go](/workspace/cmd/server/main.go#L45-L49)）。
- 模型字段均标注 `json:"-"`，通过 DTO 输出，未泄露密码哈希（[models.go](/workspace/internal/model/models.go)）。
- 容器内后端以非 root 用户运行（[backend.Dockerfile](/workspace/deploy/backend.Dockerfile#L15-L18)）。

## 建议修复顺序

1. C-2 JWT 密钥强制校验与 fail-fast。
2. C-1 移除默认管理员口令，强制一次性初始化。
3. H-1 限流与可信代理修复。
4. H-2 请求体上限与 API 端口收敛。
5. H-3 Vditor 资源自托管 + CSP。
6. 其余 Medium/Low 按迭代逐步处理。

---

## 修复记录（2026-11-17）

全部 18 项发现均已处理，验证方式见每行末尾。

| ID | 修复内容 | 验证 |
|---|---|---|
| C-1 | 移除 `ADMIN_PASSWORD` 默认值；为空或命中已知弱口令时 seed 直接报错；新增 12 位 + 三类字符的复杂度校验与邮箱格式校验；Compose 与 `.env.example` 改为必填占位符；README 删除默认账号 | `config_test.go`；运行服务分别输出“ADMIN_PASSWORD must contain at least 12 characters” |
| C-2 | 移除 `JWT_SECRET` 兜底；要求 ≥ 32 字符且不等于任何已发布值，否则启动失败 | `config_test.go`；实际启动分别输出 “JWT_SECRET is required” 与 “matches a published default value” |
| H-1 | 新增 `TRUSTED_PROXIES` 并交给 `SetTrustedProxies`；新增 Redis + 进程内兜底限流器（Redis 故障不再放行）；登录 10/min、刷新 30/min、公开接口 120/min；429 返回 `Retry-After` | `middleware_test.go`（无 Redis 时第 3 次请求 429 且带 Retry-After） |
| H-2 | 新增 2 MiB `BodyLimit` 中间件 + `MaxBytesReader`；预览正文 `max=262144`；Compose 将 8080 改为 `127.0.0.1` 绑定 | `middleware_test.go`（超限 413，分块请求同样被截断） |
| H-3 | 新增 `web/scripts/copy-vditor.mjs` 在 dev/build 前把 `vditor/dist` 复制到 `public/vditor`，并在编辑器显式配置 `cdn`；Nginx 增加 CSP 中 `script-src 'self'` | `npm run build` 后 `web/dist/vditor/dist/js` 存在；`web/src` 中无 unpkg 引用 |
| M-1 | 新增 `PublicUserDTO`（无 email），文章与公开资料接口改用该结构；前端新增 `PublicUser` 类型，关于页改用静态联系地址 | `blog.ts` / `types/index.ts`；Go 测试通过 |
| M-2 | `*` 时仅返回 `Access-Control-Allow-Origin: *` 且不设置 `Allow-Credentials`；白名单回显时才允许凭据 | `middleware_test.go` |
| M-3 | `page` 上限 200；公开读接口按 IP 限流 120/min | 编译期 binding 校验 + 限流中间件 |
| M-4 | API 新增 `SecurityHeaders`（nosniff、DENY frame、CSP `default-src 'none'`、Permissions-Policy）；Nginx 增加 CSP 与 Permissions-Policy | `middleware/security.go`、`deploy/nginx.conf` |
| M-5 | 新增 `RequireTrustedOrigin` 中间件，refresh/logout 校验 `Origin`/`Referer` 白名单 | `middleware_test.go`（跨站 Origin 403） |
| M-6 | 访问令牌加入 `jti`，注销时按剩余有效期写入 Redis 吊销名单；后台接口每次校验数据库中的角色与账号存在性 | `token_test.go`（jti 唯一） |
| M-7 | Compose 对 `JWT_SECRET`、`ADMIN_PASSWORD`、数据库口令使用 `${VAR:?}` 必填；API 端口仅绑定回环；新增 `no-new-privileges` | `docker-compose.yml` 静态检查 |
| L-1 | 用户不存在时仍执行一次 bcrypt 比较（固定 dummy hash） | 代码审查 |
| L-2 | 管理员口令复杂度与弱口令黑名单 | `config_test.go` |
| L-3 | 公开接口统一限流 120/min | `middleware_test.go` |
| L-4 | 自定义访问日志只记录 path，不记录查询串 | `middleware/logger.go` |
| L-5 | Dockerfile 基础镜像改为 `ARG`，便于运维固定 digest 并在 README 说明 | `deploy/*.Dockerfile` |
| L-6 | Nginx `server_tokens off` | `deploy/nginx.conf` |

说明：沙箱无法访问 Docker Hub，未能在本环境解析并写入镜像 digest，因此以可覆盖的 `ARG` 加部署说明落地；生产发布时请补上 `@sha256:...`。

